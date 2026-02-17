package main

import (
	"fmt"
	"os"
)

func push(ch chan<- Reading, r Reading) {
	select {
	case ch <- r:
	default:
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
