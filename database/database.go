package database

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func InitDB() {
	dbPath := "/data/tutor.db"

	if _, err := os.Stat("/data"); os.IsNotExist(err) {
		log.Println("Persistent storage (/data) not found, using local database...")
		dbPath = "tutor.db"
	} else {
		log.Println("Using persistent database:", dbPath)
	}

	var err error
	DB, err = sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT
		);
		
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			description TEXT NOT NULL,
			language TEXT,
			level TEXT,
			time_estimated INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		
		CREATE TABLE IF NOT EXISTS solutions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER,
			code TEXT,
			rating TEXT,
			mark REAL,
			ai_usage INTEGER,
			chat TEXT,
			time_spent INTEGER,
			FOREIGN KEY (task_id) REFERENCES tasks(id)
		);
		
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT
		);
		
		CREATE TABLE IF NOT EXISTS interactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			task_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			time_remaining INTEGER,
			time_spent INTEGER,
			category_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (task_id) REFERENCES tasks(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		);
	`

	if _, err := DB.Exec(schema); err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected and schema initialized successfully.")
}
