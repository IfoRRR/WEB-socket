package models

import "time"

// Message описывает структуру сообщения чата
type Message struct {
	Room      string    `json:"room"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
