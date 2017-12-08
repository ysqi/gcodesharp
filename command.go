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

package main

import (
	"go/build"
	"log"
	"os"

	"github.com/ysqi/gcodesharp/context"
	"github.com/ysqi/gcodesharp/gfmt"
	"github.com/ysqi/gcodesharp/glint"
	"github.com/ysqi/gcodesharp/gtest"
	"github.com/ysqi/gcodesharp/reporter"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gcodesharp",
	Short: "A mini sharp tool for go code review",
	Long: `GCodeSharp is a CLI library for Go language code review applications.
This application is a tool to generate the report to quickly review the golang code.`,
	Run: run,
}

var (
	junitpath string // enable save report to xml file

	selectTool  []string
	defaultTool = []string{"gtest", "gfmt", "glint"}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&junitpath, "junit", "j", "", `save report as junit xml file`)
	rootCmd.PersistentFlags().StringArrayVarP(&selectTool, "tool", "t", defaultTool, `specify which tool to exec`)
}

func include(array []string, s string) bool {
	for _, v := range array {
		if v == s {
			return true
		}
	}
	return false
}

func run(c *cobra.Command, args []string) {
	sCtx := initCtx(c, args...)
	rp, err := reporter.New(sCtx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if include(selectTool, "gfmt") {
		regGoFormatService(rp)
	}
	if include(selectTool, "glint") {
		regGolintService(rp)
	}
	if include(selectTool, "gtest") {
		regGoTestService(rp)
	}
	if rp.RegisterNumber() == 0 {
		log.Fatalf("does not contain a valid tool, stop running. all tool: %s", defaultTool)
	}
	err = rp.Start()
	if err != nil {
		log.Fatalf("start reporter:%s", err.Error())
	}
	rp.Wait()

	err = saveTestReport(rp)
	if err != nil {
		log.Fatalf("create and save junit:%s", err.Error())
	}
}

func initCtx(c *cobra.Command, packages ...string) *reporter.ServiceContext {
	if len(packages) == 0 {
		log.Println("[WARN] No package path is set and will use current dir as package path")
		packages = append(packages, ".")
	}
	added := map[string]struct{}{}
	appendPkg := func(path string) {
		p, err := build.Import(path, "", build.IgnoreVendor)
		if err != nil {
			log.Fatalf("initCtx:%s", err)
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
			log.Fatalf("initCtx:%s", err)
		}
		for _, p := range list {
			appendPkg(p)
		}
	}

	return &reporter.ServiceContext{
		GlobalCxt: ctx,
		Flagset:   c.Flags(),
		ErrH:      log.Fatalf,
	}
}

func regGolintService(rep *reporter.Reporter) {
	rep.Register(func(ctx *reporter.ServiceContext) (reporter.Service, error) {
		return glint.New(ctx.GlobalCxt, ctx.ErrH)
	})
}
func regGoFormatService(rep *reporter.Reporter) {
	rep.Register(func(ctx *reporter.ServiceContext) (reporter.Service, error) {
		return gfmt.New(ctx.GlobalCxt, ctx.ErrH)
	})
}

func regGoTestService(rep *reporter.Reporter) {
	rep.Register(func(ctx *reporter.ServiceContext) (reporter.Service, error) {
		return gtest.New(ctx.GlobalCxt, ctx.ErrH)
	})
}

func saveTestReport(report *reporter.Reporter) error {
	if junitpath == "" {
		log.Println("[WARN] No junit path is set and report will not be exported. please see -help.")
		return nil
	}

	f, err := os.Create(junitpath)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()
	return report.OutputJunit(false, f)
}
