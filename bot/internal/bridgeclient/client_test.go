package bridgeclient

import (
	bridgestate "bridge/state"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestWaitStablePositiveReading_Stable(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge_state.json")
	s := bridgestate.New(p)
	w := 1.234
	st := true
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.Update(func(snapshot *bridgestate.Snapshot) {
		snapshot.Scale.Weight = &w
		snapshot.Scale.Stable = &st
		snapshot.Scale.Unit = "kg"
		snapshot.Scale.UpdatedAt = now
	}); err != nil {
		t.Fatal(err)
	}

	c := New(p)
	r, err := c.WaitStablePositiveReading(context.Background(), 2*time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitStablePositiveReading error: %v", err)
	}
	if r.Qty != 1.234 {
		t.Fatalf("qty mismatch: %v", r.Qty)
	}
	if r.Unit != "kg" {
		t.Fatalf("unit mismatch: %q", r.Unit)
	}
}

func TestWaitEPCForReading(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge_state.json")
	s := bridgestate.New(p)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.Update(func(snapshot *bridgestate.Snapshot) {
		snapshot.Zebra.LastEPC = "3034257BF7194E406994036B"
		snapshot.Zebra.Verify = "MATCH"
		snapshot.Zebra.ReadLine1 = "ok"
		snapshot.Zebra.UpdatedAt = now
	}); err != nil {
		t.Fatal(err)
	}

	c := New(p)
	got, err := c.WaitEPCForReading(context.Background(), 500*time.Millisecond, 50*time.Millisecond, time.Now().Add(-1*time.Second), "")
	if err != nil {
		t.Fatalf("WaitEPCForReading error: %v", err)
	}
	if got.EPC != "3034257BF7194E406994036B" {
		t.Fatalf("epc mismatch: %q", got.EPC)
	}
	if got.Verify != "MATCH" {
		t.Fatalf("verify mismatch: %q", got.Verify)
	}
	if got.ReadLine1 != "ok" {
		t.Fatalf("read line mismatch: %q", got.ReadLine1)
	}
}
