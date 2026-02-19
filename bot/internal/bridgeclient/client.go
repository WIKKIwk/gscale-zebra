package bridgeclient

import (
	bridgestate "bridge/state"
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// nextCycleDropFraction: WaitForNextCycle yangi sikl uchun zarur bo'lgan
// minimal vazn tushishi ulushi (35%). Kichik tebranishlar (78.800↔78.850 = 0.06%)
// yangi siklni boshlamasligi uchun ishlatiladi.
const nextCycleDropFraction = 0.35

type Client struct {
	store *bridgestate.Store
}

type StableReading struct {
	Qty       float64
	Unit      string
	UpdatedAt time.Time
}

type EPCReading struct {
	EPC       string
	Verify    string
	ReadLine1 string
	ReadLine2 string
	UpdatedAt time.Time
}

func New(path string) *Client {
	return &Client{store: bridgestate.New(path)}
}

func (c *Client) WaitStablePositive(ctx context.Context, timeout, pollInterval time.Duration) (float64, string, error) {
	r, err := c.WaitStablePositiveReading(ctx, timeout, pollInterval)
	if err != nil {
		return 0, "", err
	}
	return r.Qty, r.Unit, nil
}

func (c *Client) WaitStablePositiveReading(ctx context.Context, timeout, pollInterval time.Duration) (StableReading, error) {
	if c == nil || c.store == nil || strings.TrimSpace(c.store.Path()) == "" {
		return StableReading{}, fmt.Errorf("bridge state path bo'sh")
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if pollInterval <= 0 {
		pollInterval = 220 * time.Millisecond
	}

	deadline := time.Now().Add(timeout)
	var lastWeight float64
	var haveLast bool
	stableCount := 0

	for {
		if time.Now().After(deadline) {
			return StableReading{}, fmt.Errorf("scale qty timeout (%s)", timeout)
		}
		select {
		case <-ctx.Done():
			return StableReading{}, ctx.Err()
		default:
		}

		snap, err := c.store.Read()
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}
		s := snap.Scale
		if strings.TrimSpace(s.Error) != "" {
			haveLast = false
			stableCount = 0
			time.Sleep(pollInterval)
			continue
		}
		if s.Weight == nil || *s.Weight <= 0 {
			haveLast = false
			stableCount = 0
			time.Sleep(pollInterval)
			continue
		}
		updatedAt, ok := parseSnapshotTime(s.UpdatedAt)
		if !ok || !isFreshTime(updatedAt, 4*time.Second) {
			time.Sleep(pollInterval)
			continue
		}

		w := *s.Weight
		if s.Stable != nil && *s.Stable {
			return StableReading{Qty: w, Unit: normalizeUnit(s.Unit), UpdatedAt: updatedAt}, nil
		}

		if haveLast && almostEqual(lastWeight, w, 0.001) {
			stableCount++
		} else {
			stableCount = 1
		}
		haveLast = true
		lastWeight = w

		if stableCount >= 4 {
			return StableReading{Qty: w, Unit: normalizeUnit(s.Unit), UpdatedAt: updatedAt}, nil
		}
		time.Sleep(pollInterval)
	}
}

func (c *Client) WaitEPCForReading(ctx context.Context, timeout, pollInterval time.Duration, after time.Time, lastEPC string) (EPCReading, error) {
	if c == nil || c.store == nil || strings.TrimSpace(c.store.Path()) == "" {
		return EPCReading{}, fmt.Errorf("bridge state path bo'sh")
	}
	if timeout <= 0 {
		timeout = 6 * time.Second
	}
	if pollInterval <= 0 {
		pollInterval = 140 * time.Millisecond
	}
	lastEPC = strings.ToUpper(strings.TrimSpace(lastEPC))

	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return EPCReading{}, fmt.Errorf("epc timeout (%s)", timeout)
		}
		select {
		case <-ctx.Done():
			return EPCReading{}, ctx.Err()
		default:
		}

		snap, err := c.store.Read()
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}

		epc := strings.ToUpper(strings.TrimSpace(snap.Zebra.LastEPC))
		if epc == "" || epc == lastEPC {
			time.Sleep(pollInterval)
			continue
		}

		epcAt, ok := parseSnapshotTime(snap.Zebra.UpdatedAt)
		if ok {
			if !isFreshTime(epcAt, 15*time.Second) {
				time.Sleep(pollInterval)
				continue
			}
			if !after.IsZero() && epcAt.Before(after.Add(-300*time.Millisecond)) {
				time.Sleep(pollInterval)
				continue
			}
		}

		verify := strings.ToUpper(strings.TrimSpace(snap.Zebra.Verify))
		if verify == "" {
			verify = "UNKNOWN"
		}

		return EPCReading{
			EPC:       epc,
			Verify:    verify,
			ReadLine1: strings.TrimSpace(snap.Zebra.ReadLine1),
			ReadLine2: strings.TrimSpace(snap.Zebra.ReadLine2),
			UpdatedAt: epcAt,
		}, nil
	}
}

// WaitForNextCycle returns when scale goes to reset (<=0) OR weight changes enough
// from last processed qty. This prevents batch from getting stuck when operator
// replaces product without hitting exact zero.
func (c *Client) WaitForNextCycle(ctx context.Context, timeout, pollInterval time.Duration, lastQty float64) error {
	if c == nil || c.store == nil || strings.TrimSpace(c.store.Path()) == "" {
		return fmt.Errorf("bridge state path bo'sh")
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	if pollInterval <= 0 {
		pollInterval = 220 * time.Millisecond
	}

	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("scale next-cycle timeout (%s)", timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		snap, err := c.store.Read()
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}
		s := snap.Scale
		if !isFreshSnapshot(s.UpdatedAt, 4*time.Second) {
			time.Sleep(pollInterval)
			continue
		}
		if s.Weight == nil || *s.Weight <= 0 {
			return nil
		}
		// Vazn 35%+ kamaysa yangi sikl boshlangan (item olib tashlangan).
		// Kichik tebranishlar (78.800↔78.850 = 0.06%) yangi siklni boshlamaydi.
		if lastQty > 0 && *s.Weight < lastQty*(1-nextCycleDropFraction) {
			return nil
		}

		time.Sleep(pollInterval)
	}
}

func isFreshSnapshot(updated string, maxAge time.Duration) bool {
	ts, ok := parseSnapshotTime(updated)
	if !ok {
		return false
	}
	return isFreshTime(ts, maxAge)
}

func parseSnapshotTime(updated string) (time.Time, bool) {
	updated = strings.TrimSpace(updated)
	if updated == "" {
		return time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return time.Time{}, false
	}
	return ts, true
}

func isFreshTime(ts time.Time, maxAge time.Duration) bool {
	age := time.Since(ts)
	if age < 0 {
		age = 0
	}
	return age <= maxAge
}

func normalizeUnit(v string) string {
	u := strings.TrimSpace(v)
	if u == "" {
		return "kg"
	}
	return u
}

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}
