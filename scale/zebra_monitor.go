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
	Attempts    int
	AutoTuned   bool
	Note        string
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
	st.DeviceState = safeText("-", queryVarRetry(p.DevicePath, "device.status", timeout, 3, 90*time.Millisecond))
	st.MediaState = safeText("-", queryVarRetry(p.DevicePath, "media.status", timeout, 3, 90*time.Millisecond))
	st.ReadLine1 = "-"
	st.ReadLine2 = "-"
	st.Verify = "-"
	return st
}

func runZebraRead(preferredDevice string, timeout time.Duration) ZebraStatus {
	st := ZebraStatus{
		Action:    "read",
		Verify:    "-",
		UpdatedAt: time.Now(),
		Attempts:  1,
	}

	p, err := SelectZebraPrinter(preferredDevice)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.Connected = true
	st.DevicePath = p.DevicePath
	st.Name = p.DisplayName()

	line1, line2, verify := readbackRFIDResult(p.DevicePath, "", timeout, 5)
	st.ReadLine1 = safeText("-", line1)
	st.ReadLine2 = safeText("-", line2)
	st.Verify = verify
	st.DeviceState = safeText("-", queryVarRetry(p.DevicePath, "device.status", timeout, 3, 90*time.Millisecond))
	st.MediaState = safeText("-", queryVarRetry(p.DevicePath, "media.status", timeout, 3, 90*time.Millisecond))
	return st
}

func runZebraEncodeAndRead(preferredDevice, epc string, timeout time.Duration) ZebraStatus {
	st := ZebraStatus{
		Action:    "encode",
		Verify:    "-",
		UpdatedAt: time.Now(),
		Attempts:  1,
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

	line1, line2, verify, err := encodeAndVerify(p.DevicePath, norm, timeout)
	if err != nil {
		st.Error = err.Error()
		applyZebraSnapshot(&st, p, timeout)
		return st
	}
	st.ReadLine1 = safeText("-", line1)
	st.ReadLine2 = safeText("-", line2)
	st.Verify = verify

	st.DeviceState = safeText("-", queryVarRetry(p.DevicePath, "device.status", timeout, 3, 90*time.Millisecond))
	st.MediaState = safeText("-", queryVarRetry(p.DevicePath, "media.status", timeout, 3, 90*time.Millisecond))
	if st.Verify != "MATCH" && strings.TrimSpace(st.Error) == "" {
		st.Note = strings.TrimSpace(strings.Join([]string{st.Note, "verify=" + st.Verify}, " "))
	}
	return st
}

func encodeAndVerify(device, epc string, timeout time.Duration) (string, string, string, error) {
	stream, err := buildRFIDEncodeCommand(epc)
	if err != nil {
		return "", "", "UNKNOWN", err
	}
	if err := sendRawRetry(device, []byte(stream), 5, 120*time.Millisecond); err != nil {
		return "", "", "UNKNOWN", err
	}
	time.Sleep(820 * time.Millisecond)

	line1, line2, verify := readbackRFIDResult(device, epc, timeout, 1)
	return line1, line2, verify, nil
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
		time.Sleep(240 * time.Millisecond)

		line1 = queryVarRetry(device, "rfid.tag.read.result_line1", timeout, 3, 100*time.Millisecond)
		line2 = queryVarRetry(device, "rfid.tag.read.result_line2", timeout, 3, 100*time.Millisecond)
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
		waitReady(device, 3*time.Second)
	}
	if err := sendRawRetry(device, []byte("^XA^HR^XZ\n"), 3, 140*time.Millisecond); err == nil {
		notes = append(notes, "^HR")
		waitReady(device, 3*time.Second)
	}
	if err := sendRawRetry(device, []byte("~JC\n"), 3, 140*time.Millisecond); err == nil {
		notes = append(notes, "~JC")
		waitReady(device, 4*time.Second)
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
		v := queryVarRetry(device, "device.status", 650*time.Millisecond, 1, 0)
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
		err := zebraSendRaw(device, payload)
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
		lastErr = errors.New("zebra: send retry failed")
	}
	return lastErr
}

func sendSGDRetry(device, command string, retries int, delay time.Duration) error {
	if retries < 1 {
		retries = 1
	}
	var lastErr error
	for i := 0; i < retries; i++ {
		err := zebraSendSGD(device, command)
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
		lastErr = errors.New("zebra: sgd retry failed")
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
		v, err := queryZebraSGDVar(device, key, timeout)
		ch <- result{v: v, err: err}
	}()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-time.After(timeout + 250*time.Millisecond):
		return "", errors.New("query timeout")
	}
}

func isBusyLikeError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "resource busy") || strings.Contains(msg, "device or resource busy") || strings.Contains(msg, "temporarily unavailable")
}

func applyZebraSnapshot(st *ZebraStatus, p ZebraPrinter, timeout time.Duration) {
	st.DeviceState = safeText("-", queryVarRetry(p.DevicePath, "device.status", timeout, 3, 90*time.Millisecond))
	st.MediaState = safeText("-", queryVarRetry(p.DevicePath, "media.status", timeout, 3, 90*time.Millisecond))
	st.ReadLine1 = "-"
	st.ReadLine2 = "-"
}

func publishZebraStatus(ch chan<- ZebraStatus, st ZebraStatus) {
	select {
	case ch <- st:
	default:
	}
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
	return "^XA\n" +
		"^RS8,,,1,N\n" +
		fmt.Sprintf("^RFW,H,,,A^FD%s^FS\n", norm) +
		"^FO28,24^A0N,30,30\n" +
		fmt.Sprintf("^FD%s^FS\n", norm) +
		"^PQ1\n" +
		"^XZ\n", nil
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
