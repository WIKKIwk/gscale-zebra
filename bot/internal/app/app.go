package app

import (
	"context"
	"log"
	"strings"
	"time"

	"bot/internal/app/commands"
	"bot/internal/config"
	"bot/internal/erp"
	"bot/internal/telegram"
)

type App struct {
	cfg                config.Config
	tg                 *telegram.Client
	erp                *erp.Client
	log                *log.Logger
	startInfoMsgByChat map[int64]int64
}

func New(cfg config.Config, logger *log.Logger) *App {
	if logger == nil {
		logger = log.Default()
	}
	return &App{
		cfg:                cfg,
		tg:                 telegram.New(cfg.TelegramBotToken),
		erp:                erp.New(cfg.ERPURL, cfg.ERPAPIKey, cfg.ERPAPISecret),
		log:                logger,
		startInfoMsgByChat: make(map[int64]int64),
	}
}

func (a *App) deps() commands.Deps {
	return commands.Deps{TG: a.tg, ERP: a.erp}
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

			if upd.InlineQuery != nil {
				if err := commands.HandleInlineQuery(ctx, a.deps(), *upd.InlineQuery); err != nil {
					a.log.Printf("handleInlineQuery error: %v", err)
				}
				continue
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

	if itemCode, ok := commands.ExtractSelectedItemCode(text); ok {
		return commands.HandleItemSelected(ctx, a.deps(), msg.Chat.ID, itemCode)
	}

	cmd := normalizeCommand(text)
	if cmd == "" {
		return nil
	}

	a.maybeDeleteCommandMessage(ctx, msg, cmd)

	switch cmd {
	case "/start":
		messageID, err := commands.HandleStart(ctx, a.deps(), msg)
		if err != nil {
			return err
		}
		a.trackStartInfoMessage(ctx, msg.Chat.ID, messageID)
		return nil
	case "/batch":
		if err := commands.HandleBatch(ctx, a.deps(), msg); err != nil {
			return err
		}
		a.deleteTrackedStartInfoMessage(ctx, msg.Chat.ID)
		return nil
	default:
		return a.tg.SendMessage(ctx, msg.Chat.ID, "Qo'llanadigan buyruqlar: /start, /batch")
	}
}

func (a *App) trackStartInfoMessage(ctx context.Context, chatID, messageID int64) {
	if messageID <= 0 {
		return
	}

	if prev := a.startInfoMsgByChat[chatID]; prev > 0 && prev != messageID {
		if err := a.tg.DeleteMessage(ctx, chatID, prev); err != nil {
			a.log.Printf("delete old start-info warning: %v", err)
		}
	}
	a.startInfoMsgByChat[chatID] = messageID
}

func (a *App) deleteTrackedStartInfoMessage(ctx context.Context, chatID int64) {
	messageID := a.startInfoMsgByChat[chatID]
	if messageID <= 0 {
		return
	}
	if err := a.tg.DeleteMessage(ctx, chatID, messageID); err != nil {
		a.log.Printf("delete start-info warning: %v", err)
	}
	delete(a.startInfoMsgByChat, chatID)
}

func (a *App) maybeDeleteCommandMessage(ctx context.Context, msg telegram.Message, cmd string) {
	if !shouldDeleteUserCommand(cmd) {
		return
	}
	if msg.MessageID <= 0 {
		return
	}

	if err := a.tg.DeleteMessage(ctx, msg.Chat.ID, msg.MessageID); err != nil {
		a.log.Printf("deleteMessage warning: %v", err)
	}
}

func shouldDeleteUserCommand(cmd string) bool {
	switch cmd {
	case "/start", "/batch":
		return true
	default:
		return false
	}
}

func normalizeCommand(text string) string {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return ""
	}
	cmd := strings.ToLower(parts[0])
	if !strings.HasPrefix(cmd, "/") {
		return ""
	}
	if i := strings.Index(cmd, "@"); i > 0 {
		cmd = cmd[:i]
	}
	return cmd
}
