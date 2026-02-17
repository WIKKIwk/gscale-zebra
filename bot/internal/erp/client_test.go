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
