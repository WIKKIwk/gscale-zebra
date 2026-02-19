package main

import (
	"testing"
	"time"
)

func TestMergeZebraStatus_PreservesOldEPCEventTimeOnHeartbeat(t *testing.T) {
	prevAt := time.Date(2026, 2, 18, 15, 20, 0, 0, time.UTC)
	prev := ZebraStatus{
		LastEPC:   "3034257BF7194E4069940A1B",
		Verify:    "WRITTEN",
		UpdatedAt: prevAt,
	}
	incoming := ZebraStatus{
		LastEPC:   "",
		Verify:    "-",
		UpdatedAt: prevAt.Add(2 * time.Second),
	}

	got := mergeZebraStatus(prev, incoming)

	if got.LastEPC != prev.LastEPC {
		t.Fatalf("last epc mismatch: got=%q want=%q", got.LastEPC, prev.LastEPC)
	}
	if got.Verify != prev.Verify {
		t.Fatalf("verify mismatch: got=%q want=%q", got.Verify, prev.Verify)
	}
	if !got.UpdatedAt.Equal(prevAt) {
		t.Fatalf("updated_at should stay on original epc event: got=%s want=%s", got.UpdatedAt.Format(time.RFC3339Nano), prevAt.Format(time.RFC3339Nano))
	}
}

func TestMergeZebraStatus_UsesIncomingWhenNewEPCArrives(t *testing.T) {
	prevAt := time.Date(2026, 2, 18, 15, 20, 0, 0, time.UTC)
	incomingAt := prevAt.Add(3 * time.Second)
	prev := ZebraStatus{
		LastEPC:   "3034257BF7194E4069940A1B",
		Verify:    "WRITTEN",
		UpdatedAt: prevAt,
	}
	incoming := ZebraStatus{
		LastEPC:   "3034257BF7194E4069940A1C",
		Verify:    "WRITTEN",
		UpdatedAt: incomingAt,
	}

	got := mergeZebraStatus(prev, incoming)

	if got.LastEPC != incoming.LastEPC {
		t.Fatalf("last epc mismatch: got=%q want=%q", got.LastEPC, incoming.LastEPC)
	}
	if !got.UpdatedAt.Equal(incomingAt) {
		t.Fatalf("updated_at mismatch: got=%s want=%s", got.UpdatedAt.Format(time.RFC3339Nano), incomingAt.Format(time.RFC3339Nano))
	}
}
