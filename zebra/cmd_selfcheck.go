package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func runSelfCheck(args []string) error {
	fs := flag.NewFlagSet("self-check", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	printOne := fs.Bool("print", false, "print one minimal test label")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	resp, err := QueryHostStatus(p.DevicePath, 1100*time.Millisecond)
	if err != nil {
		fmt.Printf("Status query: no response (%v)\n", err)
	} else {
		fmt.Printf("Status query: ok (%d bytes)\n", len(strings.TrimSpace(resp)))
	}

	if *printOne {
		stream := BuildPrintTestCommandStream("SELF CHECK", 1)
		if err := SendRaw(p.DevicePath, []byte(stream)); err != nil {
			return fmt.Errorf("self-check print xato: %w", err)
		}
		fmt.Println("Self-check print yuborildi (1 label).")
	}

	fmt.Println("Self-check tugadi.")
	return nil
}
