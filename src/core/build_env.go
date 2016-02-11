package core

import (
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

var home = os.Getenv("HOME")
var homeRex = regexp.MustCompile("(?:^|:)(~(?:[/:]|$))")

// Expand all prefixes of ~ without a user specifier TO $HOME.
// NB: this will break on recursive calls to plz
func expandHomePath(path string) string {
	return homeRex.ReplaceAllStringFunc(path, func(subpath string) string {
		return strings.Replace(subpath, "~", home, -1)
	})
}

// BuildEnvironment state target test creates the shell env vars to be passed
// into the exec.Command calls made by plz. Use test=true for plz test targets.
func BuildEnvironment(state *BuildState, target *BuildTarget, test bool) []string {
	sources := target.AllSourcePaths(state.Graph)
	env := []string{
		"PKG=" + target.Label.PackageName,
		// Need to know these for certain rules, particularly Go rules.
		"ARCH=" + runtime.GOARCH,
		"OS=" + runtime.GOOS,
		// Need this for certain tools, for example sass
		"LANG=" + state.Config.Please.Lang,
		// Use a restricted PATH; it'd be easier for the user if we pass it through
		// but really external environment variables shouldn't affect this.
		// The only concession is that ~ is expanded as the user's home directory
		// in PATH entries.
		"PATH=" + expandHomePath(strings.Join(state.Config.Build.Path, ":")),
	}
	if !test {
		env = append(env,
			"TMP_DIR="+path.Join(RepoRoot, target.TmpDir()),
			"SRCS="+strings.Join(sources, " "),
			"OUTS="+strings.Join(target.Outputs(), " "),
			"NAME="+target.Label.Name,
		)
		tools := make([]string, len(target.Tools))
		for i, tool := range target.Tools {
			tools[i] = state.Graph.TargetOrDie(tool).toolPath()
		}
		env = append(env, "TOOLS="+strings.Join(tools, " "))
		// The OUT variable is only available on rules that have a single output.
		if len(target.Outputs()) == 1 {
			env = append(env, "OUT="+path.Join(RepoRoot, target.TmpDir(), target.Outputs()[0]))
		}
		// The SRC variable is only available on rules that have a single source file.
		if len(sources) == 1 {
			env = append(env, "SRC="+sources[0])
		}
		// Similarly, TOOL is only available on rules with a single tool.
		if len(target.Tools) == 1 {
			env = append(env, "TOOL="+state.Graph.TargetOrDie(target.Tools[0]).toolPath())
		}
		// Named source groups if the target declared any.
		for name, srcs := range target.NamedSources {
			paths := target.SourcePaths(state.Graph, srcs)
			env = append(env, "SRCS_"+strings.ToUpper(name)+"="+strings.Join(paths, " "))
		}
	} else {
		env = append(env, "TEST_DIR="+path.Join(RepoRoot, target.TestDir()))
		if state.NeedCoverage {
			env = append(env, "COVERAGE=true")
		}
	}
	return env
}