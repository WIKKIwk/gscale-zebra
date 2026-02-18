# bridge

`bridge` moduli scale, zebra va bot o'rtasidagi umumiy holatni bitta faylga yig'adi.

Default state file:
- `/tmp/gscale-zebra/bridge_state.json`

Saqlanadigan bo'limlar:
- `scale` - live qty, stable, error, source, port
- `zebra` - oxirgi EPC, verify, printer holati
- `batch` - bot batch active/stop holati

Maqsad:
- `qty.json` va `batch_state.json` kabi alohida fayllarni o'qib-yurishni kamaytirish
- bot + scale avtomatizatsiyasini bitta kanalga to'plash
- race/xatolik ehtimolini pasaytirish (atomic update + file lock)
