package db

import (
	"database/sql"
)

type MangaRow struct {
	ID        int64
	Path      string
	CatalogID *int64
	Title     string
	PageCount int
	Progress  int
}

func (db *DB) IndexedPaths() (map[string]struct{}, error) {
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
		SELECT m.id, m.path, m.catalog_id, m.title, m.page_count, COALESCE(p.page, 0)
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
		if err := rows.Scan(&m.ID, &m.Path, &catalogID, &m.Title, &m.PageCount, &m.Progress); err != nil {
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
	row := db.QueryRow(`SELECT id, path, catalog_id, title, page_count FROM manga WHERE id = ?`, id)
	var m MangaRow
	var catalogID sql.NullInt64
	if err := row.Scan(&m.ID, &m.Path, &catalogID, &m.Title, &m.PageCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if catalogID.Valid {
		m.CatalogID = &catalogID.Int64
	}
	return &m, nil
}

func (db *DB) GetProgress(userID, mangaID int64) (int, error) {
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

func (db *DB) HasMangaProgress(mangaID int64) (bool, error) {
	var exists int
	err := db.QueryRow(`SELECT 1 FROM manga_progress WHERE manga_id = ? LIMIT 1`, mangaID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (db *DB) MaxMangaProgressByID() (map[int64]int, error) {
	rows, err := db.Query(`SELECT manga_id, MAX(page) FROM manga_progress GROUP BY manga_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	progress := make(map[int64]int)
	for rows.Next() {
		var mangaID int64
		var page int
		if err := rows.Scan(&mangaID, &page); err != nil {
			return nil, err
		}
		progress[mangaID] = page
	}
	return progress, rows.Err()
}

func (db *DB) UpsertProgress(userID, mangaID int64, page int) error {
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
	sq, err := db.Exec(
		`INSERT OR IGNORE INTO manga (path, title, page_count, catalog_id) VALUES (?, ?, ?, ?)`,
		path, title, pageCount, catalogID,
	)
	if err != nil {
		return 0, err
	}
	id, err := sq.LastInsertId()
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
	if _, err := tx.Exec(`DELETE FROM scan WHERE manga_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM manga WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}
