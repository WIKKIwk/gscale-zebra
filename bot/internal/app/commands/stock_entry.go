package commands

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/telegram"
)

const stockEntryCallbackMaterialIssue = "stock:material_issue"
const stockEntryCallbackReceipt = "stock:receipt"

func ExtractSelectedWarehouse(text string) (string, string, bool) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 0 {
		return "", "", false
	}

	var itemCode string
	var warehouse string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)

		if strings.HasPrefix(lower, "item:") {
			itemCode = strings.TrimSpace(line[len("item:"):])
			continue
		}
		if strings.HasPrefix(lower, "ombor:") {
			warehouse = strings.TrimSpace(line[len("ombor:"):])
		}
	}

	if itemCode == "" || warehouse == "" {
		return "", "", false
	}
	return itemCode, warehouse, true
}

func HandleWarehouseSelected(ctx context.Context, deps Deps, chatID int64, itemCode, warehouse string) error {
	itemCode = strings.TrimSpace(itemCode)
	warehouse = strings.TrimSpace(warehouse)
	if itemCode == "" || warehouse == "" {
		return nil
	}

	keyboard := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "Material Issue", CallbackData: stockEntryCallbackMaterialIssue},
				{Text: "Receipt", CallbackData: stockEntryCallbackReceipt},
			},
		},
	}

	text := fmt.Sprintf("Item tanlandi: %s\nOmbor tanlandi: %s\nStock entry tanlang:", itemCode, warehouse)
	return deps.TG.SendMessageWithInlineKeyboard(ctx, chatID, text, keyboard)
}

func HandleCallbackQuery(ctx context.Context, deps Deps, q telegram.CallbackQuery) error {
	switch strings.TrimSpace(q.Data) {
	case stockEntryCallbackMaterialIssue:
		return deps.TG.AnswerCallbackQuery(ctx, q.ID, "Material Issue tanlandi")
	case stockEntryCallbackReceipt:
		return deps.TG.AnswerCallbackQuery(ctx, q.ID, "Receipt tanlandi")
	default:
		return deps.TG.AnswerCallbackQuery(ctx, q.ID, "")
	}
}
