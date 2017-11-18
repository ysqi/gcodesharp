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
