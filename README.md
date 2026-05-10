# Awesome Chat Project

WebSocket-based real-time chat server на Go с поддержкой комнат и SQLite БД.

## 🏗️ Архитектура

Проект следует стандарту [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

- **cmd/server** - точка входа приложения
- **internal/database** - работа с SQLite БД
- **internal/hub** - управление WebSocket подключениями
- **internal/models** - структуры данных
- **internal/server** - HTTP сервер и обработчики
- **static/** - фронтенд на HTML/CSS

## ✨ Функции

✅ **WebSocket чат** - реал-тайм сообщения с поддержкой комнат  
✅ **SQLite БД** - сохранение истории сообщений  
✅ **REST API** - получение статистики и истории  
✅ **Структурированное логирование** - все события логируются с временем  
✅ **Обработка ошибок** - правильная обработка ошибок JSON и сокетов  
✅ **Таймауты** - ReadDeadline, WriteDeadline для предотвращения зависаний  
✅ **Graceful shutdown** - корректное завершение с сигналами SIGINT/SIGTERM  
✅ **Валидация** - проверка пустых сообщений и корректной JSON  
✅ **Метрики** - endpoint `/stats` для мониторинга активных соединений  
✅ **Thread-safe** - использование sync.RWMutex для безопасного доступа к данным  
✅ **Буферизация каналов** - broadcast канал с буфером для лучшей производительности  
✅ **Ping/Pong** - heartbeat для долгоживущих соединений  

## 🚀 Использование

### Сборка и запуск

**Напрямую:**
```bash
go run ./cmd/server
```

**Собрать бинарник:**
```bash
go build -o server ./cmd/server
./server
```

### Web UI
Откройте в браузере:
```
http://localhost:8080/
```

## 📡 API Endpoints

### 1. Health Check
```
GET /ping
```
Ответ: `pong`

### 2. Статистика сервера
```
GET /stats
```
Ответ:
```json
{
  "general": 5,
  "rooms": 2,
  "total": 12,
  "work": 7
}
```

### 3. История сообщений
```
GET /history?room=general&limit=20
```

**Параметры:**
- `room` (optional) - имя комнаты (по умолчанию "general")
- `limit` (optional) - количество сообщений (по умолчанию 50, макс 1000)

Ответ:
```json
[
  {
    "room": "general",
    "username": "alice",
    "text": "Hello!",
    "timestamp": "2026-05-05T10:30:45Z"
  }
]
```

### 4. Статистика комнаты
```
GET /room-stats?room=general
```

Ответ:
```json
{
  "room": "general",
  "messages": 42
}
```

### 5. WebSocket
```
WS ws://localhost:8080/ws?room=general&username=alice
```

**Параметры:**
- `room` (optional) - имя комнаты (по умолчанию "general")
- `username` (optional) - имя пользователя (по умолчанию "anonymous")

**Формат сообщения:**
```json
{
  "text": "Hello, everyone!"
}
```

**Ответ от сервера:**
```json
{
  "room": "general",
  "username": "alice",
  "text": "Hello, everyone!",
  "timestamp": "2026-05-05T10:30:45.123Z"
}
```

## 💡 Примеры использования

### JavaScript клиент
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?room=general&username=alice');

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(`[${msg.username}] ${msg.text}`);
};

ws.send(JSON.stringify({ text: 'Hello!' }));
```

### Получение истории
```javascript
async function getHistory() {
  const response = await fetch('/history?room=general&limit=50');
  const messages = await response.json();
  console.log(messages);
}
```

### cURL примеры
```bash
# Health check
curl http://localhost:8080/ping

# Статистика
curl http://localhost:8080/stats

# История
curl "http://localhost:8080/history?room=general&limit=20"

# Статистика комнаты
curl "http://localhost:8080/room-stats?room=general"
```

## 📂 Структура проекта

```
awesomeProject17/
├── cmd/server/                  # Точка входа
│   └── main.go
├── internal/                    # Приватный код
│   ├── database/               # Работа с БД
│   ├── hub/                    # Управление подключениями
│   ├── models/                 # Структуры данных
│   └── server/                 # HTTP и WebSocket
├── static/                     # Статические файлы
├── ARCHITECTURE.md             # Документация архитектуры
├── DATABASE.md                 # Документация БД
├── PROJECT_STRUCTURE.md        # Структура проекта
└── Makefile                    # Команды сборки
```

## 🔧 Разработка

### Команды

```bash
make help      # Показать все команды
make deps      # Загрузить зависимости
make build     # Собрать проект
make run       # Запустить
make dev       # Режим разработки
make clean     # Очистка
make fmt       # Форматирование
make lint      # Linting
make test      # Тесты
```

### Зависимости

- `github.com/gorilla/websocket` - WebSocket
- `modernc.org/sqlite` - SQLite БД (без CGO)

## 📊 Особенности

- **async/await** паттерны в Go (горутины и каналы)
- **type-safe** операции с БД
- **graceful shutdown** с контекстом
- **RWMutex** для потокобезопасности
- **буферизация каналов** для производительности

## 🐛 Требования

- Go 1.25+

## 📝 Лицензия

MIT
