# Scale monitor (TUI + Zebra) 📟

`scale` moduli USB serial tarozi oqimini o'qiydi, Zebra holatini kuzatadi va auto EPC encode oqimini boshqaradi.

## Ishga tushirish

```bash
cd /home/wikki/local.git/gscale-zebra/scale
go run .
```

TUI tugmalari:

- `q` - chiqish
- `e` - qo'lda encode+print yuborish
- `r` - RFID read yuborish

## Boot'da auto-start (systemd) 🚀

Repo root'dan:

```bash
cd /home/wikki/local.git/gscale-zebra
make autostart-install
```

Bu `scale` va `bot` service'larini systemd'ga o'rnatadi, enable qiladi va start beradi.

## Ishlash oqimi

1. Serial port auto-detect qilinadi (`/dev/serial/by-id/*`, `ttyUSB*`, `ttyACM*`).
2. Agar serial ishga tushmasa, HTTP bridge fallback ishlatiladi (`--bridge-url`).
3. Har reading bridge snapshot'ga yoziladi (`scale` + `zebra`).
4. `batch.active=true` bo'lsa auto encode ishlaydi, aks holda to'xtaydi.
5. Stable qty topilganda EPC yaratiladi va Zebra encode command yuboriladi.

## Auto EPC qoidalari (joriy kod)

- Faqat `> 0` qty uchun ishlaydi.
- Qty taxminan `1s` barqaror tursa trigger bo'ladi.
- Jitter filtri: epsilon `0.005`.
- Bir xil nuqtada qayta-qayta trigger qilmaydi.
- Yangi sikl ochilishi uchun qty oxirgi printed nuqtadan ma'noli o'zgarishi kerak.
- Qty o'zgarib keyin oldingi qiymatga qaytsa ham, yana stable bo'lsa yangi EPC chiqadi.

## Batch gate (`bridge_state.json`)

- default fayl: `/tmp/gscale-zebra/bridge_state.json`
- flag: `--bridge-state-file /tmp/gscale-zebra/bridge_state.json`

Bot tomonda:

- `Material Issue` => `batch.active=true`
- `Batch Stop` => `batch.active=false`

Scale TUI bu holatni `BATCH: ACTIVE/STOPPED` sifatida ko'rsatadi.

## Bot auto-start

Default holatda scale ichidan bot ham ko'tariladi (`go run ./cmd/bot` in `--bot-dir`).

O'chirish:

```bash
go run . --no-bot
```

## Parametrlar

- `--device` (example: `/dev/ttyUSB0`) - serial device'ni qo'lda berish
- `--baud` (default: `9600`) - asosiy baud
- `--baud-list` (default: `9600,19200,38400,57600,115200`) - detect uchun baudlar
- `--probe-timeout` (default: `800ms`) - port probe timeout
- `--unit` (default: `kg`) - default birlik
- `--bridge-url` (default: `http://127.0.0.1:18000/api/v1/scale`) - fallback endpoint
- `--bridge-interval` (default: `120ms`) - fallback poll interval
- `--no-bridge` - HTTP fallback'ni o'chiradi
- `--zebra-device` (example: `/dev/usb/lp0`) - printer path
- `--zebra-interval` (default: `900ms`) - Zebra monitor interval
- `--no-zebra` - Zebra monitor va `e/r` actionlarni o'chiradi
- `--bot-dir` (default: `../bot`) - bot modul yo'li
- `--no-bot` - bot auto-startni o'chiradi
- `--bridge-state-file` - shared snapshot fayli

## Loglar

Worker loglari `../logs/scale/` ichiga yoziladi.
Har restartda `logs/scale/` tozalanib, yangi sessiya boshlanadi.
