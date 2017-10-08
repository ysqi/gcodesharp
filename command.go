package main

import (
	"log"
	"os"

	"github.com/ysqi/gcodesharp/gfmt"
	"github.com/ysqi/gcodesharp/gtest"

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
	junitpath      string // enable save report to xml file
	onlyCurrentDir bool
)

func init() {
	cmd.PersistentFlags().StringVarP(&junitpath, "junit", "j", "", `save go test report to junit xml file`)
	cmd.PersistentFlags().BoolVarP(&onlyCurrentDir, "onlyself", "", false, `only run go test for current dir not contains child dir`)
}

func run(c *cobra.Command, args []string) {
	cfg := gtest.Config{}
	cfg.PackagePaths = args
	cfg.ContainImport = !onlyCurrentDir
	report, err := gtest.Run(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	gfmtTest, err := gfmtAsTestPart(args)
	if err != nil {
		log.Fatal(err)
	}
	report.Packages = append(report.Packages, gfmtTest)

	if junitpath != "" {
		err := saveTestReport(report)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
}

func gfmtAsTestPart(args []string) (*gtest.Package, error) {
	cfg := gfmt.Config{
		PackagePaths:  args,
		ContainImport: !onlyCurrentDir,
		PrintLog:      true,
	}
	report, err := gfmt.Run(ctx, &cfg)
	if err != nil {
		return nil, err
	}
	//go fmt as a part of the test result
	pkg := gtest.Package{
		Name:     report.GoFmt,
		Coverage: -1.0,
		Runtime:  report.Created,
		Cost:     report.Cost,
	}
	for _, f := range report.Files {
		u := gtest.Unit{Name: f.Name}
		if f.NeedFmt {
			u.Result = gtest.FAIL
			u.Output = f.Diff
		}
		pkg.Units = append(pkg.Units, &u)
	}
	return &pkg, nil
}

func saveTestReport(report *gtest.Report) error {
	f, err := os.Create(junitpath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = gtest.JUnitReportXML(report, false, f)
	// remove the file when error.
	if err != nil {
		f.Close()
		os.Remove(junitpath)
		return err
	}
	return nil
}
