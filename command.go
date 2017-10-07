package main

import (
	"log"

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

func init() {
	cmd.Flags().StringP("outputdir", "o", "", `Place output files from profiling in the specified directory,
by default the directory of application working.`)

}

func run(c *cobra.Command, args []string) {
	log.Println("args:", len(args), args)
	cfg := gtest.Config{}
	cfg.PackageDirs = args
	report, err := gtest.Run(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(len(report.Pkgs))
	for _, p := range report.Pkgs {
		pass, skip, fail := p.PassCount(), p.SkipCount(), p.FailCount()
		log.Println(len(p.Units), pass, skip, fail)
		log.Printf("+%v", p)
	}
}
