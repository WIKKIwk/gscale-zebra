package main

import (
	"flag"
	"fmt"
	"strings"
)

func runSetVar(args []string) error {
	fs := flag.NewFlagSet("setvar", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	key := fs.String("key", "", "SGD key (example: ezpl.print_width)")
	value := fs.String("value", "", "SGD value")
	save := fs.Bool("save", true, "save settings (^JUS)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	k := strings.TrimSpace(*key)
	v := strings.TrimSpace(*value)
	if k == "" {
		return fmt.Errorf("--key bo'sh")
	}
	if v == "" {
		return fmt.Errorf("--value bo'sh")
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("! U1 setvar \"%s\" \"%s\"\r\n", k, strings.ReplaceAll(v, "\"", ""))
	if err := SendRaw(p.DevicePath, []byte(cmd)); err != nil {
		return err
	}
	if *save {
		if err := SendRaw(p.DevicePath, []byte("^XA^JUS^XZ\n")); err != nil {
			return err
		}
	}

	readBack, readErr := QuerySGDVar(p.DevicePath, k, 1400)
	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Set: %s=%s (save=%v)\n", k, v, *save)
	if readErr != nil {
		fmt.Printf("Read-back: xato (%v)\n", readErr)
		return nil
	}
	fmt.Printf("Read-back: %s=%s\n", k, strings.Trim(strings.TrimSpace(readBack), "\""))
	return nil
}
