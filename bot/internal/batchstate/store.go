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

func (s *Store) Set(active bool, chatID int64, itemCode, itemName, warehouse string) error {
	if s == nil || s.store == nil || strings.TrimSpace(s.store.Path()) == "" {
		return nil
	}

	itemCode = strings.TrimSpace(itemCode)
	itemName = strings.TrimSpace(itemName)
	warehouse = strings.TrimSpace(warehouse)
	if itemName == "" {
		itemName = itemCode
	}

	at := time.Now().UTC().Format(time.RFC3339Nano)
	return s.store.Update(func(snapshot *bridgestate.Snapshot) {
		snapshot.Batch.Active = active
		snapshot.Batch.ChatID = chatID
		if active {
			snapshot.Batch.ItemCode = itemCode
			snapshot.Batch.ItemName = itemName
			snapshot.Batch.Warehouse = warehouse
		} else {
			snapshot.Batch.ItemCode = ""
			snapshot.Batch.ItemName = ""
			snapshot.Batch.Warehouse = ""
		}
		snapshot.Batch.UpdatedAt = at
	})
}
