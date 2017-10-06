package gtest

//go:generate stringer -type=Result

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Result represents a test result.
type Result int

// Test result constants
const (
	PASS Result = iota
	FAIL
	SKIP
)

func toResult(name string) Result {
	switch name {
	case "PASS":
		return PASS
	case "FAIL":
		return FAIL
	case "SKIP":
		return SKIP
	}
	panic(fmt.Errorf("can't parse %q to Result enum", name))
}

// Unit single test funcation
type Unit struct {
	Name   string
	Cost   float32
	Result Result
	Output string
}

// Package is a single package that contains test results
type Package struct {
	Name string
	Cost float32
	// Coverage package test coverage
	// TODO: need support with '-coverpkg' arg，the result such like :
	// 		ok      github.com/ysqi/gcodereview/gtest       0.011s  coverage: 12.5% of statements in fmt
	// Current only support that (full package testting converage):
	//		ok      github.com/ysqi/gcodereview/gtest       0.010s  coverage: 47.7% of statements
	Coverage float32
	Failed   bool
	Err      string
	Units    []*Unit
}

func (pkg *Package) getCount(r Result) int {
	count := 0
	for _, unit := range pkg.Units {
		if unit.Result == r {
			count++
		}
	}
	return count
}

// FailCount counts the number of failed tests
func (pkg *Package) FailCount() int {
	return pkg.getCount(FAIL)
}

// PassCount counts the number of pass tests
func (pkg *Package) PassCount() int {
	return pkg.getCount(PASS)

}

// SkipCount counts the number of skip tests
func (pkg *Package) SkipCount() int {
	return pkg.getCount(SKIP)
}

// HasCoverage check this package have contains coverage info.
// create new package with coverage=-1.0.
func (pkg *Package) HasCoverage() bool {
	return pkg.Coverage >= 0.00
}

// GetByResult seach the same result of all unit test
func (pkg *Package) GetByResult(r Result) []*Unit {
	s := []*Unit{}
	for _, unit := range pkg.Units {
		if unit.Result == r {
			s = append(s, unit)
		}
	}
	return s
}

// ExecGoTest run go test command.
// return the test result info and realtime print info with logger.
func ExecGoTest(pkgname string) (pkg *Package, err error) {
	var (
		stderr bytes.Buffer
		stdout io.ReadCloser
	)
	// TODO: need support more args
	cmd := exec.Command("go", "test", pkgname, "-cover", "-v", "-timeout", "3s")
	cmd.Stderr = &stderr
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdout)

	// need wait for output process compeled.
	wg := sync.WaitGroup{}
	wg.Add(1)
	setLastErr := func(err interface{}) {
		if pkg == nil {
			return
		}
		pkg.Failed = true
		if len(pkg.Units) == 0 {
			return
		}
		errStr := ""
		switch v := err.(type) {
		case error:
			errStr = v.Error()
		case string:
			errStr = v
		default:
			errStr = fmt.Sprintf("%v", v)
		}
		pkg.Units[len(pkg.Units)-1].Output = errStr
		pkg.Units[len(pkg.Units)-1].Result = FAIL
	}
	defer func() {
		if err := recover(); err != nil {
			setLastErr(err)
		}
	}()
	go func() {
		var pkgs []*Package
		pkgs, err = parse(scanner, true)
		if err == nil && len(pkgs) > 0 {
			pkg = pkgs[0]
		}
		wg.Done()
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	wg.Wait()
	if err = cmd.Wait(); err != nil {
		errStr := stderr.String()
		if pkg == nil {
			return nil, errors.New("ExecGoTest:" + errStr)
		}
		pkg.Failed = true
		if pkg != nil && regPanic.Match([]byte(errStr)) {
			// the last test panic error
			// set error info to last test
			setLastErr(errStr)
			pkg.Err = err.Error()
		} else {
			if errStr == "" {
				errStr = err.Error()
			} else {
				errStr = appendLine(errStr, err.Error())
			}
			pkg.Err = errStr
		}
	}
	return pkg, nil
}

var (
	// panic error
	regPanic = regexp.MustCompile(`^panic: (.* \[recovered\])|(test timed out after)`)
	// coverage info ,the string look like :coverage: 36.4% of statements
	regCoverage = regexp.MustCompile(`^coverage: (\d+\.{0,1}\d+)% of statements(?:\sin\s.+)?`)

	// test method pass,like:
	//	--- PASS: TestAddressHexChecksum (0.00s)
	//  --- FAIL: TestAddressHexChecksum (0.02s)
	//  --- SKIP: TestAddressHexChecksum (0.00s)
	regStatus = regexp.MustCompile(`\t*--- (PASS|FAIL|SKIP): (.+) \((\d+\.\d+)(?: seconds|s)\)`)
	// one test running,  === RUN   TestParse
	regUnitTestStart = regexp.MustCompile(`^\t*=== RUN\s+(\S+)$`)
	// package test result,like:
	//	ok          github.com/ysqi/com     1.211s
	//	ok          github.com/ysqi/com     0.00s	[no tests to run]
	//	FAIL        github.com/ysqi/com     0.005s
	//  FAIL	github.com/ysqi/com [setup failed]
	//  ?	github.com/ysqi/com 	[no test files]
	regexResult = regexp.MustCompile(`(ok|FAIL|\?)\s+([^ ]+)\s+(?:(\d+\.\d+)s|\[([\w\s]+)\])(?:\s+coverage:\s+(\d+\.{0,1}\d+)%\sof\sstatements(?:\sin\s.+)?)?`)
	// regexResult = regexp.MustCompile(`^(ok|FAIL|\?)\s+([^ ]+)\s+(?:(\d+\.\d+)\s|(\[\w+ failed\]))(?:\s+coverage:\s+(\d+\.\d+)%\sof\sstatements(?:\sin\s.+)?)?$`)
)

func parse(scanner *bufio.Scanner, logprint bool) ([]*Package, error) {
	var (
		// pakcage array
		pkgs = []*Package{}
		// current line content is package errror info flag
		nextIsPkgError bool
		// current unit test
		curUnit *Unit
		// current package
		pkg    *Package
		newPkg = func(name string) *Package {
			nextIsPkgError = false
			return &Package{
				Name:     name,
				Coverage: -1,
			}
		}
	)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if logprint {
			log.Println(line)
		}
		if pkg == nil {
			pkg = newPkg("")
		}

		data := []byte(line)
		if matches := regUnitTestStart.FindSubmatch(data); len(matches) == 2 {
			// if current package is failed ,then this package has completed test.
			// need create a new package
			if pkg.Failed {
				pkg = newPkg("")
			}
			curUnit = &Unit{
				Name: string(matches[1]),
			}
			pkg.Units = append(pkg.Units, curUnit)
			continue
		}
		if matches := regStatus.FindSubmatch(data); len(matches) == 4 {
			//e.g:	--- PASS: TestAddressHexChecksum (0.00s)
			// the unit must be added when found '=== RUN testname' line.
			curUnit = findUnitTest(pkg.Units, string(matches[2]))
			curUnit.Cost = mustFloat32(matches[3])
			curUnit.Result = toResult(string(matches[1]))
			continue
		}
		if matches := regexResult.FindSubmatch(data); len(matches) == 6 {
			pkg.Name = string(matches[2])
			if p := findPkg(pkgs, pkg.Name); p == nil {
				pkgs = append(pkgs, pkg)
			} else {
				pkg = p
			}
			// e.g:	FAIL        github.com/ysqi/com/one [build failed]
			if string(matches[1]) != "ok" {
				pkg.Failed = true
			}
			pkg.Cost = mustFloat32(matches[3])
			if pkg.Failed && len(matches[4]) > 0 {
				pkg.Err = appendLine(pkg.Err, string(matches[4]))
			}
			// e.g: ok      github.com/ysqi/gcodereview/gtest       0.024s  coverage: 60.6% of statements
			if string(matches[5]) != "" {
				pkg.Coverage = mustFloat32(matches[5])
			}
			// reset
			pkg = nil
			curUnit = nil
			continue
		}

		if line == "FAIL" || strings.HasPrefix(line, "exit status ") {
			pkg.Failed = true
			continue
		}
		if line == "PASS" {
			pkg.Failed = false
			continue
		}
		if matches := regCoverage.FindSubmatch(data); matches != nil {
			// e.g:	coverage: 36.4% of statements
			pkg.Coverage = mustFloat32(matches[1])
			continue
		}
		if strings.HasPrefix(line, "# ") {
			pkg = newPkg(strings.TrimLeft(line, "# "))
			nextIsPkgError = true
			pkgs = append(pkgs, pkg)
			continue
		}
		if nextIsPkgError {
			pkg.Err = appendLine(pkg.Err, line)
			pkg.Failed = true
			continue
		}
		if curUnit != nil {
			curUnit.Output = appendLine(curUnit.Output, line)
		}
	}
	return pkgs, nil
}

// mustFloat32 convert bytes to float number.
// os exist with error if parse failed.
func mustFloat32(b []byte) float32 {
	if len(b) == 0 {
		return 0
	}
	val, err := strconv.ParseFloat(string(b), 10)
	if err != nil {
		log.Fatal(err)
	}
	return float32(val)
}

// find pakcage from the list by name.
func findPkg(pkgs []*Package, name string) *Package {
	for _, p := range pkgs {
		if p.Name == name {
			return p
		}
	}
	return nil
}
func findUnitTest(tests []*Unit, name string) *Unit {
	for _, u := range tests {
		if u.Name == name {
			return u
		}
	}
	return nil
}

func appendLine(str, line string) string {
	if str != "" {
		str += "\n"
	}
	return str + line
}
