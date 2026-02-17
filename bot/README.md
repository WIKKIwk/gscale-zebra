# bot

Telegram bot (`/start`, `/batch`) orqali ERPNext bilan ishlaydi.

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

## Eslatma

Inline menu ishlashi uchun botda inline mode yoqilgan bo'lishi kerak (`@BotFather` -> `/setinline`).
