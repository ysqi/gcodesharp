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

package gtest

import (
	// "bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Note: change from https://github.com/jstemmer/go-junit-report/blob/master/junit-formatter.go

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []JUnitTestSuite
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Errors     int             `xml:"errors,attr"`
	Time       float32         `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Timestamp  string          `xml:"timestamp,attr"`
	Err        string          `xml:"system-err,omitempty"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        float32           `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

// JUnitReportXML writes a JUnit xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/JUnit.xsd
func JUnitReportXML(report *Report, noXMLHeader bool, w io.Writer) error {
	suites := JUnitTestSuites{}

	// convert Report to JUnit test suites
	for _, pkg := range report.Packages {
		ts := JUnitTestSuite{
			Tests:      len(pkg.Units),
			Time:       pkg.Cost,
			Name:       pkg.Name,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
			Failures:   pkg.FailCount(),
			Timestamp:  pkg.Runtime.UTC().Format("2006-01-02T15:04:05"), //ISO8601
		}
		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

		ts.Properties = append(ts.Properties,
			JUnitProperty{"go.version", report.Env.GoVersion},
			JUnitProperty{"os", report.Env.OS},
			JUnitProperty{"arch", report.Env.Arch},
		)
		if pkg.HasCoverage() {
			ts.Properties = append(ts.Properties, JUnitProperty{"coverage.statements.pct", fmt.Sprintf("%.2f", pkg.Coverage)})
		}
		if pkg.Failed {
			ts.Err = pkg.Err
		}

		// individual test cases
		for _, test := range pkg.Units {
			testCase := JUnitTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      test.Cost,
				Failure:   nil,
			}

			if test.Result == FAIL {
				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: test.Output,
				}
			}

			if test.Result == SKIP {
				testCase.SkipMessage = &JUnitSkipMessage{test.Output}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if !noXMLHeader {
		w.Write([]byte(xml.Header))
	}
	if err := enc.Encode(suites); err != nil {
		return err
	}
	w.Write([]byte("\n"))
	if err := enc.Flush(); err != nil {
		return err
	}
	return nil
}
