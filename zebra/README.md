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

# 4) Calibration (avval dry-run tavsiya)
go run . calibrate --dry-run
go run . calibrate

# 5) Self check (status + ixtiyoriy 1 dona print)
go run . self-check --print
```

## Eslatma

- `print-test` RFID encode qilmaydi.
- `calibrate` bir nechta label/tag feed qilishi mumkin.
- Taglarni tejash uchun `--copies` 1..3 oralig'ida cheklangan.
