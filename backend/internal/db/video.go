package db

import "database/sql"

type VideoRow struct {
	ID        int64
	Path      string
	Title     string
	Duration  int
	Width     int
	Height    int
	Position  int
	Completed bool
	LastError string
}

func (db *DB) ListVideo(userID int64) ([]VideoRow, error) {
	rows, err := db.Query(`
		SELECT v.id, v.path, v.title, v.duration, v.width, v.height,
		       COALESCE(p.position, 0), COALESCE(p.completed, 0), COALESCE(p.last_error, '')
		FROM video v
		LEFT JOIN video_progress p ON p.video_id = v.id AND p.user_id = ?
		ORDER BY v.title
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []VideoRow
	for rows.Next() {
		var v VideoRow
		if err := rows.Scan(&v.ID, &v.Path, &v.Title, &v.Duration, &v.Width, &v.Height, &v.Position, &v.Completed, &v.LastError); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

func (db *DB) GetVideoByID(userID, id int64) (*VideoRow, error) {
	row := db.QueryRow(`
		SELECT v.id, v.path, v.title, v.duration, v.width, v.height,
		       COALESCE(p.position, 0), COALESCE(p.completed, 0), COALESCE(p.last_error, '')
		FROM video v
		LEFT JOIN video_progress p ON p.video_id = v.id AND p.user_id = ?
		WHERE v.id = ?
	`, userID, id)
	var v VideoRow
	if err := row.Scan(&v.ID, &v.Path, &v.Title, &v.Duration, &v.Width, &v.Height, &v.Position, &v.Completed, &v.LastError); err != nil {
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

func (db *DB) UpdateVideoMetadata(id int64, duration, width, height int) error {
	_, err := db.Exec(`
		UPDATE video
		SET duration = CASE WHEN ? > 0 THEN ? ELSE duration END,
		    width = CASE WHEN ? > 0 THEN ? ELSE width END,
		    height = CASE WHEN ? > 0 THEN ? ELSE height END
		WHERE id = ?
	`, duration, duration, width, width, height, height, id)
	return err
}

func (db *DB) UpsertVideoProgress(userID, videoID int64, position int, completed bool, lastError string) error {
	_, err := db.Exec(`
		INSERT INTO video_progress (user_id, video_id, position, completed, last_error, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, video_id) DO UPDATE SET
			position = excluded.position,
			completed = excluded.completed,
			last_error = excluded.last_error,
			updated_at = CURRENT_TIMESTAMP
	`, userID, videoID, position, completed, lastError)
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
	if _, err := db.Exec(`DELETE FROM video_progress WHERE video_id = ?`, id); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM video WHERE id = ?`, id)
	return err
}
