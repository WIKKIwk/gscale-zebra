package main

import (
	"context"
	"strings"
	"time"

	"github.com/tarm/serial"
)

func startSerialReader(ctx context.Context, device string, baud int, unit string, out chan<- Reading) error {
	port, err := serial.OpenPort(&serial.Config{Name: device, Baud: baud, ReadTimeout: 250 * time.Millisecond})
	if err != nil {
		return err
	}

	go func() {
		defer port.Close()
		buf := make([]byte, 256)
		raw := ""

		push(out, Reading{
			Source:    "serial",
			Port:      device,
			Baud:      baud,
			Unit:      unit,
			UpdatedAt: time.Now(),
		})

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := port.Read(buf)
			if err != nil {
				push(out, Reading{
					Source:    "serial",
					Port:      device,
					Baud:      baud,
					Unit:      unit,
					Error:     err.Error(),
					UpdatedAt: time.Now(),
				})
				return
			}
			if n == 0 {
				continue
			}

			chunk := string(buf[:n])
			if strings.TrimSpace(chunk) == "" {
				continue
			}

			raw = appendRaw(raw, chunk, 240)
			weight, parsedUnit, stable, ok := parseWeight(raw, unit)
			if !ok {
				continue
			}
			w := weight
			push(out, Reading{
				Source:    "serial",
				Port:      device,
				Baud:      baud,
				Weight:    &w,
				Unit:      parsedUnit,
				Stable:    stable,
				Raw:       raw,
				UpdatedAt: time.Now(),
			})
		}
	}()

	return nil
}
