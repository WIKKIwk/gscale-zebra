package main

import (
	"fmt"
	"strings"
	"time"
)

func runZebraRead(preferredDevice string, timeout time.Duration) ZebraStatus {
	zebraIOMutex.Lock()
	defer zebraIOMutex.Unlock()

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

func runZebraEncodeAndRead(preferredDevice, epc, qtyText string, timeout time.Duration) ZebraStatus {
	zebraIOMutex.Lock()
	defer zebraIOMutex.Unlock()

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

	line1, line2, verify, err := encodeAndVerify(p.DevicePath, norm, qtyText, timeout)
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

func encodeAndVerify(device, epc, qtyText string, timeout time.Duration) (string, string, string, error) {
	stream, err := buildRFIDEncodeCommand(epc, qtyText)
	if err != nil {
		return "", "", "UNKNOWN", err
	}
	if err := sendRawRetry(device, []byte(stream), 18, 140*time.Millisecond); err != nil {
		if isBusyLikeError(err) {
			return "", "", "UNKNOWN", fmt.Errorf("%w (printer busy: boshqa process /dev/usb/lp0 ni band qilgan)", err)
		}
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
