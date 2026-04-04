package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	playersSchema = `
	CREATE TABLE IF NOT EXISTS players (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    name TEXT NOT NULL,
	    xuid TEXT UNIQUE,
	    guid TEXT UNIQUE,
	    level INTEGER DEFAULT 0,
		client_id INTEGER NULL,
		discord_id INTEGER NULL,
	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	walletsSchema = `
	CREATE TABLE IF NOT EXISTS wallets (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    player_id INTEGER NOT NULL UNIQUE,
	    balance INTEGER DEFAULT 0,
	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY(player_id) REFERENCES players(id) ON DELETE CASCADE
	);`

	bankSchema = `
	CREATE TABLE IF NOT EXISTS bank (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		balance INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	playerStatsSchema = `
	CREATE TABLE IF NOT EXISTS player_stats (
    	player_id INTEGER PRIMARY KEY,
    	total_gambles INTEGER DEFAULT 0,
    	total_wagered INTEGER DEFAULT 0,
    	total_won INTEGER DEFAULT 0,
    	total_lost INTEGER DEFAULT 0,
    	wins INTEGER DEFAULT 0,
    	losses INTEGER DEFAULT 0,
   		last_gamble DATETIME,
	    FOREIGN KEY(player_id) REFERENCES players(id) ON DELETE CASCADE
	);`

	gambleStatsSchema = `
	CREATE TABLE IF NOT EXISTS global_stats (
    	id INTEGER PRIMARY KEY CHECK(id = 1),
    	total_gambles INTEGER DEFAULT 0,
    	total_wagered INTEGER DEFAULT 0,
    	total_paid INTEGER DEFAULT 0,
    	last_update DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
)

func Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "hypno.db?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		fmt.Printf("storage: couldnt not enable foreign keys: %v", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	schemas := []string{
		playersSchema,
		walletsSchema,
		bankSchema,
		playerStatsSchema,
		gambleStatsSchema,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}
