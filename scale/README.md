# Scale monitor (Go)

USB serial tarozi qiymatini terminalda (TUI) ko'rsatadi va Zebra RFID printer holatini ham monitor qiladi.

## Ishga tushirish

```bash
cd scale
go run .
```

Keys:
- `q` - chiqish
- `e` - Zebra'ga test EPC encode + print
- `r` - RFID read

## Bot Auto-Start

Default holatda TUI ishga tushganda Telegram bot ham birga ishga tushadi (`../bot` dan `go run ./cmd/bot`).

Botni o'chirish:
```bash
go run . --no-bot
```

## Bridge state orqali boshqaruv

Scale `bridge_state.json` ga live `scale` va `zebra` holatini yozadi.

- default: `/tmp/gscale-zebra/bridge_state.json`
- flag: `--bridge-state-file /tmp/gscale-zebra/bridge_state.json`

Batch gate:
- botda `Material Issue` -> `batch.active=true` -> auto print ishlaydi
- botda `Batch Stop` -> `batch.active=false` -> auto print to'xtaydi

TUI ichida ham `BATCH: ACTIVE/STOPPED` ko'rinadi.

## Parametrlar

- `--device /dev/ttyUSB0` - tarozi qurilmasini qo'lda berish
- `--baud 9600` - asosiy baudrate
- `--baud-list 9600,19200,38400,57600,115200` - detect payti probelar
- `--probe-timeout 800ms` - probe davomiyligi
- `--unit kg` - default birlik
- `--bridge-url http://127.0.0.1:18000/api/v1/scale` - fallback HTTP endpoint
- `--bridge-interval 250ms` - fallback poll interval
- `--no-bridge` - HTTP fallback'ni o'chirish
- `--zebra-device /dev/usb/lp0` - Zebra printer yo'lini qo'lda berish
- `--zebra-interval 900ms` - Zebra monitor poll interval
- `--no-zebra` - Zebra monitoring va `e/r` actionlarni o'chirish
- `--bot-dir ../bot` - bot modul joylashuvi
- `--no-bot` - bot auto-start ni o'chirish
- `--bridge-state-file /tmp/gscale-zebra/bridge_state.json` - shared state fayli

`--batch-state-file` hali ishlaydi, lekin deprecated alias.
