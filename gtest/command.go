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

type errHander func(fm string, args ...interface{})

// Report go test report
type Report struct {
	Env struct {
		GoVersion string
		OS        string
		Arch      string
	}
	Creted   time.Time
	Cost     float32
	Packages []*Package
}

// Config run go test config
type Config struct {
	PackagePaths  []string //need run go test for some dir
	ContainImport bool     //need run all for child dir
}

type Service struct {
	Report

	ctx *context.Context

	errh      errHander
	completed chan struct{}
	exit      chan struct{}
}

func New(ctx *context.Context, errh errHander) (*Service, error) {
	return &Service{
		ctx:       ctx,
		errh:      errh,
		completed: make(chan struct{}, 1),
		exit:      make(chan struct{}, 3),
	}, nil
}

func (s *Service) error(msg string) {
	s.errh("gtest: %s", msg)
	s.Stop()
}

func (s *Service) Run() error {

	s.Report.Creted = time.Now()
	s.Report.Env.GoVersion = runtime.Version()
	s.Report.Env.OS = runtime.GOOS
	s.Report.Env.Arch = runtime.GOARCH

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(s.ctx.Packages))
		for _, p := range s.ctx.Packages {

			// batch gofmt
			go func(path string) {
				defer wg.Done()
				select {
				default:
				case <-s.exit:
					return
				}
				pkg, err := run(path, []string{"-cover", "-v"})
				if err != nil {
					s.error(err.Error())
					return
				}
				s.Report.Packages = append(s.Report.Packages, pkg)

			}(p.ImportPath)

			//abort the foreach if exit
			select {
			case <-s.exit:
				break
			default:
			}
		}
		go func() {
			// wait for all go test  done
			wg.Wait()
			s.Report.Cost = float32(time.Since(s.Report.Creted).Seconds())
			close(s.completed)
		}()
	}()
	return nil
}

func (s *Service) Stop() error {
	close(s.exit)
	return nil
}

func (s *Service) Wait() error {
	for {
		select {
		case <-s.exit:
			return nil
		case <-s.completed:
			return nil
		case <-time.After(1 * time.Second):
		}
	}
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
