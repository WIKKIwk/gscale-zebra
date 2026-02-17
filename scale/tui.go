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
		fmt.Print("\033[?1049h\033[H\033[2J\033[?25l")
	}

	defer func() {
		if restore != nil {
			_ = restore()
		}
		if isTTY {
			fmt.Print("\033[?25h\033[?1049l")
		}
		fmt.Print("\n")
	}()

	quit := make(chan struct{}, 1)
	if isTTY {
		go readKeys(quit)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var last Reading
	last.Unit = "kg"
	message := "scale oqimi kutilmoqda"
	if serialErr != nil {
		message = serialErr.Error()
	}

	dirty := true
	lastRender := time.Time{}

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
			dirty = true
		case <-ticker.C:
			now := time.Now()
			if dirty || now.Sub(lastRender) >= time.Second {
				render(sourceLine, message, last)
				lastRender = now
				dirty = false
			}
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
	width, height := frameSize()
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

	maxRaw := 5
	frame := make([]string, 0, 64)
	appendBox := func(title string, lines []string) {
		if len(frame) > 0 {
			frame = append(frame, "")
		}
		frame = append(frame, drawBoxLines(title, lines, width)...)
	}

	appendBox("GSCALE ZEBRA MONITOR", []string{
		"Chiqish: Q tugmasi yoki Ctrl+C",
	})
	appendBox("Connection", []string{
		"Source : " + source,
		"Port   : " + port,
		"Baud   : " + baud,
	})
	appendBox("Reading", []string{
		"QTY    : " + qty,
		"Stable : " + stableText(r.Stable),
		"Update : " + updatedAt,
	})
	appendBox("Status", []string{
		"State  : " + classifyStatus(status),
		"Detail : " + status,
	})

	baseRows := len(frame) + 1 + 4
	available := height - baseRows
	if available < 1 {
		available = 1
	}
	if available < maxRaw {
		maxRaw = available
	}
	if len(rawLines) > maxRaw {
		rawLines = rawLines[len(rawLines)-maxRaw:]
	}
	appendBox("Raw Stream", rawLines)

	if len(frame) > height {
		frame = frame[:height]
	}

	fmt.Print("\033[H\033[2J")
	fmt.Print(strings.Join(frame, "\n"))
}

func drawBoxLines(title string, lines []string, width int) []string {
	if width < 24 {
		width = 24
	}
	inner := width - 4
	if inner < 1 {
		inner = 1
	}

	out := make([]string, 0, 16)
	border := "+" + strings.Repeat("-", width-2) + "+"
	out = append(out, border)
	out = append(out, "| "+padRight(truncateText(title, inner), inner)+" |")
	out = append(out, border)

	if len(lines) == 0 {
		lines = []string{""}
	}

	for _, line := range lines {
		wrapped := wrapByWidth(line, inner)
		for _, part := range wrapped {
			out = append(out, "| "+padRight(truncateText(part, inner), inner)+" |")
		}
	}

	out = append(out, border)
	return out
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

func frameSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		w = 92
	}
	if err != nil || h <= 0 {
		h = 28
	}

	width := w - 2
	if width > 110 {
		width = 110
	}
	if width < 24 {
		width = w - 1
	}
	if width < 24 {
		width = 24
	}

	height := h - 1
	if height < 10 {
		height = 10
	}
	return width, height
}

func wrapByWidth(text string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if text == "" {
		return []string{""}
	}

	lines := make([]string, 0, 8)
	for _, row := range strings.Split(text, "\n") {
		runes := []rune(row)
		if len(runes) == 0 {
			lines = append(lines, "")
			continue
		}
		for len(runes) > width {
			lines = append(lines, string(runes[:width]))
			runes = runes[width:]
		}
		lines = append(lines, string(runes))
	}

	return lines
}

func truncateText(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func padRight(text string, width int) string {
	length := len([]rune(text))
	if length >= width {
		return text
	}
	return text + strings.Repeat(" ", width-length)
}
