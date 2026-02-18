package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

type batchStateSnapshot struct {
	Active bool `json:"active"`
}

type batchStateReader struct {
	path          string
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
	return &batchStateReader{path: path, defaultActive: defaultActive}
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

	active, err := r.read()
	if err != nil {
		if !r.cached {
			r.value = r.defaultActive
			r.cached = true
		}
		r.nextReadAt = now.Add(250 * time.Millisecond)
		return r.value
	}

	r.value = active
	r.cached = true
	r.nextReadAt = now.Add(250 * time.Millisecond)
	return r.value
}

func (r *batchStateReader) read() (bool, error) {
	b, err := os.ReadFile(r.path)
	if err != nil {
		return false, err
	}
	var s batchStateSnapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return false, err
	}
	return s.Active, nil
}
