package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
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
	case "print-test":
		if err := runPrintTest(args); err != nil {
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

func runList() error {
	printers, err := FindUSBLPPrinters()
	if err != nil {
		return err
	}
	if len(printers) == 0 {
		fmt.Println("USB printer topilmadi.")
		return nil
	}

	sort.Slice(printers, func(i, j int) bool {
		return printers[i].DevicePath < printers[j].DevicePath
	})

	fmt.Printf("Topilgan printerlar: %d\n", len(printers))
	for i, p := range printers {
		z := "no"
		if p.IsZebra() {
			z = "yes"
		}
		fmt.Printf("%d) %s\n", i+1, p.DevicePath)
		fmt.Printf("   vendor/product: %s:%s\n", p.VendorID, p.ProductID)
		fmt.Printf("   manufacturer : %s\n", safeStr(p.Manufacturer, "-"))
		fmt.Printf("   product      : %s\n", safeStr(p.Product, "-"))
		fmt.Printf("   serial       : %s\n", safeStr(p.Serial, "-"))
		fmt.Printf("   bus/dev      : %s/%s\n", safeStr(p.BusNum, "-"), safeStr(p.DevNum, "-"))
		fmt.Printf("   zebra        : %s\n", z)
	}
	return nil
}

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
	if *copies > 3 {
		return errors.New("copies 3 dan ko'p bo'lmasin (taglarni tejash uchun)")
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	zpl := BuildTestLabelZPL(*message, *copies)
	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : print-test (copies=%d, dry-run=%v)\n", *copies, *dryRun)

	if *dryRun {
		fmt.Println("--- ZPL preview ---")
		fmt.Println(zpl)
		return nil
	}

	if err := SendRaw(p.DevicePath, []byte(zpl)); err != nil {
		return err
	}
	fmt.Println("Test label yuborildi.")
	return nil
}

func runCalibrate(args []string) error {
	fs := flag.NewFlagSet("calibrate", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	dryRun := fs.Bool("dry-run", false, "show commands only, do not send")
	save := fs.Bool("save", true, "save settings after calibration")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	cmds := BuildCalibrationCommands(*save)
	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : calibrate (dry-run=%v, save=%v)\n", *dryRun, *save)
	fmt.Println("Ogohlantirish: calibration bir nechta label/tagni feed qilishi mumkin.")

	if *dryRun {
		fmt.Println("--- Command preview ---")
		for i, c := range cmds {
			fmt.Printf("%d) %q\n", i+1, c)
		}
		return nil
	}

	for i, c := range cmds {
		if err := SendRaw(p.DevicePath, []byte(c)); err != nil {
			return fmt.Errorf("calibration command #%d xato: %w", i+1, err)
		}
		time.Sleep(350 * time.Millisecond)
	}

	fmt.Println("Calibration buyruqlari yuborildi.")
	return nil
}

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
		zpl := BuildTestLabelZPL("SELF CHECK", 1)
		if err := SendRaw(p.DevicePath, []byte(zpl)); err != nil {
			return fmt.Errorf("self-check print xato: %w", err)
		}
		fmt.Println("Self-check print yuborildi (1 label).")
	}

	fmt.Println("Self-check tugadi.")
	return nil
}

func printUsage() {
	fmt.Println("Zebra USB tool (gscale-zebra/zebra)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  zebra list")
	fmt.Println("  zebra status [--device /dev/usb/lp0]")
	fmt.Println("  zebra print-test [--device /dev/usb/lp0] [--message TEXT] [--copies 1] [--dry-run]")
	fmt.Println("  zebra calibrate [--device /dev/usb/lp0] [--dry-run] [--save=true]")
	fmt.Println("  zebra self-check [--device /dev/usb/lp0] [--print]")
}

func safeStr(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
