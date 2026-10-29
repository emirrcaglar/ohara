package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"ohara/src/internal/media/manga"
)

type MangaRow struct {
	ID            int64
	Path          string
	CatalogID     *int64
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
		SELECT m.id, m.path, m.catalog_id, m.title, m.page_count, COALESCE(p.page, 0), p.updated_at,
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
		var catalogID sql.NullInt64
		var metadataJSON string
		if err := rows.Scan(&m.ID, &m.Path, &catalogID, &m.Title, &m.PageCount, &m.Progress, &m.UpdatedAt,
			&m.SeriesTitle, &metadataJSON, &m.ParserVersion); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(metadataJSON), &m.Metadata); err != nil {
			return nil, err
		}
		if catalogID.Valid {
			m.CatalogID = &catalogID.Int64
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (db *DB) GetMangaByID(id int64) (*MangaRow, error) {
	row := db.QueryRow(`SELECT id, path, catalog_id, title, page_count, series_title, metadata_json, parser_version FROM manga WHERE id = ?`, id)
	var m MangaRow
	var catalogID sql.NullInt64
	var metadataJSON string
	if err := row.Scan(&m.ID, &m.Path, &catalogID, &m.Title, &m.PageCount, &m.SeriesTitle, &metadataJSON, &m.ParserVersion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(metadataJSON), &m.Metadata); err != nil {
		return nil, err
	}
	if catalogID.Valid {
		m.CatalogID = &catalogID.Int64
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
	return db.InsertMangaWithCatalog(path, title, pageCount, nil)
}

func (db *DB) InsertMangaWithCatalog(path, title string, pageCount int, catalogID *int64) (int64, error) {
	return db.InsertMangaWithMetadata(path, title, pageCount, catalogID, "", nil, "")
}

func (db *DB) InsertMangaWithMetadata(path, title string, pageCount int, catalogID *int64, seriesTitle string, metadata *manga.Metadata, parserVersion string) (int64, error) {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return 0, err
	}
	// RETURNING avoids a stale LastInsertId when an existing path is imported again.
	var id int64
	err = db.QueryRow(`
		INSERT INTO manga (path, title, page_count, catalog_id, series_title, metadata_json, parser_version)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(path) DO NOTHING RETURNING id
	`, path, title, pageCount, catalogID, seriesTitle, string(metadataJSON), parserVersion).Scan(&id)
	if err == sql.ErrNoRows {
		err = db.QueryRow(`SELECT id FROM manga WHERE path = ?`, path).Scan(&id)
	}
	return id, err
}

func (db *DB) UpdateMangaCatalog(id int64, catalogID *int64) error {
	if catalogID == nil {
		_, err := db.Exec(`UPDATE manga SET catalog_id = NULL WHERE id = ?`, id)
		return err
	}

	_, err := db.Exec(`UPDATE manga SET catalog_id = ? WHERE id = ?`, *catalogID, id)
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
