package models

import "github.com/gorilla/websocket"

// Client представляет подключенного пользователя
type Client struct {
	Conn     *websocket.Conn
	Room     string
	Username string
}
