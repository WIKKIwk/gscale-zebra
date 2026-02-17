package commands

import "testing"

func TestExtractSelectedItemCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{name: "normal", in: "Item: ITM-001\nNomi: Apple", want: "ITM-001", ok: true},
		{name: "spaces", in: " item:  GRENKI YASHIL  \nNomi: X", want: "GRENKI YASHIL", ok: true},
		{name: "invalid", in: "Nomi: Apple", want: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ExtractSelectedItemCode(tc.in)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("got=(%q,%v) want=(%q,%v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestWarehouseInlineRoundtrip(t *testing.T) {
	seed := buildWarehouseInlineSeed("GRENKI YASHIL")
	req, ok := parseWarehouseInlineQuery(seed)
	if !ok {
		t.Fatalf("parse failed for seed=%q", seed)
	}
	if req.ItemCode != "GRENKI YASHIL" {
		t.Fatalf("item code mismatch: %q", req.ItemCode)
	}
	if req.Query != "" {
		t.Fatalf("query mismatch: %q", req.Query)
	}
}

func TestParseWarehouseInlineQuery_WithSearch(t *testing.T) {
	seed := buildWarehouseInlineSeed("ITM-001")
	seed = seed[:len(seed)-1] + "store"
	req, ok := parseWarehouseInlineQuery(seed)
	if !ok {
		t.Fatalf("parse failed for %q", seed)
	}
	if req.ItemCode != "ITM-001" {
		t.Fatalf("item code mismatch: %q", req.ItemCode)
	}
	if req.Query != "store" {
		t.Fatalf("query mismatch: %q", req.Query)
	}
}
