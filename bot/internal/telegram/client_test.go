package telegram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInlineKeyboardButton_EmptySwitchInlineQueryCurrentChatIsSerialized(t *testing.T) {
	k := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{{
			{Text: "Item tanlash", SwitchInlineQueryCurrentChat: ""},
		}},
	}

	b, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	const want = `"switch_inline_query_current_chat":""`
	if !strings.Contains(string(b), want) {
		t.Fatalf("expected %s in payload, got: %s", want, string(b))
	}
}
