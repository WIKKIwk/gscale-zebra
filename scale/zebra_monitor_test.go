package main

import (
	"strings"
	"testing"
	"time"
)

func TestBuildRFIDEncodeCommand_IncludesEPCAndQtyOnLabel(t *testing.T) {
	stream, err := buildRFIDEncodeCommand("3034ABCDEF1234567890AA", "1.250 kg", "GREEN TEA")
	if err != nil {
		t.Fatalf("buildRFIDEncodeCommand error: %v", err)
	}

	if !strings.Contains(stream, "^RFW,H,,,A^FD3034ABCDEF1234567890AA^FS") {
		t.Fatalf("rfid write command not found in stream: %s", stream)
	}
	if !strings.Contains(stream, "^FDITEM: GREEN TEA^FS") {
		t.Fatalf("human ITEM line missing: %s", stream)
	}
	if !strings.Contains(stream, "^FDEPC: 3034ABCDEF1234567890AA^FS") {
		t.Fatalf("human EPC line missing: %s", stream)
	}
	if !strings.Contains(stream, "^FDQTY: 1.250 kg^FS") {
		t.Fatalf("human QTY line missing: %s", stream)
	}
	if !strings.Contains(stream, "^BCN,44,N,N,N") {
		t.Fatalf("barcode command missing: %s", stream)
	}
	if strings.Count(stream, "3034ABCDEF1234567890AA") < 3 {
		t.Fatalf("epc should be present for rfid write, text and barcode: %s", stream)
	}
}

func TestBuildRFIDEncodeCommand_DefaultQtyWhenEmpty(t *testing.T) {
	stream, err := buildRFIDEncodeCommand("3034ABCDEF1234567890AA", "", "")
	if err != nil {
		t.Fatalf("buildRFIDEncodeCommand error: %v", err)
	}

	if !strings.Contains(stream, "^FDQTY: - kg^FS") {
		t.Fatalf("default qty missing: %s", stream)
	}
	if !strings.Contains(stream, "^FDITEM: -^FS") {
		t.Fatalf("default item missing: %s", stream)
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

func TestGenerateTestEPC_LengthAndUniq(t *testing.T) {
	t0 := time.Unix(1_700_000_000, 123_456_789)
	a := generateTestEPC(t0)
	b := generateTestEPC(t0)

	if len(a) != 24 || len(b) != 24 {
		t.Fatalf("epc len mismatch: a=%d b=%d", len(a), len(b))
	}
	if a == b {
		t.Fatalf("expected unique epc for same tick: %s", a)
	}
	if strings.HasSuffix(a, "00000000") || strings.HasSuffix(b, "00000000") {
		t.Fatalf("epc tail should not be all-zero: a=%s b=%s", a, b)
	}
	if !isUpperHexScale(a) || !isUpperHexScale(b) {
		t.Fatalf("epc must be uppercase hex: a=%s b=%s", a, b)
	}
}

func TestParseRFIDCounter(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"382", 382, true},
		{" 74 ", 74, true},
		{"\"12\"", 12, true},
		{"", 0, false},
		{"?", 0, false},
		{"abc", 0, false},
	}

	for _, tc := range cases {
		got, ok := parseRFIDCounter(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("parseRFIDCounter(%q) = (%d,%v), want (%d,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func isUpperHexScale(v string) bool {
	for _, ch := range v {
		if strings.ContainsRune("0123456789ABCDEF", ch) {
			continue
		}
		return false
	}
	return true
}
