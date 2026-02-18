package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bot/internal/app/commands"
	"bot/internal/erp"
	"bot/internal/telegram"
)

func (a *App) handleCallbackQuery(ctx context.Context, q telegram.CallbackQuery) error {
	data := strings.TrimSpace(q.Data)
	switch data {
	case commands.StockEntryCallbackMaterialIssue:
		return a.handleMaterialIssueCallback(ctx, q)
	case commands.StockEntryCallbackReceipt:
		if err := a.tg.AnswerCallbackQuery(ctx, q.ID, "Receipt tez orada qo'shiladi"); err != nil {
			return err
		}
		return nil
	default:
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "")
	}
}

func (a *App) handleMaterialIssueCallback(ctx context.Context, q telegram.CallbackQuery) error {
	if err := a.tg.AnswerCallbackQuery(ctx, q.ID, "Material Issue: qty olinmoqda..."); err != nil {
		return err
	}
	if q.Message == nil || q.Message.Chat.ID == 0 {
		return nil
	}

	chatID := q.Message.Chat.ID
	sel, ok := a.getSelection(chatID)
	if !ok {
		return a.tg.SendMessage(ctx, chatID, "Avval /batch orqali item va ombor tanlang.")
	}

	qty, unit, err := a.scale.WaitStablePositive(ctx, 35*time.Second, 220*time.Millisecond)
	if err != nil {
		return a.tg.SendMessage(ctx, chatID, "Scale qty olinmadi: "+err.Error())
	}

	draft, err := a.erp.CreateMaterialIssueDraft(ctx, erp.MaterialIssueDraftInput{
		ItemCode:  sel.ItemCode,
		Warehouse: sel.Warehouse,
		Qty:       qty,
	})
	if err != nil {
		return a.tg.SendMessage(ctx, chatID, "ERP draft yaratilmadi: "+err.Error())
	}

	msg := fmt.Sprintf(
		"Material Issue draft yaratildi:\nStock Entry: %s\nItem: %s\nOmbor: %s\nQTY: %.3f %s",
		draft.Name,
		draft.ItemCode,
		draft.Warehouse,
		draft.Qty,
		strings.ToLower(strings.TrimSpace(unit)),
	)
	return a.tg.SendMessage(ctx, chatID, msg)
}
