package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"bot/internal/config"
	"bot/internal/erp"
	"bot/internal/telegram"
)

type App struct {
	cfg config.Config
	tg  *telegram.Client
	erp *erp.Client
	log *log.Logger
}

func New(cfg config.Config, logger *log.Logger) *App {
	if logger == nil {
		logger = log.Default()
	}
	return &App{
		cfg: cfg,
		tg:  telegram.New(cfg.TelegramBotToken),
		erp: erp.New(cfg.ERPURL, cfg.ERPAPIKey, cfg.ERPAPISecret),
		log: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Printf("bot started, ERP=%s", a.cfg.ERPURL)
	var offset int64 = 0

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		updates, err := a.tg.GetUpdates(ctx, offset, 55)
		if err != nil {
			a.log.Printf("getUpdates error: %v", err)
			time.Sleep(1200 * time.Millisecond)
			continue
		}

		for _, upd := range updates {
			if upd.UpdateID >= offset {
				offset = upd.UpdateID + 1
			}
			if upd.Message == nil {
				continue
			}
			if err := a.handleMessage(ctx, *upd.Message); err != nil {
				a.log.Printf("handleMessage error: %v", err)
			}
		}
	}
}

func (a *App) handleMessage(ctx context.Context, msg telegram.Message) error {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}

	if strings.HasPrefix(strings.ToLower(text), "/start") {
		user, err := a.erp.CheckConnection(ctx)
		if err != nil {
			return a.tg.SendMessage(ctx, msg.Chat.ID, "ERPNext ulanishi xato: "+err.Error())
		}
		return a.tg.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("ERPNext ga ulandi. User: %s", user))
	}

	return a.tg.SendMessage(ctx, msg.Chat.ID, "Hozircha faqat /start buyrug'i qo'llanadi.")
}
