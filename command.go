package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ysqi/gcodesharp/context"
	"github.com/ysqi/gcodesharp/gfmt"
	"github.com/ysqi/gcodesharp/gtest"
	"github.com/ysqi/gcodesharp/reporter"

	"github.com/spf13/cobra"
)

// cmd represents the base command when called without any subcommands
var cmd = &cobra.Command{
	Use:   "gcodesharp",
	Short: "A mini sharp tool for go code review",
	Long: `GCodeSharp is a CLI library for Go language code review applications.
This application is a tool to generate the report to quickly review the golang code.`,
	Run: run,
}

var (
	junitpath string // enable save report to xml file
)

func init() {
	cmd.PersistentFlags().StringVarP(&junitpath, "junit", "j", "", `save report as junit xml file`)
}

func run(c *cobra.Command, args []string) {
	sCtx := initCtx(c, args...)

	rp, err := reporter.New(sCtx)
	if err != nil {
		Failf(err.Error())
	}
	regGoFormatServer(rp)
	regGoTestService(rp)

	err = rp.Start()
	if err != nil {
		Failf("start reporter:%s", err.Error())
	}
	rp.Wait()

	if junitpath != "" {
		err := saveTestReport(rp)
		if err != nil {
			Failf("create and save junit:%s", err.Error())
		}
	}
}

func Errorf(fmt_ string, args ...interface{}) {
	if !strings.HasSuffix(fmt_, "\n") {
		fmt_ += "\n"
	}
	fmt.Printf(fmt_, args...)
}
func Failf(fmt_ string, args ...interface{}) {
	Errorf(fmt_, args)
	os.Exit(2)
}

func initCtx(c *cobra.Command, packages ...string) *reporter.ServiceContext {
	added := map[string]struct{}{}
	appendPkg := func(path string) {
		p, err := build.Import(path, "", build.IgnoreVendor)
		if err != nil {
			Failf("initCtx:%s", err)
		}
		// repeat clear
		if _, ok := added[p.Dir]; !ok {
			ctx.Packages = append(ctx.Packages, p)
			added[p.Dir] = struct{}{}
		}
	}
	for _, p := range packages {
		// find all package in dir by go list command.
		list, err := context.GetPackagePaths(p)
		if err != nil {
			Failf("initCtx:%s", err)
		}
		for _, p := range list {
			appendPkg(p)
		}

	}

	return &reporter.ServiceContext{
		GlobalCxt: ctx,
		Flagset:   c.Flags(),
		ErrH:      Errorf,
	}
}

func regGoFormatServer(rep *reporter.Reporter) {
	rep.Register(func(ctx *reporter.ServiceContext) (reporter.Service, error) {
		return gfmt.New(ctx.GlobalCxt, ctx.ErrH)
	})
}

func regGoTestService(rep *reporter.Reporter) {
	rep.Register(func(ctx *reporter.ServiceContext) (reporter.Service, error) {
		return gtest.New(ctx.GlobalCxt, ctx.ErrH)
	})
}

func saveTestReport(report *reporter.Reporter) (err error) {
	var f *os.File
	f, err = ioutil.TempFile("", "gcodesharp_junit")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && f != nil {
			os.Remove(f.Name())
		}
	}()
	defer f.Close()

	err = report.OutputJunit(false, f)
	if err == nil {
		err = os.Rename(f.Name(), junitpath)
	}
	return
}
