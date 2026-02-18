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
	if r == nil {
		return true
	}
	if now.IsZero() {
		now = time.Now()
	}
	if r.cached && now.Before(r.nextReadAt) {
		return r.value
	}

	snap, err := r.store.Read()
	if err != nil {
		if !r.cached {
			r.value = r.defaultActive
			r.cached = true
		}
		r.nextReadAt = now.Add(250 * time.Millisecond)
		return r.value
	}

	if strings.TrimSpace(snap.Batch.UpdatedAt) == "" && snap.Batch.ChatID == 0 && !snap.Batch.Active {
		r.value = r.defaultActive
	} else {
		r.value = snap.Batch.Active
	}
	r.cached = true
	r.nextReadAt = now.Add(250 * time.Millisecond)
	return r.value
}
