# Zebra USB Tool 🖨️

`zebra` utili USB printer bilan diagnostika, SGD query/set va RFID encode testlarini qiladi.

## Ishga tushirish

```bash
cd /home/wikki/local.git/gscale-zebra/zebra
go run . help
```

## Asosiy buyruqlar

### 1) Printerlarni topish

```bash
go run . list
```

### 2) Host status (`~HS`)

```bash
go run . status --device /dev/usb/lp0
```

### 3) Sozlamalarni o'qish (`SGD getvar`)

```bash
go run . settings --device /dev/usb/lp0
go run . settings --device /dev/usb/lp0 --key print.width --key rfid.enable
```

### 4) Sozlama yozish (`SGD setvar`)

```bash
go run . setvar --device /dev/usb/lp0 --key ezpl.print_width --value 832 --save=true
```

### 5) Raw getvar (tez tekshiruv)

```bash
go run . raw-getvar --device /dev/usb/lp0 --key rfid.tag.read.result_line1 --count 1
```

### 6) Oddiy print test (RFIDsiz)

```bash
go run . print-test --device /dev/usb/lp0 --message "GSCALE TEST" --copies 1
go run . print-test --dry-run
```

### 7) RFID EPC encode test

```bash
# default: DRY-RUN (xavfsiz preview)
go run . epc-test --device /dev/usb/lp0 --epc 3034257BF7194E4000000001

# real encode (tag sarflaydi)
go run . epc-test --device /dev/usb/lp0 --epc 3034257BF7194E4000000001 --send
```

Foydali flaglar:

- `--auto-tune` - NO TAG/ERROR bo'lsa kalibratsiya+retry
- `--profile-init` - encode oldidan RFID profil o'rnatish
- `--profile-calibrate` - `rfid.tag.calibrate` bajarish
- `--label-tries` - `rfid.label_tries`
- `--error-handling` - `none|pause|error`
- `--read-power` / `--write-power` - RF power tuning

### 8) EPC o'qish testi

```bash
go run . read-epc --device /dev/usb/lp0 --tries 12
go run . read-epc --device /dev/usb/lp0 --expected 3034257BF7194E4000000001
```

### 9) Calibration

```bash
go run . calibrate --device /dev/usb/lp0 --dry-run
go run . calibrate --device /dev/usb/lp0 --save=true
```

### 10) Self-check

```bash
go run . self-check --device /dev/usb/lp0
go run . self-check --device /dev/usb/lp0 --print
```

## Muhim eslatmalar

- `epc-test` default holatda `--send=false` (ya'ni DRY-RUN).
- Real encode faqat `--send` bilan ketadi.
- `print-test` RFID encode qilmaydi, faqat oddiy label.
- `print-test --copies` maksimum `20`.
- `epc-test` xavfsizlik uchun har urinishda `1 tag` bilan ishlaydi.
- `calibrate` bir nechta label/tag feed qilishi mumkin.
- Qurilma band bo'lsa (`busy`) boshqa process printer portini ishlatayotgan bo'lishi mumkin.
