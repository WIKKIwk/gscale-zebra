SHELL := /bin/sh

COMPOSE ?= docker compose
SCALE_DEVICE ?= /dev/ttyUSB0
ZEBRA_DEVICE ?= /dev/usb/lp0

DC_ENV := SCALE_DEVICE=$(SCALE_DEVICE) ZEBRA_DEVICE=$(ZEBRA_DEVICE)

.PHONY: help check-env build run run-bg down restart ps logs-bot logs-all clean

help:
	@echo "Targets:"
	@echo "  make run       - docker orqali scale TUI + bot parallel (TUI attach)"
	@echo "  make run-bg    - stackni backgroundda ishga tushirish"
	@echo "  make build     - image build"
	@echo "  make down      - containerlarni to'xtatish/o'chirish"
	@echo "  make restart   - scale va bot restart"
	@echo "  make ps        - container holati"
	@echo "  make logs-bot  - bot loglarini live ko'rish"
	@echo "  make logs-all  - barcha loglar"
	@echo ""
	@echo "Override:"
	@echo "  make run SCALE_DEVICE=/dev/ttyUSB1 ZEBRA_DEVICE=/dev/usb/lp0"

check-env:
	@test -f bot/.env || (echo "xato: bot/.env topilmadi (bot/.env.example dan nusxa oling)"; exit 1)

build:
	$(DC_ENV) $(COMPOSE) build

run: check-env
	@set -e; \
	cleanup() { $(COMPOSE) stop bot >/dev/null 2>&1 || true; }; \
	trap cleanup EXIT INT TERM; \
	$(DC_ENV) $(COMPOSE) up -d --build --no-deps bot; \
	$(DC_ENV) $(COMPOSE) run --rm --no-deps --service-ports scale

run-bg: check-env
	$(DC_ENV) $(COMPOSE) up -d --build

down:
	$(COMPOSE) down --remove-orphans

restart:
	$(COMPOSE) restart scale bot

ps:
	$(COMPOSE) ps

logs-bot:
	$(COMPOSE) logs -f bot

logs-all:
	$(COMPOSE) logs -f

clean: down
	$(COMPOSE) down -v --remove-orphans
