package main

import "testing"

func TestParseWeightNegativeFormats(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "minus-attached", raw: "ST,-13kg", want: -13},
		{name: "minus-spaced", raw: "ST, - 13 kg", want: -13},
		{name: "minus-suffix", raw: "ST, 13 kg-", want: -13},
		{name: "unicode-minus", raw: "ST, −13.5kg", want: -13.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, unit, _, ok := parseWeight(tc.raw, "kg")
			if !ok {
				t.Fatalf("parseWeight returned ok=false")
			}
			if w != tc.want {
				t.Fatalf("weight mismatch: got=%v want=%v", w, tc.want)
			}
			if unit != "kg" {
				t.Fatalf("unit mismatch: got=%q want=%q", unit, "kg")
			}
		})
	}
}

func TestParseWeightPrefersMatchWithUnit(t *testing.T) {
	raw := "x=123 ST - 13 kg"
	w, unit, _, ok := parseWeight(raw, "kg")
	if !ok {
		t.Fatalf("parseWeight returned ok=false")
	}
	if w != -13 {
		t.Fatalf("weight mismatch: got=%v want=%v", w, -13)
	}
	if unit != "kg" {
		t.Fatalf("unit mismatch: got=%q want=%q", unit, "kg")
	}
}
