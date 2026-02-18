package app

import (
	"context"
	"fmt"

	"bot/internal/telegram"
)

func (a *App) handleIncomingPhoto(ctx context.Context, msg telegram.Message) error {
	chatID := msg.Chat.ID
	if !a.isImageAwaiting(chatID) {
		return nil
	}

	photo, ok := pickLargestPhoto(msg.Photo)
	if !ok {
		return a.tg.SendMessage(ctx, chatID, "Rasm topilmadi. Iltimos, qayta yuboring.")
	}

	filePath, err := a.tg.GetFilePath(ctx, photo.FileID)
	if err != nil {
		return a.tg.SendMessage(ctx, chatID, "Rasm faylini olishda xato: "+err.Error())
	}

	payload, err := a.tg.DownloadFile(ctx, filePath)
	if err != nil {
		return a.tg.SendMessage(ctx, chatID, "Rasmni yuklab olishda xato: "+err.Error())
	}

	if err := a.imagePrinter.PrintImageBytes(ctx, payload); err != nil {
		return a.tg.SendMessage(ctx, chatID, "Rasm print xato: "+err.Error())
	}

	a.setImageAwaiting(chatID, false)
	return a.tg.SendMessage(ctx, chatID, fmt.Sprintf("Rasm chop etildi (%dx%d).", photo.Width, photo.Height))
}

func pickLargestPhoto(photos []telegram.PhotoSize) (telegram.PhotoSize, bool) {
	if len(photos) == 0 {
		return telegram.PhotoSize{}, false
	}

	best := photos[0]
	bestScore := photoScore(best)
	for _, p := range photos[1:] {
		score := photoScore(p)
		if score > bestScore {
			best = p
			bestScore = score
		}
	}

	if best.FileID == "" {
		return telegram.PhotoSize{}, false
	}
	return best, true
}

func photoScore(p telegram.PhotoSize) int64 {
	area := int64(p.Width) * int64(p.Height)
	if p.FileSize > 0 {
		return area*10 + p.FileSize
	}
	return area * 10
}
