package app

func (a *App) setBatchChangePending(chatID, statusMessageID int64) {
	if chatID == 0 || statusMessageID <= 0 {
		return
	}
	a.batchChangeMsgByChat[chatID] = statusMessageID
}

func (a *App) consumeBatchChangePending(chatID int64) (int64, bool) {
	if chatID == 0 {
		return 0, false
	}
	id, ok := a.batchChangeMsgByChat[chatID]
	if !ok || id <= 0 {
		return 0, false
	}
	delete(a.batchChangeMsgByChat, chatID)
	return id, true
}

func (a *App) clearBatchChangePending(chatID int64) {
	if chatID == 0 {
		return
	}
	delete(a.batchChangeMsgByChat, chatID)
}
