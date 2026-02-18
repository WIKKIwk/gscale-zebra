package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: strings.TrimSpace(path)}
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(s.path)
}

func (s *Store) Read() (Snapshot, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return Snapshot{}, fmt.Errorf("bridge state path bo'sh")
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return Snapshot{}, err
	}
	var out Snapshot
	if err := json.Unmarshal(b, &out); err != nil {
		return Snapshot{}, err
	}
	return out, nil
}

func (s *Store) Update(mutator func(*Snapshot)) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("mkdir bridge dir: %w", err)
	}

	unlock, err := lockFile(s.path + ".lock")
	if err != nil {
		return fmt.Errorf("lock bridge state: %w", err)
	}
	defer unlock()

	cur := Snapshot{}
	if b, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(b, &cur)
	}

	if mutator != nil {
		mutator(&cur)
	}
	cur.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	b, err := json.Marshal(cur)
	if err != nil {
		return fmt.Errorf("marshal bridge state: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return fmt.Errorf("write temp bridge state: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename bridge state: %w", err)
	}
	return nil
}

func lockFile(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}
	unlock := func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
	return unlock, nil
}
