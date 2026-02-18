package scaleapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaitStablePositive_StableTrue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"weight":1.234,"unit":"kg","stable":true}`))
	}))
	defer ts.Close()

	c := New(ts.URL)
	qty, unit, err := c.WaitStablePositive(context.Background(), 2*time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitStablePositive error: %v", err)
	}
	if qty != 1.234 {
		t.Fatalf("qty mismatch: %v", qty)
	}
	if unit != "kg" {
		t.Fatalf("unit mismatch: %q", unit)
	}
}

func TestWaitStablePositive_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"weight":0.0,"unit":"kg","stable":false}`))
	}))
	defer ts.Close()

	c := New(ts.URL)
	_, _, err := c.WaitStablePositive(context.Background(), 250*time.Millisecond, 50*time.Millisecond)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}
