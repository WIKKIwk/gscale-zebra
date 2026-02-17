package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		exitErr(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	updates := make(chan Reading, 32)
	var zebraUpdates <-chan ZebraStatus
	var sourceLine string
	var serialErr error
	started := false

	port, usedBaud, err := detectScalePort(cfg.device, cfg.bauds, cfg.probeTimeout, cfg.unit)
	if err == nil {
		if startErr := startSerialReader(ctx, port, usedBaud, cfg.unit, updates); startErr == nil {
			sourceLine = fmt.Sprintf("serial (%s @ %d)", port, usedBaud)
			started = true
		} else {
			serialErr = startErr
		}
	} else {
		serialErr = err
	}

	if !started && !cfg.disableBridge && strings.TrimSpace(cfg.bridgeURL) != "" {
		startBridgeReader(ctx, strings.TrimSpace(cfg.bridgeURL), cfg.bridgeInterval, updates)
		sourceLine = fmt.Sprintf("bridge (%s)", strings.TrimSpace(cfg.bridgeURL))
		started = true
	}

	if !started {
		if serialErr != nil {
			exitErr(serialErr)
		}
		exitErr(errors.New("scale source not available"))
	}

	if !cfg.disableZebra {
		zch := make(chan ZebraStatus, 16)
		startZebraMonitor(ctx, cfg.zebraDevice, cfg.zebraInterval, zch)
		zebraUpdates = zch
	}

	if !cfg.disableBot {
		if err := startBotProcess(ctx, cfg.botDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: bot auto-start bo'lmadi: %v\n", err)
		}
	}

	if err := runTUI(ctx, updates, zebraUpdates, sourceLine, cfg.zebraDevice, serialErr); err != nil {
		exitErr(err)
	}
}
