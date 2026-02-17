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
	lastGap    time.Duration
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

		if !m.last.UpdatedAt.IsZero() && !upd.UpdatedAt.IsZero() {
			gap := upd.UpdatedAt.Sub(m.last.UpdatedAt)
			if gap > 0 && gap < 10*time.Second {
				m.lastGap = gap
			}
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
	viewWidth, viewHeight := viewSize(m.width, m.height)
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
	statusKind := classifyStatus(status)

	port := "-"
	if strings.TrimSpace(m.last.Port) != "" {
		port = strings.TrimSpace(m.last.Port)
	}

	stable := strings.ToUpper(stableText(m.last.Stable))
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

	rate := "-"
	if m.lastGap > 0 {
		rate = fmt.Sprintf("%.1f Hz", 1.0/m.lastGap.Seconds())
	}

	minV, maxV := historyRange(m.history)
	trend := sparkline(m.history, 28)
	if trend == "" {
		trend = "-"
	}

	leftW, rightW := splitWidths(viewWidth)

	titleLine := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("GSCALE ZEBRA MONITOR")
	helpLine := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Q: chiqish  |  Ctrl+C: chiqish")
	header := renderPanel("Dashboard", []string{titleLine, helpLine}, viewWidth, "63", 2)

	qtyLine := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(qty)
	trendLine := lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Render("Trend: " + trend)
	rangeLine := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(fmt.Sprintf("Range: %.3f .. %.3f", minV, maxV))
	readingPanel := renderPanel("Live Reading", []string{
		qtyLine,
		trendLine,
		rangeLine,
		"Stable: " + stable,
		"Updated: " + updated,
	}, leftW, "45", 3)

	sourceShort := elideMiddle(m.sourceLine, rightW-15)
	statusBadge := renderBadge(statusKind)
	metaPanel := renderPanel("Connection", []string{
		"Source: " + sourceShort,
		"Port: " + elideMiddle(port, rightW-12),
		"Rate: " + rate,
		"Lag: " + lag,
		"Status: " + statusBadge,
	}, rightW, "69", 2)

	top := lipgloss.JoinHorizontal(lipgloss.Top, readingPanel, " ", metaPanel)

	statusDetail := elideMiddle(status, viewWidth-18)
	statusPanel := renderPanel("System", []string{
		"State: " + renderBadge(statusKind),
		"Detail: " + statusDetail,
	}, viewWidth, "99", 2)

	rawLines := normalizeRawLines(m.last.Raw)
	rawLimit := rawLineLimit(viewHeight, header, top, statusPanel)
	if len(rawLines) > rawLimit {
		rawLines = rawLines[len(rawLines)-rawLimit:]
	}
	if len(rawLines) == 0 {
		rawLines = []string{"(empty)"}
	}
	rawPanel := renderPanel("Raw Stream", rawLines, viewWidth, "240", 1)

	layout := strings.Join([]string{header, "", top, "", statusPanel, "", rawPanel}, "\n")
	return lipgloss.NewStyle().Padding(0, 1).Render(layout)
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
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
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
		w = 100
	}
	if h <= 0 {
		h = 32
	}

	width := w - 4
	if width > 118 {
		width = 118
	}
	if width < 72 {
		width = 72
	}

	height := h - 2
	if height < 20 {
		height = 20
	}
	return width, height
}

func splitWidths(total int) (int, int) {
	left := int(math.Round(float64(total) * 0.56))
	if left < 38 {
		left = 38
	}
	right := total - left - 1
	if right < 28 {
		right = 28
		left = total - right - 1
	}
	if left < 24 {
		left = 24
	}
	return left, right
}

func renderPanel(title string, lines []string, width int, borderColor string, titleColor int) string {
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

	titleStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(fmt.Sprintf("%d", titleColor))).Render(title)
	body := strings.Join(normalized, "\n")
	content := titleStyled + "\n" + body

	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))
	return style.Render(content)
}

func renderBadge(kind string) string {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	s := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	switch kind {
	case "OK":
		return s.Foreground(lipgloss.Color("46")).Background(lipgloss.Color("22")).Render("OK")
	case "WARN":
		return s.Foreground(lipgloss.Color("228")).Background(lipgloss.Color("94")).Render("WARN")
	default:
		return s.Foreground(lipgloss.Color("231")).Background(lipgloss.Color("160")).Render("ERROR")
	}
}

func normalizeRawLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r", "\n")
	raw = strings.ReplaceAll(raw, "\t", " ")
	rows := strings.Split(raw, "\n")
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		out = append(out, row)
	}
	return out
}

func rawLineLimit(totalHeight int, header, top, status string) int {
	used := lineCount(header) + lineCount(top) + lineCount(status) + 6
	free := totalHeight - used
	if free < 2 {
		return 2
	}
	if free > 8 {
		return 8
	}
	return free
}

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
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
