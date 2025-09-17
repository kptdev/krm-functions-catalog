package e2etest

import (
	"testing"

	"github.com/kptdev/kpt/pkg/test/runner"
)

// TestE2E accepts a path and scans the path to find all available packages that can
// be tested. A package which contains a directory '.expected' is considered testable.
//
// The kpt package should contain a declarative function that will be tested.
//
// This driver expects a directory '.expected' in the top level of the package and
// '.expected' should contain 3 files:
//  - 'config.yaml' has the configuration information for this test. It should contain a
//    map of following fields:
//    - 'exitCode': Contains a single number which is expected exit code of command
//      'kpt fn run'. If this field is missed, driver will assume the expected exit code
//      is 0.
//    - 'network': A boolean which indicates whether network should be enabled
//    	for this test. If this field exists and the content in it is 'true' then the
//    	network is accessible. Otherwise the function cannot access network.
//    - 'runCount': A number which sepcifies the times of function running. If this field
//      missing then the function will be run *twice*.
//  - 'diff.patch' is the expected diff output between original package files and
//    files after function running. The diff will be compared only when the exit code
//    matches expected and is zero.
//  - 'results.yaml' is the expected results output from the command 'kpt fn eval'.
//    The results will be compared only when the exit code matches expected and is not
//    zero.
//
// Given a package's name is 'my-pkg', this driver will copy the package into a temporary
// directory and then run command 'kpt fn run my-pkg --results-dir results'. The test
// will pass when the diff output, results output and exit code are all matched with
// expected.
//
// Git is required to generate diff output.

func TestE2E(t *testing.T) {
	runTests(t, "../..")
}

// runTests will scan test cases in 'path', run the command
// on all of the packages in path, and test that
// the diff between the results and the original package is as
// expected
func runTests(t *testing.T, path string) {
	cases, err := runner.ScanTestCases(path)
	if err != nil {
		t.Fatalf("failed to scan test cases: %s", err)
	}
	setImagePullPolicyToIfNotPresent(*cases)
	for _, c := range *cases {
		c := c // capture range variable
		t.Run(c.Path, func(t *testing.T) {
			t.Parallel()
			r, err := runner.NewRunner(t, c, c.Config.TestType)
			if err != nil {
				t.Fatalf("failed to create test runner: %s", err)
			}
			if r.Skip() {
				t.Skip()
			}
			err = r.Run()
			if err != nil {
				t.Fatalf("failed when running test: %s", err)
			}
		})
	}
}

func setImagePullPolicyToIfNotPresent(testcases []runner.TestCase) {
	for i := range testcases {
		testcases[i].Config.ImagePullPolicy = "never"
	}
}
