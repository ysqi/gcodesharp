// Copyright (C) 2017. author ysqi(devysq@gmail.com).
//
// The gcodesharp is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gcodesharp is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package glint

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
		Name:      r.ExecPath,
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
	className := filepath.Base(r.ExecPath)
	// individual test cases
	for _, test := range r.Files {
		if !test.HasProblem() {
			// don't the go files about gofmt check success.
			// will have a big test case in Junit if do that.
			continue
		}
		testCase := formater.JUnitTestCase{
			Classname: className,
			Name:      test.Name,
		}

		testCase.Failure = &formater.JUnitFailure{
			Message:  fmt.Sprintf("golint %s", filepath.Base(test.Name)),
			Type:     "WARNING",
			Contents: test.ProblemContent(),
		}

		ts.TestCases = append(ts.TestCases, testCase)
		ts.Failures++
	}
	ts.Tests = len(ts.TestCases)
	return formater.JUnitTestSuites{
		Suites: []formater.JUnitTestSuite{ts},
	}, nil

}
