package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type readingMsg struct {
	reading Reading
}

type quitMsg struct{}

type clockMsg time.Time

type tuiModel struct {
	ctx        context.Context
	updates    <-chan Reading
	sourceLine string
	message    string
	last       Reading
	width      int
	height     int
}

func runTUI(ctx context.Context, updates <-chan Reading, sourceLine string, serialErr error) error {
	m := tuiModel{
		ctx:        ctx,
		updates:    updates,
		sourceLine: sourceLine,
		last:       Reading{Unit: "kg"},
		message:    "scale oqimi kutilmoqda",
	}
	if serialErr != nil {
		m.message = serialErr.Error()
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(waitForReadingCmd(m.ctx, m.updates), clockTickCmd())
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		s := strings.ToLower(strings.TrimSpace(msg.String()))
		if s == "q" || s == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	case readingMsg:
		upd := msg.reading
		if upd.Unit == "" && m.last.Unit != "" {
			upd.Unit = m.last.Unit
		}
		m.last = upd
		if upd.Error != "" {
			m.message = upd.Error
		} else {
			m.message = "ok"
		}
		return m, waitForReadingCmd(m.ctx, m.updates)
	case quitMsg:
		return m, tea.Quit
	case clockMsg:
		return m, clockTickCmd()
	default:
		return m, nil
	}
}

func (m tuiModel) View() string {
	width, height := viewSize(m.width, m.height)
	unit := strings.TrimSpace(m.last.Unit)
	if unit == "" {
		unit = "kg"
	}

	qty := "-- " + unit
	if m.last.Weight != nil {
		qty = fmt.Sprintf("%.3f %s", *m.last.Weight, unit)
	}

	port := "-"
	if strings.TrimSpace(m.last.Port) != "" {
		port = strings.TrimSpace(m.last.Port)
	}

	baud := "-"
	if m.last.Baud > 0 {
		baud = fmt.Sprintf("%d", m.last.Baud)
	}

	updatedAt := "-"
	if !m.last.UpdatedAt.IsZero() {
		updatedAt = m.last.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	status := strings.TrimSpace(m.message)
	if status == "" {
		status = "ok"
	}

	raw := strings.TrimSpace(sanitizeInline(m.last.Raw, 700))
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
		"Source : " + m.sourceLine,
		"Port   : " + port,
		"Baud   : " + baud,
	})
	appendBox("Reading", []string{
		"QTY    : " + qty,
		"Stable : " + stableText(m.last.Stable),
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
	return strings.Join(frame, "\n")
}

func waitForReadingCmd(ctx context.Context, updates <-chan Reading) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return quitMsg{}
		case upd, ok := <-updates:
			if !ok {
				return quitMsg{}
			}
			return readingMsg{reading: upd}
		}
	}
}

func clockTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockMsg(t)
	})
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

func viewSize(w, h int) (int, int) {
	if w <= 0 {
		w = 92
	}
	if h <= 0 {
		h = 28
	}

	width := 76
	if w-4 < width {
		width = w - 4
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
