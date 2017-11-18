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
