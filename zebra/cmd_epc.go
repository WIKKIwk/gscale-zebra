package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func runEPCTest(args []string) error {
	fs := flag.NewFlagSet("epc-test", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	epc := fs.String("epc", "3034257BF7194E4000000001", "EPC hex")
	feed := fs.Bool("feed", false, "feed label after encode")
	printHuman := fs.Bool("print-human", false, "print EPC text on label")
	send := fs.Bool("send", false, "actually send encode command (consumes tag)")
	timeout := fs.Duration("timeout", 1500*time.Millisecond, "status query timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	stream, err := BuildRFIDEncodeCommandStream(*epc, 1, *feed, *printHuman)
	if err != nil {
		return err
	}

	norm, _ := NormalizeEPC(*epc)
	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : epc-test (epc=%s, send=%v, feed=%v, print-human=%v)\n", norm, *send, *feed, *printHuman)

	if !*send {
		fmt.Println("Ogohlantirish: hozircha DRY-RUN. Real yuborish uchun --send qo'shing.")
		fmt.Println("--- RFID command preview ---")
		fmt.Println(stream)
		return nil
	}

	beforeCount := queryVarRetry(p.DevicePath, "odometer.total_label_count", *timeout, 5, 150*time.Millisecond)
	beforeMedia := queryVarRetry(p.DevicePath, "media.status", *timeout, 5, 150*time.Millisecond)
	beforeDevice := queryVarRetry(p.DevicePath, "device.status", *timeout, 5, 150*time.Millisecond)

	if err := SendRaw(p.DevicePath, []byte(stream)); err != nil {
		return err
	}

	pause := 700 * time.Millisecond
	if *feed || *printHuman {
		pause = 1500 * time.Millisecond
	}
	time.Sleep(pause)

	read1, read2, verify := readbackRFIDResult(p.DevicePath, norm, *timeout, 5)

	afterCount := queryVarRetry(p.DevicePath, "odometer.total_label_count", *timeout, 6, 180*time.Millisecond)
	afterMedia := queryVarRetry(p.DevicePath, "media.status", *timeout, 6, 180*time.Millisecond)
	afterDevice := queryVarRetry(p.DevicePath, "device.status", *timeout, 6, 180*time.Millisecond)
	hs, hsErr := queryHostRetry(p.DevicePath, *timeout, 5, 180*time.Millisecond)

	fmt.Printf("Before: label_count=%s media=%s device=%s\n", safeStr(beforeCount, "?"), safeStr(beforeMedia, "?"), safeStr(beforeDevice, "?"))
	fmt.Printf("After : label_count=%s media=%s device=%s\n", safeStr(afterCount, "?"), safeStr(afterMedia, "?"), safeStr(afterDevice, "?"))
	fmt.Printf("Read  : line1=%s line2=%s verify=%s\n", safeStr(read1, "-"), safeStr(read2, "-"), verify)
	if hsErr != nil {
		fmt.Printf("~HS   : no response (%v)\n", hsErr)
	} else {
		if len(hs) > 260 {
			hs = hs[:260] + "..."
		}
		fmt.Printf("~HS   : %s\n", strings.ReplaceAll(hs, "\n", " | "))
	}

	fmt.Println("EPC test command yuborildi.")
	return nil
}

func readbackRFIDResult(device, expected string, timeout time.Duration, retries int) (string, string, string) {
	if retries < 1 {
		retries = 1
	}

	var line1 string
	var line2 string
	verify := "UNKNOWN"

	for i := 0; i < retries; i++ {
		_ = SendRaw(device, []byte("! U1 setvar \"rfid.tag.read.content\" \"epc\"\r\n"))
		time.Sleep(80 * time.Millisecond)
		_ = SendRaw(device, []byte("! U1 do \"rfid.tag.read.execute\"\r\n"))
		time.Sleep(240 * time.Millisecond)

		line1 = queryVarRetry(device, "rfid.tag.read.result_line1", timeout, 4, 120*time.Millisecond)
		line2 = queryVarRetry(device, "rfid.tag.read.result_line2", timeout, 4, 120*time.Millisecond)
		verify = inferVerify(line1, line2, expected)

		if verify == "MATCH" || verify == "MISMATCH" || verify == "OK" {
			break
		}
	}

	return line1, line2, verify
}

func queryVarRetry(device, key string, timeout time.Duration, retries int, delay time.Duration) string {
	if retries < 1 {
		retries = 1
	}
	for i := 0; i < retries; i++ {
		v, err := QuerySGDVar(device, key, timeout)
		if err == nil && strings.TrimSpace(v) != "" {
			return v
		}
		time.Sleep(delay)
	}
	return ""
}

func queryHostRetry(device string, timeout time.Duration, retries int, delay time.Duration) (string, error) {
	if retries < 1 {
		retries = 1
	}
	var lastErr error
	for i := 0; i < retries; i++ {
		v, err := QueryHostStatus(device, timeout)
		if err == nil && strings.TrimSpace(v) != "" {
			return v, nil
		}
		lastErr = err
		time.Sleep(delay)
	}
	return "", lastErr
}
