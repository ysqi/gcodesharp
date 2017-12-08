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

package reporter

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
)

// Reporter is a report service manager.
// responsible for the operation of the register service.
type Reporter struct {
	context *ServiceContext

	serviceFuncs []ServiceConstructor
	services     map[bool][]Service // key=true,value=running service,key=false,value=all service

	running        bool
	serviceProcess sync.WaitGroup

	sync.RWMutex
}

// New return a new report manager
func New(ctx *ServiceContext) (*Reporter, error) {
	r := Reporter{
		context: ctx,
	}
	return &r, nil
}

func (r *Reporter) RegisterNumber() int {
	return len(r.serviceFuncs)
}

// Register is add a server constructor to reporter
func (r *Reporter) Register(constructor ServiceConstructor) error {
	if constructor == nil {
		return errors.New("constructor cannot be set nil")
	}
	r.Lock()
	defer r.Unlock()
	if r.running {
		return errors.New("cannot register server to running reporter")
	}
	r.serviceFuncs = append(r.serviceFuncs, constructor)
	return nil
}

// Start reporter and service.
// disable start the report without service or is running,otherwise return error.
func (r *Reporter) Start() error {
	r.Lock()
	defer r.Unlock()
	if r.running {
		return ErrRepeatedStart
	}
	if len(r.serviceFuncs) == 0 {
		return ReportActionError{action: "start report", err: errors.New("cannot start for empty service list")}
	}
	r.services = make(map[bool][]Service, 2)

	// create each of the services
	checkKind := make(map[reflect.Type]struct{}, len(r.serviceFuncs))
	for _, s := range r.serviceFuncs {
		service, err := s(r.context)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := checkKind[kind]; exists {
			return DuplicateError{kind: kind}
		}
		checkKind[kind] = struct{}{}
		r.services[false] = append(r.services[false], service)
	}

	r.running = true

	// run service one by one
	r.serviceProcess = sync.WaitGroup{}
	for _, s := range r.services[false] {
		r.serviceProcess.Add(1)
		r.services[true] = append(r.services[true], s)
		go func(s Service) {
			defer r.serviceProcess.Done()
			if !r.running {
				return
			}
			if err := trycatch(s.Run); err != nil {
				go r.Stop()
				return
			}
			s.Wait()
		}(s)
	}

	return nil
}

// Stop the reporter and each of running service.
// disable stop the not running reporter,otherwise return error.
// return the stop error if one service stop failed.
// Note: this stop action will wait for all service stop done.
func (r *Reporter) Stop() error {
	r.Lock()
	defer r.Unlock()
	if !r.running {
		return ErrNotRunning
	}
	r.running = false
	runningError := ReportActionError{action: "stop report"}
	// stop the each of running service
	for _, s := range r.services[true] {
		if err := trycatch(s.Stop); err != nil {
			runningError.services[reflect.TypeOf(s)] = err
		}
	}

	r.serviceProcess.Wait()
	//r.services = nil

	if len(runningError.services) > 0 {
		return runningError
	}

	return nil
}

func (r *Reporter) OutputHTML(writer io.Writer) error {
	r.Lock()
	defer r.Unlock()
	if r.running {
		return ErrIsRunning
	}

	return nil
}

// Wait blocks the thread until the each of services is stopped.
func (r *Reporter) Wait() {
	r.Lock()
	if !r.running {
		r.Unlock()
		return
	}
	r.Unlock()
	r.serviceProcess.Wait()
	r.running = false
}

// trycatch try catch panic error
func trycatch(do func() error) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			switch v := err2.(type) {
			case error:
				err = v
			case string:
				err = fmt.Errorf("panic error: %s", v)
			default:
				err = fmt.Errorf("panice error: %v", v)
			}
		}
	}()
	err = do()
	return
}
