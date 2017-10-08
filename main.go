package main

import (
	"fmt"
	"os"

	"github.com/ysqi/gcodesharp/context"
)

var ctx *context.Context

func init() {
	c, err := context.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx = c
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
