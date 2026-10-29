package db

import (
	"database/sql"
	"fmt"
)

// Keep the old names only here so existing installations retain their hierarchy,
// media assignments, and resumable uploads when upgrading.
func migrateFolderNaming(conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'catalog')`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		if _, err := tx.Exec(`ALTER TABLE catalog RENAME TO folder`); err != nil {
			return fmt.Errorf("rename legacy folder table: %w", err)
		}
	}

	for _, table := range []string{"manga", "audio", "video", "upload_sessions"} {
		rows, err := tx.Query(`PRAGMA table_info(` + table + `)`)
		if err != nil {
			return err
		}
		found := false
		for rows.Next() {
			var cid, notNull, pk int
			var name, columnType string
			var defaultValue any
			if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
				rows.Close()
				return err
			}
			found = found || name == "catalog_id"
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		if found {
			if _, err := tx.Exec(`ALTER TABLE ` + table + ` RENAME COLUMN catalog_id TO folder_id`); err != nil {
				return fmt.Errorf("rename folder assignment on %s: %w", table, err)
			}
		}
	}

	for _, index := range []string{
		"idx_catalog_parent_id", "idx_catalog_root_name", "idx_catalog_parent_name",
		"idx_manga_catalog_id", "idx_audio_catalog_id", "idx_video_catalog_id", "idx_upload_sessions_catalog_id",
	} {
		if _, err := tx.Exec(`DROP INDEX IF EXISTS ` + index); err != nil {
			return err
		}
	}
	return tx.Commit()
}
