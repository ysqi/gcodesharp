package gtest

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
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
	PackageDirs   []string //need run go test for some dir
	ContainImport bool     //need run all for child dir

	importpaths []string
}

// check and find package import path.
func check(ctx *context.Context, cfg *Config) error {
	if len(cfg.PackageDirs) == 0 {
		cfg.PackageDirs = append(cfg.PackageDirs, ".")
	}
	for _, dir := range cfg.PackageDirs {
		importPath, _, err := ctx.FindImportPath(dir)
		if err != nil {
			return err
		}
		cfg.importpaths = append(cfg.importpaths, importPath)
	}
	return nil
}

// Run go test command.
// return the test result info and realtime print info with logger.
func Run(ctx *context.Context, cfg *Config) (report *Report, err error) {
	if err = check(ctx, cfg); err != nil {
		return nil, err
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
	for _, p := range cfg.importpaths {
		if cfg.ContainImport {
			list, err := getPackageList(p)
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
		if pkg != nil && regPanic.Match([]byte(errStr)) {
			// the last test panic error
			// set error info to last test
			setLastErr(errStr)
			errStr = ""
		}
		if errStr == "" {
			errStr = err.Error()
		} else {
			errStr = appendLine(errStr, err.Error())
		}
		pkg.Err = errStr
	}
	return pkg, nil
}

// getPackageList get all import path prefixed with input
func getPackageList(pkgpath string) ([]string, error) {
	if pkgpath == "" {
		pkgpath = "./..."
	} else if pkgpath == "." {
		pkgpath = "./..."
	} else {
		pkgpath += "..."
	}
	list := []string{}
	cmd := exec.Command("go", "list", pkgpath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return list, err
	}
	outputStr := string(output)
	// e.g warning:,
	if strings.Contains(outputStr, ":") {
		return list, errors.New(outputStr)
	}
	for _, line := range strings.Split(outputStr, "\n") {
		// ignore vendor path
		if strings.Contains(line, "/vendor/") {
			continue
		}
		list = append(list, line)
	}
	return list, nil
}
