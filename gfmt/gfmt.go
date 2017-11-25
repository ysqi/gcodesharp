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

package gfmt

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ysqi/gcodesharp/context"
)

type errHander func(fm string, args ...interface{})

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
	SysErr error
}

var gofmtpath string

func init() {
	gofmtpath = context.GofmtPath()
}

// File need format go file
type File struct {
	// Name file name
	Name string
	// Diff instead of rewriting file
	Diff    string
	NeedFmt bool
}

type Service struct {
	Report

	ctx *context.Context

	running   bool
	completed chan struct{}
	errh      errHander
	exit      chan struct{}
	waitGroup sync.WaitGroup

	sync.Mutex
}

func New(ctx *context.Context, errh errHander) (*Service, error) {
	return &Service{
		ctx:  ctx,
		errh: errh,

		completed: make(chan struct{}, 1),
		exit:      make(chan struct{}, 3),
	}, nil
}
func (s *Service) error(msg string) {
	s.SysErr = errors.New(msg)
	s.errh("gfmt: %s", msg)
	s.Stop()
}

// Run go fmt
func (s *Service) Run() error {
	if s.running {
		return errors.New("gfmt is running")
	}
	s.Lock()
	defer s.Unlock()

	s.Created = time.Now()
	s.Env.GoVersion = runtime.Version()
	s.Env.OS = runtime.GOOS
	s.Env.Arch = runtime.GOARCH
	s.GoFmt = gofmtpath

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(s.ctx.Packages))
		for _, p := range s.ctx.Packages {
			files := p.GoFiles
			// absolute path.
			for i := 0; i < len(files); i++ {
				if !filepath.IsAbs(files[i]) {
					files[i] = filepath.Join(p.Dir, files[i])
				}
			}
			// batch gofmt
			go func(files []string) {
				defer wg.Done()
				select {
				default:
				case <-s.exit:
					return
				}
				result := s.gofmt(files)
				s.Report.Files = append(s.Report.Files, result...)

			}(files)

			//abort the foreach if exit
			select {
			case <-s.exit:
				break
			default:
			}
		}
		go func() {
			// wait for all go fmt done
			wg.Wait()
			s.Cost = float32(time.Since(s.Created).Seconds())
			close(s.completed)
		}()
	}()
	s.running = true
	return nil
}

func (s *Service) Stop() error {
	if !s.running {
		return nil
	}
	s.Lock()
	defer s.Unlock()
	close(s.exit)
	s.running = false
	return nil
}

func (s *Service) Wait() error {
	if !s.running {
		return nil
	}
	for {
		select {
		case <-s.exit:
			return nil
		case <-s.completed:
			return nil
		case <-time.After(1 * time.Second):
		}
	}
	s.running = false
	return nil
}
func (s *Service) gofmt(files []string) []*File {
	result, err := runGoFmt(files...)
	if err != nil {
		s.error(err.Error())
	}
	return result
}

var (
	// Match diff print info
	//
	// diff testdata/needFmt.go gofmt/./testdata/needFmt.go
	//
	// diff -u ./testdata/needFmt.go.orig  ./testdata/needFmt.go
	regDiffHead = regexp.MustCompile(`^diff(?: -u){0,1} \S+\s(?:gofmt\/){0,1}(\S+)$`)
)

func runGoFmt(files ...string) ([]*File, error) {

	var (
		stderr bytes.Buffer
		stdout bytes.Buffer
	)
	cmd := exec.Command(gofmtpath, "-d", "-e", "-s")
	cmd.Args = append(cmd.Args, files...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		s := stderr.String()
		if s != "" {
			return nil, errors.New(stderr.String() + "\n" + err.Error())
		}
		return nil, err
	}
	var result []*File
	for _, f := range files {
		result = append(result, &File{
			Name: f,
		})
	}
	// read output add add to diffrent file
	var file *File
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line != "" {
			log.Println(line)
		}
		if matches := regDiffHead.FindSubmatch([]byte(line)); len(matches) == 2 {
			// find file
			name := string(matches[1])
			for _, f := range result {
				if f.Name == name {
					file = f
					file.NeedFmt = true
					break
				}

			}
			continue
		}
		// add content as diff body
		if file != nil {
			file.Diff += line + "\n"
		}
	}
	return result, nil
}
