# bot

Telegram bot (`/start`) orqali ERPNext ulanishini tekshiradi.

## Ishga tushirish

```bash
cd /home/wikki/local.git/gscale-zebra/bot
cp .env.example .env
# .env ichiga real TELEGRAM_BOT_TOKEN yozing
go run ./cmd/bot
```

## Hozirgi vazifa

- `/start` kelganda ERPNext API bilan auth tekshiradi.
- Muvaffaqiyatli bo'lsa foydalanuvchiga `ERPNext ga ulandi` degan xabar yuboradi.
- Xato bo'lsa xato matnini yuboradi.
