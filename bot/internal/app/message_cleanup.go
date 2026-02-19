package app

import "context"

func (a *App) trackStartInfoMessage(ctx context.Context, chatID, messageID int64) {
	if messageID <= 0 {
		return
	}
	if prev := a.startInfoMsgByChat[chatID]; prev > 0 && prev != messageID {
		a.deleteMessageBestEffort(ctx, chatID, prev, "delete old start-info warning")
	}
	a.startInfoMsgByChat[chatID] = messageID
}

func (a *App) deleteTrackedStartInfoMessage(ctx context.Context, chatID int64) {
	messageID := a.startInfoMsgByChat[chatID]
	if messageID <= 0 {
		return
	}
	a.deleteMessageBestEffort(ctx, chatID, messageID, "delete start-info warning")
	delete(a.startInfoMsgByChat, chatID)
}

func (a *App) trackBatchPromptMessage(ctx context.Context, chatID, messageID int64) {
	if messageID <= 0 {
		return
	}
	if prev := a.batchPromptMsgByChat[chatID]; prev > 0 && prev != messageID {
		a.deleteMessageBestEffort(ctx, chatID, prev, "delete old batch-prompt warning")
	}
	a.batchPromptMsgByChat[chatID] = messageID
}

func (a *App) deleteTrackedBatchPromptMessage(ctx context.Context, chatID int64) {
	messageID := a.batchPromptMsgByChat[chatID]
	if messageID <= 0 {
		return
	}
	a.deleteMessageBestEffort(ctx, chatID, messageID, "delete batch-prompt warning")
	delete(a.batchPromptMsgByChat, chatID)
}

func (a *App) trackWarehousePromptMessage(ctx context.Context, chatID, messageID int64) {
	if messageID <= 0 {
		return
	}
	if prev := a.warehousePromptMsgByChat[chatID]; prev > 0 && prev != messageID {
		a.deleteMessageBestEffort(ctx, chatID, prev, "delete old warehouse-prompt warning")
	}
	a.warehousePromptMsgByChat[chatID] = messageID
}

func (a *App) deleteTrackedWarehousePromptMessage(ctx context.Context, chatID int64) {
	messageID := a.warehousePromptMsgByChat[chatID]
	if messageID <= 0 {
		return
	}
	a.deleteMessageBestEffort(ctx, chatID, messageID, "delete warehouse-prompt warning")
	delete(a.warehousePromptMsgByChat, chatID)
}

func (a *App) deleteMessageBestEffort(ctx context.Context, chatID, messageID int64, logPrefix string) {
	if messageID <= 0 {
		return
	}
	if err := a.tg.DeleteMessage(ctx, chatID, messageID); err != nil {
		a.logCleanup.Printf("%s: %v", logPrefix, err)
	}
}
