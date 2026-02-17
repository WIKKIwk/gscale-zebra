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

type zebraMsg struct {
	status ZebraStatus
}

type quitMsg struct{}

type clockMsg time.Time

type tuiModel struct {
	ctx            context.Context
	updates        <-chan Reading
	zebraUpdates   <-chan ZebraStatus
	sourceLine     string
	zebraPreferred string
	message        string
	info           string
	last           Reading
	zebra          ZebraStatus
	width          int
	height         int
	now            time.Time
	history        []float64
}

func runTUI(ctx context.Context, updates <-chan Reading, zebraUpdates <-chan ZebraStatus, sourceLine string, zebraPreferred string, serialErr error) error {
	m := tuiModel{
		ctx:            ctx,
		updates:        updates,
		zebraUpdates:   zebraUpdates,
		sourceLine:     sourceLine,
		zebraPreferred: zebraPreferred,
		last:           Reading{Unit: "kg"},
		message:        "scale oqimi kutilmoqda",
		info:           "keys: q quit | e encode test epc | r read epc",
		now:            time.Now(),
		history:        make([]float64, 0, 256),
		zebra: ZebraStatus{
			Connected: false,
			Verify:    "-",
			ReadLine1: "-",
			ReadLine2: "-",
			UpdatedAt: time.Now(),
		},
	}
	if serialErr != nil {
		m.message = serialErr.Error()
	}
	if zebraUpdates == nil {
		m.zebra.Error = "disabled"
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m tuiModel) Init() tea.Cmd {
	cmds := []tea.Cmd{waitForReadingCmd(m.ctx, m.updates), clockTickCmd()}
	if m.zebraUpdates != nil {
		cmds = append(cmds, waitForZebraCmd(m.ctx, m.zebraUpdates))
	}
	return tea.Batch(cmds...)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		s := strings.ToLower(strings.TrimSpace(msg.String()))
		switch s {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e":
			if m.zebraUpdates == nil {
				m.info = "zebra monitor o'chirilgan (--no-zebra)"
				return m, nil
			}
			m.info = "epc encode yuborilmoqda (1 tag)..."
			return m, runEncodeEPCCmd(m.zebraPreferred)
		case "r":
			if m.zebraUpdates == nil {
				m.info = "zebra monitor o'chirilgan (--no-zebra)"
				return m, nil
			}
			m.info = "rfid read yuborilmoqda..."
			return m, runRFIDReadCmd(m.zebraPreferred)
		default:
			return m, nil
		}
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
	case zebraMsg:
		st := msg.status
		if st.UpdatedAt.IsZero() {
			st.UpdatedAt = time.Now()
		}
		m.zebra = st
		if st.Action != "" {
			m.info = zebraActionSummary(st)
		}
		if st.Error != "" && st.Action != "" {
			m.info = zebraActionSummary(st)
		}
		if m.zebraUpdates != nil {
			return m, waitForZebraCmd(m.ctx, m.zebraUpdates)
		}
		return m, nil
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

	scaleConnected := isConnected(status, m.last, now)
	scaleBadge := renderConnectedBadge(scaleConnected)

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

	trendW := (w / 2) - 18
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

	scaleLines := []string{
		"Connection: " + scaleBadge,
		qtyLine,
		trendLine,
		rangeLine,
		stableLine,
		"Updated: " + updated,
		"Lag: " + lag,
		"Source: " + elideMiddle(m.sourceLine, 42),
		"Port: " + elideMiddle(port, 42),
	}

	zebraConnected := m.zebra.Connected && strings.TrimSpace(m.zebra.Error) == ""
	zebraBadge := renderConnectedBadge(zebraConnected)
	if strings.EqualFold(strings.TrimSpace(m.zebra.Error), "disabled") {
		zebraBadge = renderDisabledBadge()
	}

	zebraName := strings.TrimSpace(m.zebra.Name)
	if zebraName == "" {
		zebraName = "-"
	}
	zebraDevice := strings.TrimSpace(m.zebra.DevicePath)
	if zebraDevice == "" {
		zebraDevice = "-"
	}
	deviceState := safeText("-", m.zebra.DeviceState)
	mediaState := safeText("-", m.zebra.MediaState)
	read1 := safeText("-", m.zebra.ReadLine1)
	read2 := safeText("-", m.zebra.ReadLine2)
	verify := safeText("-", m.zebra.Verify)
	lastEPC := safeText("-", m.zebra.LastEPC)
	zebraUpdated := "-"
	if !m.zebra.UpdatedAt.IsZero() {
		zebraUpdated = m.zebra.UpdatedAt.Format("15:04:05.000")
	}
	zebraErr := safeText("-", m.zebra.Error)

	zebraLines := []string{
		"Connection: " + zebraBadge,
		"Printer: " + elideMiddle(zebraName, 40),
		"Device: " + elideMiddle(zebraDevice, 40),
		"Device state: " + strings.ToUpper(deviceState),
		"Media state: " + strings.ToUpper(mediaState),
		"Read line1: " + elideMiddle(read1, 40),
		"Read line2: " + elideMiddle(read2, 40),
		"Verify: " + strings.ToUpper(verify),
		"Last EPC: " + elideMiddle(lastEPC, 40),
		"Updated: " + zebraUpdated,
		"Error: " + elideMiddle(zebraErr, 40),
	}

	leftW := w
	rightW := w
	if w >= 112 {
		leftW = (w - 2) / 2
		rightW = w - leftW - 1
	}

	scalePanel := renderPanel("Live Reading", scaleLines, leftW, "45")
	zebraPanel := renderPanel("Zebra RFID", zebraLines, rightW, "63")

	body := scalePanel
	if w >= 112 {
		body = lipgloss.JoinHorizontal(lipgloss.Top, scalePanel, " ", zebraPanel)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, scalePanel, zebraPanel)
	}

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(elideMiddle(m.info, w-2))
	return lipgloss.NewStyle().Padding(0, 1).Render(body + "\n" + footer)
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

func waitForZebraCmd(ctx context.Context, updates <-chan ZebraStatus) tea.Cmd {
	if updates == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return quitMsg{}
		case upd, ok := <-updates:
			if !ok {
				return quitMsg{}
			}
			return zebraMsg{status: upd}
		}
	}
}

func runEncodeEPCCmd(preferredDevice string) tea.Cmd {
	return func() tea.Msg {
		epc := generateTestEPC(time.Now())
		st := runZebraEncodeAndRead(preferredDevice, epc, 1400*time.Millisecond)
		st.UpdatedAt = time.Now()
		return zebraMsg{status: st}
	}
}

func runRFIDReadCmd(preferredDevice string) tea.Cmd {
	return func() tea.Msg {
		st := runZebraRead(preferredDevice, 1400*time.Millisecond)
		st.UpdatedAt = time.Now()
		return zebraMsg{status: st}
	}
}

func zebraActionSummary(st ZebraStatus) string {
	a := strings.ToUpper(strings.TrimSpace(st.Action))
	if a == "" {
		a = "MONITOR"
	}
	if strings.TrimSpace(st.Error) != "" {
		return fmt.Sprintf("zebra %s xato: %s", strings.ToLower(a), st.Error)
	}
	if a == "ENCODE" {
		return fmt.Sprintf("zebra encode: epc=%s verify=%s line1=%s", safeText("-", st.LastEPC), safeText("UNKNOWN", st.Verify), safeText("-", st.ReadLine1))
	}
	if a == "READ" {
		return fmt.Sprintf("zebra read: verify=%s line1=%s", safeText("UNKNOWN", st.Verify), safeText("-", st.ReadLine1))
	}
	return fmt.Sprintf("zebra %s: ok", strings.ToLower(a))
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
		w = 110
	}
	if h <= 0 {
		h = 30
	}

	width := w - 4
	if width > 132 {
		width = 132
	}
	if width < 64 {
		width = 64
	}
	return width, h
}

func renderPanel(title string, lines []string, width int, borderColor string) string {
	if width < 24 {
		width = 24
	}
	inner := width - 4

	titleStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render(title)
	rows := make([]string, 0, len(lines)+1)
	rows = append(rows, titleStyled)
	lineStyle := lipgloss.NewStyle().Width(inner).MaxWidth(inner)
	for _, line := range lines {
		rows = append(rows, lineStyle.Render(line))
	}
	body := strings.Join(rows, "\n")

	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))
	return style.Render(body)
}

func renderConnectedBadge(connected bool) string {
	s := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if connected {
		return s.Foreground(lipgloss.Color("46")).Background(lipgloss.Color("22")).Render("CONNECTED")
	}
	return s.Foreground(lipgloss.Color("231")).Background(lipgloss.Color("160")).Render("DISCONNECTED")
}

func renderDisabledBadge() string {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("238")).
		Render("DISABLED")
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

func elideMiddle(text string, max int) string {
	runes := []rune(strings.TrimSpace(text))
	if max <= 0 {
		return ""
	}
	if len(runes) <= max {
		return string(runes)
	}
	if max <= 5 {
		return string(runes[:max])
	}
	keep := (max - 3) / 2
	left := string(runes[:keep])
	right := string(runes[len(runes)-keep:])
	return left + "..." + right
}
