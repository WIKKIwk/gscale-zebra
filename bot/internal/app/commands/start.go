package commands

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/telegram"
)

func HandleStart(ctx context.Context, deps Deps, msg telegram.Message) (int64, error) {
	user, err := deps.ERP.CheckConnection(ctx)
	if err != nil {
		return 0, deps.TG.SendMessage(ctx, msg.Chat.ID, "ERPNext ulanishi xato: "+err.Error())
	}

	info := strings.Join([]string{
		fmt.Sprintf("ERPNext ga ulandi. User: %s", user),
		"",
		"Men scale + zebra jarayonlari uchun yordamchi botman.",
		"Davom etish uchun /batch ni bosing.",
	}, "\n")

	return deps.TG.SendMessageWithInlineKeyboardAndReturnID(ctx, msg.Chat.ID, info, nil)
}
