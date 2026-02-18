package app

func (a *App) setImageAwaiting(chatID int64, waiting bool) {
	if waiting {
		a.imageAwaitByChat[chatID] = true
		return
	}
	delete(a.imageAwaitByChat, chatID)
}

func (a *App) isImageAwaiting(chatID int64) bool {
	return a.imageAwaitByChat[chatID]
}
