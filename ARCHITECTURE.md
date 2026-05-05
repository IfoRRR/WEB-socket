# Архитектура проекта

Проект следует стандарту [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

## Структура директорий

```
awesomeProject17/
├── cmd/
│   └── server/
│       └── main.go                 # Точка входа приложения
├── internal/
│   ├── database/
│   │   └── database.go             # Логика работы с БД
│   ├── hub/
│   │   └── hub.go                  # Управление подключениями
│   ├── models/
│   │   ├── client.go               # Модель клиента
│   │   └── message.go              # Модель сообщения
│   └── server/
│       ├── handlers.go             # HTTP обработчики
│       ├── server.go               # Основной сервер
│       └── websocket.go            # WebSocket логика
├── static/
│   ├── index.html
│   └── style.css
├── DATABASE.md                     # Документация по БД
├── ARCHITECTURE.md                 # Этот файл
├── README.md
├── go.mod
├── go.sum
└── .gitignore
```

## Описание компонентов

### cmd/server/
- **main.go** - точка входа приложения
- Инициализирует все компоненты и запускает сервер
- Отвечает за graceful shutdown

### internal/

#### database/
- **database.go** - работа с SQLite
  - `InitDatabase()` - инициализация БД
  - `SaveMessage()` - сохранение сообщений
  - `GetMessageHistory()` - получение истории
  - `GetMessageCount()` - подсчет сообщений
  - `DeleteOldMessages()` - очистка данных
  - `Close()` - закрытие соединения

#### hub/
- **hub.go** - управление подключениями и трансляция сообщений
  - `Hub` - центральный компонент для управления всеми клиентами
  - `NewHub()` - создание нового Hub
  - `Run()` - основной цикл обработки событий
  - `registerClient()` - регистрация клиента
  - `unregisterClient()` - удаление клиента
  - `broadcastMessage()` - отправка сообщения всем клиентам
  - `GetStats()` - статистика сервера

#### models/
- **message.go** - структура сообщения
  - `Message` - структура с полями Room, Username, Text, Timestamp

- **client.go** - структура клиента
  - `Client` - представление подключенного пользователя

#### server/
- **server.go** - основной HTTP сервер
  - `Server` - структура сервера
  - `NewServer()` - создание сервера с регистрацией маршрутов
  - `Start()` - запуск сервера
  - `Stop()` - остановка с graceful shutdown

- **handlers.go** - HTTP обработчики
  - `handlePing()` - проверка здоровья
  - `handleStats()` - статистика сервера
  - `handleHistory()` - история сообщений
  - `handleRoomStats()` - статистика комнаты

- **websocket.go** - WebSocket обработчики
  - `handleWebSocket()` - принятие WebSocket соединений
  - `handleClientMessages()` - обработка входящих сообщений

### static/
- Фронтенд приложения (HTML, CSS, JavaScript)

## Принципы архитектуры

### 1. Разделение ответственности
- **cmd/server** - точка входа
- **internal/database** - только работа с БД
- **internal/hub** - только управление подключениями
- **internal/server** - только HTTP и WebSocket логика
- **internal/models** - только структуры данных

### 2. Чистые границы пакетов
- `internal` пакеты не могут быть импортированы из других модулей
- `models` используются всеми другими пакетами
- Каждый компонент имеет четкий API

### 3. Горутины и каналы
- Hub запускается в отдельной горутине
- Сервер запускается в отдельной горутине
- Каждый клиент обслуживается отдельной горутиной
- Используются каналы для безопасной коммуникации

### 4. Управление ресурсами
- Graceful shutdown со списанием всех ресурсов
- Правильное закрытие БД и соединений
- Контекст с таймаутом при выключении

## Зависимости между компонентами

```
cmd/server
    ↓
    ├→ internal/database
    ├→ internal/hub ─→ internal/database
    │                 ├→ internal/models
    │                 └→ github.com/gorilla/websocket
    └→ internal/server ─→ internal/hub
                         ├→ internal/models
                         ├→ internal/database
                         └→ github.com/gorilla/websocket
```

## Как запустить

```bash
# Загрузить зависимости
go mod download

# Запустить сервер
go run ./cmd/server

# Или собрать бинарик
go build -o server ./cmd/server
```

## Добавление новых функций

Если нужно добавить новый обработчик:
1. Добавить метод в `internal/server/handlers.go`
2. Зарегистрировать маршрут в `internal/server/server.go`

Если нужно добавить новую операцию с БД:
1. Добавить метод в `internal/database/database.go`

Если нужно расширить логику Hub:
1. Добавить метод в `internal/hub/hub.go`

Новые модели данных:
1. Добавить в `internal/models/`
