package reporter

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

type DuplicateError struct {
	kind reflect.Type
}

func (d DuplicateError) Error() string {
	return fmt.Sprintf("found a duplicate service: %s", d.kind.String())
}

type ReportActionError struct {
	action   string
	services map[reflect.Type]error
	err      error
}

func (r ReportActionError) Error() string {
	str := r.action + " error:"
	if r.err != nil {
		str += r.err.Error() + "."
	}
	for k, err := range r.services {
		str += fmt.Sprintf("service %q:%s,", k.Name(), err.Error())
	}
	return str
}

var (
	ErrNotRunning    = errors.New("cannot do this if the reporter is not running")
	ErrIsRunning     = errors.New("cannot do this if the reporter is running")
	ErrRepeatedStart = errors.New("cannot repeated start the reporter")
)
