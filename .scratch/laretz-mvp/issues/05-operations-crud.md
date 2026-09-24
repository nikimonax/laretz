# 05: Операции: CRUD в портфеле

**What to build:** Пользователь вручную заносит сделки: модалка на странице портфеля. Тройка price×qty=total решается JS на странице; сервер получает ровно `quantity + total_usd`, поле price игнорирует. Полный CRUD с независимой пагинацией таблицы операций. Страница `/portfolio/:id` появляется здесь (позиции и график — плейсхолдеры). См. спеку: «Зафиксированные решения», «Формы и страницы», ADR-0003.

**Blocked by:** 02 (Портфели: CRUD end-to-end + sqlite фундамент), 04 (Импорт Assets: поиск по каталогу — нужен селектор актива).

**Status:** ready-for-agent

- [ ] Форма операции: тип (Buy/Sell), Asset (select), время (`datetime-local`), quantity + total_usd, fee (amount + Asset, nullable), note
- [ ] JS уравнивает тройку на странице; сервер читает только qty + total_usd; при выключенном JS форма остаётся рабочей
- [ ] Валидация: время не в будущем, quantity и total_usd > 0; децималы shopspring, в БД TEXT
- [ ] Таблица операций на `/portfolio/:id`: сортировка по времени desc, независимая пагинация `?ops_page=`
- [ ] Редактирование с префиллом (GetOperation), удаление — spoofing + `<dialog>`; минус-балансы не валидируются
- [ ] Команды Create/Update/DeleteOperation + GetOperation с тестами
