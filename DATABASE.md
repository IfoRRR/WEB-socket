# Взаимодействие с базой данных

## Обзор

Проект интегрирует SQLite для сохранения истории сообщений чата. База данных автоматически инициализируется при запуске сервера.

## Функции базы данных

### 1. Сохранение сообщений
Каждое сообщение, отправленное через WebSocket, автоматически сохраняется в базу данных с информацией:
- `room` - название комнаты
- `username` - имя отправителя
- `text` - текст сообщения
- `timestamp` - время отправки

### 2. Таблица данных

```sql
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room TEXT NOT NULL,
    username TEXT NOT NULL,
    text TEXT NOT NULL,
    timestamp DATETIME NOT NULL
);
```

Также создается индекс для быстрого поиска:
```sql
CREATE INDEX IF NOT EXISTS idx_room_timestamp ON messages(room, timestamp DESC);
```

## REST API для работы с БД

### 1. Получить историю сообщений
**Эндпоинт:** `GET /history`

**Параметры:**
- `room` (опционально) - название комнаты (по умолчанию "general")
- `limit` (опционально) - количество сообщений (по умолчанию 50, максимум 1000)

**Пример:**
```bash
curl "http://localhost:8080/history?room=general&limit=20"
```

**Ответ:**
```json
[
  {
    "room": "general",
    "username": "Alice",
    "text": "Привет, мир!",
    "timestamp": "2025-05-05T12:30:45Z"
  },
  {
    "room": "general",
    "username": "Bob",
    "text": "Привет, Alice!",
    "timestamp": "2025-05-05T12:31:20Z"
  }
]
```

### 2. Получить статистику по комнате
**Эндпоинт:** `GET /room-stats`

**Параметры:**
- `room` (опционально) - название комнаты (по умолчанию "general")

**Пример:**
```bash
curl "http://localhost:8080/room-stats?room=general"
```

**Ответ:**
```json
{
  "room": "general",
  "messages": 42
}
```

### 3. Получить общую статистику сервера
**Эндпоинт:** `GET /stats`

**Пример:**
```bash
curl "http://localhost:8080/stats"
```

**Ответ:**
```json
{
  "general": 3,
  "random": 1,
  "total": 4,
  "rooms": 2
}
```

## Структура Database

```go
type Database struct {
    db *sql.DB
}
```

### Методы Database

#### InitDatabase(filepath string) (*Database, error)
Инициализирует подключение к БД и создает таблицу.
```go
db, err := InitDatabase("./messages.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

#### SaveMessage(msg *Message) error
Сохраняет сообщение в БД.
```go
msg := &Message{
    Room: "general",
    Username: "Alice",
    Text: "Hello!",
    Timestamp: time.Now(),
}
err := db.SaveMessage(msg)
```

#### GetMessageHistory(room string, limit int) ([]Message, error)
Получает историю сообщений комнаты.
```go
messages, err := db.GetMessageHistory("general", 50)
if err != nil {
    log.Fatal(err)
}
for _, msg := range messages {
    fmt.Printf("%s: %s\n", msg.Username, msg.Text)
}
```

#### GetMessageCount(room string) (int, error)
Возвращает количество сообщений в комнате.
```go
count, err := db.GetMessageCount("general")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Всего сообщений: %d\n", count)
```

#### DeleteOldMessages(days int) error
Удаляет сообщения старше указанного количества дней.
```go
err := db.DeleteOldMessages(30) // Удалить старше 30 дней
```

#### Close() error
Закрывает соединение с БД.
```go
err := db.Close()
```

## Примеры использования

### JavaScript клиент - получение истории
```javascript
async function loadHistory(room, limit = 20) {
    const response = await fetch(`/history?room=${room}&limit=${limit}`);
    const messages = await response.json();
    messages.forEach(msg => {
        console.log(`[${msg.timestamp}] ${msg.username}: ${msg.text}`);
    });
}

loadHistory('general', 50);
```

### Python скрипт - анализ данных
```python
import requests
import json

# Получить историю
response = requests.get('http://localhost:8080/history?room=general&limit=100')
messages = response.json()

# Статистика по пользователям
user_stats = {}
for msg in messages:
    user = msg['username']
    user_stats[user] = user_stats.get(user, 0) + 1

print("Статистика по пользователям:")
for user, count in sorted(user_stats.items(), key=lambda x: x[1], reverse=True):
    print(f"  {user}: {count} сообщений")
```

## Особенности и оптимизация

### Автоматическое сохранение
- Все сообщения сохраняются в реальном времени при отправке
- Ошибки при сохранении логируются, но не влияют на доставку сообщения

### Индексирование
- Создается индекс по (room, timestamp) для быстрого поиска истории
- Это позволяет эффективно получать последние сообщения

### Ограничения
- Максимальное количество сообщений при запросе истории - 1000
- БД хранится локально в файле `messages.db`

## Управление базой данных

### Просмотр содержимого БД
```bash
sqlite3 messages.db
```

```sql
-- Просмотр всех сообщений
SELECT * FROM messages;

-- Сообщения конкретной комнаты
SELECT * FROM messages WHERE room = 'general' ORDER BY timestamp DESC;

-- Количество сообщений по комнатам
SELECT room, COUNT(*) as message_count FROM messages GROUP BY room;

-- Сообщения за последний час
SELECT * FROM messages 
WHERE timestamp > datetime('now', '-1 hour')
ORDER BY timestamp DESC;
```

### Очистка старых данных
Вы можете вызвать `DeleteOldMessages` периодически в фоне:
```go
// Очищаем сообщения старше 90 дней каждый час
ticker := time.NewTicker(time.Hour)
go func() {
    for range ticker.C {
        if err := db.DeleteOldMessages(90); err != nil {
            logger.Printf("Ошибка при очистке БД: %v", err)
        }
    }
}()
```

## Файлы проекта

- `main.go` - основной код сервера с функциями БД
- `messages.db` - автоматически создаваемый файл БД SQLite
- `go.mod` - зависимости (включает github.com/mattn/go-sqlite3)
