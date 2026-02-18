package batchstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetWritesSnapshot(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "batch_state.json")

	s := New(p)
	if err := s.Set(true, 123); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got["active"] != true {
		t.Fatalf("active mismatch: %v", got["active"])
	}
	if got["chat_id"] != float64(123) {
		t.Fatalf("chat_id mismatch: %v", got["chat_id"])
	}
	if got["updated_at"] == nil || got["updated_at"] == "" {
		t.Fatalf("updated_at missing")
	}
}
