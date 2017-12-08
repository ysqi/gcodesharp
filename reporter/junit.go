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
	"encoding/xml"
	"fmt"
	"io"
	"reflect"

	"github.com/ysqi/gcodesharp/reporter/formater"
)

type JunitFormater interface {
	ToJunit() (formater.JUnitTestSuites, error)
}

func (r *Reporter) OutputJunit(noXMLHeader bool, w io.Writer) error {
	if r.running {
		return ErrIsRunning
	}
	// find support junit service
	suites := formater.JUnitTestSuites{}
	for _, s := range r.services[false] {
		js, ok := s.(JunitFormater)
		if !ok {
			fmt.Printf("SKIP: service %s is not JunitFormater", reflect.TypeOf(s).String())
			continue
		}
		s, err := js.ToJunit()
		if err != nil {
			return err
		}
		suites.Suites = append(suites.Suites, s.Suites...)
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if !noXMLHeader {
		w.Write([]byte(xml.Header))
	}
	if err := enc.Encode(suites); err != nil {
		return err
	}
	w.Write([]byte("\n"))
	return enc.Flush()
}
