package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

// Database управляет подключением к базе данных и операциями с сообщениями
type Database struct {
	db *sql.DB
}

// InitDatabase инициализирует базу данных и создает таблицу если её нет
func InitDatabase(filepath string) (*Database, error) {
	// Используем файловый путь с префиксом file: для modernc/sqlite
	dsn := "file:" + filepath + "?cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Создание таблицы для сообщений
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room TEXT NOT NULL,
		username TEXT NOT NULL,
		text TEXT NOT NULL,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_room_timestamp ON messages(room, timestamp DESC);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// SaveMessage сохраняет сообщение в базу данных
func (d *Database) SaveMessage(msg *Message) error {
	if d.db == nil {
		return nil
	}

	query := `INSERT INTO messages (room, username, text, timestamp) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, msg.Room, msg.Username, msg.Text, msg.Timestamp)
	return err
}

// GetMessageHistory возвращает историю сообщений по комнате
func (d *Database) GetMessageHistory(room string, limit int) ([]Message, error) {
	if d.db == nil {
		return []Message{}, nil
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `SELECT room, username, text, timestamp FROM messages WHERE room = ? ORDER BY timestamp DESC LIMIT ?`
	rows, err := d.db.Query(query, room, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Room, &msg.Username, &msg.Text, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Разворачиваем слайс, чтобы сообщения были в хронологическом порядке
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteOldMessages удаляет сообщения старше указанного дня
func (d *Database) DeleteOldMessages(days int) error {
	if d.db == nil {
		return nil
	}

	query := `DELETE FROM messages WHERE timestamp < datetime('now', ? || ' days')`
	_, err := d.db.Exec(query, -days)
	return err
}

// Close закрывает соединение с базой данных
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetMessageCount возвращает количество сообщений в комнате
func (d *Database) GetMessageCount(room string) (int, error) {
	if d.db == nil {
		return 0, nil
	}

	query := `SELECT COUNT(*) FROM messages WHERE room = ?`
	var count int
	err := d.db.QueryRow(query, room).Scan(&count)
	return count, err
}

type Hub struct {
	clients    map[string]map[*websocket.Conn]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *log.Logger
	db         *Database
}

// Client представляет подключенного пользователя
type Client struct {
	conn     *websocket.Conn
	room     string
	username string
	hub      *Hub
}

// Message описывает структуру сообщения чата
type Message struct {
	Room      string    `json:"room"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// Run запускает hub для обработки сообщений и управления клиентами
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.room] == nil {
				h.clients[client.room] = make(map[*websocket.Conn]bool)
			}
			h.clients[client.room][client.conn] = true
			h.mu.Unlock()
			h.logger.Printf("[%s] Клиент подключился (активных: %d)", client.room, len(h.clients[client.room]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.room]; ok {
				if _, exists := clients[client.conn]; exists {
					delete(clients, client.conn)
					h.logger.Printf("[%s] Клиент отключился (активных: %d)", client.room, len(clients))
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			// Сохраняем сообщение в базу данных
			if err := h.db.SaveMessage(&msg); err != nil {
				h.logger.Printf("Ошибка сохранения в БД: %v", err)
			}

			h.mu.RLock()
			clients := h.clients[msg.Room]
			h.mu.RUnlock()

			data, err := json.Marshal(msg)
			if err != nil {
				h.logger.Printf("Ошибка маршализации: %v", err)
				continue
			}

			for conn := range clients {
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					h.logger.Printf("Ошибка отправки сообщения: %v", err)
				}
			}
		}
	}
}

// GetStats возвращает статистику сервера
func (h *Hub) GetStats() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]int)
	totalClients := 0
	for room, clients := range h.clients {
		stats[room] = len(clients)
		totalClients += len(clients)
	}
	stats["total"] = totalClients
	stats["rooms"] = len(h.clients)
	return stats
}

func main() {
	logger := log.New(os.Stdout, "[чат-сервер] ", log.LstdFlags)

	// Инициализируем базу данных
	db, err := InitDatabase("./messages.db")
	if err != nil {
		logger.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	hub := &Hub{
		clients:    make(map[string]map[*websocket.Conn]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		db:         db,
	}

	go hub.Run()

	// Статический фронтенд на Vue
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Проверка здоровья сервера
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Статистика сервера
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := hub.GetStats()
		json.NewEncoder(w).Encode(stats)
	})

	// История сообщений комнаты
	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "general"
		}
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if n, err := strconv.Atoi(limitStr); err == nil {
				limit = n
			}
		}

		messages, err := db.GetMessageHistory(room, limit)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if messages == nil {
			messages = []Message{}
		}
		json.NewEncoder(w).Encode(messages)
	})

	// Статистика по комнате
	http.HandleFunc("/room-stats", func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "general"
		}

		count, err := db.GetMessageCount(room)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"room":     room,
			"messages": count,
		})
	})

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// WebSocket обработчик
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "general"
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			username = "anonymous"
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Printf("Ошибка upgrade: %v", err)
			return
		}

		client := &Client{
			conn:     conn,
			room:     room,
			username: username,
			hub:      hub,
		}

		hub.register <- client

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		go func() {
			defer func() {
				hub.unregister <- client
				conn.Close()
			}()

			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logger.Printf("Ошибка сокета: %v", err)
					}
					break
				}

				var msg Message
				if err := json.Unmarshal(data, &msg); err != nil {
					logger.Printf("Ошибка парсинга JSON: %v", err)
					continue
				}

				// Валидация
				if msg.Text == "" {
					continue
				}

				msg.Room = room
				msg.Username = username
				msg.Timestamp = time.Now()

				hub.broadcast <- msg
			}
		}()
	})

	// Graceful shutdown
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Println("Сервер запущен на :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Завершение работы сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Ошибка при выключении сервера: %v", err)
	}

	logger.Println("Сервер остановлен")
}
