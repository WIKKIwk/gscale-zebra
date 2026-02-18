package main

import (
	bridgestate "bridge/state"
	"strings"
	"time"
)

type batchStateReader struct {
	store         *bridgestate.Store
	defaultActive bool
	cached        bool
	value         bool
	nextReadAt    time.Time
	itemCode      string
	itemName      string
	warehouse     string
}

func newBatchStateReader(path string, defaultActive bool) *batchStateReader {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	return &batchStateReader{
		store:         bridgestate.New(path),
		defaultActive: defaultActive,
	}
}

func (r *batchStateReader) Active(now time.Time) bool {
	r.refresh(now)
	return r.value
}

func (r *batchStateReader) ItemLabel(now time.Time) string {
	r.refresh(now)
	if !r.value {
		return ""
	}
	if strings.TrimSpace(r.itemName) != "" {
		return strings.TrimSpace(r.itemName)
	}
	return strings.TrimSpace(r.itemCode)
}

func (r *batchStateReader) refresh(now time.Time) {
	if r == nil {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	if r.cached && now.Before(r.nextReadAt) {
		return
	}

	snap, err := r.store.Read()
	if err != nil {
		if !r.cached {
			r.value = r.defaultActive
			r.cached = true
		}
		r.nextReadAt = now.Add(250 * time.Millisecond)
		return
	}

	if strings.TrimSpace(snap.Batch.UpdatedAt) == "" && snap.Batch.ChatID == 0 && !snap.Batch.Active {
		r.value = r.defaultActive
		r.itemCode = ""
		r.itemName = ""
		r.warehouse = ""
	} else {
		r.value = snap.Batch.Active
		r.itemCode = strings.TrimSpace(snap.Batch.ItemCode)
		r.itemName = strings.TrimSpace(snap.Batch.ItemName)
		if r.itemName == "" {
			r.itemName = r.itemCode
		}
		r.warehouse = strings.TrimSpace(snap.Batch.Warehouse)
		if !r.value {
			r.itemCode = ""
			r.itemName = ""
			r.warehouse = ""
		}
	}

	r.cached = true
	r.nextReadAt = now.Add(250 * time.Millisecond)
}
