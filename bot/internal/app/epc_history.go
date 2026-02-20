package app

import (
	"strings"
	"sync"
)

// EPCHistory keeps EPC values used in successful draft creation for current bot runtime.
type EPCHistory struct {
	mu   sync.RWMutex
	list []string
}

func NewEPCHistory() *EPCHistory {
	return &EPCHistory{
		list: make([]string, 0, 64),
	}
}

func (h *EPCHistory) Add(epc string) {
	if h == nil {
		return
	}

	v := strings.ToUpper(strings.TrimSpace(epc))
	if v == "" {
		return
	}

	h.mu.Lock()
	h.list = append(h.list, v)
	h.mu.Unlock()
}

func (h *EPCHistory) Snapshot() []string {
	if h == nil {
		return nil
	}

	h.mu.RLock()
	out := make([]string, len(h.list))
	copy(out, h.list)
	h.mu.RUnlock()
	return out
}
