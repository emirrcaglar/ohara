package db

import (
	"database/sql"
	"ohara/src/internal/media/audio"
)

type MangaRow struct {
	ID        int64
	Path      string
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
		SELECT m.id, m.path, m.title, m.page_count, COALESCE(p.page, 0)
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
		if err := rows.Scan(&m.ID, &m.Path, &m.Title, &m.PageCount, &m.Progress); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (db *DB) GetMangaByID(id int64) (*MangaRow, error) {
	row := db.QueryRow(`SELECT id, path, title, page_count FROM manga WHERE id = ?`, id)
	var m MangaRow
	if err := row.Scan(&m.ID, &m.Path, &m.Title, &m.PageCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
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

func (db *DB) UpsertProgress(userID, mangaID int64, page int) error {
	_, err := db.Exec(`
		INSERT INTO manga_progress (user_id, manga_id, page, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, manga_id) DO UPDATE SET page = excluded.page, updated_at = excluded.updated_at
	`, userID, mangaID, page)
	return err
}

func (db *DB) InsertManga(path, title string, pageCount int) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO manga (path, title, page_count) VALUES (?, ?, ?)`,
		path, title, pageCount,
	)
	return err
}

func (db *DB) IndexedAudioPaths() (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT path FROM audio`)
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

func (db *DB) InsertAudio(track *audio.Track) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO audio (path, title, artist, album, duration) VALUES (?, ?, ?, ?, ?)`,
		track.FilePath, track.Title, track.Artist, track.Album, track.Duration,
	)
	return err
}
