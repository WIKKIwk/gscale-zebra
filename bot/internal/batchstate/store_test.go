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
	if err := s.Set(true, 123, "ITM-001", "GRENKI YASHIL", "Stores - A"); err != nil {
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
	if got.Batch.ItemCode != "ITM-001" {
		t.Fatalf("item_code mismatch: %q", got.Batch.ItemCode)
	}
	if got.Batch.ItemName != "GRENKI YASHIL" {
		t.Fatalf("item_name mismatch: %q", got.Batch.ItemName)
	}
	if got.Batch.Warehouse != "Stores - A" {
		t.Fatalf("warehouse mismatch: %q", got.Batch.Warehouse)
	}
	if got.Batch.UpdatedAt == "" {
		t.Fatalf("updated_at missing")
	}
}

func TestSetInactiveClearsItemFields(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge_state.json")

	s := New(p)
	if err := s.Set(true, 123, "ITM-001", "ITEM", "Stores - A"); err != nil {
		t.Fatalf("Set active error: %v", err)
	}
	if err := s.Set(false, 123, "", "", ""); err != nil {
		t.Fatalf("Set inactive error: %v", err)
	}

	got, err := bridgestate.New(p).Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if got.Batch.Active {
		t.Fatalf("active should be false")
	}
	if got.Batch.ItemCode != "" || got.Batch.ItemName != "" || got.Batch.Warehouse != "" {
		t.Fatalf("item fields not cleared: %+v", got.Batch)
	}
}
