package app

import (
	"reflect"
	"testing"
	"time"
)

func TestEPCHistorySnapshot(t *testing.T) {
	t.Parallel()

	h := NewEPCHistory()
	h.Add("")
	h.Add("  ")
	h.Add("3034257bf7194e4000000001")
	h.Add("  3034257BF7194E4000000002  ")

	got := h.Snapshot()
	want := []string{
		"3034257BF7194E4000000001",
		"3034257BF7194E4000000002",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot mismatch\ngot : %#v\nwant: %#v", got, want)
	}
}

func TestBuildEPCDocument(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.February, 20, 12, 34, 56, 0, time.UTC)
	filename, content := buildEPCDocument([]string{"AAA", "BBB"}, now)

	if filename != "epc-history-20260220-123456.txt" {
		t.Fatalf("filename mismatch: %q", filename)
	}
	if string(content) != "AAA\nBBB\n" {
		t.Fatalf("content mismatch: %q", string(content))
	}
}
