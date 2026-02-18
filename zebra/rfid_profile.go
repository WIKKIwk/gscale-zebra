package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type rfidProfileOptions struct {
	LabelTries    int
	ErrorHandling string
	ReadPower     int
	WritePower    int
}

func defaultRFIDProfileOptions() rfidProfileOptions {
	return rfidProfileOptions{
		LabelTries:    1,
		ErrorHandling: "none",
		ReadPower:     30,
		WritePower:    30,
	}
}

func applyRFIDProfile(device string, timeout time.Duration, opt rfidProfileOptions) string {
	notes := make([]string, 0, 8)

	if opt.LabelTries < 1 {
		opt.LabelTries = 1
	}
	if opt.LabelTries > 10 {
		opt.LabelTries = 10
	}
	opt.ReadPower = clampInt(opt.ReadPower, 0, 30)
	opt.WritePower = clampInt(opt.WritePower, 0, 30)

	if err := sendRawRetry(device, []byte("~PS\n"), 3, 90*time.Millisecond); err == nil {
		notes = append(notes, "resume=ok")
	} else {
		notes = append(notes, "resume=warn")
	}

	if setRFIDVar(device, []string{"rfid.enable"}, "on", timeout) {
		notes = append(notes, "enable=on")
	} else {
		notes = append(notes, "enable=warn")
	}

	if setRFIDVar(device, []string{"rfid.label_tries"}, strconv.Itoa(opt.LabelTries), timeout) {
		notes = append(notes, fmt.Sprintf("tries=%d", opt.LabelTries))
	} else {
		notes = append(notes, "tries=warn")
	}

	errMode := normalizeRFIDErrorHandling(opt.ErrorHandling)
	if setRFIDVar(device, []string{"rfid.error_handling"}, errMode, timeout) {
		notes = append(notes, "error="+errMode)
	} else {
		notes = append(notes, "error=warn")
	}

	if setRFIDVar(device, []string{"rfid.tag.read.content"}, "epc", timeout) {
		notes = append(notes, "read_content=epc")
	} else {
		notes = append(notes, "read_content=warn")
	}

	if setRFIDVar(device, []string{"rfid.reader_power.read", "rfid.read_power"}, strconv.Itoa(opt.ReadPower), timeout) {
		notes = append(notes, fmt.Sprintf("read_pwr=%d", opt.ReadPower))
	} else {
		notes = append(notes, "read_pwr=warn")
	}

	if setRFIDVar(device, []string{"rfid.reader_power.write", "rfid.write_power"}, strconv.Itoa(opt.WritePower), timeout) {
		notes = append(notes, fmt.Sprintf("write_pwr=%d", opt.WritePower))
	} else {
		notes = append(notes, "write_pwr=warn")
	}

	return strings.Join(notes, ", ")
}

func setRFIDVar(device string, keys []string, value string, timeout time.Duration) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(keys) == 0 {
		return false
	}

	// Known keylarni tekshirib birinchisini tanlaymiz; query bo'lmasa default birinchisi.
	key := strings.TrimSpace(keys[0])
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if got := queryVarRetry(device, k, timeout, 1, 0); got != "" {
			key = k
			break
		}
	}

	cmd := fmt.Sprintf("! U1 setvar \"%s\" \"%s\"\r\n", key, strings.ReplaceAll(value, "\"", ""))
	if err := sendRawRetry(device, []byte(cmd), 3, 90*time.Millisecond); err != nil {
		return false
	}

	// Read-back har modelda ham bo'lmasligi mumkin; set yuborilganini success deb olamiz.
	return true
}

func normalizeRFIDErrorHandling(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "pause", "p":
		return "pause"
	case "error", "e":
		return "error"
	default:
		return "none"
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
