package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	timeout := fs.Duration("timeout", 1200*time.Millisecond, "status read timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	resp, err := QueryHostStatus(p.DevicePath, *timeout)
	if err != nil {
		fmt.Printf("Host status: query yuborildi, lekin javob olinmadi (%v)\n", err)
		return nil
	}

	preview := strings.TrimSpace(resp)
	if preview == "" {
		fmt.Println("Host status: bo'sh javob")
		return nil
	}
	if len(preview) > 300 {
		preview = preview[:300] + "..."
	}
	fmt.Printf("Host status response:\n%s\n", preview)
	return nil
}
