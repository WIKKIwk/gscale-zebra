package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var hexOnlyRegex = regexp.MustCompile(`^[0-9A-F]+$`)

func BuildPrintTestCommandStream(message string, copies int) string {
	return "~PS\n" + BuildTestLabelZPL(message, copies)
}

func BuildTestLabelZPL(message string, copies int) string {
	if copies < 1 {
		copies = 1
	}
	if copies > 3 {
		copies = 3
	}
	msg := sanitizeZPLText(message)
	now := time.Now().Format("2006-01-02 15:04:05")

	var b strings.Builder
	b.WriteString("^XA\n")
	b.WriteString("^PW560\n")
	b.WriteString("^LL260\n")
	b.WriteString("^LH0,0\n")
	b.WriteString("^CF0,36\n")
	b.WriteString(fmt.Sprintf("^FO28,24^FD%s^FS\n", msg))
	b.WriteString("^CF0,24\n")
	b.WriteString(fmt.Sprintf("^FO28,78^FD%s^FS\n", sanitizeZPLText(now)))
	b.WriteString("^CF0,22\n")
	b.WriteString("^FO28,122^FDRFID encode yo'q (safe test)^FS\n")
	b.WriteString(fmt.Sprintf("^PQ%d,0,1,N\n", copies))
	b.WriteString("^XZ\n")
	return b.String()
}

func BuildRFIDEncodeCommandStream(epc string, copies int, feedAfter bool, printHuman bool) (string, error) {
	norm, err := NormalizeEPC(epc)
	if err != nil {
		return "", err
	}
	if copies < 1 {
		copies = 1
	}
	if copies > 1 {
		// safety: avoid wasting many tags
		copies = 1
	}

	var b strings.Builder
	b.WriteString("~PS\n") // resume printing
	b.WriteString("^XA\n")
	if feedAfter {
		b.WriteString("^MMT\n")
	}
	b.WriteString("^RS8,,,1,N\n")
	b.WriteString(fmt.Sprintf("^RFW,H,,,A^FD%s^FS\n", norm))
	if printHuman {
		b.WriteString("^FO28,24^A0N,30,30\n")
		b.WriteString(fmt.Sprintf("^FD%s^FS\n", norm))
	}
	b.WriteString(fmt.Sprintf("^PQ%d\n", copies))
	b.WriteString("^XZ\n")
	if feedAfter {
		b.WriteString("~PH\n")
	}
	return b.String(), nil
}

func BuildCalibrationCommands(save bool) []string {
	cmds := []string{
		"~JC\n", // media sensor auto calibration
	}
	if save {
		cmds = append(cmds, "^XA^JUS^XZ\n") // save current settings
	}
	return cmds
}

func NormalizeEPC(epc string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(epc))
	v = strings.TrimPrefix(v, "0X")
	v = strings.ReplaceAll(v, " ", "")
	v = strings.ReplaceAll(v, "-", "")

	if v == "" {
		return "", errors.New("epc bo'sh")
	}
	if !hexOnlyRegex.MatchString(v) {
		return "", errors.New("epc faqat hex bo'lishi kerak")
	}
	if len(v)%4 != 0 {
		return "", errors.New("epc uzunligi 4 ga bo'linishi kerak (word boundary)")
	}
	if len(v) < 8 || len(v) > 64 {
		return "", errors.New("epc uzunligi 8..64 hex oralig'ida bo'lsin")
	}
	return v, nil
}

func sanitizeZPLText(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "TEST"
	}
	replacer := strings.NewReplacer(
		"^", " ",
		"~", " ",
		"\n", " ",
		"\r", " ",
	)
	v = replacer.Replace(v)
	if len(v) > 80 {
		v = v[:80]
	}
	return v
}
