package app

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (a *App) handleEPCCommand(ctx context.Context, chatID int64) error {
	epcs := a.epcHistory.Snapshot()
	if len(epcs) == 0 {
		return a.tg.SendMessage(ctx, chatID, "Hozircha draft uchun EPC yozilmagan.")
	}

	filename, content := buildEPCDocument(epcs, time.Now().UTC())
	caption := fmt.Sprintf("Draft EPC ro'yxati (session boshidan): %d ta", len(epcs))
	return a.tg.SendDocument(ctx, chatID, filename, content, caption)
}

func buildEPCDocument(epcs []string, now time.Time) (string, []byte) {
	ts := now.UTC().Format("20060102-150405")
	name := "epc-history-" + ts + ".txt"
	body := strings.Join(epcs, "\n")
	if body != "" {
		body += "\n"
	}
	return name, []byte(body)
}
