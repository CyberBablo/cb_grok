// Package main provides a unified CLI for unit testing commands.
// Available commands: test, place_order
package main

import (
	"flag"
	"fmt"
	"os"

	"cb_grok/cmd/unitest/place_order"
	"cb_grok/cmd/unitest/test"
)

const (
	usageText = `Unit Test CLI

Usage:
  unitest -cmd=<command> [options]

Available commands:
  test         Test Bybit exchange functionality and wallet balance
  place_order  Test placing orders on Bybit exchange

Examples:
  unitest -cmd=test
  unitest -cmd=place_order

For command-specific options, run:
  unitest -cmd=<command> -help
`
)

func main() {
	var (
		cmd  = flag.String("cmd", "", "Command to run: test, place_order")
		help = flag.Bool("help", false, "Show help information")
	)

	flag.Usage = func() {
		fmt.Print(usageText)
	}

	flag.Parse()

	if *help || *cmd == "" {
		flag.Usage()
		return
	}

	switch *cmd {
	case "test":
		test.CMD()
	case "place_order":
		place_order.CMD()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", *cmd)
		flag.Usage()
		os.Exit(1)
	}
}
