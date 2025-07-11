package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var Connection *sql.DB

func Setup() {
	var err error
	Connection, err = sql.Open("sqlite", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create users table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE
    );`)
	if err != nil {
		log.Fatal(err)
	}

	// Create magic_links table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS magic_links (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        token TEXT UNIQUE,
        email TEXT,
        expires_at INTEGER,
        used INTEGER DEFAULT 0
    );`)
	if err != nil {
		log.Fatal(err)
	}

	// create shows DB
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS shows (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		show_id INTEGER, 
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
		air_date TEXT
    );`)
	if err != nil {
		log.Fatal(err)
	}

	// Create user_shows table
	_, err = Connection.Exec(`CREATE TABLE IF NOT EXISTS user_shows (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER, 
		show_id INTEGER
    );`)
	if err != nil {
		log.Fatal(err)
	}
}
