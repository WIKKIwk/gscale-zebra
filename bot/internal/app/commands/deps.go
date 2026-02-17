package commands

import (
	"context"

	"bot/internal/erp"
	"bot/internal/telegram"
)

type TelegramService interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendMessageWithInlineKeyboard(ctx context.Context, chatID int64, text string, keyboard *telegram.InlineKeyboardMarkup) error
	AnswerInlineQuery(ctx context.Context, inlineQueryID string, results []telegram.InlineQueryResultArticle, cacheSeconds int) error
}

type ERPService interface {
	CheckConnection(ctx context.Context) (string, error)
	SearchItems(ctx context.Context, query string, limit int) ([]erp.Item, error)
}

type Deps struct {
	TG  TelegramService
	ERP ERPService
}
