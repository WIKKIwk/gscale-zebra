package erp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckConnection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "token k:s" {
			t.Fatalf("auth header mismatch: %q", got)
		}
		if r.URL.Path != "/api/method/frappe.auth.get_logged_user" {
			t.Fatalf("path mismatch: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Administrator"}`))
	}))
	defer ts.Close()

	c := New(ts.URL, "k", "s")
	user, err := c.CheckConnection(context.Background())
	if err != nil {
		t.Fatalf("CheckConnection error: %v", err)
	}
	if user != "Administrator" {
		t.Fatalf("user mismatch: %q", user)
	}
}

func TestSearchItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "token k:s" {
			t.Fatalf("auth header mismatch: %q", got)
		}
		if r.URL.Path != "/api/resource/Item" {
			t.Fatalf("path mismatch: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"name":"ITM-001","item_code":"ITM-001","item_name":"Apple"},{"name":"ITM-002","item_name":"Banana"}]}`))
	}))
	defer ts.Close()

	c := New(ts.URL, "k", "s")
	items, err := c.SearchItems(context.Background(), "", 10)
	if err != nil {
		t.Fatalf("SearchItems error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items len mismatch: got=%d want=2", len(items))
	}
	if items[0].ItemCode != "ITM-001" || items[0].ItemName != "Apple" {
		t.Fatalf("item[0] mismatch: %+v", items[0])
	}
	if items[1].ItemCode != "ITM-002" || items[1].ItemName != "Banana" {
		t.Fatalf("item[1] mismatch: %+v", items[1])
	}
}
