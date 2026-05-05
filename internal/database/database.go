package database

import (
	"database/sql"

	"awesomeProject17/internal/models"

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
func (d *Database) SaveMessage(msg *models.Message) error {
	if d.db == nil {
		return nil
	}

	query := `INSERT INTO messages (room, username, text, timestamp) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, msg.Room, msg.Username, msg.Text, msg.Timestamp)
	return err
}

// GetMessageHistory возвращает историю сообщений по комнате
func (d *Database) GetMessageHistory(room string, limit int) ([]models.Message, error) {
	if d.db == nil {
		return []models.Message{}, nil
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

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
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
