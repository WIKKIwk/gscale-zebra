package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		exitErr(errors.New("buyruq berilmadi"))
	}

	cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
	args := os.Args[2:]

	switch cmd {
	case "list":
		if err := runList(); err != nil {
			exitErr(err)
		}
	case "status":
		if err := runStatus(args); err != nil {
			exitErr(err)
		}
	case "settings", "config":
		if err := runSettings(args); err != nil {
			exitErr(err)
		}
	case "print-test":
		if err := runPrintTest(args); err != nil {
			exitErr(err)
		}
	case "epc-test", "rfid-test", "encode":
		if err := runEPCTest(args); err != nil {
			exitErr(err)
		}
	case "calibrate", "auto-calibrate":
		if err := runCalibrate(args); err != nil {
			exitErr(err)
		}
	case "self-check":
		if err := runSelfCheck(args); err != nil {
			exitErr(err)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		printUsage()
		exitErr(fmt.Errorf("noma'lum buyruq: %s", cmd))
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
