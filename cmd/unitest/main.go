package main

import (
	"cb_grok/cmd/unitest/place_order"
	"cb_grok/cmd/unitest/test"
	"flag"
)

func main() {
	var cmd string

	flag.StringVar(&cmd, "cmd", "", "CMD name: test, place_order")
	flag.Parse()

	switch cmd {
	case "test":
		test.CMD()
	case "place_order":
		place_order.CMD()
	}
}
