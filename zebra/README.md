# Zebra USB Tool

`gscale-zebra/zebra` ichidagi ushbu util USB orqali ulangan Zebra printer bilan ishlaydi.

## Buyruqlar

```bash
cd /home/wikki/local.git/gscale-zebra/zebra

# 1) Printerlarni topish
go run . list

# 2) Status query (~HS)
go run . status

# 3) Minimal test print (1 dona)
go run . print-test --copies 1 --message "GSCALE TEST"

# 4) EPC encode test (default dry-run)
go run . epc-test --epc 3034257BF7194E4000000001
# real yuborish (1 tag):
go run . epc-test --epc 3034257BF7194E4000000001 --send

# 5) Calibration (avval dry-run tavsiya)
go run . calibrate --dry-run
go run . calibrate

# 6) Self check (status + ixtiyoriy 1 dona print)
go run . self-check --print
```

## Eslatma

- `print-test` RFID encode qilmaydi.
- `epc-test` default `dry-run`; real encode uchun `--send` kerak.
- `--send` rejimida default holatda EPC tagga yoziladi va o'sha EPC matni label ustiga ham print qilinadi (`--print-human=true`, `--feed=true`).
- `calibrate` bir nechta label/tag feed qilishi mumkin.
- Taglarni tejash uchun `print-test`da `--copies` 1..20, `epc-test` esa 1 taga cheklangan.
