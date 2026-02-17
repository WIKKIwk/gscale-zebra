package main

import (
	"context"
	corepkg "core"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	autoDetector   *corepkg.StableEPCDetector
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
		info:           "ready",
		now:            time.Now(),
		autoDetector:   corepkg.NewStableEPCDetector(corepkg.DefaultStableEPCConfig()),
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
			m.info = "encode+print yuborildi"
			return m, runEncodeEPCCmd(m.zebraPreferred)
		case "r":
			if m.zebraUpdates == nil {
				m.info = "zebra monitor o'chirilgan (--no-zebra)"
				return m, nil
			}
			m.info = "rfid read yuborildi"
			return m, runRFIDReadCmd(m.zebraPreferred)
		default:
			return m, nil
		}
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

		cmd := waitForReadingCmd(m.ctx, m.updates)
		if m.zebraUpdates != nil && m.autoDetector != nil {
			if epc, ok := m.autoDetector.Observe(upd.Weight, upd.UpdatedAt); ok {
				m.info = fmt.Sprintf("auto encode queued: epc=%s", epc)
				cmd = tea.Batch(cmd, runEncodeEPCCmdWithEPC(m.zebraPreferred, epc))
			}
		}
		return m, cmd
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
	scaleState := stateText(scaleConnected)

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

	panelW := w

	zebraDisabled := strings.EqualFold(strings.TrimSpace(m.zebra.Error), "disabled")
	zebraConnected := m.zebra.Connected && strings.TrimSpace(m.zebra.Error) == "" && !zebraDisabled
	zebraState := "DOWN"
	if zebraDisabled {
		zebraState = "DISABLED"
	} else if zebraConnected {
		zebraState = "UP"
	}

	zebraName := strings.TrimSpace(m.zebra.Name)
	if zebraName == "" {
		zebraName = "-"
	}
	zebraDevice := strings.TrimSpace(m.zebra.DevicePath)
	if zebraDevice == "" {
		zebraDevice = "-"
	}
	deviceState := strings.ToUpper(safeText("-", m.zebra.DeviceState))
	mediaState := strings.ToUpper(safeText("-", m.zebra.MediaState))
	read1 := safeText("-", m.zebra.ReadLine1)
	verify := strings.ToUpper(safeText("-", m.zebra.Verify))
	lastEPC := safeText("-", m.zebra.LastEPC)
	zebraUpdated := "-"
	if !m.zebra.UpdatedAt.IsZero() {
		zebraUpdated = m.zebra.UpdatedAt.Format("15:04:05.000")
	}
	zebraErr := safeText("-", m.zebra.Error)

	scaleLines := []string{
		kv("STATUS", scaleState),
		kv("QTY", qty),
		kv("STABLE", strings.ToUpper(stableText(m.last.Stable))),
		kv("UPDATED", updated),
		kv("LAG", lag),
		kv("SOURCE", elideMiddle(m.sourceLine, maxInt(20, panelW-16))),
		kv("PORT", elideMiddle(port, maxInt(20, panelW-16))),
	}

	zebraLines := []string{
		kv("STATUS", zebraState),
		kv("PRINTER", elideMiddle(zebraName, maxInt(18, panelW-16))),
		kv("DEVICE", elideMiddle(zebraDevice, maxInt(18, panelW-16))),
		kv("DEVICE ST", deviceState),
		kv("MEDIA ST", mediaState),
		kv("VERIFY", verify),
		kv("LAST EPC", elideMiddle(lastEPC, maxInt(18, panelW-16))),
		kv("READ", elideMiddle(read1, maxInt(18, panelW-16))),
		kv("UPDATED", zebraUpdated),
		kv("ERROR", elideMiddle(zebraErr, maxInt(18, panelW-16))),
	}

	header := renderHeader(w, now, scaleState, zebraState)
	panel := renderUnifiedPanel("GSCALE-ZEBRA MONITOR", "SCALE", scaleLines, "ZEBRA", zebraLines, panelW)
	footer := renderFooter(w, m.info)
	return header + "\n" + panel + "\n" + footer
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
	epc := generateTestEPC(time.Now())
	return runEncodeEPCCmdWithEPC(preferredDevice, epc)
}

func runEncodeEPCCmdWithEPC(preferredDevice, epc string) tea.Cmd {
	return func() tea.Msg {
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
	return tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg {
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

	width := w - 2
	if width > 136 {
		width = 136
	}
	if width < 68 {
		width = 68
	}
	return width, h
}

func panelWidths(total int) (int, int) {
	left := total
	right := total
	if total >= 110 {
		left = (total - 1) / 2
		right = total - left - 1
	}
	return left, right
}

func renderHeader(width int, now time.Time, scaleState, zebraState string) string {
	text := fmt.Sprintf("GSCALE-ZEBRA CONSOLE | %s | SCALE=%s | ZEBRA=%s", now.Format("2006-01-02 15:04:05"), scaleState, zebraState)
	return fitLineRaw(text, width)
}

func renderFooter(width int, info string) string {
	left := "keys: [q] quit [e] encode+print [r] read"
	text := left + " | " + strings.TrimSpace(info)
	if strings.TrimSpace(info) == "" {
		text = left
	}
	return fitLineRaw(text, width)
}

func renderUnifiedPanel(title, scaleTitle string, scaleLines []string, zebraTitle string, zebraLines []string, width int) string {
	if width < 68 {
		width = 68
	}
	inner := width - 2
	rows := make([]string, 0, len(scaleLines)+len(zebraLines)+6)
	rows = append(rows, "┌"+centerTitle(title, inner)+"┐")
	rows = append(rows, "│"+fitPanelLine("["+strings.ToUpper(strings.TrimSpace(scaleTitle))+"]", inner)+"│")
	for _, line := range scaleLines {
		rows = append(rows, "│"+fitPanelLine(line, inner)+"│")
	}
	rows = append(rows, "├"+strings.Repeat("─", inner)+"┤")
	rows = append(rows, "│"+fitPanelLine("["+strings.ToUpper(strings.TrimSpace(zebraTitle))+"]", inner)+"│")
	for _, line := range zebraLines {
		rows = append(rows, "│"+fitPanelLine(line, inner)+"│")
	}
	rows = append(rows, "└"+strings.Repeat("─", inner)+"┘")
	return strings.Join(rows, "\n")
}

func renderUnixPanel(title string, lines []string, width int) string {
	if width < 32 {
		width = 32
	}
	inner := width - 2
	rows := make([]string, 0, len(lines)+2)
	rows = append(rows, "┌"+centerTitle(title, inner)+"┐")
	for _, line := range lines {
		rows = append(rows, "│"+fitPanelLine(line, inner)+"│")
	}
	rows = append(rows, "└"+strings.Repeat("─", inner)+"┘")
	return strings.Join(rows, "\n")
}

func joinHorizontalPanels(left, right string, leftW, rightW int) string {
	lrows := strings.Split(left, "\n")
	rrows := strings.Split(right, "\n")
	n := len(lrows)
	if len(rrows) > n {
		n = len(rrows)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		l := strings.Repeat(" ", leftW)
		if i < len(lrows) {
			l = fitLineRaw(lrows[i], leftW)
		}
		r := strings.Repeat(" ", rightW)
		if i < len(rrows) {
			r = fitLineRaw(rrows[i], rightW)
		}
		out = append(out, l+" "+r)
	}
	return strings.Join(out, "\n")
}

func kv(label, value string) string {
	label = strings.ToUpper(strings.TrimSpace(label))
	value = strings.TrimSpace(value)
	if value == "" {
		value = "-"
	}
	return fmt.Sprintf("%-10s : %s", label, value)
}

func stateText(connected bool) string {
	if connected {
		return "UP"
	}
	return "DOWN"
}

func centerTitle(title string, width int) string {
	t := " " + strings.ToUpper(strings.TrimSpace(title)) + " "
	if width <= 0 {
		return ""
	}
	if runeLen(t) > width {
		return truncateRunes(t, width)
	}
	left := (width - runeLen(t)) / 2
	right := width - runeLen(t) - left
	return strings.Repeat("─", left) + t + strings.Repeat("─", right)
}

func fitPanelLine(text string, width int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.TrimSpace(text)
	if width <= 0 {
		return ""
	}
	if runeLen(text) > width {
		text = elideMiddle(text, width)
	}
	return padRight(text, width)
}

func fitLineRaw(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if runeLen(text) > width {
		text = truncateRunes(text, width)
	}
	return padRightRaw(text, width)
}

func padRight(text string, width int) string {
	if runeLen(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-runeLen(text))
}

func padRightRaw(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if runeLen(text) > width {
		return truncateRunes(text, width)
	}
	if runeLen(text) < width {
		text += strings.Repeat(" ", width-runeLen(text))
	}
	return text
}

func runeLen(text string) int {
	return len([]rune(text))
}

func truncateRunes(text string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(text)
	if len(r) <= max {
		return text
	}
	return string(r[:max])
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
