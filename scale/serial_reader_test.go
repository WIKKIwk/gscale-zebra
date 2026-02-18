package main

import "testing"

func TestPopSerialFrame(t *testing.T) {
	frame, rest, ok := popSerialFrame("-  2.05\r-  1.00\r")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if frame != "-  2.05" {
		t.Fatalf("frame mismatch: got=%q", frame)
	}
	if rest != "-  1.00\r" {
		t.Fatalf("rest mismatch: got=%q", rest)
	}
}

func TestPopSerialFrameConsumesCRLF(t *testing.T) {
	frame, rest, ok := popSerialFrame("0.00\r\n- 0.50\n")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if frame != "0.00" {
		t.Fatalf("frame mismatch: got=%q", frame)
	}
	if rest != "- 0.50\n" {
		t.Fatalf("rest mismatch: got=%q", rest)
	}
}

func TestPopSerialFrameNoDelimiter(t *testing.T) {
	frame, rest, ok := popSerialFrame("-  2.05")
	if ok {
		t.Fatalf("expected ok=false, frame=%q rest=%q", frame, rest)
	}
	if rest != "-  2.05" {
		t.Fatalf("rest mismatch: got=%q", rest)
	}
}
