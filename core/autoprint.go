package core

import (
	"fmt"
	"math"
	"time"
)

type StableEPCConfig struct {
	StableFor time.Duration
	Epsilon   float64
	MinWeight float64
}

func DefaultStableEPCConfig() StableEPCConfig {
	return StableEPCConfig{
		StableFor: 1 * time.Second,
		Epsilon:   0.005,
		MinWeight: 0.0,
	}
}

type StableEPCDetector struct {
	cfg StableEPCConfig

	active    bool
	candidate float64
	since     time.Time
	printed   bool

	lastNS int64
	seq    uint32
}

func NewStableEPCDetector(cfg StableEPCConfig) *StableEPCDetector {
	if cfg.StableFor <= 0 {
		cfg.StableFor = 1 * time.Second
	}
	if cfg.Epsilon <= 0 {
		cfg.Epsilon = 0.005
	}
	if cfg.MinWeight < 0 {
		cfg.MinWeight = 0
	}
	return &StableEPCDetector{cfg: cfg}
}

func (d *StableEPCDetector) Observe(weight *float64, at time.Time) (string, bool) {
	if at.IsZero() {
		at = time.Now()
	}
	if weight == nil {
		d.reset()
		return "", false
	}

	w := *weight
	if math.IsNaN(w) || math.IsInf(w, 0) || w <= d.cfg.MinWeight {
		d.reset()
		return "", false
	}

	if !d.active {
		d.active = true
		d.candidate = w
		d.since = at
		d.printed = false
		return "", false
	}

	if math.Abs(w-d.candidate) > d.cfg.Epsilon {
		d.candidate = w
		d.since = at
		d.printed = false
		return "", false
	}

	if d.printed {
		return "", false
	}

	if at.Sub(d.since) < d.cfg.StableFor {
		return "", false
	}

	d.printed = true
	return d.nextEPC24(at), true
}

func (d *StableEPCDetector) reset() {
	d.active = false
	d.printed = false
	d.candidate = 0
	d.since = time.Time{}
}

// nextEPC24 returns a 24-char uppercase hex EPC-like id:
// 30 + 14 hex chars (unix ns low 56-bit) + 8 hex chars (sequence).
func (d *StableEPCDetector) nextEPC24(t time.Time) string {
	ns := t.UnixNano()
	if ns != d.lastNS {
		d.lastNS = ns
		d.seq = 0
	} else {
		d.seq++
	}
	return fmt.Sprintf("30%014X%08X", uint64(ns)&0x00FFFFFFFFFFFFFF, d.seq)
}
