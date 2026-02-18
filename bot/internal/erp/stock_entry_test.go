package erp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateMaterialIssueDraft(t *testing.T) {
	t.Helper()

	type posted struct {
		StockEntryType string `json:"stock_entry_type"`
		Company        string `json:"company"`
		FromWarehouse  string `json:"from_warehouse"`
		Items          []struct {
			ItemCode   string  `json:"item_code"`
			Warehouse  string  `json:"s_warehouse"`
			Qty        float64 `json:"qty"`
			UOM        string  `json:"uom"`
			StockUOM   string  `json:"stock_uom"`
			Conversion float64 `json:"conversion_factor"`
		} `json:"items"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "token k:s" {
			t.Fatalf("auth header mismatch: %q", got)
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/resource/Warehouse":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"name":"Stores - A","company":"Accord"}]}`))
			return
		case r.Method == http.MethodGet && r.URL.Path == "/api/resource/Item":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"name":"ITEM-1","stock_uom":"Kg"}]}`))
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api/resource/Stock Entry":
			var p posted
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				t.Fatalf("decode post payload: %v", err)
			}
			if p.StockEntryType != "Material Issue" {
				t.Fatalf("stock_entry_type mismatch: %q", p.StockEntryType)
			}
			if p.Company != "Accord" || p.FromWarehouse != "Stores - A" {
				t.Fatalf("header fields mismatch: %+v", p)
			}
			if len(p.Items) != 1 {
				t.Fatalf("items len mismatch: %d", len(p.Items))
			}
			if p.Items[0].ItemCode != "ITEM-1" || p.Items[0].Warehouse != "Stores - A" || p.Items[0].Qty != 2.5 {
				t.Fatalf("item payload mismatch: %+v", p.Items[0])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"name":"MAT-STE-2026-00001"}}`))
			return
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New(ts.URL, "k", "s")
	draft, err := c.CreateMaterialIssueDraft(context.Background(), MaterialIssueDraftInput{
		ItemCode:  "ITEM-1",
		Warehouse: "Stores - A",
		Qty:       2.5,
	})
	if err != nil {
		t.Fatalf("CreateMaterialIssueDraft error: %v", err)
	}
	if draft.Name != "MAT-STE-2026-00001" {
		t.Fatalf("draft name mismatch: %q", draft.Name)
	}
	if draft.UOM != "Kg" {
		t.Fatalf("draft uom mismatch: %q", draft.UOM)
	}
}

func TestCreateMaterialIssueDraft_Validate(t *testing.T) {
	c := New("https://example.invalid", "k", "s")
	_, err := c.CreateMaterialIssueDraft(context.Background(), MaterialIssueDraftInput{ItemCode: "", Warehouse: "W", Qty: 1})
	if err == nil || !strings.Contains(err.Error(), "item code") {
		t.Fatalf("expected item code error, got: %v", err)
	}
}
