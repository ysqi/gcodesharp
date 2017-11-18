package gfmt

import (
	"fmt"
	"path/filepath"

	"github.com/ysqi/gcodesharp/reporter/formater"
)

// ToJunit convert Report to JUnit test suites.
// Just add need format go file as test case in Junit.
// and the failure type is Warning
func (r *Report) ToJunit() (formater.JUnitTestSuites, error) {
	ts := formater.JUnitTestSuite{
		Time:      r.Cost,
		Name:      r.GoFmt,
		Timestamp: r.Created.UTC().Format("2006-01-02T15:04:05"), //ISO8601
	}
	ts.Properties = []formater.JUnitProperty{
		{"go.version", r.Env.GoVersion},
		{"os", r.Env.OS},
		{"arch", r.Env.Arch},
	}

	if r.SysErr != nil {
		ts.Err = r.SysErr.Error()
	}
	className := filepath.Base(r.GoFmt)
	// individual test cases
	for _, test := range r.Files {
		if !test.NeedFmt {
			// don't the go files about gofmt check success.
			// will have a big test case in Junit if do that.
			continue
		}
		testCase := formater.JUnitTestCase{
			Classname: className,
			Name:      test.Name,
		}
		testCase.Failure = &formater.JUnitFailure{
			Message:  fmt.Sprintf("gofmt -d -e %s", filepath.Base(test.Name)),
			Type:     "WARNING",
			Contents: test.Diff,
		}

		ts.TestCases = append(ts.TestCases, testCase)
		ts.Failures++
	}
	ts.Tests = len(ts.TestCases)
	return formater.JUnitTestSuites{
		Suites: []formater.JUnitTestSuite{ts},
	}, nil

}
