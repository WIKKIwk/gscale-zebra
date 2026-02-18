package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func safeText(fallback, v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func sanitizeZPLText(v string) string {
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.ReplaceAll(v, "^", " ")
	v = strings.ReplaceAll(v, "~", " ")
	return strings.TrimSpace(v)
}

func buildRFIDEncodeCommand(epc, qtyText string) (string, error) {
	norm, err := normalizeEPC(epc)
	if err != nil {
		return "", err
	}

	qty := sanitizeZPLText(strings.TrimSpace(qtyText))
	if qty == "" {
		qty = "- kg"
	}

	return "^XA\n" +
		"^RS8,,,1,N\n" +
		fmt.Sprintf("^RFW,H,,,A^FD%s^FS\n", norm) +
		"^FO28,24^A0N,30,30\n" +
		fmt.Sprintf("^FDEPC: %s^FS\n", sanitizeZPLText(norm)) +
		"^FO28,68^A0N,30,30\n" +
		fmt.Sprintf("^FDQTY: %s^FS\n", qty) +
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
	if len(v)%2 != 0 {
		return "", errors.New("epc uzunligi juft bo'lishi kerak")
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
