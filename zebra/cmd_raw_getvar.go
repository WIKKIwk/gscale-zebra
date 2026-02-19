package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"
	"time"
)

func runRawGetVar(args []string) error {
	fs := flag.NewFlagSet("raw-getvar", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	key := fs.String("key", "", "SGD key")
	timeout := fs.Duration("timeout", 2*time.Second, "transceive timeout")
	count := fs.Int("count", 1, "repeat count")
	if err := fs.Parse(args); err != nil {
		return err
	}

	k := strings.TrimSpace(*key)
	if k == "" {
		return fmt.Errorf("--key bo'sh")
	}
	if *count < 1 {
		*count = 1
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : raw-getvar (key=%s, count=%d)\n", k, *count)

	cmd := fmt.Sprintf("! U1 getvar \"%s\"\r\n", k)
	for i := 1; i <= *count; i++ {
		resp, rerr := transceiveRaw(p.DevicePath, []byte(cmd), *timeout)
		if rerr != nil {
			fmt.Printf("Try %02d: err=%v\n", i, rerr)
			continue
		}
		fmt.Printf("Try %02d RAW HEX: %s\n", i, strings.ToUpper(hex.EncodeToString(resp)))
		fmt.Printf("Try %02d RAW TXT: %q\n", i, string(resp))
		fmt.Printf("Try %02d NORM   : %q\n", i, normalizeStatusResponse(resp))
	}
	return nil
}
