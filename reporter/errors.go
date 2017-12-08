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
	"fmt"
	"reflect"

	"errors"
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
