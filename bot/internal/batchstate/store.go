package batchstate

import (
	bridgestate "bridge/state"
	"strings"
	"time"
)

type Store struct {
	store *bridgestate.Store
}

func New(path string) *Store {
	path = strings.TrimSpace(path)
	if path == "" {
		return &Store{}
	}
	return &Store{store: bridgestate.New(path)}
}

func (s *Store) Set(active bool, chatID int64) error {
	if s == nil || s.store == nil || strings.TrimSpace(s.store.Path()) == "" {
		return nil
	}
	at := time.Now().UTC().Format(time.RFC3339Nano)
	return s.store.Update(func(snapshot *bridgestate.Snapshot) {
		snapshot.Batch.Active = active
		snapshot.Batch.ChatID = chatID
		snapshot.Batch.UpdatedAt = at
	})
}
