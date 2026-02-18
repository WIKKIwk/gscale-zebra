# core

`core` ichida scale -> zebra avtomatik EPC logikasi saqlanadi.

Qoidalar:
- QTY 1 soniya davomida bir nuqtada barqaror qolsa EPC generatsiya qilinadi.
- Faqat musbat (`> 0`) qiymatlar uchun ishlaydi.
- `0` yoki minus bo'lsa EPC hech qachon yuborilmaydi.
- Har bir yangi barqaror nuqta uchun yangi (unikal) EPC yaratiladi.
- EPC 24 xonali hex formatda yaratiladi.
- EPC oxiri oddiy `00000000` emas: vaqt atomi (`unix nano`) + sequence + process `salt` aralashmasidan olinadi.
- Shu sabab dastur qayta ishga tushganda ham collision ehtimoli ancha past bo'ladi.
