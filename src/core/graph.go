// Representation of the build graph.
// The graph of build targets forms a DAG which we discover from the top
// down and then build bottom-up.

package core

import (
	"sort"
	"sync"

	"github.com/streamrail/concurrent-map"
)

type BuildGraph struct {
	// Map of all currently known targets by their label.
	targets cmap.ConcurrentMap
	// Map of all currently known packages.
	packages cmap.ConcurrentMap
	// Reverse dependencies that are pending on targets actually being added to the graph.
	pendingRevDeps cmap.ConcurrentMap
	// Actual reverse dependencies
	revDeps cmap.ConcurrentMap
	// Used to arbitrate access to the graph. We parallelise most build operations
	// and Go maps aren't natively threadsafe so this is needed.
	mutex sync.RWMutex
}

type revdepMap map[BuildLabel]*BuildTarget

// Adds a new target to the graph.
func (graph *BuildGraph) AddTarget(target *BuildTarget) *BuildTarget {
	if !graph.targets.SetIfAbsent(target.Label.String(), target) {
		panic("Attempted to re-add existing target to build graph: " + target.Label.String())
	}
	// Check these reverse deps which may have already been added against this target.
	if revdeps, present := graph.pendingRevDeps.Get(target.Label.String()); present {
		for revdep, originalTarget := range revdeps.(revdepMap) {
			if originalTarget != nil {
				graph.linkDependencies(graph.Target(revdep), originalTarget)
			} else {
				graph.linkDependencies(graph.Target(revdep), target)
			}
		}
		graph.pendingRevDeps.Remove(target.Label.String())
	}
	return target
}

// Adds a new package to the graph with given name.
func (graph *BuildGraph) AddPackage(pkg *Package) {
	if !graph.packages.SetIfAbsent(pkg.Name, pkg) {
		panic("Attempt to readd existing package: " + pkg.Name)
	}
}

// Target retrieves a target from the graph by label
func (graph *BuildGraph) Target(label BuildLabel) *BuildTarget {
	target, present := graph.targets.Get(label.String())
	if !present {
		return nil
	}
	return target.(*BuildTarget)
}

// TargetOrDie retrieves a target from the graph by label. Dies if the target doesn't exist.
func (graph *BuildGraph) TargetOrDie(label BuildLabel) *BuildTarget {
	target := graph.Target(label)
	if target == nil {
		log.Fatalf("Target %s not found in build graph\n", label)
	}
	return target
}

// Package retrieves a package from the graph by name
func (graph *BuildGraph) Package(name string) *Package {
	pkg, present := graph.packages.Get(name)
	if !present {
		return nil
	}
	return pkg.(*Package)
}

// PackageOrDie retrieves a package by name, and dies if it can't be found.
func (graph *BuildGraph) PackageOrDie(name string) *Package {
	pkg := graph.Package(name)
	if pkg == nil {
		log.Fatalf("Package %s doesn't exist in graph", name)
	}
	return pkg
}

func (graph *BuildGraph) Len() int {
	return graph.targets.Count()
}

// Returns a sorted slice of all the targets in the graph.
func (graph *BuildGraph) AllTargets() BuildTargets {
	targets := make(BuildTargets, 0, graph.Len())
	for kv := range graph.targets.IterBuffered() {
		targets = append(targets, kv.Val.(*BuildTarget))
	}
	sort.Sort(targets)
	return targets
}

// Used for getting a local copy of the package map without having to expose it publicly.
func (graph *BuildGraph) PackageMap() map[string]*Package {
	packages := make(map[string]*Package)
	for kv := range graph.packages.IterBuffered() {
		packages[kv.Key] = kv.Val.(*Package)
	}
	return packages
}

func (graph *BuildGraph) AddDependency(from BuildLabel, to BuildLabel) {
	fromTarget := graph.Target(from)
	// We might have done this already; do a quick check here first.
	if fromTarget.hasResolvedDependency(to) {
		return
	}
	// The dependency may not exist yet if we haven't parsed its package.
	// In that case we stash it away for later.
	if toTarget := graph.Target(to); toTarget == nil {
		graph.addPendingRevDep(from, to, nil)
	} else {
		graph.linkDependencies(fromTarget, toTarget)
	}
}

func NewGraph() *BuildGraph {
	return &BuildGraph{
		targets:        cmap.New(),
		packages:       cmap.New(),
		pendingRevDeps: cmap.New(),
		revDeps:        cmap.New(),
	}
}

// ReverseDependencies returns the set of revdeps on the given target.
func (graph *BuildGraph) ReverseDependencies(target *BuildTarget) []*BuildTarget {
	if revdeps, present := graph.revDeps.Get(target.Label.String()); present {
		return revdeps.(BuildTargets)[:]
	}
	return nil
}

// AllDepsBuilt returns true if all the dependencies of a target are built.
func (graph *BuildGraph) AllDepsBuilt(target *BuildTarget) bool {
	graph.mutex.RLock()
	defer graph.mutex.RUnlock()
	return target.allDepsBuilt()
}

// AllDependenciesResolved returns true once all the dependencies of a target have been
// parsed and resolved to real targets.
func (graph *BuildGraph) AllDependenciesResolved(target *BuildTarget) bool {
	graph.mutex.RLock()
	defer graph.mutex.RUnlock()
	return target.allDependenciesResolved()
}

// linkDependencies adds the dependency of fromTarget on toTarget and the corresponding
// reverse dependency in the other direction.
// This is complicated somewhat by the require/provide mechanism which is resolved at this
// point, but some of the dependencies may not yet exist.
func (graph *BuildGraph) linkDependencies(fromTarget, toTarget *BuildTarget) {
	for _, label := range toTarget.ProvideFor(fromTarget) {
		if target := graph.Target(label); target != nil {
			fromTarget.resolveDependency(toTarget.Label, target)
			if revdeps, present := graph.revDeps.Get(label.String()); present {
				graph.revDeps.Set(label.String(), append(revdeps.(BuildTargets), fromTarget))
			} else {
				graph.revDeps.Set(label.String(), BuildTargets{fromTarget})
			}
		} else {
			graph.addPendingRevDep(fromTarget.Label, label, toTarget)
		}
	}
}

func (graph *BuildGraph) addPendingRevDep(from, to BuildLabel, orig *BuildTarget) {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()
	if deps, present := graph.pendingRevDeps.Get(to.String()); present {
		deps.(revdepMap)[from] = orig
	} else {
		graph.pendingRevDeps.Set(to.String(), revdepMap{from: orig})
	}
}

// DependentTargets returns the labels that 'from' should actually depend on when it declared a dependency on 'to'.
// This is normally just 'to' but could be otherwise given require/provide shenanigans.
func (graph *BuildGraph) DependentTargets(from, to BuildLabel) []BuildLabel {
	fromTarget := graph.Target(from)
	if toTarget := graph.Target(to); fromTarget != nil && toTarget != nil {
		graph.mutex.Lock()
		defer graph.mutex.Unlock()
		return toTarget.ProvideFor(fromTarget)
	}
	return []BuildLabel{to}
}
