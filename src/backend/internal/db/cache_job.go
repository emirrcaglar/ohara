package db

type CacheJobStatus string

const (
	CacheJobStatusPending    CacheJobStatus = "pending"
	CacheJobStatusProcessing CacheJobStatus = "processing"
	CacheJobStatusFailed     CacheJobStatus = "failed"
)

// CacheJob is a page-cache task. The legacy scan table remains its storage format.
// Successful tasks are deleted; failed tasks retain their error and attempt count.
type CacheJob struct {
	ID           int64          `db:"id"`
	MangaID      int64          `db:"manga_id"`
	PageIndex    int            `db:"page_idx"`
	Status       CacheJobStatus `db:"status"`
	Attempts     int            `db:"attempts"`
	ErrorMessage string         `db:"error_message"`
	CreatedAt    string         `db:"created_at"`
	UpdatedAt    string         `db:"updated_at"`
}

func (db *DB) InsertCacheJob(mangaID int64, pageIndex int) error {
	_, err := db.Exec(`
		INSERT INTO scan (manga_id, page_idx, status)
		VALUES (?, ?, ?)
		`, mangaID, pageIndex, CacheJobStatusPending)
	return err
}

func (db *DB) ClaimNextCacheJob() (*CacheJob, error) {
	row := db.QueryRow(`
		UPDATE scan
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM scan
			WHERE status = ?
			ORDER BY created_at ASC
			LIMIT 1
		)
		RETURNING id, manga_id, page_idx, attempts
	`, CacheJobStatusProcessing, CacheJobStatusPending)

	var job CacheJob
	err := row.Scan(&job.ID, &job.MangaID, &job.PageIndex, &job.Attempts)
	if err != nil {
		return nil, err
	}
	job.Status = CacheJobStatusProcessing
	return &job, nil
}

func (db *DB) MarkCacheJobFailed(id int64, errMsg string) error {
	_, err := db.Exec(`
		UPDATE scan
		SET status = ?, attempts = attempts + 1, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		`, CacheJobStatusFailed, errMsg, id)
	return err
}

func (db *DB) DeleteCacheJob(id int64) error {
	_, err := db.Exec("DELETE FROM scan WHERE id = ?", id)
	return err
}
