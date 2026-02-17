package main

import (
	"flag"
	"fmt"
	"time"
)

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
