package commands

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/telegram"
)

const StockEntryCallbackMaterialIssue = "stock:material_issue"
const StockEntryCallbackReceipt = "stock:receipt"
const StockEntryCallbackBatchStop = "stock:batch_stop"

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
				{Text: "Material Issue", CallbackData: StockEntryCallbackMaterialIssue},
				{Text: "Receipt", CallbackData: StockEntryCallbackReceipt},
			},
		},
	}

	text := fmt.Sprintf("Item tanlandi: %s\nOmbor tanlandi: %s\nStock entry tanlang:", itemCode, warehouse)
	return deps.TG.SendMessageWithInlineKeyboard(ctx, chatID, text, keyboard)
}

func BuildBatchStopKeyboard() *telegram.InlineKeyboardMarkup {
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "Batch Stop", CallbackData: StockEntryCallbackBatchStop},
			},
		},
	}
}
