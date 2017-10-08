package gtest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/ysqi/gcodesharp/context"
)

// Report go test report
type Report struct {
	Env struct {
		GoVersion string
		OS        string
		Arch      string
	}
	Creted   time.Time
	Packages []*Package
}

// Config run go test config
type Config struct {
	PackagePaths  []string //need run go test for some dir
	ContainImport bool     //need run all for child dir
}

// Run go test command.
// return the test result info and realtime print info with logger.
func Run(ctx *context.Context, cfg *Config) (report *Report, err error) {
	if len(cfg.PackagePaths) == 0 {
		cfg.PackagePaths = append(cfg.PackagePaths, ".")
	}
	report = &Report{
		Packages: []*Package{},
		Creted:   time.Now(),
	}
	report.Env.GoVersion = runtime.Version()
	report.Env.OS = runtime.GOOS
	report.Env.Arch = runtime.GOARCH

	packagepaths := []string{}
	// add path
	for _, p := range cfg.PackagePaths {
		if cfg.ContainImport {
			list, err := context.GetPackagePaths(p)
			if err != nil {
				return nil, err
			}
			// add all child import path to cmd.
			// note:contains self.
			packagepaths = append(packagepaths, list...)
		} else {
			packagepaths = append(packagepaths, p)
		}
	}

	for _, p := range packagepaths {
		args := []string{"-cover", "-v", "-timeout", "3s"}
		pkg, err := run(p, args)
		if err != nil {
			return nil, err
		}
		report.Packages = append(report.Packages, pkg)
	}

	return report, nil
}

func run(packagepath string, args []string) (pkg *Package, err error) {
	var (
		stderr bytes.Buffer
		stdout io.ReadCloser
	)
	// TODO: need support more args
	cmd := exec.Command("go", "test", packagepath)
	cmd.Args = append(cmd.Args, args...)

	cmd.Stderr = &stderr
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdout)

	pkg = &Package{
		Name: packagepath,
	}
	// need wait for output process compeled.
	wg := sync.WaitGroup{}
	wg.Add(1)
	setLastErr := func(err interface{}) {
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
		pkg.Failed = true
		if pkg != nil {
			if regPanic.Match([]byte(errStr)) {
				// the last test panic error
				// set error info to last test
				setLastErr(errStr)
			}
			errStr = ""
		}
		if _, ok := err.(*exec.ExitError); !ok {
			errStr = appendLine(errStr, err.Error())
		}
		pkg.Err = appendLine(pkg.Err, errStr)
	}
	return pkg, nil
}
