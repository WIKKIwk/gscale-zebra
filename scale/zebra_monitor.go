package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var zebraHexOnlyRegex = regexp.MustCompile(`^[0-9A-F]+$`)

type ZebraStatus struct {
	Connected   bool
	DevicePath  string
	Name        string
	DeviceState string
	MediaState  string
	ReadLine1   string
	ReadLine2   string
	LastEPC     string
	Verify      string
	Action      string
	Error       string
	UpdatedAt   time.Time
}

func startZebraMonitor(ctx context.Context, preferredDevice string, interval time.Duration, out chan<- ZebraStatus) {
	if out == nil {
		return
	}
	if interval < 300*time.Millisecond {
		interval = 300 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		publishZebraStatus(out, collectZebraStatus(preferredDevice, 900*time.Millisecond))
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				publishZebraStatus(out, collectZebraStatus(preferredDevice, 900*time.Millisecond))
			}
		}
	}()
}

func collectZebraStatus(preferredDevice string, timeout time.Duration) ZebraStatus {
	st := ZebraStatus{
		Connected: false,
		Verify:    "-",
		UpdatedAt: time.Now(),
	}

	p, err := SelectZebraPrinter(preferredDevice)
	if err != nil {
		st.Error = err.Error()
		return st
	}

	st.Connected = true
	st.DevicePath = p.DevicePath
	st.Name = p.DisplayName()
	st.DeviceState = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "device.status", timeout)
	}))
	st.MediaState = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "media.status", timeout)
	}))
	st.ReadLine1 = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "rfid.tag.read.result_line1", timeout)
	}))
	st.ReadLine2 = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "rfid.tag.read.result_line2", timeout)
	}))
	st.Verify = inferVerify(st.ReadLine1, st.ReadLine2, "")
	return st
}

func runZebraRead(preferredDevice string, timeout time.Duration) ZebraStatus {
	st := ZebraStatus{
		Action:    "read",
		Verify:    "-",
		UpdatedAt: time.Now(),
	}

	p, err := SelectZebraPrinter(preferredDevice)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.Connected = true
	st.DevicePath = p.DevicePath
	st.Name = p.DisplayName()

	if err := zebraSendSGD(p.DevicePath, `! U1 setvar "rfid.tag.read.content" "epc"`); err != nil {
		st.Error = err.Error()
		return st
	}
	if err := zebraSendSGD(p.DevicePath, `! U1 do "rfid.tag.read.execute"`); err != nil {
		st.Error = err.Error()
		return st
	}

	time.Sleep(260 * time.Millisecond)
	applyZebraSnapshot(&st, p, timeout)
	st.Verify = inferVerify(st.ReadLine1, st.ReadLine2, "")
	return st
}

func runZebraEncodeAndRead(preferredDevice, epc string, timeout time.Duration) ZebraStatus {
	st := ZebraStatus{
		Action:    "encode",
		Verify:    "-",
		UpdatedAt: time.Now(),
	}

	norm, err := normalizeEPC(epc)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.LastEPC = norm

	p, err := SelectZebraPrinter(preferredDevice)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.Connected = true
	st.DevicePath = p.DevicePath
	st.Name = p.DisplayName()

	stream, err := buildRFIDEncodeCommand(norm)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	if err := zebraSendRaw(p.DevicePath, []byte(stream)); err != nil {
		st.Error = err.Error()
		return st
	}

	time.Sleep(900 * time.Millisecond)
	if err := zebraSendSGD(p.DevicePath, `! U1 setvar "rfid.tag.read.content" "epc"`); err != nil {
		st.Error = err.Error()
		applyZebraSnapshot(&st, p, timeout)
		return st
	}
	if err := zebraSendSGD(p.DevicePath, `! U1 do "rfid.tag.read.execute"`); err != nil {
		st.Error = err.Error()
		applyZebraSnapshot(&st, p, timeout)
		return st
	}

	time.Sleep(260 * time.Millisecond)
	applyZebraSnapshot(&st, p, timeout)
	st.Verify = inferVerify(st.ReadLine1, st.ReadLine2, norm)
	return st
}

func applyZebraSnapshot(st *ZebraStatus, p ZebraPrinter, timeout time.Duration) {
	st.DeviceState = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "device.status", timeout)
	}))
	st.MediaState = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "media.status", timeout)
	}))
	st.ReadLine1 = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "rfid.tag.read.result_line1", timeout)
	}))
	st.ReadLine2 = safeText("-", queryOrFallback(func() (string, error) {
		return queryZebraSGDVar(p.DevicePath, "rfid.tag.read.result_line2", timeout)
	}))
}

func publishZebraStatus(ch chan<- ZebraStatus, st ZebraStatus) {
	select {
	case ch <- st:
	default:
	}
}

func queryOrFallback(fn func() (string, error)) string {
	v, err := fn()
	if err != nil {
		return ""
	}
	return v
}

func safeText(fallback, v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func buildRFIDEncodeCommand(epc string) (string, error) {
	norm, err := normalizeEPC(epc)
	if err != nil {
		return "", err
	}
	return "~PS\n" +
		"^XA\n" +
		"^MMT\n" +
		"^PW560\n" +
		"^LL260\n" +
		"^RS8,,,1,N\n" +
		fmt.Sprintf("^RFW,H,,,A^FD%s^FS\n", norm) +
		"^FO28,24^A0N,32,32\n" +
		fmt.Sprintf("^FD%s^FS\n", norm) +
		"^PQ1\n" +
		"^XZ\n" +
		"~PH\n", nil
}

func normalizeEPC(epc string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(epc))
	v = strings.TrimPrefix(v, "0X")
	v = strings.ReplaceAll(v, " ", "")
	v = strings.ReplaceAll(v, "-", "")

	if v == "" {
		return "", errors.New("epc bo'sh")
	}
	if !zebraHexOnlyRegex.MatchString(v) {
		return "", errors.New("epc faqat hex bo'lishi kerak")
	}
	if len(v)%4 != 0 {
		return "", errors.New("epc uzunligi 4 ga bo'linishi kerak")
	}
	if len(v) < 8 || len(v) > 64 {
		return "", errors.New("epc uzunligi 8..64 oralig'ida bo'lsin")
	}
	return v, nil
}

func generateTestEPC(t time.Time) string {
	sec := uint64(t.Unix()) & 0xFFFFFFFF
	ns := uint64(t.Nanosecond()) & 0xFFFFFFFF
	return fmt.Sprintf("3034%08X%08X", sec, ns)
}

func inferVerify(line1, line2, expected string) string {
	line1 = strings.TrimSpace(strings.Trim(line1, "\""))
	line2 = strings.TrimSpace(strings.Trim(line2, "\""))
	text := strings.ToUpper(strings.TrimSpace(line1 + " " + line2))

	if text == "" || text == "-" {
		return "UNKNOWN"
	}
	if strings.Contains(strings.ToLower(text), "no tag") {
		return "NO TAG"
	}
	if strings.Contains(strings.ToLower(text), "error") {
		return "ERROR"
	}

	expected = strings.ToUpper(strings.TrimSpace(expected))
	if expected != "" {
		if strings.Contains(strings.ReplaceAll(text, " ", ""), expected) {
			return "MATCH"
		}
		return "MISMATCH"
	}

	return "OK"
}
