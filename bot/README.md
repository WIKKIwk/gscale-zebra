# bot

Telegram bot (`/start`, `/batch`, `/image`) orqali ERPNext va Zebra bilan ishlaydi.

## Ishga tushirish

```bash
cd /home/wikki/local.git/gscale-zebra/bot
cp .env.example .env
# .env ichiga real TELEGRAM_BOT_TOKEN yozing
go run ./cmd/bot
```

## Buyruqlar

- `/start`
  - ERPNext API auth tekshiradi.
  - Ulangan bo'lsa `ERPNext ga ulandi` xabarini yuboradi.

- `/batch`
  - `Item tanlang:` xabarini inline tugma bilan yuboradi.
  - Tugma bosilganda current chat'da Telegram inline menu ochiladi.
  - Inline query natijalarida ERPNext'dagi itemlar chiqadi.

- `/image`
  - `Rasm tashlang.` xabarini yuboradi.
  - User rasm yuborsa, bot rasmni label o'lchamiga sig'dirib Zebra printerga chiqaradi.

## Bridge state

Bot batch session holatini shared bridge faylga yozadi:

- default: `/tmp/gscale-zebra/bridge_state.json`
- `.env`: `BRIDGE_STATE_FILE=...`

`Material Issue` bosilganda `batch.active=true`, `Batch Stop` bosilganda `batch.active=false` bo'ladi.
Scale TUI shu holatga qarab auto printni yoqadi/o'chiradi.

## Printer sozlamalari (`/image`)

- `PRINTER_DEVICE` default: `/dev/usb/lp0`
- `LABEL_WIDTH_DOTS` default: `560`
- `LABEL_HEIGHT_DOTS` default: `320`

## Eslatma

Inline menu ishlashi uchun botda inline mode yoqilgan bo'lishi kerak (`@BotFather` -> `/setinline`).
