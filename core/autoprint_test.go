package core

import (
	"strings"
	"testing"
	"time"
)

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
