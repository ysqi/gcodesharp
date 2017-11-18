package gtest

import (
	"fmt"
	"strings"

	"github.com/ysqi/gcodesharp/reporter/formater"
)

// ToJunit convert Report to JUnit test suites.
// Just add need format go file as test case in Junit.
// and the failure type is Warning
func (r *Report) ToJunit() (formater.JUnitTestSuites, error) {
	suites := formater.JUnitTestSuites{}

	// convert Report to JUnit test suites
	for _, pkg := range r.Packages {

		ts := formater.JUnitTestSuite{
			Tests: len(pkg.Units),
			Time:  pkg.Cost,
			Name:  pkg.Name,
			//Properties: []JUnitProperty{},
			//TestCases:  []JUnitTestCase{},
			Failures:  pkg.FailCount(),
			Timestamp: pkg.Runtime.UTC().Format("2006-01-02T15:04:05"), //ISO8601
		}
		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

		// just add info to first test suite.
		if len(suites.Suites) == 0 {
			ts.Properties = []formater.JUnitProperty{
				{"go.version", r.Env.GoVersion},
				{"os", r.Env.OS},
				{"arch", r.Env.Arch},
			}
		}
		if pkg.HasCoverage() {
			ts.Properties = append(ts.Properties,
				formater.JUnitProperty{
					Name:  "coverage.statements.pct",
					Value: fmt.Sprintf("%.2f", pkg.Coverage)})
		}
		if pkg.Failed {
			ts.Err = pkg.Err
		}

		// individual test cases
		for _, test := range pkg.Units {
			testCase := formater.JUnitTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      test.Cost,
				Failure:   nil,
			}

			if test.Result == FAIL {
				testCase.Failure = &formater.JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: test.Output,
				}
			}

			if test.Result == SKIP {
				testCase.SkipMessage = &formater.JUnitSkipMessage{test.Output}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}
		suites.Suites = append(suites.Suites, ts)
	}

	return suites, nil
}
