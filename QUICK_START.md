# 🚀 QUICK START - Быстрый старт

## 3️⃣ Команды для запуска

### Быстрый запуск (разработка)
```bash
go run ./cmd/server
```

### Собрать и запустить
```bash
make run
```

### Режим разработки (рекомендуется)
```bash
make dev
```

**Сервер запустится на:** `http://localhost:8080`

---

## 📂 Структура за 30 секунд

```
cmd/server/main.go        ← Точка входа
internal/models/          ← Структуры данных
internal/database/        ← Работа с БД
internal/hub/             ← Управление WebSocket
internal/server/          ← HTTP сервер
static/                   ← Веб интерфейс
```

---

## 🔗 Основные ссылки

| Документ | Описание |
|----------|---------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | 📖 Полная архитектура |
| [SETUP_GUIDE.md](SETUP_GUIDE.md) | 🛠️ Полное руководство |
| [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) | 📁 Структура файлов |
| [README.md](README.md) | 📚 Основной README |

---

## 💻 API в кратце

```
GET  /ping                    → pong
GET  /stats                   → Статистика сервера
GET  /history?room=X&limit=Y  → История сообщений
GET  /room-stats?room=X       → Статистика комнаты
WS   /ws?room=X&username=Y    → WebSocket
```

---

## 🔨 Команды Make

```bash
make help      # Показать справку
make deps      # Загрузить зависимости
make build     # Собрать
make run       # Запустить
make dev       # Разработка
make clean     # Очистить
make fmt       # Форматировать
make lint      # Проверить
```

---

## ✨ Что улучшилось

✅ Из 1 файла (main.go) → 9 файлов в 6 пакетах  
✅ Слабая архитектура → Стандарт golang-standards  
✅ Все в main → Разделено по ответственности  
✅ Сложно расширять → Легко добавлять функции  

---

## 🎯 Следующие шаги

1. Запустить: `make dev`
2. Открыть: `http://localhost:8080`
3. Прочитать: [ARCHITECTURE.md](ARCHITECTURE.md)
4. Кодить!

---

**Готово к работе!** 🚀
