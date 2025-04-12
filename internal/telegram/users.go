package telegram

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Users представляет хранилище идентификаторов пользователей
type Users struct {
	db    *sql.DB
	mu    sync.Mutex
	logger *log.Logger
}

// NewUsers создает новое хранилище пользователей с поддержкой базы данных SQLite
func NewUsers() *Users {
	logger := log.New(os.Stdout, "UsersDB: ", log.LstdFlags)
	
	// Создаем директорию для БД, если она не существует
	dbDir := "/app/data"
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			logger.Printf("Failed to create data directory: %v", err)
			// Если не получается создать директорию, используем текущую
			dbDir = "."
		}
	}
	
	dbPath := filepath.Join(dbDir, "users.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Printf("Error opening database: %v, using in-memory storage", err)
		return &Users{
			mu: sync.Mutex{},
			logger: logger,
		}
	}
	
	// Создаем таблицу пользователей, если она не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			chat_id INTEGER PRIMARY KEY,
			username TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		logger.Printf("Error creating users table: %v, using in-memory storage", err)
		db.Close()
		return &Users{
			mu: sync.Mutex{},
			logger: logger,
		}
	}
	
	logger.Println("SQLite database initialized successfully")
	return &Users{
		db: db,
		mu: sync.Mutex{},
		logger: logger,
	}
}

// Add добавляет пользователя в хранилище
func (u *Users) Add(chatID int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	// Если БД не инициализирована, используем in-memory хранилище
	if u.db == nil {
		u.logger.Printf("Warning: Database not initialized, user %d might not be persisted", chatID)
		return
	}
	
	// Добавляем пользователя, игнорируем, если уже существует
	_, err := u.db.Exec("INSERT OR IGNORE INTO users (chat_id) VALUES (?)", chatID)
	if err != nil {
		u.logger.Printf("Error adding user %d to database: %v", chatID, err)
	} else {
		u.logger.Printf("User %d added to database", chatID)
	}
}

// AddWithUsername добавляет пользователя с именем пользователя
func (u *Users) AddWithUsername(chatID int64, username string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	// Если БД не инициализирована, используем in-memory хранилище
	if u.db == nil {
		u.logger.Printf("Warning: Database not initialized, user %d might not be persisted", chatID)
		return
	}
	
	// Добавляем пользователя с именем, обновляем имя, если пользователь уже существует
	_, err := u.db.Exec(`
		INSERT INTO users (chat_id, username) 
		VALUES (?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET username = ?
	`, chatID, username, username)
	
	if err != nil {
		u.logger.Printf("Error adding user %d with username %s to database: %v", chatID, username, err)
	} else {
		u.logger.Printf("User %d with username %s added to database", chatID, username)
	}
}

// GetAll возвращает список всех пользователей
func (u *Users) GetAll() []int64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	var users []int64
	
	// Если БД не инициализирована, возвращаем пустой список
	if u.db == nil {
		u.logger.Printf("Warning: Database not initialized, returning empty users list")
		return users
	}
	
	// Получаем всех пользователей из базы
	rows, err := u.db.Query("SELECT chat_id FROM users")
	if err != nil {
		u.logger.Printf("Error retrieving users from database: %v", err)
		return users
	}
	defer rows.Close()
	
	// Обрабатываем результаты
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			u.logger.Printf("Error scanning user row: %v", err)
			continue
		}
		users = append(users, chatID)
	}
	
	if err := rows.Err(); err != nil {
		u.logger.Printf("Error iterating user rows: %v", err)
	}
	
	u.logger.Printf("Retrieved %d users from database", len(users))
	return users
}

// Count возвращает количество пользователей
func (u *Users) Count() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	// Если БД не инициализирована, возвращаем 0
	if u.db == nil {
		return 0
	}
	
	var count int
	err := u.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		u.logger.Printf("Error counting users: %v", err)
		return 0
	}
	
	return count
}

// Close закрывает соединение с базой данных
func (u *Users) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	if u.db != nil {
		return u.db.Close()
	}
	
	return nil
} 