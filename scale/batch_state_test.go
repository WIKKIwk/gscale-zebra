package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBatchStateReader_DefaultWhenMissing(t *testing.T) {
	r := newBatchStateReader(filepath.Join(t.TempDir(), "missing.json"), false)
	if got := r.Active(time.Now()); got {
		t.Fatalf("expected false default, got true")
	}
}

func TestBatchStateReader_ReadsActive(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge_state.json")
	if err := os.WriteFile(p, []byte(`{"batch":{"active":true,"updated_at":"2026-02-18T00:00:00Z"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r := newBatchStateReader(p, false)
	if got := r.Active(time.Now()); !got {
		t.Fatalf("expected true")
	}
}
