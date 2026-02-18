package commands

import (
	"context"

	"bot/internal/telegram"
)

func HandleImage(ctx context.Context, deps Deps, msg telegram.Message) (int64, error) {
	return deps.TG.SendMessageWithInlineKeyboardAndReturnID(ctx, msg.Chat.ID, "Rasm tashlang.", nil)
}
