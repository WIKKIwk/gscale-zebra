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
	case commands.StockEntryCallbackBatchChangeItem:
		return a.handleBatchChangeItemCallback(ctx, q)
	case commands.StockEntryCallbackBatchStop:
		return a.handleBatchStopCallback(ctx, q)
	default:
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "")
	}
}

func (a *App) handleMaterialIssueCallback(ctx context.Context, q telegram.CallbackQuery) error {
	if q.Message == nil || q.Message.Chat.ID == 0 {
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch boshlandi")
	}

	chatID := q.Message.Chat.ID
	sel, ok := a.getSelection(chatID)
	if !ok {
		if err := a.tg.AnswerCallbackQuery(ctx, q.ID, "Avval item va ombor tanlang"); err != nil {
			return err
		}
		return a.tg.SendMessage(ctx, chatID, "Avval /batch orqali item va ombor tanlang.")
	}

	a.clearBatchChangePending(chatID)
	_ = a.startMaterialIssueBatch(ctx, chatID, sel, q.Message.MessageID, "Scale qty kutilmoqda...")
	return a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch boshlandi")
}

func (a *App) handleBatchChangeItemCallback(ctx context.Context, q telegram.CallbackQuery) error {
	if q.Message == nil || q.Message.Chat.ID == 0 {
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "Item almashtirish")
	}

	chatID := q.Message.Chat.ID
	_ = a.stopBatchSession(chatID)
	a.setBatchChangePending(chatID, q.Message.MessageID)

	pausedText := formatPausedStatus(q.Message.Text)
	if err := a.tg.EditMessageText(ctx, chatID, q.Message.MessageID, pausedText, commands.BuildBatchControlKeyboard()); err != nil && !isMessageNotModifiedError(err) {
		a.log.Printf("edit paused status warning: %v", err)
	}

	promptID, err := commands.HandleBatch(ctx, a.deps(), telegram.Message{Chat: telegram.Chat{ID: chatID}})
	if err != nil {
		if cbErr := a.tg.AnswerCallbackQuery(ctx, q.ID, "Pause qilindi, lekin item tanlashda xato"); cbErr != nil {
			return cbErr
		}
		return err
	}
	a.trackBatchPromptMessage(ctx, chatID, promptID)
	a.deleteTrackedWarehousePromptMessage(ctx, chatID)

	return a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch pause. Yangi item tanlang")
}

func (a *App) handleBatchStopCallback(ctx context.Context, q telegram.CallbackQuery) error {
	if q.Message == nil || q.Message.Chat.ID == 0 {
		return a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch to'xtatildi")
	}

	chatID := q.Message.Chat.ID
	a.clearBatchChangePending(chatID)
	stopped := a.stopBatchSession(chatID)
	if stopped {
		if err := a.tg.AnswerCallbackQuery(ctx, q.ID, "Batch to'xtatildi"); err != nil {
			return err
		}
		stoppedText := formatStoppedStatus(q.Message.Text)
		if err := a.tg.EditMessageText(ctx, chatID, q.Message.MessageID, stoppedText, nil); err != nil && !isMessageNotModifiedError(err) {
			a.log.Printf("edit stopped status warning: %v", err)
		}
		return nil
	}

	return a.tg.AnswerCallbackQuery(ctx, q.ID, "Aktiv batch yo'q")
}

func (a *App) startMaterialIssueBatch(ctx context.Context, chatID int64, sel SelectedContext, statusMessageID int64, note string) int64 {
	initial := formatBatchStatusText(sel, 0, "", 0, "", "", strings.TrimSpace(note))
	statusMessageID = a.upsertBatchStatusMessage(ctx, chatID, statusMessageID, initial)

	a.startBatchSession(ctx, chatID, func(batchCtx context.Context) {
		a.runMaterialIssueBatchLoop(batchCtx, chatID, sel, statusMessageID)
	})
	return statusMessageID
}

func (a *App) runMaterialIssueBatchLoop(ctx context.Context, chatID int64, sel SelectedContext, statusMessageID int64) {
	draftCount := 0
	lastEPC := ""

	for {
		reading, err := a.qtyReader.WaitStablePositiveReading(ctx, 35*time.Second, 220*time.Millisecond)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				continue
			}
			statusMessageID = a.upsertBatchStatusMessage(ctx, chatID, statusMessageID, formatBatchStatusText(sel, draftCount, "", 0, "", "", "Scale xato: "+err.Error()))
			continue
		}

		epc := ""
		epcNote := ""
		epc, err = a.qtyReader.WaitEPCForReading(ctx, 6*time.Second, 140*time.Millisecond, reading.UpdatedAt, lastEPC)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			epcNote = err.Error()
		}

		draft, err := a.erp.CreateMaterialIssueDraft(ctx, erp.MaterialIssueDraftInput{
			ItemCode:  sel.ItemCode,
			Warehouse: sel.Warehouse,
			Qty:       reading.Qty,
			Barcode:   epc,
		})
		if err != nil {
			statusMessageID = a.upsertBatchStatusMessage(ctx, chatID, statusMessageID, formatBatchStatusText(sel, draftCount, "", 0, "", epc, "ERP xato: "+err.Error()))
			continue
		}

		draftCount++
		if epc != "" {
			lastEPC = epc
		}

		note := "Batch davom etmoqda"
		if strings.TrimSpace(epcNote) != "" {
			note = note + " | EPC ogohlantirish: " + strings.TrimSpace(epcNote)
		}
		statusMessageID = a.upsertBatchStatusMessage(ctx, chatID, statusMessageID, formatBatchStatusText(sel, draftCount, draft.Name, draft.Qty, reading.Unit, epc, note))

		for {
			err := a.qtyReader.WaitForNextCycle(ctx, 10*time.Minute, 220*time.Millisecond, draft.Qty)
			if err == nil {
				break
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			statusMessageID = a.upsertBatchStatusMessage(ctx, chatID, statusMessageID, formatBatchStatusText(sel, draftCount, draft.Name, draft.Qty, reading.Unit, epc, "Keyingi mahsulotni qo'ying (yoki 0 kg)"))
		}
	}
}

func (a *App) upsertBatchStatusMessage(ctx context.Context, chatID, messageID int64, text string) int64 {
	if messageID > 0 {
		err := a.tg.EditMessageText(ctx, chatID, messageID, text, commands.BuildBatchControlKeyboard())
		if err == nil || isMessageNotModifiedError(err) {
			return messageID
		}
		a.log.Printf("edit batch status warning: %v", err)
	}

	newID, err := a.tg.SendMessageWithInlineKeyboardAndReturnID(ctx, chatID, text, commands.BuildBatchControlKeyboard())
	if err != nil {
		a.log.Printf("send batch status warning: %v", err)
		return messageID
	}
	return newID
}

func formatBatchStatusText(sel SelectedContext, draftCount int, draftName string, qty float64, unit, epc, note string) string {
	lines := []string{
		"Batch ishlayapti",
		fmt.Sprintf("Item: %s", strings.TrimSpace(sel.ItemCode)),
		fmt.Sprintf("Ombor: %s", strings.TrimSpace(sel.Warehouse)),
		fmt.Sprintf("Draftlar: %d", draftCount),
	}

	if draftCount > 0 {
		u := strings.ToLower(strings.TrimSpace(unit))
		if u == "" {
			u = "kg"
		}
		lines = append(lines, fmt.Sprintf("Oxirgi draft: %s", strings.TrimSpace(draftName)))
		lines = append(lines, fmt.Sprintf("Oxirgi QTY: %.3f %s", qty, u))
		epc = strings.ToUpper(strings.TrimSpace(epc))
		if epc == "" {
			epc = "-"
		}
		lines = append(lines, "Oxirgi EPC: "+epc)
	}

	note = strings.TrimSpace(note)
	if note != "" {
		lines = append(lines, "Holat: "+note)
	}

	return strings.Join(lines, "\n")
}

func formatPausedStatus(current string) string {
	base := strings.TrimSpace(current)
	if base == "" {
		return "Batch pausa qilindi. Yangi item tanlang."
	}
	if strings.Contains(strings.ToUpper(base), "PAUSE") {
		return base
	}
	return base + "\n\nStatus: PAUSE (yangi item tanlanmoqda)"
}

func formatStoppedStatus(current string) string {
	base := strings.TrimSpace(current)
	if base == "" {
		base = "Batch"
	}
	if strings.Contains(strings.ToUpper(base), "TO'XTATILDI") {
		return base
	}
	return base + "\n\nStatus: TO'XTATILDI"
}

func isMessageNotModifiedError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "message is not modified")
}
