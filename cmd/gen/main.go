// Package main provides a unified CLI for various data generation commands.
// Available commands: candle, metrics, ohlc
package main

import (
	"flag"
	"fmt"
	"os"

	"cb_grok/cmd/gen/candle"
	"cb_grok/cmd/gen/metrics"
	"cb_grok/cmd/gen/ohlc"
)

const (
	usageText = `Data Generation CLI

Usage:
  gen -cmd=<command> [options]

Available commands:
  candle   Generate and manage candle data
  metrics  Generate trading metrics data
  ohlc     Generate OHLC charts and visualizations

Examples:
  gen -cmd=candle -op=generate -symbol=BTCUSDT -count=100
  gen -cmd=metrics -symbol=ETHUSDT -days=30 -trades=200
  gen -cmd=ohlc

For command-specific options, run:
  gen -cmd=<command> -help
`
)

func main() {
	var (
		cmd  = flag.String("cmd", "", "Command to run: candle, metrics, ohlc")
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
	case "candle":
		candle.CMD()
	case "metrics":
		metrics.CMD()
	case "ohlc":
		ohlc.CMD()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", *cmd)
		flag.Usage()
		os.Exit(1)
	}
}
