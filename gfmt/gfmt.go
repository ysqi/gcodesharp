package gfmt

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/ysqi/gcodesharp/context"
)

// Report gofmt result
type Report struct {
	Files   []*File
	GoFmt   string
	Created time.Time
	Cost    float32
	Env     struct {
		GoVersion string
		OS        string
		Arch      string
	}
}

var gofmtpath string

// Config run go test config
type Config struct {
	PackagePaths  []string //need run go test for some dir
	ContainImport bool     //need run all for child dir
	PrintLog      bool
}

// File need format go file
type File struct {
	// Name file name
	Name string
	// Diff instead of rewriting file
	Diff    string
	NeedFmt bool
}

//func New(ctx reporter.ServiceContext) (reporter.Service,error){
//	return
//}

// Run run go fmt and return report
func Run(ctx *context.Context, cfg *Config) (report *Report, err error) {
	if len(cfg.PackagePaths) == 0 {
		cfg.PackagePaths = append(cfg.PackagePaths, ".")
	}
	report = &Report{
		Created: time.Now(),
	}
	report.Env.GoVersion = runtime.Version()
	report.Env.OS = runtime.GOOS
	report.Env.Arch = runtime.GOARCH
	report.GoFmt = gofmtpath
	for _, p := range cfg.PackagePaths {
		files, err := getGoFiles(p)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			report.Files = append(report.Files, &File{
				Name: f,
			})
		}
	}

	if len(report.Files) == 0 {
		return
	}
	// run go fmt  with 20 files every times
	once := 20
	max := len(report.Files)
	for i := 0; i < max; i += once {
		right := i + once
		if right >= max {
			right = max
		}
		err = runGoFmt(report.Files[i:right])
		if err != nil {
			return nil, err
		}
	}
	report.Cost = float32(time.Since(report.Created).Seconds())
	return report, nil
}

var (
	// Match diff print info
	//
	// diff testdata/needFmt.go gofmt/./testdata/needFmt.go
	//
	// diff -u ./testdata/needFmt.go.orig  ./testdata/needFmt.go
	regDiffHead = regexp.MustCompile(`^diff(?: -u){0,1} \S+\s(?:gofmt\/){0,1}(\S+)$`)
)

func runGoFmt(files []*File) error {
	args := []string{}
	for _, f := range files {
		args = append(args, f.Name)
	}
	var (
		stderr bytes.Buffer
		stdout bytes.Buffer
	)
	cmd := exec.Command(gofmtpath, "-d", "-e", "-s")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		s := stderr.String()
		if s != "" {
			return errors.New(stderr.String() + "\n" + err.Error())
		}
		return err
	}

	// read output add add to diffrent file
	var file *File
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line != "" {
			log.Println(line)
		}
		if matches := regDiffHead.FindSubmatch([]byte(line)); len(matches) == 2 {
			//add to file diff
			name := string(matches[1])
			fmt.Println("find diff===>", name)
			for _, f := range files {
				fmt.Println(f.Name == name, f.Name)
				if f.Name == name {
					file = f
					file.NeedFmt = true
				}
			}
			continue
		}
		if file != nil {
			file.Diff += line + "\n"
		}
	}
	return nil
}

// getGoFiles get go file that would be executed by go fmt
// gofmt is gofmt application full path
func getGoFiles(packagepath string) (files []string, err error) {
	if packagepath == "" || packagepath == "." {
		packagepath = "./..."
	} else {
		packagepath += "..."
	}

	cmd := exec.Command("go", "fmt", "-n", packagepath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return files, err
	}

	//output such like this:
	//	/usr/local/go/bin/gofmt -l -w command.go main.go
	// 	/usr/local/go/bin/gofmt -l -w context/context.go context/path.go context/path_test.go
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		arr := strings.Split(line, " ")
		if len(arr) <= 3 {
			continue
		}
		for _, f := range arr[3:] {
			if strings.Contains(f, "vendor/") {
				continue
			}
			files = append(files, f)
		}
	}
	return
}

func init() {
	gofmtpath = context.GofmtPath()
}
