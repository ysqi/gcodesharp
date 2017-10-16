package reporter

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type HelloService struct{}

func (h *HelloService) Run() error {
	return nil
}
func (h *HelloService) Stop() error {
	return nil
}
func (h *HelloService) Wait() error {
	return nil
}

func TestReporter_Default(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	r.Register(func(ctx *ServiceContext) (Service, error) {
		return &HelloService{}, nil
	})
	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = r.Stop()
	if err != nil {
		t.Fatal(err)
	}
	r.Wait()
}

func TestReporter_DuplicateReg(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	r.Register(func(ctx *ServiceContext) (Service, error) { return &HelloService{}, nil })
	r.Register(func(ctx *ServiceContext) (Service, error) { return &HelloService{}, nil })
	err = r.Start()
	if err == nil {
		t.Fatal("want get duplicate error")
	} else if _, ok := err.(DuplicateError); !ok {
		t.Fatalf("want get duplicate error,but got %s", err)
	}
}

func TestReporter_ErrorStart(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Start(); err == nil {
		t.Fatal("want get start failed error")
	} else if _, ok := err.(ReportActionError); !ok {
		t.Fatalf("want error type is report action error,but got %+v", err)
	}
}

type TestService struct {
	ID       string
	runHook  func() error
	stopHook func() error
	waitHook func() error
}

func (h *TestService) Run() error {
	return h.runHook()
}
func (h *TestService) Stop() error {
	return h.stopHook()
}
func (h *TestService) Wait() error {
	return h.waitHook()
}

type TestServerA struct {
	*TestService
}
type TestServerB struct {
	*TestService
}
type TestServerC struct {
	*TestService
}

type hook func() error

func loadHook(ins Service, run, stop, wait hook) {
	switch v := ins.(type) {
	case *TestServerA:
		v.TestService = &TestService{
			stopHook: stop,
			runHook:  run,
			waitHook: wait,
		}
	case *TestServerB:
		v.TestService = &TestService{
			stopHook: stop,
			runHook:  run,
			waitHook: wait,
		}
	case *TestServerC:
		v.TestService = &TestService{
			stopHook: stop,
			runHook:  run,
			waitHook: wait,
		}
	default:
		fmt.Println("unkonwn!!!")
	}
}
func TestReporter_ServiceStart(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	services := []Service{&TestServerA{}, &TestServerB{}, &TestServerC{}}
	started := make(map[Service]bool)
	stopped := make(map[Service]bool)

	for _, s := range services {
		ser := s
		r.Register(func(ctx *ServiceContext) (Service, error) {
			loadHook(ser,
				func() error {
					started[ser] = true
					return nil
				},
				func() error {
					stopped[ser] = true
					return nil
				},
				func() error {
					return nil
				},
			)
			return ser, nil
		})
	}
	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}
	r.Wait()

	for _, s := range services {
		if !started[s] {
			t.Fatal(reflect.TypeOf(s).String(), "service is not running")
		}
	}
}

func TestReporter_ServiceErrorStart(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	services := []Service{&TestServerA{}, &TestServerB{}, &TestServerC{}}
	started := make(map[Service]bool)
	stopped := make(map[Service]bool)

	for i, s := range services {
		ser := s
		i := i
		r.Register(func(ctx *ServiceContext) (Service, error) {
			loadHook(ser,
				func() error {
					if i == len(services)-1 {
						return fmt.Errorf("error %d", i)
					}
					started[ser] = true
					return nil
				},
				func() error {
					stopped[ser] = true
					return nil
				},
				func() error {
					return nil
				},
			)
			return ser, nil
		})
	}
	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}
	r.Wait()
	for _, s := range services {
		if !stopped[s] {
			t.Fatal(reflect.TypeOf(s).String(), "service is not running")
		}
	}
}

func TestReporter_CatchServicePanicErrorStart(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	services := []Service{&TestServerA{}, &TestServerB{}, &TestServerC{}}
	started := MyMap{}
	stopped := MyMap{}

	for i, s := range services {
		ser := s
		i := i
		r.Register(func(ctx *ServiceContext) (Service, error) {
			loadHook(ser,
				func() error {
					if i == len(services)-1 {
						panic("ha ha")
					}
					started.Set(ser)
					return nil
				},
				func() error {
					stopped.Set(ser)
					return nil
				},
				func() error {
					return nil
				},
			)
			return ser, nil
		})
	}
	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}
	r.Wait()
	for _, s := range services {
		if !stopped[s] {
			t.Fatal(reflect.TypeOf(s).String(), "service is not running")
		}
	}
}

var mapLock sync.Mutex

type MyMap map[Service]bool

func (m MyMap) Set(s Service) {
	mapLock.Lock()
	m[s] = true
	mapLock.Unlock()
}

func TestReporter_ServiceStop(t *testing.T) {
	ctx := ServiceContext{}
	r, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	services := []Service{&TestServerA{}, &TestServerB{}, &TestServerC{}}
	started := MyMap{}
	stopped := MyMap{}

	for i, s := range services {
		ser := s
		i := i
		r.Register(func(ctx *ServiceContext) (Service, error) {
			loadHook(ser,
				func() error {
					started.Set(ser)
					return nil
				},
				func() error {
					stopped.Set(ser)
					return nil
				},
				func() error {
					if i == len(services)-1 {
						time.Sleep(10 * time.Minute)
					}
					return nil
				},
			)
			return ser, nil
		})
	}
	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}

	err = r.Stop()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range services {
		if !stopped[s] {
			t.Fatal(reflect.TypeOf(s).String(), "service is not running")
		}
	}
}
