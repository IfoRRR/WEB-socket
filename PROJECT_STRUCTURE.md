# Project Structure

## Directory Tree

```
awesomeProject17/
├── cmd/                           # Приложения (точки входа)
│   └── server/
│       └── main.go               # Основная функция сервера
│
├── internal/                      # Приватный код приложения
│   ├── database/
│   │   └── database.go           # SQLite операции
│   │
│   ├── hub/
│   │   └── hub.go                # Управление подключениями
│   │
│   ├── models/
│   │   ├── client.go             # Структура Client
│   │   └── message.go            # Структура Message
│   │
│   └── server/
│       ├── handlers.go           # HTTP обработчики
│       ├── server.go             # HTTP сервер
│       └── websocket.go          # WebSocket логика
│
├── static/                        # Статические файлы
│   ├── index.html
│   └── style.css
│
├── .git/                          # Git репозиторий
├── .gitignore                     # Git игнор список
├── .idea/                         # IDE конфигурация
├── ARCHITECTURE.md                # Документация архитектуры
├── DATABASE.md                    # Документация БД
├── Makefile                       # Команды для сборки
├── PROJECT_STRUCTURE.md           # Этот файл
├── README.md                      # Основной README
├── go.mod                         # Go модуль
├── go.sum                         # Go зависимости (checksum)
├── messages.db                    # SQLite база (генерируется)
└── server.exe                     # Собранный бинарник (генерируется)
```

## Компоненты и их роли

| Компонент | Язык | Назначение |
|-----------|------|-----------|
| `cmd/server` | Go | Точка входа приложения |
| `internal/database` | Go | Работа с SQLite БД |
| `internal/hub` | Go | Управление WebSocket подключениями |
| `internal/models` | Go | Структуры данных (Message, Client) |
| `internal/server` | Go | HTTP сервер и обработчики |
| `static/` | HTML/CSS | Фронтенд приложения |

## Иерархия зависимостей

```
cmd/server/main.go
    ↓
    ├─→ internal/database
    │   └─→ modernc.org/sqlite
    │
    ├─→ internal/hub
    │   ├─→ internal/database
    │   ├─→ internal/models
    │   └─→ github.com/gorilla/websocket
    │
    └─→ internal/server
        ├─→ internal/hub
        ├─→ internal/database
        ├─→ internal/models
        └─→ github.com/gorilla/websocket
```

## API Endpoints

| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/` | Статический фронтенд |
| GET | `/ping` | Проверка здоровья |
| GET | `/stats` | Статистика сервера |
| GET | `/history?room=X&limit=Y` | История сообщений |
| GET | `/room-stats?room=X` | Статистика комнаты |
| WS | `/ws?room=X&username=Y` | WebSocket соединение |

## Как использовать структуру

### Добавление нового HTTP эндпоинта

1. Добавить функцию-обработчик в `internal/server/handlers.go`:
```go
func (s *Server) handleNewEndpoint(w http.ResponseWriter, r *http.Request) {
    // логика обработчика
}
```

2. Зарегистрировать в `internal/server/server.go`:
```go
mux.HandleFunc("/new-endpoint", s.handleNewEndpoint)
```

### Добавление операции с БД

1. Добавить метод в `internal/database/database.go`:
```go
func (d *Database) NewOperation() (result, error) {
    // логика
}
```

2. Использовать в других пакетах:
```go
result, err := s.db.NewOperation()
```

### Добавление нового типа данных

1. Создать новый файл в `internal/models/`:
```go
// internal/models/newtype.go
package models

type NewType struct {
    Field string
}
```

2. Использовать в других пакетах:
```go
import "awesomeProject17/internal/models"

item := models.NewType{Field: "value"}
```

## Соглашения о наименовании

- **Пакеты**: lowercase, one word если возможно
- **Типы**: CamelCase (Message, Hub, Client)
- **Функции**: CamelCase (SaveMessage, GetStats)
- **Переменные**: camelCase (db, hub, logger)
- **Константы**: UPPER_CASE (если нужны)

## Команды для работы

```bash
# Сборка
make build

# Запуск
make run

# Разработка (с автоперезагрузкой)
make dev

# Форматирование кода
make fmt

# Лinting
make lint

# Очистка
make clean

# Загрузка зависимостей
make deps
```

## Стандарты и лучшие практики

1. **Разделение ответственности** - каждый пакет отвечает за одно
2. **Чистые границы** - `internal/` пакеты не экспортируются
3. **Явные зависимости** - используется dependency injection
4. **Обработка ошибок** - ошибки проверяются везде
5. **Логирование** - используется `log.Logger`
6. **Graceful shutdown** - корректное закрытие ресурсов

## Запуск локально

```bash
# Установить зависимости
go mod download

# Запустить сервер
go run ./cmd/server

# Открыть браузер на http://localhost:8080
```

## Дополнительные ресурсы

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
