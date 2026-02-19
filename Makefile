SHELL := /bin/sh

MKFILE_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
COMPOSE ?= docker compose -f $(MKFILE_DIR)docker-compose.yml
SCALE_DEVICE ?= /dev/ttyUSB0
ZEBRA_DEVICE ?= /dev/usb/lp0

DC_ENV := SCALE_DEVICE=$(SCALE_DEVICE) ZEBRA_DEVICE=$(ZEBRA_DEVICE)

.PHONY: help check-env build run run-bg down restart ps logs-bot logs-all clean

help:
	@echo "Targets:"
	@echo "  make run       - eskisini o'chirib, cache'siz qayta build va run (TUI attach)"
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
	docker ps -aq --filter ancestor=gscale-zebra:local | xargs -r docker rm -f 2>/dev/null || true; \
	$(COMPOSE) down --remove-orphans >/dev/null 2>&1 || true; \
	$(DC_ENV) $(COMPOSE) build --no-cache; \
	$(DC_ENV) $(COMPOSE) up -d --no-deps --force-recreate bot; \
	$(DC_ENV) $(COMPOSE) run --rm --no-deps --service-ports scale

run-bg: check-env
	@echo "==> Barcha eski gscale-zebra containerlarni o'chirish..."
	@docker ps -aq --filter ancestor=gscale-zebra:local | xargs -r docker rm -f 2>/dev/null || true
	@$(COMPOSE) down --remove-orphans 2>/dev/null || true
	@echo "==> Build..."
	$(DC_ENV) $(COMPOSE) build --no-cache
	@echo "==> Run..."
	$(DC_ENV) $(COMPOSE) up -d --force-recreate

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
