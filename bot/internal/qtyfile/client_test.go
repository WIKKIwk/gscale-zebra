package qtyfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWaitStablePositive_StableTrue(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "qty.json")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := os.WriteFile(p, []byte(`{"weight":1.234,"unit":"kg","stable":true,"updated_at":"`+now+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	c := New(p)
	qty, unit, err := c.WaitStablePositive(context.Background(), 2*time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitStablePositive error: %v", err)
	}
	if qty != 1.234 {
		t.Fatalf("qty mismatch: %v", qty)
	}
	if unit != "kg" {
		t.Fatalf("unit mismatch: %q", unit)
	}
}

func TestWaitStablePositive_Timeout(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "qty.json")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := os.WriteFile(p, []byte(`{"weight":0,"unit":"kg","stable":false,"updated_at":"`+now+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	c := New(p)
	_, _, err := c.WaitStablePositive(context.Background(), 250*time.Millisecond, 50*time.Millisecond)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestWaitForNextCycle_ByReset(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "qty.json")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := os.WriteFile(p, []byte(`{"weight":0,"unit":"kg","stable":false,"updated_at":"`+now+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	c := New(p)
	if err := c.WaitForNextCycle(context.Background(), 500*time.Millisecond, 50*time.Millisecond, 1.0); err != nil {
		t.Fatalf("WaitForNextCycle reset error: %v", err)
	}
}

func TestWaitForNextCycle_ByChange(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "qty.json")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := os.WriteFile(p, []byte(`{"weight":1.500,"unit":"kg","stable":true,"updated_at":"`+now+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	c := New(p)
	if err := c.WaitForNextCycle(context.Background(), 500*time.Millisecond, 50*time.Millisecond, 1.000); err != nil {
		t.Fatalf("WaitForNextCycle change error: %v", err)
	}
}
