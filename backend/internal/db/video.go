package db

import "database/sql"

type VideoRow struct {
	ID       int64
	Path     string
	Title    string
	Duration int
}

func (db *DB) ListVideo() ([]VideoRow, error) {
	rows, err := db.Query(`SELECT id, path, title, duration FROM video ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []VideoRow
	for rows.Next() {
		var v VideoRow
		if err := rows.Scan(&v.ID, &v.Path, &v.Title, &v.Duration); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

func (db *DB) GetVideoByID(id int64) (*VideoRow, error) {
	row := db.QueryRow(`SELECT id, path, title, duration FROM video WHERE id = ?`, id)
	var v VideoRow
	if err := row.Scan(&v.ID, &v.Path, &v.Title, &v.Duration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (db *DB) InsertVideo(path, title string, duration int) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO video (path, title, duration) VALUES (?, ?, ?)`,
		path, title, duration,
	)
	return err
}

func (db *DB) IndexedVideoPaths() (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT path FROM video`)
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

func (db *DB) DeleteVideo(id int64) error {
	_, err := db.Exec(`DELETE FROM video WHERE id = ?`, id)
	return err
}
