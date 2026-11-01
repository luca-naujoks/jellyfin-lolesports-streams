package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
)

type Store struct {
	db *sql.DB
}

func NewSqliteStore() (*Store, error) {
	dsn := "./db/sqlite3.db"

	dir := filepath.Dir(dsn)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	absPath, err := filepath.Abs(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	database, err := sql.Open("sqlite3", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = pingTestDatabase(database)
	if err != nil {
		return nil, err
	}

	err = initTables(database)
	if err != nil {
		return nil, err
	}

	return &Store{db: database}, nil
}

func pingTestDatabase(db *sql.DB) error {
	err := db.Ping()
	return err
}

func initTables(db *sql.DB) error {
	if db == nil {
		fmt.Println("Database not jet Initialized")
		return errors.New("database not jet Initialized")
	}

	sqlStatements :=
		[]string{
			`CREATE TABLE IF NOT EXISTS channels (
				id TEXT PRIMARY KEY,
				name     TEXT NOT NULL,
                provider TEXT,
                language TEXT,
                icon_url TEXT
    );`,
			`CREATE TABLE IF NOT EXISTS channel_leagues (
				channel_id TEXT NOT NULL,
				league_id TEXT NOT NULL,
    
    			PRIMARY KEY (channel_id, league_id),
    			
                FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
                FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
    );`,
			`CREATE TABLE IF NOT EXISTS leagues (
				id TEXT PRIMARY KEY,
				slug     TEXT,
                name TEXT,
                region TEXT,
                image TEXT
    );`,
		}

	for _, statement := range sqlStatements {
		_, err := db.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}
