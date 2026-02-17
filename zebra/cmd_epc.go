package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"
)

func runEPCTest(args []string) error {
	fs := flag.NewFlagSet("epc-test", flag.ContinueOnError)
	device := fs.String("device", "", "printer device path (example: /dev/usb/lp0)")
	epc := fs.String("epc", "3034257BF7194E4000000001", "EPC hex")
	feed := fs.Bool("feed", true, "feed label after encode")
	printHuman := fs.Bool("print-human", true, "print EPC text on label")
	autoTune := fs.Bool("auto-tune", true, "auto calibrate and retry on NO TAG")
	send := fs.Bool("send", false, "actually send encode command (consumes tag)")
	timeout := fs.Duration("timeout", 1500*time.Millisecond, "status query timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, err := SelectPrinter(*device)
	if err != nil {
		return err
	}

	// Verify oldin bo'lishi uchun birinchi urinishda feed qilmaymiz.
	stream, err := BuildRFIDEncodeCommandStream(*epc, 1, false, *printHuman)
	if err != nil {
		return err
	}

	norm, _ := NormalizeEPC(*epc)
	fmt.Printf("Printer: %s (%s)\n", p.DevicePath, p.DisplayName())
	fmt.Printf("Action : epc-test (epc=%s, send=%v, feed=%v, print-human=%v, auto-tune=%v)\n", norm, *send, *feed, *printHuman, *autoTune)

	if !*send {
		fmt.Println("Ogohlantirish: hozircha DRY-RUN. Real yuborish uchun --send qo'shing.")
		fmt.Println("--- RFID command preview ---")
		fmt.Println(stream)
		return nil
	}

	beforeCount := queryVarRetry(p.DevicePath, "odometer.total_label_count", *timeout, 4, 120*time.Millisecond)
	beforeMedia := queryVarRetry(p.DevicePath, "media.status", *timeout, 4, 120*time.Millisecond)
	beforeDevice := queryVarRetry(p.DevicePath, "device.status", *timeout, 4, 120*time.Millisecond)

	attempts := 1
	auto := "no"
	note := "-"
	read1, read2, verify, runErr := runEPCAttempt(p.DevicePath, stream, norm, *timeout)
	if runErr != nil {
		return runErr
	}

	if *autoTune && shouldAutoTune(verify) {
		auto = "yes"
		attempts = 2
		note = runAutoTuneSequence(p.DevicePath)
		read1, read2, verify, runErr = runEPCAttempt(p.DevicePath, stream, norm, *timeout)
		if runErr != nil {
			return runErr
		}
	}

	if *feed {
		_ = sendRawRetry(p.DevicePath, []byte("~PH\n"), 4, 120*time.Millisecond)
		time.Sleep(120 * time.Millisecond)
	}

	afterCount := queryVarRetry(p.DevicePath, "odometer.total_label_count", *timeout, 5, 150*time.Millisecond)
	afterMedia := queryVarRetry(p.DevicePath, "media.status", *timeout, 5, 150*time.Millisecond)
	afterDevice := queryVarRetry(p.DevicePath, "device.status", *timeout, 5, 150*time.Millisecond)
	hs, hsErr := queryHostRetry(p.DevicePath, *timeout, 3, 120*time.Millisecond)

	fmt.Printf("Before: label_count=%s media=%s device=%s\n", safeStr(beforeCount, "?"), safeStr(beforeMedia, "?"), safeStr(beforeDevice, "?"))
	fmt.Printf("After : label_count=%s media=%s device=%s\n", safeStr(afterCount, "?"), safeStr(afterMedia, "?"), safeStr(afterDevice, "?"))
	fmt.Printf("Read  : line1=%s line2=%s verify=%s attempts=%d auto_tune=%s\n", safeStr(read1, "-"), safeStr(read2, "-"), verify, attempts, auto)
	if note != "-" {
		fmt.Printf("Tune  : %s\n", note)
	}
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

func runEPCAttempt(device, stream, expected string, timeout time.Duration) (string, string, string, error) {
	if err := sendRawRetry(device, []byte(stream), 5, 120*time.Millisecond); err != nil {
		return "", "", "UNKNOWN", err
	}
	time.Sleep(820 * time.Millisecond)

	read1, read2, verify := readbackRFIDResult(device, expected, timeout, 4)
	return read1, read2, verify, nil
}

func readbackRFIDResult(device, expected string, timeout time.Duration, retries int) (string, string, string) {
	if retries < 1 {
		retries = 1
	}

	var line1 string
	var line2 string
	verify := "UNKNOWN"

	for i := 0; i < retries; i++ {
		_ = sendSGDRetry(device, `! U1 setvar "rfid.tag.read.content" "epc"`, 3, 90*time.Millisecond)
		time.Sleep(70 * time.Millisecond)
		_ = sendSGDRetry(device, `! U1 do "rfid.tag.read.execute"`, 3, 90*time.Millisecond)
		time.Sleep(220 * time.Millisecond)

		line1 = queryVarRetry(device, "rfid.tag.read.result_line1", timeout, 3, 90*time.Millisecond)
		line2 = queryVarRetry(device, "rfid.tag.read.result_line2", timeout, 3, 90*time.Millisecond)
		verify = inferVerify(line1, line2, expected)

		if verify == "MATCH" || verify == "MISMATCH" || verify == "OK" {
			break
		}
	}

	return line1, line2, verify
}

func shouldAutoTune(verify string) bool {
	v := strings.ToUpper(strings.TrimSpace(verify))
	return v == "NO TAG" || v == "UNKNOWN" || v == "ERROR"
}

func runAutoTuneSequence(device string) string {
	notes := make([]string, 0, 4)

	if err := sendSGDRetry(device, `! U1 do "rfid.calibrate"`, 3, 120*time.Millisecond); err == nil {
		notes = append(notes, "rfid.calibrate")
		waitReady(device, 2*time.Second)
	}
	if err := sendRawRetry(device, []byte("^XA^HR^XZ\n"), 3, 140*time.Millisecond); err == nil {
		notes = append(notes, "^HR")
		waitReady(device, 2*time.Second)
	}
	if err := sendRawRetry(device, []byte("~JC\n"), 3, 140*time.Millisecond); err == nil {
		notes = append(notes, "~JC")
		waitReady(device, 3*time.Second)
	}
	if err := sendRawRetry(device, []byte("^XA^JUS^XZ\n"), 2, 120*time.Millisecond); err == nil {
		notes = append(notes, "save")
	}

	if len(notes) == 0 {
		return "auto-tune command yuborilmadi"
	}
	return "auto-tune: " + strings.Join(notes, ",")
}

func waitReady(device string, wait time.Duration) {
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		v := queryVarRetry(device, "device.status", 500*time.Millisecond, 1, 0)
		if strings.EqualFold(strings.TrimSpace(v), "ready") {
			return
		}
		time.Sleep(120 * time.Millisecond)
	}
}

func sendRawRetry(device string, payload []byte, retries int, delay time.Duration) error {
	if retries < 1 {
		retries = 1
	}
	var lastErr error
	for i := 0; i < retries; i++ {
		err := SendRaw(device, payload)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isBusyLikeError(err) {
			return err
		}
		time.Sleep(delay)
	}
	if lastErr == nil {
		lastErr = errors.New("send retry failed")
	}
	return lastErr
}

func sendSGDRetry(device, command string, retries int, delay time.Duration) error {
	if retries < 1 {
		retries = 1
	}
	if !strings.HasSuffix(command, "\r\n") {
		command += "\r\n"
	}
	var lastErr error
	for i := 0; i < retries; i++ {
		err := SendRaw(device, []byte(command))
		if err == nil {
			return nil
		}
		lastErr = err
		if !isBusyLikeError(err) {
			return err
		}
		time.Sleep(delay)
	}
	if lastErr == nil {
		lastErr = errors.New("sgd retry failed")
	}
	return lastErr
}

func queryVarRetry(device, key string, timeout time.Duration, retries int, delay time.Duration) string {
	if retries < 1 {
		retries = 1
	}
	for i := 0; i < retries; i++ {
		v, err := queryVarSoft(device, key, timeout)
		if err == nil {
			v = strings.TrimSpace(strings.Trim(v, "\""))
			if v != "" && v != "?" {
				return v
			}
		}
		time.Sleep(delay)
	}
	return ""
}

func queryVarSoft(device, key string, timeout time.Duration) (string, error) {
	type result struct {
		v   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := QuerySGDVar(device, key, timeout)
		ch <- result{v: v, err: err}
	}()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-time.After(timeout + 250*time.Millisecond):
		return "", errors.New("query timeout")
	}
}

func queryHostRetry(device string, timeout time.Duration, retries int, delay time.Duration) (string, error) {
	if retries < 1 {
		retries = 1
	}
	var lastErr error
	for i := 0; i < retries; i++ {
		v, err := queryHostSoft(device, timeout)
		if err == nil && strings.TrimSpace(v) != "" {
			return v, nil
		}
		lastErr = err
		time.Sleep(delay)
	}
	return "", lastErr
}

func queryHostSoft(device string, timeout time.Duration) (string, error) {
	type result struct {
		v   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := QueryHostStatus(device, timeout)
		ch <- result{v: v, err: err}
	}()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-time.After(timeout + 250*time.Millisecond):
		return "", errors.New("host status timeout")
	}
}

func isBusyLikeError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "resource busy") || strings.Contains(msg, "device or resource busy") || strings.Contains(msg, "temporarily unavailable")
}
