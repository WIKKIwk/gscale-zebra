package commands

import (
	"context"
	"fmt"
	"strings"

	"bot/internal/telegram"
)

func HandleStart(ctx context.Context, deps Deps, msg telegram.Message) error {
	user, err := deps.ERP.CheckConnection(ctx)
	if err != nil {
		return deps.TG.SendMessage(ctx, msg.Chat.ID, "ERPNext ulanishi xato: "+err.Error())
	}

	info := strings.Join([]string{
		fmt.Sprintf("ERPNext ga ulandi. User: %s", user),
		"",
		"Men scale + zebra jarayonlari uchun yordamchi botman.",
		"Davom etish uchun /batch ni bosing.",
	}, "\n")

	return deps.TG.SendMessage(ctx, msg.Chat.ID, info)
}
