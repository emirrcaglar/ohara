package db

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"ohara/src/internal/media/manga"
)

type MangaRow struct {
	ID            int64
	Path          string
	FolderID      *int64
	Title         string
	PageCount     int
	Progress      int
	UpdatedAt     *time.Time
	SeriesTitle   string
	Metadata      *manga.Metadata
	ParserVersion string
}

func (db *DB) IndexedMangaPaths() (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT path FROM manga`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paths := make(map[string]struct{})
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths[p] = struct{}{}
	}
	return paths, rows.Err()
}

func (db *DB) ListManga(userID int64) ([]MangaRow, error) {
	rows, err := db.Query(`
		SELECT m.id, m.path, m.folder_id, m.title, m.page_count, COALESCE(p.page, 0), p.updated_at,
		       m.series_title, m.metadata_json, m.parser_version
		FROM manga m
		LEFT JOIN manga_progress p ON p.manga_id = m.id AND p.user_id = ?
		ORDER BY m.title
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []MangaRow
	for rows.Next() {
		var m MangaRow
		var folderID sql.NullInt64
		var metadataJSON string
		if err := rows.Scan(&m.ID, &m.Path, &folderID, &m.Title, &m.PageCount, &m.Progress, &m.UpdatedAt,
			&m.SeriesTitle, &metadataJSON, &m.ParserVersion); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(metadataJSON), &m.Metadata); err != nil {
			return nil, err
		}
		if folderID.Valid {
			m.FolderID = &folderID.Int64
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (db *DB) GetMangaByID(id int64) (*MangaRow, error) {
	row := db.QueryRow(`SELECT id, path, folder_id, title, page_count, series_title, metadata_json, parser_version FROM manga WHERE id = ?`, id)
	var m MangaRow
	var folderID sql.NullInt64
	var metadataJSON string
	if err := row.Scan(&m.ID, &m.Path, &folderID, &m.Title, &m.PageCount, &m.SeriesTitle, &metadataJSON, &m.ParserVersion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(metadataJSON), &m.Metadata); err != nil {
		return nil, err
	}
	if folderID.Valid {
		m.FolderID = &folderID.Int64
	}
	return &m, nil
}

func (db *DB) GetMangaProgress(userID, mangaID int64) (int, error) {
	var page int
	err := db.QueryRow(
		`SELECT page FROM manga_progress WHERE user_id = ? AND manga_id = ?`,
		userID, mangaID,
	).Scan(&page)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return page, err
}

func (db *DB) UpsertMangaProgress(userID, mangaID int64, page int) error {
	_, err := db.Exec(`
		INSERT INTO manga_progress (user_id, manga_id, page, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, manga_id) DO UPDATE SET page = excluded.page, updated_at = excluded.updated_at
	`, userID, mangaID, page)
	return err
}

func (db *DB) InsertManga(path, title string, pageCount int) (int64, error) {
	return db.InsertMangaWithFolder(path, title, pageCount, nil)
}

func (db *DB) InsertMangaWithFolder(path, title string, pageCount int, folderID *int64) (int64, error) {
	return db.InsertMangaWithMetadata(path, title, pageCount, folderID, "", nil, "")
}

func (db *DB) InsertMangaWithMetadata(path, title string, pageCount int, folderID *int64, seriesTitle string, metadata *manga.Metadata, parserVersion string) (int64, error) {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Re-imports preserve manual placement and must not create unused folders.
	var id int64
	err = tx.QueryRow(`SELECT id FROM manga WHERE path = ?`, path).Scan(&id)
	if err == nil {
		return id, tx.Commit()
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	if name := strings.TrimSpace(seriesTitle); name != "" {
		seriesFolderID, err := mangaSeriesFolder(tx, folderID, name)
		if err != nil {
			return 0, err
		}
		folderID = &seriesFolderID
	}

	err = tx.QueryRow(`
		INSERT INTO manga (path, title, page_count, folder_id, series_title, metadata_json, parser_version)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id
	`, path, title, pageCount, folderID, seriesTitle, string(metadataJSON), parserVersion).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// mangaSeriesFolder groups releases within their import destination. Importing
// directly into a matching series folder must not create another nested copy.
func mangaSeriesFolder(tx *sql.Tx, parentID *int64, name string) (int64, error) {
	if parentID != nil {
		var parentName string
		if err := tx.QueryRow(`SELECT name FROM folder WHERE id = ?`, *parentID).Scan(&parentName); err != nil {
			return 0, err
		}
		if parentName == name {
			return *parentID, nil
		}
	}

	if _, err := tx.Exec(`INSERT INTO folder (parent_id, name) VALUES (?, ?) ON CONFLICT DO NOTHING`, parentID, name); err != nil {
		return 0, err
	}
	var id int64
	err := tx.QueryRow(`SELECT id FROM folder WHERE parent_id IS ? AND name = ?`, parentID, name).Scan(&id)
	return id, err
}

func (db *DB) UpdateMangaFolder(id int64, folderID *int64) error {
	if folderID == nil {
		_, err := db.Exec(`UPDATE manga SET folder_id = NULL WHERE id = ?`, id)
		return err
	}

	_, err := db.Exec(`UPDATE manga SET folder_id = ? WHERE id = ?`, *folderID, id)
	return err
}

func (db *DB) DeleteManga(id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM manga_progress WHERE manga_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM manga WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}
