package core

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"os"
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

	active        bool
	candidate     float64
	since         time.Time
	printed       bool
	printedWeight float64

	lastNS int64
	seq    uint32
	salt   uint32
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
	return &StableEPCDetector{cfg: cfg, salt: newEPCSalt()}
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

	// Agar allaqachon chop etilgan bo'lsa:
	// - Kichik tebranishlar (masalan 78.800↔78.850) yangi EPC hosil qilmaydi.
	// - Lekin vazn 35%+ kamaysa (item olib tashlangan), yangi siklni boshlaymiz.
	//   Bu holda foydalanuvchi tarozini 0 ga olib borishga majbur emas.
	if d.printed {
		if d.printedWeight > 0 && w < d.printedWeight*0.65 {
			d.reset()
			// fall through: yangi sikl boshlanadi
		} else {
			return "", false
		}
	}

	if !d.active {
		d.active = true
		d.candidate = w
		d.since = at
		return "", false
	}

	if math.Abs(w-d.candidate) > d.cfg.Epsilon {
		d.candidate = w
		d.since = at
		return "", false
	}

	if at.Sub(d.since) < d.cfg.StableFor {
		return "", false
	}

	d.printed = true
	d.printedWeight = w
	return d.nextEPC24(at), true
}

func (d *StableEPCDetector) reset() {
	d.active = false
	d.printed = false
	d.candidate = 0
	d.since = time.Time{}
	d.printedWeight = 0
}

// nextEPC24 returns a 24-char uppercase hex EPC-like id:
// 30 + 14 hex chars (unix ns low 56-bit) + 8 hex chars (time-atom mix tail).
func (d *StableEPCDetector) nextEPC24(t time.Time) string {
	ns := t.UnixNano()
	if ns != d.lastNS {
		d.lastNS = ns
		d.seq = 0
	} else {
		d.seq++
	}
	return formatEPC24(ns, d.seq, d.salt)
}

func formatEPC24(ns int64, seq, salt uint32) string {
	atom := uint32((uint64(ns) / 1_000) & 0xFFFFFFFF)
	tail := atom ^ bits.RotateLeft32(uint32(ns), 13) ^ bits.RotateLeft32(seq, 7) ^ salt
	tail |= 1
	return fmt.Sprintf("30%014X%08X", uint64(ns)&0x00FFFFFFFFFFFFFF, tail)
}

func newEPCSalt() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err == nil {
		return binary.BigEndian.Uint32(b[:]) | 1
	}
	fallback := uint32(time.Now().UnixNano()) ^ (uint32(os.Getpid()) << 16)
	return fallback | 1
}
