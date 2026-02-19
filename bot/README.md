# Bot (Telegram + ERPNext) 🤖

`bot` moduli Telegram orqali batch workflow'ni boshqaradi va ERPNext'ga draft yaratadi.

## Ishga tushirish

```bash
cd /home/wikki/local.git/gscale-zebra/bot
cp .env.example .env
# .env ichiga real token va ERP credentials yozing
go run ./cmd/bot
```

## Qo'llab-quvvatlanadigan buyruqlar

- `/start` - ERPNext ulanishini tekshiradi va botni tayyor holatga olib keladi.
- `/batch` - batch oqimini boshlash uchun item/ombor tanlash jarayonini ochadi.
- `/image` - rasm yuborilganda Zebra printerga image print qiladi.
- `/log` - `logs/bot` va `logs/scale` fayllarini Telegram chatga yuboradi.

## Batch workflow (hozirgi amaliy oqim) ✅

1. `/batch` beriladi.
2. `Item tanlash` inline tugmasi orqali ERP item tanlanadi.
3. `Ombor tanlash` inline tugmasi orqali warehouse tanlanadi.
4. Bot `Material Issue` yoki `Receipt` tugmalarini ko'rsatadi.
5. `Material Issue` bosilganda batch session ishga tushadi.
6. Scale'dan `stable + musbat qty` keladi.
7. Zebra'dan EPC olinadi va `VERIFY` tekshiriladi.
8. Faqat `VERIFY=MATCH|OK|WRITTEN` bo'lsa ERPNext draft yaratiladi.

`Receipt` hozir placeholder holatda (`tez orada qo'shiladi`).

## Batch boshqaruv tugmalari

- `Item almashtirish` - joriy batchni pause qiladi va yangi item tanlashga qaytaradi.
- `Batch Start` - tanlangan item/ombor bilan batchni qayta boshlaydi.
- `Batch Stop` - batchni to'xtatadi (`batch.active=false`).

## Bridge state

Shared snapshot fayl:

- default: `/tmp/gscale-zebra/bridge_state.json`
- config: `BRIDGE_STATE_FILE`

Bot `batch` bo'limini shu faylga yozadi, `scale` esa shu holatga qarab auto-print gate qiladi.

## Konfiguratsiya (`.env`)

Majburiy:

- `TELEGRAM_BOT_TOKEN`
- `ERP_URL`
- `ERP_API_KEY`
- `ERP_API_SECRET`

Ixtiyoriy/asosiy:

- `BRIDGE_STATE_FILE` (default: `/tmp/gscale-zebra/bridge_state.json`)
- `PRINTER_DEVICE` (default: `/dev/usb/lp0`)
- `LABEL_WIDTH_DOTS` (default: `560`)
- `LABEL_HEIGHT_DOTS` (default: `320`)

## Loglar

Ish jarayonida worker loglari `../logs/bot/` ichiga yoziladi.
Har restartda `logs/bot/` tozalanib, yangi sessiya boshlanadi.

## Eslatma

Inline qidiruv ishlashi uchun `@BotFather -> /setinline` yoqilgan bo'lishi kerak.
