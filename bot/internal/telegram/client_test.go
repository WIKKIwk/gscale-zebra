package telegram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInlineKeyboardButton_EmptySwitchInlineQueryCurrentChatIsOmitted(t *testing.T) {
	k := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{{
			{Text: "Item tanlash"},
		}},
	}

	b, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	const bad = `"switch_inline_query_current_chat"`
	if strings.Contains(string(b), bad) {
		t.Fatalf("did not expect %s in payload, got: %s", bad, string(b))
	}
}

func TestInlineKeyboardButton_CallbackDataIsSerialized(t *testing.T) {
	k := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{{
			{Text: "Material Issue", CallbackData: "stock:material_issue"},
		}},
	}

	b, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	const want = `"callback_data":"stock:material_issue"`
	if !strings.Contains(string(b), want) {
		t.Fatalf("expected %s in payload, got: %s", want, string(b))
	}
}
