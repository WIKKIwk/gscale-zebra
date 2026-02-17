package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func runTUI(ctx context.Context, updates <-chan Reading, sourceLine string, serialErr error) error {
	stdinFD := int(os.Stdin.Fd())
	isTTY := term.IsTerminal(stdinFD)
	var restore func() error
	if isTTY {
		oldState, err := term.MakeRaw(stdinFD)
		if err == nil {
			restore = func() error { return term.Restore(stdinFD, oldState) }
		}
	}

	if restore != nil {
		defer func() {
			_ = restore()
			fmt.Print("\n")
		}()
	}

	quit := make(chan struct{}, 1)
	if isTTY {
		go readKeys(quit)
	}

	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	var last Reading
	last.Unit = "kg"
	message := "scale oqimi kutilmoqda"
	if serialErr != nil {
		message = serialErr.Error()
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-quit:
			return nil
		case upd := <-updates:
			if upd.Unit == "" && last.Unit != "" {
				upd.Unit = last.Unit
			}
			last = upd
			if upd.Error != "" {
				message = upd.Error
			} else {
				message = "ok"
			}
		case <-ticker.C:
			render(sourceLine, message, last)
		}
	}
}

func readKeys(quit chan<- struct{}) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 3 {
			select {
			case quit <- struct{}{}:
			default:
			}
			return
		}
	}
}

func render(source, message string, r Reading) {
	fmt.Print("\033[2J\033[H")
	fmt.Println("Scale Monitor (Go)")
	fmt.Println("Q tugmasi bilan chiqish")
	fmt.Println()
	fmt.Printf("Source : %s\n", source)
	if r.Port != "" {
		fmt.Printf("Port   : %s\n", r.Port)
	}
	if r.Baud > 0 {
		fmt.Printf("Baud   : %d\n", r.Baud)
	}
	if r.Weight != nil {
		fmt.Printf("QTY    : %.3f %s\n", *r.Weight, strings.TrimSpace(r.Unit))
	} else {
		fmt.Printf("QTY    : -- %s\n", strings.TrimSpace(r.Unit))
	}
	fmt.Printf("Stable : %s\n", stableText(r.Stable))
	if !r.UpdatedAt.IsZero() {
		fmt.Printf("Update : %s\n", r.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Status : %s\n", message)
	if strings.TrimSpace(r.Raw) != "" {
		fmt.Printf("Raw    : %s\n", sanitizeInline(r.Raw, 90))
	}
}
