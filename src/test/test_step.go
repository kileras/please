package test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/op/go-logging"

	"build"
	"core"
)

var log = logging.MustGetLogger("test")

const dummyOutput = "=== RUN DummyTest\n--- PASS: DummyTest (0.00s)\nPASS\n"
const dummyCoverage = "<?xml version=\"1.0\" ?><coverage></coverage>"

func Test(tid int, state *core.BuildState, label core.BuildLabel) {
	state.LogBuildResult(tid, label, core.TargetTesting, "Testing...")
	startTime := time.Now()
	target := state.Graph.TargetOrDie(label)
	hash, err := build.RuntimeHash(state, target)
	if err != nil {
		state.LogBuildError(tid, label, core.TargetTestFailed, err, "Failed to calculate target hash")
		return
	}
	// Check the cached output files if the target wasn't rebuilt.
	hashStr := base64.RawURLEncoding.EncodeToString(core.CollapseHash(hash))
	resultsFileName := fmt.Sprintf(".test_results_%s_%s", label.Name, hashStr)
	coverageFileName := fmt.Sprintf(".test_coverage_%s_%s", label.Name, hashStr)
	outputFile := path.Join(target.TestDir(), "test.results")
	coverageFile := path.Join(target.TestDir(), "test.coverage")
	cachedOutputFile := path.Join(target.OutDir(), resultsFileName)
	cachedCoverageFile := path.Join(target.OutDir(), coverageFileName)
	needCoverage := state.NeedCoverage && !target.NoTestOutput

	cachedTest := func() {
		log.Debug("Not re-running test %s; got cached results.", label)
		coverage := parseCoverageFile(target, cachedCoverageFile)
		results, err := parseTestResults(target, cachedOutputFile, 0, true)
		target.Results.Duration = time.Since(startTime).Seconds()
		target.Results.Cached = true
		if err != nil {
			state.LogBuildError(tid, label, core.TargetTestFailed, err, "Failed to parse cached test file %s", cachedOutputFile)
		} else if results.Failed > 0 {
			panic("Test results with failures shouldn't be cached.")
		} else {
			logTestSuccess(state, tid, label, results, coverage)
		}
	}

	moveAndCacheOutputFiles := func(results core.TestResults, coverage core.TestCoverage) bool {
		// Never cache test results when given arguments; the results may be incomplete.
		if len(state.TestArgs) > 0 {
			log.Debug("Not caching results for %s, we passed it arguments", label)
			return true
		}
		if err := moveAndCacheOutputFile(state, target, hash, outputFile, cachedOutputFile, resultsFileName, dummyOutput); err != nil {
			state.LogTestResult(tid, label, core.TargetTestFailed, results, coverage, err, "Failed to move test output file")
			return false
		}
		if needCoverage || core.PathExists(coverageFile) {
			if err := moveAndCacheOutputFile(state, target, hash, coverageFile, cachedCoverageFile, coverageFileName, dummyCoverage); err != nil {
				state.LogTestResult(tid, label, core.TargetTestFailed, results, coverage, err, "Failed to move test coverage file")
				return false
			}
		}
		for _, output := range target.TestOutputs {
			tmpFile := path.Join(target.TestDir(), output)
			outFile := path.Join(target.OutDir(), output)
			if err := moveAndCacheOutputFile(state, target, hash, tmpFile, outFile, output, ""); err != nil {
				state.LogTestResult(tid, label, core.TargetTestFailed, results, coverage, err, "Failed to move test output file")
				return false
			}
		}
		return true
	}

	needToRun := func() bool {
		if target.State() == core.Unchanged && core.PathExists(cachedOutputFile) {
			// Output file exists already and appears to be valid. We might still need to rerun though
			// if the coverage files aren't available.
			if needCoverage && !core.PathExists(cachedCoverageFile) {
				return true
			}
			return false
		}
		// Check the cache for these artifacts.
		if state.Cache == nil {
			return true
		}
		cache := *state.Cache
		if !cache.RetrieveExtra(target, hash, resultsFileName) {
			return true
		}
		if needCoverage && !cache.RetrieveExtra(target, hash, coverageFileName) {
			return true
		}
		for _, output := range target.TestOutputs {
			if !cache.RetrieveExtra(target, hash, output) {
				return true
			}
		}
		return false
	}

	// Don't cache when doing multiple runs, presumably the user explicitly wants to check it.
	if state.NumTestRuns == 1 && !needToRun() {
		cachedTest()
		return
	}
	// Remove any cached test result file.
	if err := RemoveCachedTestFiles(target); err != nil {
		state.LogBuildError(tid, label, core.TargetTestFailed, err, "Failed to remove cached test files")
		return
	}
	numSucceeded := 0
	for i := 0; i < state.NumTestRuns; i++ {
		if state.NumTestRuns > 1 {
			state.LogBuildResult(tid, label, core.TargetTesting, fmt.Sprintf("Testing (%d of %d)...", i, state.NumTestRuns))
		}
		out, err, flakes := runTestWithRetries(tid, state, target)
		duration := time.Since(startTime).Seconds()
		startTime = time.Now() // reset this for next time

		// This is all pretty involved; there are lots of different possibilities of what could happen.
		// The contract is that the test must return zero on success or non-zero on failure (Unix FTW).
		// If it's successful, it must produce a parseable file named "test.results" in its temp folder.
		// (alternatively, this can be a directory containing parseable files).
		// Tests can opt out of the file requirement individually, in which case they're judged only
		// by their return value.
		// But of course, we still have to consider all the alternatives here and handle them nicely.
		target.Results.Output = string(out)
		if err != nil && target.Results.Output == "" {
			target.Results.Output = err.Error()
		}
		if err != nil {
			_, target.Results.TimedOut = err.(core.TimeoutError)
		}
		coverage := parseCoverageFile(target, coverageFile)
		if !core.PathExists(outputFile) {
			target.Results.Duration += duration
			if err == nil && target.NoTestOutput {
				results := core.TestResults{NumTests: 1, Passed: 1, Flakes: flakes}
				if moveAndCacheOutputFiles(results, coverage) {
					target.Results.NumTests = 1
					target.Results.Passed = 1
					logTestSuccess(state, tid, label, results, coverage)
				}
			} else if err == nil {
				target.Results.NumTests++
				target.Results.Failed++
				err = fmt.Errorf("Test failed to produce output results file")
				state.LogBuildError(tid, label, core.TargetTestFailed, err,
					"Test apparently succeeded but failed to produce %s. Output: %s", outputFile, string(out))
			} else {
				target.Results.NumTests++
				target.Results.Failed++
				state.LogBuildError(tid, label, core.TargetTestFailed, err,
					fmt.Sprintf("Test failed with no results. Output: %s", string(out)))
			}
		} else {
			results, err2 := parseTestResults(target, outputFile, flakes, false)
			target.Results.Duration += duration
			if err2 != nil {
				state.LogBuildError(tid, label, core.TargetTestFailed, err2,
					"Couldn't parse test output file: %s. Stdout: %s", err2, string(out))
			} else if err != nil && results.Failed == 0 {
				// Add a failure result to the test so it shows up in the final aggregation.
				results.Failed = 1
				results.Failures = append(results.Failures, core.TestFailure{
					Name:   "Return value",
					Type:   fmt.Sprintf("%s", err),
					Stdout: string(out),
				})
				state.LogTestResult(tid, label, core.TargetTestFailed, results, coverage, err,
					"Test returned nonzero but reported no errors: %s. Output: %s", err, string(out))
			} else if err == nil && results.Failed != 0 {
				err = fmt.Errorf("Test returned 0 but still reported failures")
				state.LogTestResult(tid, label, core.TargetTestFailed, results, coverage, err,
					"Test returned 0 but still reported failures. Stdout: %s", string(out))
			} else if results.Failed != 0 {
				err = fmt.Errorf("Tests failed")
				state.LogTestResult(tid, label, core.TargetTestFailed, results, coverage, err,
					fmt.Sprintf("Tests failed. Stdout: %s", string(out)))
			} else {
				logTestSuccess(state, tid, label, results, coverage)
				numSucceeded++
				// Cache only on the last run.
				if numSucceeded == state.NumTestRuns {
					moveAndCacheOutputFiles(results, coverage)
				}
				// Clean up the test directory.
				if state.CleanWorkdirs {
					if err := os.RemoveAll(target.TestDir()); err != nil {
						log.Warning("Failed to remove test directory for %s: %s", target.Label, err)
					}
				}
			}
		}
	}
}

func logTestSuccess(state *core.BuildState, tid int, label core.BuildLabel, results core.TestResults, coverage core.TestCoverage) {
	var description string
	tests := pluralise("test", results.NumTests)
	if results.Skipped != 0 || results.ExpectedFailures != 0 {
		failures := pluralise("failure", results.ExpectedFailures)
		description = fmt.Sprintf("%d %s passed. %d skipped, %d expected %s",
			results.NumTests, tests, results.Skipped, results.ExpectedFailures, failures)
	} else {
		description = fmt.Sprintf("%d %s passed.", results.NumTests, tests)
	}
	state.LogTestResult(tid, label, core.TargetTested, results, coverage, nil, description)
}

func pluralise(word string, quantity int) string {
	if quantity == 1 {
		return word
	}
	return word + "s"
}

func prepareTestDir(graph *core.BuildGraph, target *core.BuildTarget) error {
	if err := os.RemoveAll(target.TestDir()); err != nil {
		return err
	}
	if err := os.MkdirAll(target.TestDir(), core.DirPermissions); err != nil {
		return err
	}
	for out := range core.IterRuntimeFiles(graph, target, true) {
		if err := core.PrepareSourcePair(out); err != nil {
			return err
		}
	}
	return nil
}

func runTest(state *core.BuildState, target *core.BuildTarget, timeout int) ([]byte, error) {
	replacedCmd := build.ReplaceTestSequences(target, target.TestCommand)
	env := core.BuildEnvironment(state, target, true)
	if len(state.TestArgs) > 0 {
		args := strings.Join(state.TestArgs, " ")
		replacedCmd += " " + args
		env = append(env, "TESTS="+args)
	}
	cmd := exec.Command("bash", "-c", replacedCmd)
	cmd.Dir = target.TestDir()
	cmd.Env = env
	log.Debug("Running test %s\nENVIRONMENT:\n%s\n%s", target.Label, strings.Join(cmd.Env, "\n"), replacedCmd)
	if state.PrintCommands {
		log.Notice("Running test %s: %s", target.Label, replacedCmd)
	}
	return core.ExecWithTimeout(cmd, target.TestTimeout, timeout)
}

// Runs a test some number of times as indicated by its flakiness.
func runTestWithRetries(tid int, state *core.BuildState, target *core.BuildTarget) (out []byte, err error, flakiness int) {
	flakiness = target.Flakiness
	if flakiness == 0 {
		flakiness = 1
	} // Flakiness of 0 -> run it once. Equivalent behaviour to 1 but more intuitive.
	if state.MaxFlakes > 0 && flakiness > state.MaxFlakes {
		flakiness = state.MaxFlakes // Cap max flakes by flag
	}
	for i := 0; i < flakiness; i++ {
		// Re-prepare test directory between each attempt so they can't accidentally contaminate each other.
		if err = prepareTestDir(state.Graph, target); err != nil {
			state.LogBuildError(tid, target.Label, core.TargetTestFailed, err, "Failed to prepare test directory for %s: %s", target.Label, err)
			return []byte{}, err, i
		}
		out, err = runPossiblyContainerisedTest(state, target)
		if err == nil {
			return out, err, i
		} else if i < flakiness-1 {
			log.Warning("%s failed on attempt %d (%d more to go).", target.Label, i+1, flakiness-i-1)
		}
	}
	if target.Flakiness == 0 {
		flakiness = 0
	} // Reset this again so non-flaky targets don't appear so.
	return out, err, flakiness
}

// Parses the coverage output for a single target.
func parseCoverageFile(target *core.BuildTarget, coverageFile string) core.TestCoverage {
	coverage, err := parseTestCoverage(target, coverageFile)
	if err != nil {
		log.Errorf("Failed to parse coverage file for %s: %s", target.Label, err)
	}
	return coverage
}

// RemoveCachedTestFiles removes any cached test or coverage result files for a target.
func RemoveCachedTestFiles(target *core.BuildTarget) error {
	if err := removeAnyFilesWithPrefix(target.OutDir(), ".test_results_"+target.Label.Name); err != nil {
		return err
	}
	if err := removeAnyFilesWithPrefix(target.OutDir(), ".test_coverage_"+target.Label.Name); err != nil {
		return err
	}
	for _, output := range target.TestOutputs {
		if err := os.RemoveAll(path.Join(target.OutDir(), output)); err != nil {
			return err
		}
	}
	return nil
}

// removeAnyFilesWithPrefix deletes any files in a directory matching a given prefix.
func removeAnyFilesWithPrefix(dir, prefix string) error {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, info := range infos {
		if strings.HasPrefix(info.Name(), prefix) {
			if err := os.RemoveAll(path.Join(dir, info.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// Attempt to write a dummy coverage file to record that it's been done for a test.
func moveAndCacheOutputFile(state *core.BuildState, target *core.BuildTarget, hash []byte, from, to, filename, dummy string) error {
	if !core.PathExists(from) {
		if dummy == "" {
			return nil
		}
		if err := ioutil.WriteFile(to, []byte(dummy), 0644); err != nil {
			return err
		}
	} else if err := os.Rename(from, to); err != nil {
		return err
	}
	if state.Cache != nil {
		(*state.Cache).StoreExtra(target, hash, filename)
	}
	return nil
}