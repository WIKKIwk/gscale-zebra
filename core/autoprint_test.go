package core

import (
	"strings"
	"testing"
	"time"
)

func TestStableEPCDetector_ReTriggersAfterMovementAndRestabilize(t *testing.T) {
	d := NewStableEPCDetector(DefaultStableEPCConfig())
	t0 := time.Unix(1_700_000_000, 0)

	w1 := 1.250
	// Stabilize
	d.Observe(&w1, t0)
	epc, ok := d.Observe(&w1, t0.Add(1100*time.Millisecond))
	if !ok || epc == "" {
		t.Fatal("first stable trigger expected")
	}

	// O'sha nuqtada qolsa qayta trigger bo'lmasligi kerak.
	_, okSame := d.Observe(&w1, t0.Add(2200*time.Millisecond))
	if okSame {
		t.Fatal("should NOT re-trigger while staying on same stable weight")
	}

	// Ma'noli o'zgarish bo'ldi (epsilon dan katta) -> yangi sikl armed, lekin hali trigger yo'q.
	w2 := 1.256
	_, ok2 := d.Observe(&w2, t0.Add(2300*time.Millisecond))
	if ok2 {
		t.Fatal("should NOT trigger immediately on movement")
	}

	// Yangi nuqta stable bo'lgach trigger bo'lishi kerak.
	epc2, ok3 := d.Observe(&w2, t0.Add(3400*time.Millisecond))
	if !ok3 || epc2 == "" {
		t.Fatal("should trigger after movement and next stable period")
	}
	if epc2 == epc {
		t.Fatal("new stable period should produce new EPC")
	}

	// Yana o'zgarib, aynan avvalgi qty ga qaytsa ham yangi trigger bo'lishi kerak.
	_, ok4 := d.Observe(&w1, t0.Add(3500*time.Millisecond))
	if ok4 {
		t.Fatal("should NOT trigger immediately on return movement")
	}
	epc3, ok5 := d.Observe(&w1, t0.Add(4600*time.Millisecond))
	if !ok5 || epc3 == "" {
		t.Fatal("should trigger when returning to previous weight after movement")
	}
	if epc3 == epc2 {
		t.Fatal("each stable cycle should produce a unique EPC")
	}
}

func TestStableEPCDetector_TinyJitterDoesNotRearm(t *testing.T) {
	d := NewStableEPCDetector(DefaultStableEPCConfig())
	t0 := time.Unix(1_700_000_000, 0)

	w1 := 2.000
	d.Observe(&w1, t0)
	if epc, ok := d.Observe(&w1, t0.Add(1100*time.Millisecond)); !ok || epc == "" {
		t.Fatal("first stable trigger expected")
	}

	// epsilon (0.005) dan kichik tebranish rearm qilmasligi kerak.
	w2 := 2.003
	_, ok2 := d.Observe(&w2, t0.Add(1300*time.Millisecond))
	if ok2 {
		t.Fatal("tiny jitter should not trigger")
	}
	_, ok3 := d.Observe(&w2, t0.Add(2500*time.Millisecond))
	if ok3 {
		t.Fatal("tiny jitter should not rearm new cycle")
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
