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
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ysqi/gcodesharp/context"
)

type errHander func(fm string, args ...interface{})

// Report gofmt result
type Report struct {
	Files    []*File
	ExecPath string
	Created  time.Time
	Cost     float32
	Env      struct {
		GoVersion string
		OS        string
		Arch      string
	}
	SysErr error
}

type Problem struct {
	Line int
	Cell int
	Info string
}

// File need format go file
type File struct {
	// Name file name
	Name string
	// Diff instead of rewriting file
	Problem []Problem
}

func (f *File) HasProblem() bool {
	return len(f.Problem) > 0
}

func (f *File) ProblemContent() string {
	if !f.HasProblem() {
		return ""
	}
	str := bytes.NewBufferString("")
	for _, p := range f.Problem {
		str.WriteString(fmt.Sprintf("line:%d:%d ", p.Line, p.Cell))
		str.WriteString(p.Info)
		str.WriteString("\n")
	}
	return str.String()
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
	s.errh("glint: %s", msg)
	s.Stop()
}

// Run go lint
func (s *Service) Run() error {
	if s.running {
		return errors.New("glint is running")
	}
	s.Lock()
	defer s.Unlock()

	s.Created = time.Now()
	s.Env.GoVersion = runtime.Version()
	s.Env.OS = runtime.GOOS
	s.Env.Arch = runtime.GOARCH
	s.ExecPath = "golint"
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
			// batch golint
			go func(files []string) {
				defer wg.Done()
				select {
				default:
				case <-s.exit:
					return
				}
				result := s.golint(files)
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
			// wait for all go lint done
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
func (s *Service) golint(files []string) []*File {
	result, err := runGolint(files...)
	if err != nil {
		s.error(err.Error())
	}
	return result
}

var (
	// Match problem print info
	regLine = regexp.MustCompile(`^(.+\.go):(\d+):(\d+):(.*)$`)
)

func runGolint(files ...string) ([]*File, error) {

	var (
		stderr bytes.Buffer
		stdout bytes.Buffer
	)
	cmd := exec.Command("golint")
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
	// read output add add to file
	var file *File
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line != "" {
			log.Println(line)
		}
		if matches := regLine.FindSubmatch([]byte(line)); len(matches) == 5 {
			// find file
			name := string(matches[1])
			for _, f := range result {
				if f.Name == name {
					file = f
					break
				}
			}
			if file != nil {
				p := Problem{
					Line: mustInt(matches[2]),
					Cell: mustInt(matches[3]),
					Info: string(matches[4]),
				}
				file.Problem = append(file.Problem, p)
			}
		}
	}
	return result, nil
}

// mustInt convert bytes to float number.
// os exist with error if parse failed.
func mustInt(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	val, err := strconv.ParseInt(string(b), 10, 32)
	if err != nil {
		panic("mustInt:" + err.Error())
	}
	return int(val)
}
