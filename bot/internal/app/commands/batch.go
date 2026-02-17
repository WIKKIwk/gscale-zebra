package commands

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/erp"
	"bot/internal/telegram"
)

func HandleBatch(ctx context.Context, deps Deps, msg telegram.Message) error {
	keyboard := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "Item tanlash", SwitchInlineQueryCurrentChat: ""},
			},
		},
	}

	err := deps.TG.SendMessageWithInlineKeyboard(ctx, msg.Chat.ID, "Item tanlang:", keyboard)
	if err == nil {
		return nil
	}

	if isInlineButtonUnsupported(err) {
		return deps.TG.SendMessage(ctx, msg.Chat.ID, "Inline mode o'chirilgan. BotFather'da /setinline ni yoqing, keyin /batch ni qayta bering.")
	}

	return err
}

func HandleBatchInlineQuery(ctx context.Context, deps Deps, q telegram.InlineQuery) error {
	items, err := deps.ERP.SearchItems(ctx, q.Query, 50)
	if err != nil {
		// Inline query spinner tugashi uchun empty result qaytaramiz.
		return deps.TG.AnswerInlineQuery(ctx, q.ID, []telegram.InlineQueryResultArticle{}, 1)
	}

	results := buildItemResults(items)
	if len(results) == 0 {
		results = []telegram.InlineQueryResultArticle{
			{
				Type:        "article",
				ID:          "empty",
				Title:       "Item topilmadi",
				Description: "ERPNext'da item topilmadi",
				InputMessageContent: telegram.InputTextMessageContent{
					MessageText: "Item topilmadi",
				},
			},
		}
	}

	return deps.TG.AnswerInlineQuery(ctx, q.ID, results, 1)
}

func buildItemResults(items []erp.Item) []telegram.InlineQueryResultArticle {
	results := make([]telegram.InlineQueryResultArticle, 0, len(items))
	for i, it := range items {
		code := strings.TrimSpace(it.ItemCode)
		if code == "" {
			continue
		}
		name := strings.TrimSpace(it.ItemName)
		if name == "" {
			name = code
		}

		results = append(results, telegram.InlineQueryResultArticle{
			Type:        "article",
			ID:          fmt.Sprintf("%d-%s", i+1, code),
			Title:       code,
			Description: name,
			InputMessageContent: telegram.InputTextMessageContent{
				MessageText: fmt.Sprintf("Item: %s\nNomi: %s", code, name),
			},
		})
	}
	return results
}

func isInlineButtonUnsupported(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "button_type_invalid") || strings.Contains(s, "inline mode")
}
