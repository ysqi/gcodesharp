package reporter

import (
	"go/build"
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
	GlobalCxt context.Context
	// Packages is list of need handle package
	Packages []*build.Package
	// sub package path will be processed if set true
	ContainSubDir bool
	// list of program run arg
	Flagset *pflag.FlagSet
}

// An service constructor
type ServiceConstructor func(ctx *ServiceContext) (Service, error)
