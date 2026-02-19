SHELL := /bin/sh

SCALE_DEVICE ?= /dev/ttyUSB0
ZEBRA_DEVICE ?= /dev/usb/lp0
BRIDGE_STATE_FILE ?= /tmp/gscale-zebra/bridge_state.json

.PHONY: help check-env build build-bot build-scale build-zebra run run-scale run-bot test clean release release-all

help:
	@echo "Targets:"
	@echo "  make run        - scale TUI ni ishga tushiradi (bot auto-start bilan)"
	@echo "  make run-scale  - faqat scale TUI (bot auto-startsiz)"
	@echo "  make run-bot    - faqat telegram bot"
	@echo "  make build      - bot + scale + zebra binary build (./bin)"
	@echo "  make test       - barcha modullarda test"
	@echo "  make release    - linux/amd64 tar release"
	@echo "  make release-all - linux/amd64 + linux/arm64 tar release"
	@echo "  make clean      - local build papkalarini tozalash"
	@echo ""
	@echo "Override:"
	@echo "  make run SCALE_DEVICE=/dev/ttyUSB1 ZEBRA_DEVICE=/dev/usb/lp0"

check-env:
	@test -f bot/.env || (echo "xato: bot/.env topilmadi (bot/.env.example dan nusxa oling)"; exit 1)

build: build-bot build-scale build-zebra

build-bot:
	@mkdir -p bin
	go build -o ./bin/bot ./bot/cmd/bot

build-scale:
	@mkdir -p bin
	go build -o ./bin/scale ./scale

build-zebra:
	@mkdir -p bin
	go build -o ./bin/zebra ./zebra

run: check-env
	cd scale && go run . --no-bridge --device "$(SCALE_DEVICE)" --zebra-device "$(ZEBRA_DEVICE)" --bridge-state-file "$(BRIDGE_STATE_FILE)"

run-scale:
	cd scale && go run . --no-bot --no-bridge --device "$(SCALE_DEVICE)" --zebra-device "$(ZEBRA_DEVICE)" --bridge-state-file "$(BRIDGE_STATE_FILE)"

run-bot: check-env
	cd bot && go run ./cmd/bot

test:
	cd bot && go test ./...
	cd bridge && go test ./...
	cd scale && go test ./...
	cd core && GOWORK=off go test ./...

clean:
	rm -rf ./bin ./dist

release:
	./scripts/release.sh --arch amd64

release-all:
	./scripts/release.sh --arch amd64 --arch arm64
