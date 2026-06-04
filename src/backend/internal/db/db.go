package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	defaultAdminCreated bool
}

func Init(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %q: %w", dataDir, err)
	}

	dbPath := filepath.Join(dataDir, "ohara.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %q: %w", dbPath, err)
	}

	if _, err := conn.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		conn.Close()
		return nil, err
	}

	defaultAdminCreated, err := migrate(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &DB{DB: conn, defaultAdminCreated: defaultAdminCreated}, nil
}

func (db *DB) DefaultAdminCreated() bool {
	return db.defaultAdminCreated
}

func migrate(conn *sql.DB) (bool, error) {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			id            INTEGER  PRIMARY KEY AUTOINCREMENT,
			username      TEXT     NOT NULL UNIQUE,
			password_hash TEXT     NOT NULL,
			role          TEXT     NOT NULL DEFAULT 'user',
			is_approved   BOOLEAN  NOT NULL DEFAULT 0,
			pfp           INTEGER  NOT NULL DEFAULT 0,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return false, err
	}

	defaultAdminCreated, err := seedDefaultAdmin(conn)
	if err != nil {
		return false, err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS catalog (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			parent_id  INTEGER  REFERENCES catalog(id) ON DELETE CASCADE,
			name       TEXT     NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS manga (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			path       TEXT     NOT NULL UNIQUE,
			catalog_id INTEGER  REFERENCES catalog(id) ON DELETE SET NULL,
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

		CREATE TABLE IF NOT EXISTS preferences (
			user_id    INTEGER  NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			key        TEXT     NOT NULL,
			value      TEXT     NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, key)
		);

		CREATE TABLE IF NOT EXISTS audio (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			path       TEXT     NOT NULL UNIQUE,
			catalog_id INTEGER  REFERENCES catalog(id) ON DELETE SET NULL,
			title      TEXT     NOT NULL,
			artist     TEXT     NOT NULL DEFAULT '',
			album      TEXT     NOT NULL DEFAULT '',
			duration   INTEGER  NOT NULL DEFAULT 0,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS video (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			path       TEXT     NOT NULL UNIQUE,
			catalog_id INTEGER  REFERENCES catalog(id) ON DELETE SET NULL,
			title      TEXT     NOT NULL,
			duration   INTEGER  NOT NULL DEFAULT 0,
			width      INTEGER  NOT NULL DEFAULT 0,
			height     INTEGER  NOT NULL DEFAULT 0,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS video_progress (
			user_id       INTEGER  NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			video_id      INTEGER  NOT NULL REFERENCES video(id) ON DELETE CASCADE,
			position      INTEGER  NOT NULL DEFAULT 0,
			completed     BOOLEAN  NOT NULL DEFAULT 0,
			last_error    TEXT     NOT NULL DEFAULT '',
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, video_id)
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

		CREATE TABLE IF NOT EXISTS upload_sessions (
			id            TEXT     PRIMARY KEY,
			user_id       INTEGER  NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			filename      TEXT     NOT NULL,
			size          INTEGER  NOT NULL,
			profile       TEXT     NOT NULL DEFAULT '',
			last_modified INTEGER  NOT NULL DEFAULT 0,
			chunk_size    INTEGER  NOT NULL,
			total_chunks  INTEGER  NOT NULL,
			status        TEXT     NOT NULL DEFAULT 'active',
			target_path    TEXT     NOT NULL DEFAULT '',
			catalog_id    INTEGER  REFERENCES catalog(id) ON DELETE SET NULL,
			error_message TEXT,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at  DATETIME
		);

		CREATE INDEX IF NOT EXISTS idx_upload_sessions_resume
		ON upload_sessions(user_id, filename, size, last_modified, profile, status, updated_at);

		CREATE TABLE IF NOT EXISTS upload_chunks (
			upload_id   TEXT     NOT NULL REFERENCES upload_sessions(id) ON DELETE CASCADE,
			chunk_index INTEGER  NOT NULL,
			size        INTEGER  NOT NULL,
			path        TEXT     NOT NULL DEFAULT '',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (upload_id, chunk_index)
		);

		CREATE TABLE IF NOT EXISTS deployments (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			deployed_at TEXT    NOT NULL UNIQUE,
			created_at  TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at
			ON deployments(deployed_at DESC);
	`)
	if err != nil {
		return false, err
	}

	if err := addColumnIfMissing(conn, "user", "pfp", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "video", "width", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "video", "height", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "catalog", "parent_id", "INTEGER REFERENCES catalog(id) ON DELETE CASCADE"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "manga", "catalog_id", "INTEGER REFERENCES catalog(id) ON DELETE SET NULL"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "audio", "catalog_id", "INTEGER REFERENCES catalog(id) ON DELETE SET NULL"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "video", "catalog_id", "INTEGER REFERENCES catalog(id) ON DELETE SET NULL"); err != nil {
		return false, err
	}
	if err := addColumnIfMissing(conn, "upload_sessions", "catalog_id", "INTEGER REFERENCES catalog(id) ON DELETE SET NULL"); err != nil {
		return false, err
	}

	_, err = conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_catalog_parent_id
		ON catalog(parent_id);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_catalog_root_name
		ON catalog(name)
		WHERE parent_id IS NULL;

		CREATE UNIQUE INDEX IF NOT EXISTS idx_catalog_parent_name
		ON catalog(parent_id, name)
		WHERE parent_id IS NOT NULL;

		CREATE INDEX IF NOT EXISTS idx_manga_catalog_id
		ON manga(catalog_id);

		CREATE INDEX IF NOT EXISTS idx_audio_catalog_id
		ON audio(catalog_id);

		CREATE INDEX IF NOT EXISTS idx_video_catalog_id
		ON video(catalog_id);

		CREATE INDEX IF NOT EXISTS idx_upload_sessions_catalog_id
		ON upload_sessions(catalog_id);
	`)
	if err != nil {
		return false, err
	}

	return defaultAdminCreated, nil
}

func seedDefaultAdmin(conn *sql.DB) (bool, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	result, err := conn.Exec(`
		INSERT INTO user (username, password_hash, role, is_approved)
		VALUES ('admin', ?, 'admin', 1)
		ON CONFLICT(username) DO NOTHING
	`, string(hash))
	if err != nil {
		return false, fmt.Errorf("failed to seed default admin user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to verify default admin seed result: %w", err)
	}

	return rowsAffected > 0, nil
}

func addColumnIfMissing(conn *sql.DB, table, column, definition string) error {
	rows, err := conn.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return rows.Err()
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = conn.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}
