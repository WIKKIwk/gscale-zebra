package main

import (
	"flag"
	"fmt"
)

const maxPrintCopies = 20

func runPrintTest(args []string) error {
	fs := flag.NewFlagSet("print-test", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	message := fs.String("message", "GSCALE ZEBRA TEST", "line text to print")
	copies := fs.Int("copies", 1, "label copies")
	dryRun := fs.Bool("dry-run", false, "show ZPL only, do not send")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *copies < 1 {
		*copies = 1
	}
	if *copies > maxPrintCopies {
		return fmt.Errorf("copies %d dan ko'p bo'lmasin", maxPrintCopies)
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	stream := BuildPrintTestCommandStream(*message, *copies)
	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : print-test (copies=%d, dry-run=%v)\n", *copies, *dryRun)

	if *dryRun {
		fmt.Println("--- ZPL preview ---")
		fmt.Println(stream)
		return nil
	}

	if err := SendRaw(p.DevicePath, []byte(stream)); err != nil {
		return err
	}
	fmt.Println("Test label yuborildi.")
	return nil
}
