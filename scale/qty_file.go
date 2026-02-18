package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type qtySnapshot struct {
	Source    string   `json:"source,omitempty"`
	Port      string   `json:"port,omitempty"`
	Weight    *float64 `json:"weight"`
	Unit      string   `json:"unit"`
	Stable    *bool    `json:"stable"`
	Error     string   `json:"error,omitempty"`
	EPC       string   `json:"epc,omitempty"`
	EPCVerify string   `json:"epc_verify,omitempty"`
	EPCAt     string   `json:"epc_updated_at,omitempty"`
	UpdatedAt string   `json:"updated_at"`
}

func writeQtySnapshot(path string, rd Reading, zebra ZebraStatus) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	ts := rd.UpdatedAt
	if ts.IsZero() {
		ts = time.Now()
	}

	snap := qtySnapshot{
		Source:    strings.TrimSpace(rd.Source),
		Port:      strings.TrimSpace(rd.Port),
		Weight:    rd.Weight,
		Unit:      strings.TrimSpace(rd.Unit),
		Stable:    rd.Stable,
		Error:     strings.TrimSpace(rd.Error),
		UpdatedAt: ts.UTC().Format(time.RFC3339Nano),
	}
	if snap.Unit == "" {
		snap.Unit = "kg"
	}
	if epc := strings.ToUpper(strings.TrimSpace(zebra.LastEPC)); epc != "" {
		snap.EPC = epc
		snap.EPCVerify = strings.ToUpper(strings.TrimSpace(zebra.Verify))
		zts := zebra.UpdatedAt
		if zts.IsZero() {
			zts = ts
		}
		snap.EPCAt = zts.UTC().Format(time.RFC3339Nano)
	}

	b, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshal qty snapshot: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir qty dir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return fmt.Errorf("write temp qty file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename qty file: %w", err)
	}
	return nil
}
