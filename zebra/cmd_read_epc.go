package main

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var readHexRegex = regexp.MustCompile(`[0-9A-F]{8,64}`)

func runReadEPC(args []string) error {
	fs := flag.NewFlagSet("read-epc", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	expected := fs.String("expected", "", "expected EPC hex (optional)")
	tries := fs.Int("tries", 12, "read attempts")
	interval := fs.Duration("interval", 180*time.Millisecond, "wait between attempts")
	timeout := fs.Duration("timeout", 1500*time.Millisecond, "SGD query timeout")
	readPower := fs.Int("read-power", -1, "temporary RFID read power (0..30), -1 keeps current")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	expect := strings.TrimSpace(strings.ToUpper(*expected))
	if expect != "" {
		norm, nerr := NormalizeEPC(expect)
		if nerr != nil {
			return nerr
		}
		expect = norm
	}

	if *tries < 1 {
		*tries = 1
	}

	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : read-epc (tries=%d, expected=%s)\n", *tries, safeStr(expect, "-"))

	restorePower := ""
	if *readPower >= 0 {
		old := strings.TrimSpace(queryVarRetry(p.DevicePath, "rfid.reader_1.power.read", *timeout, 2, 60*time.Millisecond))
		if old != "" && old != "?" {
			restorePower = old
		}
		if setRFIDVar(p.DevicePath, []string{"rfid.reader_1.power.read", "rfid.reader_power.read", "rfid.read_power"}, strconv.Itoa(*readPower), *timeout) {
			fmt.Printf("Read power: set to %d\n", *readPower)
		} else {
			fmt.Printf("Read power: set warning (wanted=%d)\n", *readPower)
		}
	}
	if restorePower != "" {
		defer func() {
			_ = setRFIDVar(p.DevicePath, []string{"rfid.reader_1.power.read", "rfid.reader_power.read", "rfid.read_power"}, restorePower, *timeout)
		}()
	}

	var lastHex string
	var lastResp string
	var lastL1 string
	var lastL2 string

	for i := 1; i <= *tries; i++ {
		_ = sendSGDRetry(p.DevicePath, `! U1 setvar "rfid.tag.read.content" "epc"`, 2, 70*time.Millisecond)
		_ = sendSGDRetry(p.DevicePath, `! U1 do "rfid.tag.read.execute"`, 2, 90*time.Millisecond)
		time.Sleep(140 * time.Millisecond)

		l1 := strings.TrimSpace(queryVarRetry(p.DevicePath, "rfid.tag.read.result_line1", *timeout, 2, 60*time.Millisecond))
		l2 := strings.TrimSpace(queryVarRetry(p.DevicePath, "rfid.tag.read.result_line2", *timeout, 2, 60*time.Millisecond))
		resp := strings.TrimSpace(queryVarRetry(p.DevicePath, "rfid.error.response", *timeout, 2, 60*time.Millisecond))
		hex := extractReadHex(l1, l2)

		lastHex = hex
		lastResp = resp
		lastL1 = l1
		lastL2 = l2

		fmt.Printf("Try %02d: resp=%s line1=%s line2=%s epc=%s\n", i, safeStr(resp, "-"), safeStr(l1, "-"), safeStr(l2, "-"), safeStr(hex, "-"))

		if hex == "" {
			time.Sleep(*interval)
			continue
		}

		if expect == "" {
			fmt.Printf("Read EPC: %s\n", hex)
			return nil
		}
		if strings.EqualFold(hex, expect) {
			fmt.Printf("MATCH: %s\n", hex)
			return nil
		}

		time.Sleep(*interval)
	}

	if expect != "" {
		return fmt.Errorf("expected EPC topilmadi/mos kelmadi: expected=%s got=%s resp=%s line1=%s line2=%s", expect, safeStr(lastHex, "-"), safeStr(lastResp, "-"), safeStr(lastL1, "-"), safeStr(lastL2, "-"))
	}
	return errors.New("EPC o'qilmadi (NO TAG yoki bo'sh)")
}

func extractReadHex(line1, line2 string) string {
	text := strings.ToUpper(strings.TrimSpace(strings.Trim(line1, "\"")) + strings.TrimSpace(strings.Trim(line2, "\"")))
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "-", "")
	if text == "" {
		return ""
	}
	if strings.Contains(text, "NOTAG") {
		return ""
	}

	m := readHexRegex.FindString(text)
	if m == "" {
		return ""
	}
	// odd length bo'lsa oxirgi belgini tashlaymiz (byte-aligned uchun)
	if len(m)%2 != 0 {
		m = m[:len(m)-1]
	}
	return m
}
