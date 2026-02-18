package batchstate

import (
	bridgestate "bridge/state"
	"path/filepath"
	"testing"
)

func TestSetWritesSnapshot(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge_state.json")

	s := New(p)
	if err := s.Set(true, 123); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	got, err := bridgestate.New(p).Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if !got.Batch.Active {
		t.Fatalf("active mismatch: %v", got.Batch.Active)
	}
	if got.Batch.ChatID != 123 {
		t.Fatalf("chat_id mismatch: %v", got.Batch.ChatID)
	}
	if got.Batch.UpdatedAt == "" {
		t.Fatalf("updated_at missing")
	}
}
