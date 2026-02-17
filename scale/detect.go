package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tarm/serial"
)

func detectScalePort(device string, bauds []int, probeTimeout time.Duration, unit string) (string, int, error) {
	if strings.TrimSpace(device) != "" {
		return strings.TrimSpace(device), bauds[0], nil
	}

	candidates := listCandidates()
	if len(candidates) == 0 {
		return "", 0, errors.New("serial device topilmadi (/dev/ttyUSB* yoki /dev/ttyACM*)")
	}

	var lastBusy error
	for _, dev := range candidates {
		for _, b := range bauds {
			found, hasData, err := probePort(dev, b, probeTimeout, unit)
			if err != nil {
				if isBusyErr(err) {
					lastBusy = fmt.Errorf("%s band: %w", dev, err)
					continue
				}
				continue
			}
			if found {
				return dev, b, nil
			}
			if hasData {
				return dev, b, nil
			}
		}
	}

	if lastBusy != nil {
		return "", 0, fmt.Errorf("serial port band: %w", lastBusy)
	}

	return candidates[0], bauds[0], nil
}

func listCandidates() []string {
	seen := map[string]bool{}
	out := make([]string, 0, 16)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		out = append(out, v)
	}

	if byID, err := filepath.Glob("/dev/serial/by-id/*"); err == nil {
		sort.Strings(byID)
		for _, path := range byID {
			target, err := filepath.EvalSymlinks(path)
			if err == nil {
				add(target)
				continue
			}
			add(path)
		}
	}

	for _, pattern := range []string{"/dev/ttyUSB*", "/dev/ttyACM*"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		sort.Strings(matches)
		for _, path := range matches {
			add(path)
		}
	}

	return out
}

func probePort(device string, baud int, timeout time.Duration, unit string) (bool, bool, error) {
	port, err := serial.OpenPort(&serial.Config{Name: device, Baud: baud, ReadTimeout: timeout})
	if err != nil {
		return false, false, err
	}
	defer port.Close()

	deadline := time.Now().Add(timeout)
	buf := make([]byte, 256)
	raw := ""
	hasData := false
	for time.Now().Before(deadline) {
		n, err := port.Read(buf)
		if err != nil {
			return false, hasData, err
		}
		if n == 0 {
			continue
		}
		hasData = true
		raw = appendRaw(raw, string(buf[:n]), 240)
		if _, _, _, ok := parseWeight(raw, unit); ok {
			return true, true, nil
		}
	}

	return false, hasData, nil
}

func isBusyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "resource busy") || strings.Contains(msg, "device or resource busy") || strings.Contains(msg, "permission denied")
}
