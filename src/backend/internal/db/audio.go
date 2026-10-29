package db

import (
	"database/sql"
	"ohara/src/internal/media/audio"
)

type AudioRow struct {
	ID       int64
	Path     string
	FolderID *int64
	Title    string
	Artist   string
	Album    string
	Duration int
}

func (db *DB) ListAudio() ([]AudioRow, error) {
	rows, err := db.Query(`SELECT id, path, folder_id, title, artist, album, duration FROM audio ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AudioRow
	for rows.Next() {
		var a AudioRow
		var folderID sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Path, &folderID, &a.Title, &a.Artist, &a.Album, &a.Duration); err != nil {
			return nil, err
		}
		if folderID.Valid {
			a.FolderID = &folderID.Int64
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (db *DB) GetAudioByID(id int64) (*AudioRow, error) {
	row := db.QueryRow(`SELECT id, path, folder_id, title, artist, album, duration FROM audio WHERE id = ?`, id)
	var a AudioRow
	var folderID sql.NullInt64
	if err := row.Scan(&a.ID, &a.Path, &folderID, &a.Title, &a.Artist, &a.Album, &a.Duration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if folderID.Valid {
		a.FolderID = &folderID.Int64
	}
	return &a, nil
}

func (db *DB) InsertAudio(track *audio.Track) error {
	return db.InsertAudioWithFolder(track, nil)
}

func (db *DB) InsertAudioWithFolder(track *audio.Track, folderID *int64) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO audio (path, title, artist, album, duration, folder_id) VALUES (?, ?, ?, ?, ?, ?)`,
		track.FilePath, track.Title, track.Artist, track.Album, track.Duration, folderID,
	)
	return err
}

func (db *DB) UpdateAudioFolder(id int64, folderID *int64) error {
	if folderID == nil {
		_, err := db.Exec(`UPDATE audio SET folder_id = NULL WHERE id = ?`, id)
		return err
	}

	_, err := db.Exec(`UPDATE audio SET folder_id = ? WHERE id = ?`, *folderID, id)
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
