package commands

import (
	"context"
	"fmt"

	"bot/internal/telegram"
)

func HandleStart(ctx context.Context, deps Deps, msg telegram.Message) error {
	user, err := deps.ERP.CheckConnection(ctx)
	if err != nil {
		return deps.TG.SendMessage(ctx, msg.Chat.ID, "ERPNext ulanishi xato: "+err.Error())
	}
	return deps.TG.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("ERPNext ga ulandi. User: %s", user))
}
