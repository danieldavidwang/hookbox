package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDatabase(path string) error {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	dsn := url.URL{Scheme: "file", Path: filepath.ToSlash(absolutePath)}
	options := url.Values{}
	options.Set("_pragma", "busy_timeout(5000)")
	dsn.RawQuery = options.Encode()

	database, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return err
	}
	// Serialize database access because SQLite permits only one writer at a time.
	database.SetMaxOpenConns(1)
	if err := database.Ping(); err != nil {
		database.Close()
		return fmt.Errorf("open database: %w", err)
	}
	// specified DB schema to store requests
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		query TEXT NOT NULL,
		headers TEXT NOT NULL,
		body BLOB NOT NULL,
		received_at TEXT NOT NULL
	)`)
	if err != nil {
		database.Close()
		return fmt.Errorf("create requests table: %w", err)
	}
	db = database
	return nil
}
