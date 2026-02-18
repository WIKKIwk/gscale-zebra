package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreUpdateAndRead(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge_state.json")
	s := New(p)

	weight := 1.25
	stable := true
	if err := s.Update(func(snap *Snapshot) {
		snap.Scale.Weight = &weight
		snap.Scale.Stable = &stable
		snap.Scale.Unit = "kg"
		snap.Zebra.LastEPC = "3034257BF7194E406994036B"
		snap.Batch.Active = true
		snap.Batch.ChatID = 42
	}); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	got, err := s.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if got.Scale.Weight == nil || *got.Scale.Weight != 1.25 {
		t.Fatalf("weight mismatch: %+v", got.Scale.Weight)
	}
	if got.Zebra.LastEPC != "3034257BF7194E406994036B" {
		t.Fatalf("epc mismatch: %s", got.Zebra.LastEPC)
	}
	if !got.Batch.Active || got.Batch.ChatID != 42 {
		t.Fatalf("batch mismatch: %+v", got.Batch)
	}
	if got.UpdatedAt == "" {
		t.Fatalf("updated_at empty")
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("state file missing: %v", err)
	}
}
