package core

import (
	"strings"
	"testing"
	"time"
)

func TestStableEPCDetector_NoDoubleTriggerOnFluctuation(t *testing.T) {
	d := NewStableEPCDetector(DefaultStableEPCConfig())
	t0 := time.Unix(1_700_000_000, 0)

	w1 := 1.250
	// Stabilize
	d.Observe(&w1, t0)
	epc, ok := d.Observe(&w1, t0.Add(1100*time.Millisecond))
	if !ok || epc == "" {
		t.Fatal("first stable trigger expected")
	}

	// Slight fluctuation — must NOT trigger again
	w2 := 1.256 // > epsilon (0.005)
	_, ok2 := d.Observe(&w2, t0.Add(1200*time.Millisecond))
	if ok2 {
		t.Fatal("should NOT re-trigger on fluctuation after print")
	}
	_, ok3 := d.Observe(&w2, t0.Add(2300*time.Millisecond))
	if ok3 {
		t.Fatal("should NOT re-trigger even after new stable period with fluctuation")
	}

	// Remove item (nil weight) → reset
	d.Observe(nil, t0.Add(3000*time.Millisecond))

	// Place again → should trigger
	w3 := 1.250
	d.Observe(&w3, t0.Add(3100*time.Millisecond))
	epc2, ok4 := d.Observe(&w3, t0.Add(4200*time.Millisecond))
	if !ok4 || epc2 == "" {
		t.Fatal("should trigger after item removed and placed again")
	}
	if epc2 == epc {
		t.Fatal("new trigger should produce new EPC")
	}
}

func TestNextEPC24_LengthAndUniq(t *testing.T) {
	d := NewStableEPCDetector(DefaultStableEPCConfig())
	t0 := time.Unix(1_700_000_000, 123_456_789)

	a := d.nextEPC24(t0)
	b := d.nextEPC24(t0)

	if len(a) != 24 {
		t.Fatalf("epc len mismatch: got=%d epc=%s", len(a), a)
	}
	if len(b) != 24 {
		t.Fatalf("epc len mismatch: got=%d epc=%s", len(b), b)
	}
	if a == b {
		t.Fatalf("epc should be unique on same ns tick: %s", a)
	}
	if strings.HasSuffix(a, "00000000") || strings.HasSuffix(b, "00000000") {
		t.Fatalf("epc tail should not be all-zero: a=%s b=%s", a, b)
	}
	if !isUpperHex(a) || !isUpperHex(b) {
		t.Fatalf("epc must be uppercase hex: a=%s b=%s", a, b)
	}
}

func isUpperHex(v string) bool {
	for _, ch := range v {
		if strings.ContainsRune("0123456789ABCDEF", ch) {
			continue
		}
		return false
	}
	return true
}
