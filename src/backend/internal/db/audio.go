package db

import (
	"database/sql"
	"ohara/src/internal/media/audio"
)

type AudioRow struct {
	ID        int64
	Path      string
	CatalogID *int64
	Title     string
	Artist    string
	Album     string
	Duration  int
}

func (db *DB) ListAudio() ([]AudioRow, error) {
	rows, err := db.Query(`SELECT id, path, catalog_id, title, artist, album, duration FROM audio ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AudioRow
	for rows.Next() {
		var a AudioRow
		var catalogID sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Path, &catalogID, &a.Title, &a.Artist, &a.Album, &a.Duration); err != nil {
			return nil, err
		}
		if catalogID.Valid {
			a.CatalogID = &catalogID.Int64
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (db *DB) GetAudioByID(id int64) (*AudioRow, error) {
	row := db.QueryRow(`SELECT id, path, catalog_id, title, artist, album, duration FROM audio WHERE id = ?`, id)
	var a AudioRow
	var catalogID sql.NullInt64
	if err := row.Scan(&a.ID, &a.Path, &catalogID, &a.Title, &a.Artist, &a.Album, &a.Duration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if catalogID.Valid {
		a.CatalogID = &catalogID.Int64
	}
	return &a, nil
}

func (db *DB) InsertAudio(track *audio.Track) error {
	return db.InsertAudioWithCatalog(track, nil)
}

func (db *DB) InsertAudioWithCatalog(track *audio.Track, catalogID *int64) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO audio (path, title, artist, album, duration, catalog_id) VALUES (?, ?, ?, ?, ?, ?)`,
		track.FilePath, track.Title, track.Artist, track.Album, track.Duration, catalogID,
	)
	return err
}

func (db *DB) UpdateAudioCatalog(id int64, catalogID *int64) error {
	if catalogID == nil {
		_, err := db.Exec(`UPDATE audio SET catalog_id = NULL WHERE id = ?`, id)
		return err
	}

	_, err := db.Exec(`UPDATE audio SET catalog_id = ? WHERE id = ?`, *catalogID, id)
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
