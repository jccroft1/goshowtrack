package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var Connection *sql.DB

func Setup() func() {
	var err error
	Connection, err = sql.Open("sqlite3", "file:./data/data.db?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	Connection.SetMaxOpenConns(1)
	Connection.SetMaxIdleConns(1)
	Connection.SetConnMaxLifetime(10 * time.Minute)

	_, err = Connection.Exec(`
		PRAGMA synchronous = NORMAL;
		PRAGMA temp_store = MEMORY;
		PRAGMA cache_size = -8192;
	`)
	if err != nil {
		log.Fatalf("Failed to set PRAGMAs: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Connection.PingContext(ctx); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	// Create users table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE
    );`)
	if err != nil {
		log.Fatal(err)
	}

	// create shows DB
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS shows (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		show_id INTEGER UNIQUE, 
		name TEXT,
		status TEXT, 
        air_date TEXT,
        description TEXT,
        poster_path TEXT
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// create seasons table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS seasons (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        show_id INTEGER, 
		name TEXT, 
		season_number INTEGER,
        episode_count INTEGER, 
		air_date TEXT, 
		last_air_date TEXT, 
		UNIQUE(show_id, season_number)
    );`)
	if err != nil {
		log.Fatal(err)
	}

	// Create user_shows table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS user_shows (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER, 
		show_id INTEGER, 
		UNIQUE(user_id, show_id)
    );`)
	if err != nil {
		log.Fatal(err)
	}

	// Create user_seasons table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS user_seasons (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER, 
		show_id INTEGER, 
		season_number INTEGER, 
		UNIQUE(user_id, show_id, season_number)
    );`)
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		log.Println("Closing DB...")
		Connection.Close()
	}
}
