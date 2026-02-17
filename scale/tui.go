package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	now        time.Time
	history    []float64
}

func runTUI(ctx context.Context, updates <-chan Reading, sourceLine string, serialErr error) error {
	m := tuiModel{
		ctx:        ctx,
		updates:    updates,
		sourceLine: sourceLine,
		last:       Reading{Unit: "kg"},
		message:    "scale oqimi kutilmoqda",
		now:        time.Now(),
		history:    make([]float64, 0, 256),
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
		if upd.Weight != nil {
			m.history = append(m.history, *upd.Weight)
			if len(m.history) > 240 {
				m.history = m.history[len(m.history)-240:]
			}
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
		m.now = time.Time(msg)
		return m, clockTickCmd()
	default:
		return m, nil
	}
}

func (m tuiModel) View() string {
	w, _ := viewSize(m.width, m.height)
	now := m.now
	if now.IsZero() {
		now = time.Now()
	}

	unit := strings.TrimSpace(m.last.Unit)
	if unit == "" {
		unit = "kg"
	}
	qty := "-- " + unit
	if m.last.Weight != nil {
		qty = fmt.Sprintf("%.3f %s", *m.last.Weight, unit)
	}

	status := strings.TrimSpace(m.message)
	if status == "" {
		status = "ok"
	}

	connected := isConnected(status, m.last, now)
	connectedBadge := renderConnectedBadge(connected)

	port := strings.TrimSpace(m.last.Port)
	if port == "" {
		port = "-"
	}

	updated := "-"
	lag := "-"
	if !m.last.UpdatedAt.IsZero() {
		updated = m.last.UpdatedAt.Format("15:04:05.000")
		d := now.Sub(m.last.UpdatedAt)
		if d < 0 {
			d = 0
		}
		lag = fmt.Sprintf("%d ms", d.Milliseconds())
	}

	trendW := w - 24
	if trendW < 12 {
		trendW = 12
	}
	trend := sparkline(m.history, trendW)
	if trend == "" {
		trend = "-"
	}
	minV, maxV := historyRange(m.history)

	qtyLine := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(qty)
	trendLine := lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Render("Trend: " + trend)
	rangeLine := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(fmt.Sprintf("Range: %.3f .. %.3f", minV, maxV))
	stableLine := "Stable: " + strings.ToUpper(stableText(m.last.Stable))

	panel := renderPanel("Live Reading", []string{
		qtyLine,
		trendLine,
		rangeLine,
		stableLine,
		"Updated: " + updated,
		"Connection: " + connectedBadge,
		"Lag: " + lag,
		"Source: " + elideMiddle(m.sourceLine, w-15),
		"Port: " + elideMiddle(port, w-13),
	}, w, "45")

	return lipgloss.NewStyle().Padding(0, 1).Render(panel)
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
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return clockMsg(t)
	})
}

func isConnected(status string, last Reading, now time.Time) bool {
	if strings.TrimSpace(last.Error) != "" {
		return false
	}
	if strings.ToLower(strings.TrimSpace(status)) != "ok" {
		return false
	}
	if last.UpdatedAt.IsZero() {
		return false
	}
	if now.Sub(last.UpdatedAt) > 3*time.Second {
		return false
	}
	if strings.TrimSpace(last.Port) == "" {
		return false
	}
	return true
}

func viewSize(w, h int) (int, int) {
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 28
	}

	width := w - 4
	if width > 110 {
		width = 110
	}
	if width < 58 {
		width = 58
	}
	return width, h
}

func renderPanel(title string, lines []string, width int, borderColor string) string {
	if width < 24 {
		width = 24
	}
	inner := width - 4

	normalized := make([]string, 0, len(lines)+2)
	for _, line := range lines {
		wrapped := wrapByWidth(line, inner)
		for _, part := range wrapped {
			normalized = append(normalized, truncateText(part, inner))
		}
	}
	if len(normalized) == 0 {
		normalized = []string{""}
	}

	titleStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render(title)
	body := strings.Join(normalized, "\n")
	content := titleStyled + "\n" + body

	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))
	return style.Render(content)
}

func renderConnectedBadge(connected bool) string {
	s := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if connected {
		return s.Foreground(lipgloss.Color("46")).Background(lipgloss.Color("22")).Render("CONNECTED")
	}
	return s.Foreground(lipgloss.Color("231")).Background(lipgloss.Color("160")).Render("DISCONNECTED")
}

func historyRange(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	minV, maxV := values[0], values[0]
	for _, v := range values[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	return minV, maxV
}

func sparkline(values []float64, width int) string {
	if width <= 0 || len(values) == 0 {
		return ""
	}
	blocks := []rune("▁▂▃▄▅▆▇█")
	if len(values) > width {
		values = values[len(values)-width:]
	}
	minV, maxV := historyRange(values)
	if math.Abs(maxV-minV) < 1e-9 {
		return strings.Repeat(string(blocks[0]), len(values))
	}

	var b strings.Builder
	for _, v := range values {
		norm := (v - minV) / (maxV - minV)
		idx := int(math.Round(norm * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteRune(blocks[idx])
	}
	return b.String()
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

func elideMiddle(text string, max int) string {
	runes := []rune(strings.TrimSpace(text))
	if max <= 0 {
		return ""
	}
	if len(runes) <= max {
		return string(runes)
	}
	if max <= 5 {
		return truncateText(string(runes), max)
	}
	keep := (max - 3) / 2
	left := string(runes[:keep])
	right := string(runes[len(runes)-keep:])
	return left + "..." + right
}
