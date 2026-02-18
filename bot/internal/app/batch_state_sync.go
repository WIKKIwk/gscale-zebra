package app

func (a *App) setBatchState(active bool, chatID int64) {
	if a == nil || a.batchState == nil {
		return
	}
	if err := a.batchState.Set(active, chatID); err != nil {
		a.log.Printf("batch state write error: %v", err)
	}
}

func (a *App) syncBatchStateFromSessions(chatHint int64) {
	a.batchMu.Lock()
	active := len(a.batchByChat) > 0
	if chatHint == 0 {
		for chatID := range a.batchByChat {
			chatHint = chatID
			break
		}
	}
	a.batchMu.Unlock()
	a.setBatchState(active, chatHint)
}
