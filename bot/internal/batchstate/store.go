package batchstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store struct {
	path string
}

type snapshot struct {
	Active    bool   `json:"active"`
	ChatID    int64  `json:"chat_id,omitempty"`
	UpdatedAt string `json:"updated_at"`
}

func New(path string) *Store {
	return &Store{path: strings.TrimSpace(path)}
}

func (s *Store) Set(active bool, chatID int64) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return nil
	}

	snap := snapshot{
		Active:    active,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if chatID != 0 {
		snap.ChatID = chatID
	}

	b, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshal batch state: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir batch state dir: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return fmt.Errorf("write temp batch state: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename batch state: %w", err)
	}
	return nil
}
