package reporter

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/pkg/errors"
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
func New(ctx ServiceContext) (*Reporter, error) {
	r := Reporter{
		context: &ctx,
	}
	return &r, nil
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
	r.serviceProcess.Add(len(r.services[false]))
	fmt.Println(len(r.services[false]))
	for _, s := range r.services[false] {
		r.services[true] = append(r.services[true], s)
		go func(s Service) {
			if !r.running {
				return
			}
			if err := trycatch(s.Run); err != nil {
				r.Stop()
				return
			}
			s.Wait()
			if r.running {
				r.serviceProcess.Done()
			}
		}(s)
	}

	return nil
}

// Stop the reporter and each of running service.
// disable stop the not running reporter,otherwise return error.
// return the stop error if one service stop failed.
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
		r.serviceProcess.Done()
	}

	r.services = nil

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
