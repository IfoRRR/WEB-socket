package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"awesomeProject17/internal/database"
	"awesomeProject17/internal/models"
)

// Hub управляет всеми подключениями клиентов и сообщениями
type Hub struct {
	clients    map[string]map[*websocket.Conn]bool
	broadcast  chan models.Message
	register   chan *models.Client
	unregister chan *models.Client
	mu         sync.RWMutex
	logger     *log.Logger
	db         *database.Database
}

// NewHub создает новый Hub с инициализированными каналами
func NewHub(logger *log.Logger, db *database.Database) *Hub {
	return &Hub{
		clients:    make(map[string]map[*websocket.Conn]bool),
		broadcast:  make(chan models.Message, 256),
		register:   make(chan *models.Client),
		unregister: make(chan *models.Client),
		logger:     logger,
		db:         db,
	}
}

// Run запускает hub для обработки сообщений и управления клиентами
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case msg := <-h.broadcast:
			h.broadcastMessage(msg)
		}
	}
}

// registerClient регистрирует нового клиента
func (h *Hub) registerClient(client *models.Client) {
	h.mu.Lock()
	if h.clients[client.Room] == nil {
		h.clients[client.Room] = make(map[*websocket.Conn]bool)
	}
	h.clients[client.Room][client.Conn] = true
	h.mu.Unlock()
	h.logger.Printf("[%s] Клиент подключился (активных: %d)", client.Room, len(h.clients[client.Room]))
}

// unregisterClient удаляет клиента
func (h *Hub) unregisterClient(client *models.Client) {
	h.mu.Lock()
	if clients, ok := h.clients[client.Room]; ok {
		if _, exists := clients[client.Conn]; exists {
			delete(clients, client.Conn)
			h.logger.Printf("[%s] Клиент отключился (активных: %d)", client.Room, len(clients))
		}
	}
	h.mu.Unlock()
}

// broadcastMessage сохраняет и отправляет сообщение всем клиентам комнаты
func (h *Hub) broadcastMessage(msg models.Message) {
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
		return
	}

	for conn := range clients {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			h.logger.Printf("Ошибка отправки сообщения: %v", err)
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

// RegisterChan возвращает канал регистрации
func (h *Hub) RegisterChan() chan *models.Client {
	return h.register
}

// UnregisterChan возвращает канал дерегистрации
func (h *Hub) UnregisterChan() chan *models.Client {
	return h.unregister
}

// BroadcastChan возвращает канал трансляции
func (h *Hub) BroadcastChan() chan models.Message {
	return h.broadcast
}
