package main

import (
	"strings"
	"testing"
)

func TestBuildRFIDEncodeCommand_IncludesEPCAndQtyOnLabel(t *testing.T) {
	stream, err := buildRFIDEncodeCommand("3034ABCDEF1234567890AA", "1.250 kg")
	if err != nil {
		t.Fatalf("buildRFIDEncodeCommand error: %v", err)
	}

	if !strings.Contains(stream, "^RFW,H,,,A^FD3034ABCDEF1234567890AA^FS") {
		t.Fatalf("rfid write command not found in stream: %s", stream)
	}
	if !strings.Contains(stream, "^FDEPC: 3034ABCDEF1234567890AA^FS") {
		t.Fatalf("human EPC line missing: %s", stream)
	}
	if !strings.Contains(stream, "^FDQTY: 1.250 kg^FS") {
		t.Fatalf("human QTY line missing: %s", stream)
	}
}

func TestBuildRFIDEncodeCommand_DefaultQtyWhenEmpty(t *testing.T) {
	stream, err := buildRFIDEncodeCommand("3034ABCDEF1234567890AA", "")
	if err != nil {
		t.Fatalf("buildRFIDEncodeCommand error: %v", err)
	}

	if !strings.Contains(stream, "^FDQTY: - kg^FS") {
		t.Fatalf("default qty missing: %s", stream)
	}
}

func TestFormatLabelQty(t *testing.T) {
	w := 2.3456
	got := formatLabelQty(&w, "kg")
	if got != "2.346 kg" {
		t.Fatalf("formatLabelQty mismatch: got=%q", got)
	}

	got = formatLabelQty(nil, "kg")
	if got != "- kg" {
		t.Fatalf("formatLabelQty nil mismatch: got=%q", got)
	}
}
