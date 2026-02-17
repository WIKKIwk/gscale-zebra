package main

import (
	"fmt"
	"strings"
	"time"
)

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

func BuildCalibrationCommands(save bool) []string {
	cmds := []string{
		"~JC\n", // media sensor auto calibration
	}
	if save {
		cmds = append(cmds, "^XA^JUS^XZ\n") // save current settings
	}
	return cmds
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
