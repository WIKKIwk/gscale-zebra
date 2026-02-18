package app

import (
	"context"
	"errors"
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
	case commands.StockEntryCallbackBatchStop:
		return a.handleBatchStopCallback(ctx, q)
	default:
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "")
	}
}

func (a *App) handleMaterialIssueCallback(ctx context.Context, q telegram.CallbackQuery) error {
	if err := a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch boshlandi"); err != nil {
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

	stockPromptMessageID := q.Message.MessageID
	a.startBatchSession(ctx, chatID, func(batchCtx context.Context) {
		a.runMaterialIssueBatchLoop(batchCtx, chatID, sel, stockPromptMessageID)
	})
	return nil
}

func (a *App) handleBatchStopCallback(ctx context.Context, q telegram.CallbackQuery) error {
	if q.Message == nil || q.Message.Chat.ID == 0 {
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch to'xtatildi")
	}

	chatID := q.Message.Chat.ID
	stopped := a.stopBatchSession(chatID)
	if stopped {
		if err := a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch to'xtatildi"); err != nil {
			return err
		}
		a.deleteMessageBestEffort(ctx, chatID, q.Message.MessageID, "delete batch-stop message warning")
		return a.tg.SendMessage(ctx, chatID, "Batch to'xtatildi.")
	}

	return a.tg.AnswerCallbackQuery(ctx, q.ID, "Aktiv batch yo'q")
}

func (a *App) runMaterialIssueBatchLoop(ctx context.Context, chatID int64, sel SelectedContext, stockPromptMessageID int64) {
	deletedStockPrompt := false

	for {
		qty, unit, err := a.qtyReader.WaitStablePositive(ctx, 35*time.Second, 220*time.Millisecond)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				continue
			}
			a.sendMessageNoFail(ctx, chatID, "Scale qty olinmadi (fayl): "+err.Error())
			continue
		}

		draft, err := a.erp.CreateMaterialIssueDraft(ctx, erp.MaterialIssueDraftInput{
			ItemCode:  sel.ItemCode,
			Warehouse: sel.Warehouse,
			Qty:       qty,
		})
		if err != nil {
			a.sendMessageNoFail(ctx, chatID, "ERP draft yaratilmadi: "+err.Error())
			continue
		}

		msg := fmt.Sprintf(
			"Material Issue draft yaratildi:\nStock Entry: %s\nItem: %s\nOmbor: %s\nQTY: %.3f %s",
			draft.Name,
			draft.ItemCode,
			draft.Warehouse,
			draft.Qty,
			strings.ToLower(strings.TrimSpace(unit)),
		)
		if err := a.tg.SendMessageWithInlineKeyboard(ctx, chatID, msg, commands.BuildBatchStopKeyboard()); err != nil {
			a.log.Printf("send draft message warning: %v", err)
		}

		if !deletedStockPrompt && stockPromptMessageID > 0 {
			a.deleteMessageBestEffort(ctx, chatID, stockPromptMessageID, "delete stock-entry-prompt warning")
			deletedStockPrompt = true
		}

		if err := a.qtyReader.WaitForReset(ctx, 10*time.Minute, 220*time.Millisecond); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				continue
			}
			continue
		}
	}
}

func (a *App) sendMessageNoFail(ctx context.Context, chatID int64, text string) {
	if err := a.tg.SendMessage(ctx, chatID, text); err != nil {
		a.log.Printf("send message warning: %v", err)
	}
}
