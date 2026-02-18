package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteQtySnapshot(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "qty.json")
	w := 1.234
	st := true

	err := writeQtySnapshot(p, Reading{
		Source:    "serial",
		Port:      "/dev/ttyUSB0",
		Weight:    &w,
		Unit:      "kg",
		Stable:    &st,
		UpdatedAt: time.Now(),
	}, ZebraStatus{LastEPC: "3034ABCDEF1234567890AA", Verify: "MATCH", UpdatedAt: time.Now()})
	if err != nil {
		t.Fatalf("writeQtySnapshot error: %v", err)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got["unit"] != "kg" {
		t.Fatalf("unit mismatch: %v", got["unit"])
	}
	if got["weight"] == nil {
		t.Fatalf("weight missing")
	}
	if got["epc"] != "3034ABCDEF1234567890AA" {
		t.Fatalf("epc mismatch: %v", got["epc"])
	}
}
