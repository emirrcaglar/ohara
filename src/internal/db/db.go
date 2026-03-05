package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Init(dataDir string) (*DB, error) {
	conn, err := sql.Open("sqlite", dataDir+"/ohara.db")
	if err != nil {
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
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			username   TEXT     NOT NULL UNIQUE,
			role       TEXT     NOT NULL DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		INSERT OR IGNORE INTO user (id, username, role) VALUES (1, 'admin', 'admin');

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
	`)
	return err
}
