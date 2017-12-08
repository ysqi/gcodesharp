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
	"io"

	"github.com/ysqi/gcodesharp/context"

	"github.com/spf13/pflag"
)

// HTMLGennerate a html gennerate inferface.
// reporter service need implement if can provide html report
type HTMLGennerate interface {
	HTitle() string
	HSummary() string
	HGroupDetail() []string
}

// TextPlainGenerate a text plain generate interface.
// reporter service need implement to print text plain report
type TextPlainGenerate interface {
	TOutput(writer io.Writer) error
}

// Service a report service interface
type Service interface {
	Run() error
	Stop() error
	Wait() error
}

// ServiceContext is a context for report work
type ServiceContext struct {
	// Global context contains os info
	GlobalCxt *context.Context

	// list of program run arg
	Flagset *pflag.FlagSet

	ErrH func(fm string, args ...interface{})
}

// An service constructor
type ServiceConstructor func(ctx *ServiceContext) (Service, error)
