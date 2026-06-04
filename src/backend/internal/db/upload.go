package db

import "database/sql"

const (
	UploadStatusActive     = "active"
	UploadStatusPaused     = "paused"
	UploadStatusAssembling = "assembling"
	UploadStatusComplete   = "complete"
	UploadStatusFailed     = "failed"
	UploadStatusCancelled  = "cancelled"
)

type UploadSession struct {
	ID           string
	UserID       int64
	Filename     string
	Size         int64
	Profile      string
	LastModified int64
	ChunkSize    int64
	TotalChunks  int64
	Status       string
	TargetPath   string
	CatalogID    *int64
	ErrorMessage sql.NullString
	CreatedAt    string
	UpdatedAt    string
	CompletedAt  sql.NullString
}

type PendingUploadSession struct {
	UploadSession
	UploadedCount int64
}

func (db *DB) CreateUploadSession(s UploadSession) error {
	_, err := db.Exec(`
		INSERT INTO upload_sessions (
			id, user_id, filename, size, profile, last_modified, chunk_size, total_chunks,
			status, target_path, catalog_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, s.ID, s.UserID, s.Filename, s.Size, s.Profile, s.LastModified, s.ChunkSize, s.TotalChunks, s.Status, s.TargetPath, s.CatalogID)
	return err
}

func (db *DB) FindResumableUploadSession(userID int64, filename string, size, lastModified int64, profile string, catalogID *int64) (*UploadSession, error) {
	row := db.QueryRow(`
		SELECT id, user_id, filename, size, profile, last_modified, chunk_size, total_chunks,
			status, target_path, catalog_id, error_message, created_at, updated_at, completed_at
		FROM upload_sessions
		WHERE user_id = ?
			AND filename = ?
			AND size = ?
			AND last_modified = ?
			AND profile = ?
			AND (catalog_id = ? OR (catalog_id IS NULL AND ? IS NULL))
			AND status IN ('active', 'paused', 'failed')
		ORDER BY updated_at DESC
		LIMIT 1
	`, userID, filename, size, lastModified, profile, catalogID, catalogID)
	return scanUploadSession(row)
}

func (db *DB) GetUploadSession(id string, userID int64) (*UploadSession, error) {
	row := db.QueryRow(`
		SELECT id, user_id, filename, size, profile, last_modified, chunk_size, total_chunks,
			status, target_path, catalog_id, error_message, created_at, updated_at, completed_at
		FROM upload_sessions
		WHERE id = ? AND user_id = ?
	`, id, userID)
	return scanUploadSession(row)
}

func (db *DB) ListPendingUploadSessions(userID int64) ([]PendingUploadSession, error) {
	rows, err := db.Query(`
		SELECT s.id, s.user_id, s.filename, s.size, s.profile, s.last_modified, s.chunk_size,
			s.total_chunks, s.status, s.target_path, s.catalog_id, s.error_message, s.created_at, s.updated_at,
			s.completed_at, COUNT(c.chunk_index) AS uploaded_count
		FROM upload_sessions s
		LEFT JOIN upload_chunks c ON c.upload_id = s.id
		WHERE s.user_id = ?
			AND s.status IN ('active', 'paused', 'failed', 'assembling')
		GROUP BY s.id
		ORDER BY s.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	uploads := make([]PendingUploadSession, 0)
	for rows.Next() {
		var upload PendingUploadSession
		if err := rows.Scan(
			&upload.ID,
			&upload.UserID,
			&upload.Filename,
			&upload.Size,
			&upload.Profile,
			&upload.LastModified,
			&upload.ChunkSize,
			&upload.TotalChunks,
			&upload.Status,
			&upload.TargetPath,
			&upload.CatalogID,
			&upload.ErrorMessage,
			&upload.CreatedAt,
			&upload.UpdatedAt,
			&upload.CompletedAt,
			&upload.UploadedCount,
		); err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, rows.Err()
}

func (db *DB) UpsertUploadChunk(uploadID string, chunkIndex int, size int64, path string) error {
	_, err := db.Exec(`
		INSERT INTO upload_chunks (upload_id, chunk_index, size, path, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(upload_id, chunk_index) DO UPDATE SET
			size = excluded.size,
			path = excluded.path,
			created_at = CURRENT_TIMESTAMP
	`, uploadID, chunkIndex, size, path)
	return err
}

func (db *DB) ListUploadChunkIndexes(uploadID string) ([]int, error) {
	rows, err := db.Query(`
		SELECT chunk_index
		FROM upload_chunks
		WHERE upload_id = ?
		ORDER BY chunk_index
	`, uploadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make([]int, 0)
	for rows.Next() {
		var index int
		if err := rows.Scan(&index); err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}
	return indexes, rows.Err()
}

func (db *DB) CountUploadChunks(uploadID string) (int64, error) {
	var count int64
	err := db.QueryRow(`SELECT COUNT(*) FROM upload_chunks WHERE upload_id = ?`, uploadID).Scan(&count)
	return count, err
}

func (db *DB) UpdateUploadSessionStatus(uploadID, status string) error {
	_, err := db.Exec(`
		UPDATE upload_sessions
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, uploadID)
	return err
}

func (db *DB) FailUploadSession(uploadID string, uploadErr error) error {
	message := ""
	if uploadErr != nil {
		message = uploadErr.Error()
	}
	_, err := db.Exec(`
		UPDATE upload_sessions
		SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status NOT IN ('paused', 'cancelled', 'complete')
	`, UploadStatusFailed, message, uploadID)
	return err
}

func (db *DB) CompleteUploadSession(uploadID string) error {
	_, err := db.Exec(`
		UPDATE upload_sessions
		SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, UploadStatusComplete, uploadID)
	return err
}

func (db *DB) CancelUploadSession(uploadID string) error {
	_, err := db.Exec(`
		UPDATE upload_sessions
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, UploadStatusCancelled, uploadID)
	return err
}

func (db *DB) DeleteUploadChunks(uploadID string) error {
	_, err := db.Exec(`DELETE FROM upload_chunks WHERE upload_id = ?`, uploadID)
	return err
}

func scanUploadSession(row *sql.Row) (*UploadSession, error) {
	var s UploadSession
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.Filename,
		&s.Size,
		&s.Profile,
		&s.LastModified,
		&s.ChunkSize,
		&s.TotalChunks,
		&s.Status,
		&s.TargetPath,
		&s.CatalogID,
		&s.ErrorMessage,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
