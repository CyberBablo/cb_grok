package main

import (
	"cb_grok/cmd/gen/candle"
	"cb_grok/cmd/gen/metrics"
	"cb_grok/cmd/gen/ohlc"
	"flag"
)

func main() {
	var cmd string

	flag.StringVar(&cmd, "cmd", "", "CMD name: candle, metrics, ohlc")
	flag.Parse()

	switch cmd {
	case "candle":
		candle.CMD()
	case "metrics":
		metrics.CMD()
	case "ohlc":
		ohlc.CMD()
	}
}
