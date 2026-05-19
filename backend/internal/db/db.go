package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Init(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", dataDir+"/ohara.db")
	if err != nil {
		return nil, err
	}

	if _, err := conn.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		conn.Close()
		return nil, err
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &DB{conn}, nil
}
func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			id            INTEGER  PRIMARY KEY AUTOINCREMENT,
			username      TEXT     NOT NULL UNIQUE,
			password_hash TEXT     NOT NULL,
			role          TEXT     NOT NULL DEFAULT 'user',
			is_approved   BOOLEAN  NOT NULL DEFAULT 0,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS manga (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			path       TEXT     NOT NULL UNIQUE,
			title      TEXT     NOT NULL,
			page_count INTEGER  NOT NULL DEFAULT 0,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS manga_progress (
			user_id    INTEGER  NOT NULL REFERENCES user(id),
			manga_id   INTEGER  NOT NULL REFERENCES manga(id),
			page       INTEGER  NOT NULL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, manga_id)
		);

		CREATE TABLE IF NOT EXISTS audio (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			path       TEXT     NOT NULL UNIQUE,
			title      TEXT     NOT NULL,
			artist     TEXT     NOT NULL DEFAULT '',
			album      TEXT     NOT NULL DEFAULT '',
			duration   INTEGER  NOT NULL DEFAULT 0,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS scan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			manga_id INTEGER NOT NULL,
			page_idx INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(manga_id) REFERENCES manga(id) ON DELETE CASCADE
		);
	`)
	return err
}
