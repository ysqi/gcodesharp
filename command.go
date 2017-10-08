package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/ysqi/gcodesharp/gtest"
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
	cfg.PackageDirs = args
	cfg.ContainImport = !onlyCurrentDir
	report, err := gtest.Run(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	if junitpath != "" {
		err := saveTestReport(report)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
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
