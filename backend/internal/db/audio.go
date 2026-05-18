package db

import (
	"database/sql"
	"ohara/src/internal/media/audio"
)

type AudioRow struct {
	ID       int64
	Path     string
	Title    string
	Artist   string
	Album    string
	Duration int
}

func (db *DB) ListAudio() ([]AudioRow, error) {
	rows, err := db.Query(`SELECT id, path, title, artist, album, duration FROM audio ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AudioRow
	for rows.Next() {
		var a AudioRow
		if err := rows.Scan(&a.ID, &a.Path, &a.Title, &a.Artist, &a.Album, &a.Duration); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (db *DB) GetAudioByID(id int64) (*AudioRow, error) {
	row := db.QueryRow(`SELECT id, path, title, artist, album, duration FROM audio WHERE id = ?`, id)
	var a AudioRow
	if err := row.Scan(&a.ID, &a.Path, &a.Title, &a.Artist, &a.Album, &a.Duration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (db *DB) InsertAudio(track *audio.Track) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO audio (path, title, artist, album, duration) VALUES (?, ?, ?, ?, ?)`,
		track.FilePath, track.Title, track.Artist, track.Album, track.Duration,
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
