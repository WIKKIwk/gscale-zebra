package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type repeatedFlag []string

func (r *repeatedFlag) String() string {
	if r == nil {
		return ""
	}
	return strings.Join(*r, ",")
}

func (r *repeatedFlag) Set(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	*r = append(*r, v)
	return nil
}

func runSettings(args []string) error {
	fs := flag.NewFlagSet("settings", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	timeout := fs.Duration("timeout", 1200*time.Millisecond, "SGD query timeout")
	retries := fs.Int("retries", 3, "retry count per key")
	delay := fs.Duration("retry-delay", 120*time.Millisecond, "retry delay")

	var keys repeatedFlag
	fs.Var(&keys, "key", "SGD key (repeatable). Example: --key print.width")

	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		keys = defaultSettingsKeys()
	}

	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Println("Settings:")
	for _, key := range keys {
		value, qerr := queryVarWithRetries(p.DevicePath, key, *timeout, *retries, *delay)
		if qerr != nil {
			fmt.Printf("- %s = (xato: %v)\n", key, qerr)
			continue
		}
		fmt.Printf("- %s = %s\n", key, safeStr(value, "-"))
	}
	return nil
}

func queryVarWithRetries(device, key string, timeout time.Duration, retries int, delay time.Duration) (string, error) {
	if retries < 1 {
		retries = 1
	}

	var lastErr error
	for i := 0; i < retries; i++ {
		v, err := QuerySGDVar(device, key, timeout)
		if err == nil {
			return strings.TrimSpace(strings.Trim(v, "\"")), nil
		}
		lastErr = err
		time.Sleep(delay)
	}

	return "", lastErr
}

func defaultSettingsKeys() []string {
	return []string{
		"device.product_name",
		"device.friendly_name",
		"device.unique_id",
		"apl.name",
		"apl.version",
		"print.width",
		"label.length",
		"media.type",
		"media.sense_mode",
		"media.printmode",
		"print.speed",
		"print.tone",
		"zpl.label_length",
		"zpl.print_width",
		"rfid.enable",
		"rfid.tag.read.content",
	}
}
