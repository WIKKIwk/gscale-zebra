package commands

import "testing"

func TestExtractSelectedWarehouse(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		itemCode  string
		warehouse string
		ok        bool
	}{
		{
			name:      "normal",
			in:        "Item: ITM-001\nOmbor: Stores - A\nQoldiq: 12.5",
			itemCode:  "ITM-001",
			warehouse: "Stores - A",
			ok:        true,
		},
		{
			name:      "spaces",
			in:        " item:  GRENKI YASHIL \n ombor:  Main WH  ",
			itemCode:  "GRENKI YASHIL",
			warehouse: "Main WH",
			ok:        true,
		},
		{
			name:      "missing warehouse",
			in:        "Item: ITM-001\nNomi: Apple",
			itemCode:  "",
			warehouse: "",
			ok:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemCode, warehouse, ok := ExtractSelectedWarehouse(tc.in)
			if ok != tc.ok || itemCode != tc.itemCode || warehouse != tc.warehouse {
				t.Fatalf("got=(%q,%q,%v) want=(%q,%q,%v)", itemCode, warehouse, ok, tc.itemCode, tc.warehouse, tc.ok)
			}
		})
	}
}
