package main

import (
	"context"
	"time"
)

func startZebraMonitor(ctx context.Context, preferredDevice string, interval time.Duration, out chan<- ZebraStatus) {
	if out == nil {
		return
	}
	if interval < 300*time.Millisecond {
		interval = 300 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		publishZebraStatus(out, collectZebraStatus(preferredDevice, 900*time.Millisecond))
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				publishZebraStatus(out, collectZebraStatus(preferredDevice, 900*time.Millisecond))
			}
		}
	}()
}

func collectZebraStatus(preferredDevice string, timeout time.Duration) ZebraStatus {
	zebraIOMutex.Lock()
	defer zebraIOMutex.Unlock()

	st := ZebraStatus{
		Connected: false,
		Verify:    "-",
		UpdatedAt: time.Now(),
	}

	p, err := SelectZebraPrinter(preferredDevice)
	if err != nil {
		st.Error = err.Error()
		return st
	}

	st.Connected = true
	st.DevicePath = p.DevicePath
	st.Name = p.DisplayName()
	st.DeviceState = safeText("-", queryVarRetry(p.DevicePath, "device.status", timeout, 3, 90*time.Millisecond))
	st.MediaState = safeText("-", queryVarRetry(p.DevicePath, "media.status", timeout, 3, 90*time.Millisecond))
	st.ReadLine1 = "-"
	st.ReadLine2 = "-"
	st.Verify = "-"
	return st
}

func applyZebraSnapshot(st *ZebraStatus, p ZebraPrinter, timeout time.Duration) {
	st.DeviceState = safeText("-", queryVarRetry(p.DevicePath, "device.status", timeout, 3, 90*time.Millisecond))
	st.MediaState = safeText("-", queryVarRetry(p.DevicePath, "media.status", timeout, 3, 90*time.Millisecond))
	st.ReadLine1 = "-"
	st.ReadLine2 = "-"
}

func publishZebraStatus(ch chan<- ZebraStatus, st ZebraStatus) {
	select {
	case ch <- st:
	default:
	}
}
