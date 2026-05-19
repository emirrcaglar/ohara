package db

type ScanStatus string

const (
	ScanStatusPending    ScanStatus = "pending"
	ScanStatusProcessing ScanStatus = "processing"
	ScanStatusCompleted  ScanStatus = "completed"
	ScanStatusError      ScanStatus = "error"
)

type Scan struct {
	ID           int64      `db:"id"`
	MangaID      int64      `db:"manga_id"`
	PageIdx      int        `db:"page_idx"`
	Status       ScanStatus `db:"status"`
	Attempts     int        `db:"attempts"`
	ErrorMessage string     `db:"error_message"`
	CreatedAt    string     `db:"created_at"`
	UpdatedAt    string     `db:"updated_at"`
}

func (db *DB) InsertScanJob(mangaID int64, pageIdx int) error {
	_, err := db.Exec(`
		INSERT INTO scan (manga_id, page_idx, status)
		VALUES (?, ?, 'pending')
		`, mangaID, pageIdx)
	return err
}

func (db *DB) ClaimNextScanJob() (*Scan, error) {
	row := db.QueryRow(`
		UPDATE scan
		SET status = 'processing', updated_at = CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM scan
			WHERE status = 'pending'
			ORDER BY created_at ASC
			LIMIT 1
		)
		RETURNING id, manga_id, page_idx, attempts
	`)

	var s Scan
	err := row.Scan(&s.ID, &s.MangaID, &s.PageIdx, &s.Attempts)
	if err != nil {
		return nil, err
	}
	s.Status = ScanStatusProcessing
	return &s, nil
}

func (db *DB) MarkScanJobFailed(id int64, errMsg string) error {
	_, err := db.Exec(`
		UPDATE scan
		SET status = 'failed', attempts = attempts + 1, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		`, errMsg, id)
	return err
}

func (db *DB) DeleteScan(id int64) error {
	_, err := db.Exec("DELETE FROM scan WHERE id = ?", id)
	return err
}
