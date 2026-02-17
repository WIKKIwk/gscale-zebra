package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func runTUI(ctx context.Context, updates <-chan Reading, sourceLine string, serialErr error) error {
	stdinFD := int(os.Stdin.Fd())
	isTTY := term.IsTerminal(stdinFD)
	var restore func() error
	if isTTY {
		oldState, err := term.MakeRaw(stdinFD)
		if err == nil {
			restore = func() error { return term.Restore(stdinFD, oldState) }
		}
	}

	if restore != nil {
		defer func() {
			_ = restore()
			fmt.Print("\n")
		}()
	}

	quit := make(chan struct{}, 1)
	if isTTY {
		go readKeys(quit)
	}

	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	var last Reading
	last.Unit = "kg"
	message := "scale oqimi kutilmoqda"
	if serialErr != nil {
		message = serialErr.Error()
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-quit:
			return nil
		case upd := <-updates:
			if upd.Unit == "" && last.Unit != "" {
				upd.Unit = last.Unit
			}
			last = upd
			if upd.Error != "" {
				message = upd.Error
			} else {
				message = "ok"
			}
		case <-ticker.C:
			render(sourceLine, message, last)
		}
	}
}

func readKeys(quit chan<- struct{}) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 3 {
			select {
			case quit <- struct{}{}:
			default:
			}
			return
		}
	}
}

func render(source, message string, r Reading) {
	width := terminalWidth()
	unit := strings.TrimSpace(r.Unit)
	if unit == "" {
		unit = "kg"
	}

	qty := "-- " + unit
	if r.Weight != nil {
		qty = fmt.Sprintf("%.3f %s", *r.Weight, unit)
	}

	port := "-"
	if strings.TrimSpace(r.Port) != "" {
		port = strings.TrimSpace(r.Port)
	}

	baud := "-"
	if r.Baud > 0 {
		baud = fmt.Sprintf("%d", r.Baud)
	}

	updatedAt := "-"
	if !r.UpdatedAt.IsZero() {
		updatedAt = r.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	status := strings.TrimSpace(message)
	if status == "" {
		status = "ok"
	}

	raw := strings.TrimSpace(sanitizeInline(r.Raw, 700))
	if raw == "" {
		raw = "(empty)"
	}
	rawLines := wrapByWidth(raw, width-4)
	if len(rawLines) > 5 {
		rawLines = rawLines[len(rawLines)-5:]
	}

	fmt.Print("\033[2J\033[H")
	fmt.Println(drawBox("GSCALE ZEBRA MONITOR", []string{
		"Chiqish: Q tugmasi yoki Ctrl+C",
	}, width))
	fmt.Println()
	fmt.Println(drawBox("Connection", []string{
		"Source : " + source,
		"Port   : " + port,
		"Baud   : " + baud,
	}, width))
	fmt.Println()
	fmt.Println(drawBox("Reading", []string{
		"QTY    : " + qty,
		"Stable : " + stableText(r.Stable),
		"Update : " + updatedAt,
	}, width))
	fmt.Println()
	fmt.Println(drawBox("Status", []string{
		"State  : " + classifyStatus(status),
		"Detail : " + status,
	}, width))
	fmt.Println()
	fmt.Println(drawBox("Raw Stream", rawLines, width))
}

func drawBox(title string, lines []string, width int) string {
	if width < 60 {
		width = 60
	}
	inner := width - 4
	if inner < 1 {
		inner = 1
	}

	var b strings.Builder
	b.WriteString("+" + strings.Repeat("-", width-2) + "+\n")
	b.WriteString("| " + padRight(truncateText(title, inner), inner) + " |\n")
	b.WriteString("+" + strings.Repeat("-", width-2) + "+\n")

	if len(lines) == 0 {
		lines = []string{""}
	}

	for _, line := range lines {
		wrapped := wrapByWidth(line, inner)
		for _, part := range wrapped {
			b.WriteString("| " + padRight(truncateText(part, inner), inner) + " |\n")
		}
	}

	b.WriteString("+" + strings.Repeat("-", width-2) + "+")
	return b.String()
}

func classifyStatus(status string) string {
	v := strings.ToLower(status)
	switch {
	case v == "ok":
		return "OK"
	case strings.Contains(v, "busy") || strings.Contains(v, "timeout"):
		return "WARN"
	default:
		return "ERROR"
	}
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 92
	}
	if w < 70 {
		return 70
	}
	if w > 140 {
		return 140
	}
	return w
}

func wrapByWidth(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if text == "" {
		return []string{""}
	}

	lines := make([]string, 0, 4)
	for _, row := range strings.Split(text, "\n") {
		remaining := row
		for len(remaining) > width {
			lines = append(lines, remaining[:width])
			remaining = remaining[width:]
		}
		lines = append(lines, remaining)
	}

	return lines
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}

func padRight(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}
