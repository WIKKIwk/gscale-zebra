# Scale monitor (Go)

USB serial tarozi qiymatini terminalda (TUI) ko'rsatadi va Zebra RFID printer holatini ham monitor qiladi.

## Aniqlangan holat

Hozirgi kompyuterda tarozi `CH340` sifatida ko'rindi va `/dev/ttyUSB0` ga tushgan:

- `lsusb`: `1a86:7523 QinHeng CH340 serial converter`
- `by-id`: `usb-1a86_USB2.0-Ser_-if00-port0 -> /dev/ttyUSB0`

## Ishga tushirish

```bash
cd scale
go run .
```

Chiqish:

- `q` - chiqish
- `e` - Zebra'ga 1 ta test EPC encode + readback urinish
- `r` - faqat RFID read urinish

## Muhim

Hozir `/dev/ttyUSB0` ni `ZebraBridge.Web` jarayoni ishlatayotgani aniqlandi, shu sabab dastur avtomatik ravishda HTTP bridge fallback'ga o'tadi:

- `http://127.0.0.1:18000/api/v1/scale`

Agar to'g'ridan-to'g'ri serial o'qimoqchi bo'lsangiz:

```bash
go run . --no-bridge
```

Agar port band bo'lmasa, auto-detect bilan serialdan bevosita o'qiydi.

## Bot Auto-Start

Default holatda TUI ishga tushganda Telegram bot ham birga ishga tushadi (`../bot` papkadan `go run ./cmd/bot`).

Agar botni o'chirmoqchi bo'lsangiz:

```bash
go run . --no-bot
```

Agar bot papkasi boshqa joyda bo'lsa:

```bash
go run . --bot-dir /home/wikki/local.git/gscale-zebra/bot
```

## Batch stop bilan auto print boshqaruvi

Auto EPC/print oqimi `batch_state.json` orqali boshqariladi.

- Botda `Material Issue` bosilganda `active=true` bo'ladi va scale auto print ishlaydi.
- Botda `Batch Stop` bosilganda `active=false` bo'ladi va scale endi zebra'ga auto print yubormaydi.

## Parametrlar

- `--device /dev/ttyUSB0` - tarozi qurilmasini qo'lda berish
- `--baud 9600` - asosiy baudrate
- `--baud-list 9600,19200,38400,57600,115200` - detect payti probelar
- `--probe-timeout 800ms` - probe davomiyligi
- `--unit kg` - default birlik
- `--bridge-url http://127.0.0.1:18000/api/v1/scale` - fallback endpoint
- `--bridge-interval 250ms` - fallback poll interval
- `--no-bridge` - fallback'ni o'chirish
- `--zebra-device /dev/usb/lp0` - Zebra printer yo'lini qo'lda berish
- `--zebra-interval 900ms` - Zebra monitor poll interval
- `--no-zebra` - Zebra monitoring va `e/r` actionlarni o'chirish
- `--bot-dir ../bot` - bot modul joylashuvi
- `--no-bot` - bot auto-start ni o'chirish
- `--qty-file /tmp/gscale-zebra/qty.json` - bot o'qiydigan shared qty fayli
- `--batch-state-file /tmp/gscale-zebra/batch_state.json` - bot boshqaradigan batch holati fayli

## EPC tasdiq haqida

Printer ichidan `rfid.tag.read.result_line1/line2` orqali readback olinadi.

- `MATCH` - qaytgan qatorda kutilgan EPC topildi
- `NO TAG` - printer o'qish zonasida tag topmadi
- `MISMATCH` - javob bor, lekin kutilgan EPC topilmadi
- `UNKNOWN` - javob bo'sh yoki noaniq

Amaliyotda 100% yakuniy tasdiq uchun tashqi RFID reader bilan ham tekshirish tavsiya qilinadi.
