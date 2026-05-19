package worker

import (
	"bytes"
	"database/sql"
	"fmt"
	"ohara/src/internal/db"
	"ohara/src/internal/media/cbz"
	"ohara/src/internal/utils/imgutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const MANGA_NOT_FOUND = "manga not found"

type CacheWorker struct {
	dataDir string
	db      *db.DB
	cbz     cbz.CBZService
	wakeUp  chan struct{}
}

type CompressJob struct {
	MangaID int64
	PageIdx int
}

func NewCacheWorker(dataDir string, db *db.DB, cbz cbz.CBZService) *CacheWorker {
	return &CacheWorker{
		dataDir: dataDir,
		db:      db,
		cbz:     cbz,
		wakeUp:  make(chan struct{}, 1),
	}
}

func (cw *CacheWorker) Start() {
	go func() {
		for {
			job, err := cw.db.ClaimNextScanJob()
			if err != nil {
				if err == sql.ErrNoRows {
					<-cw.wakeUp
					continue
				}
				fmt.Println("DB Claim Error:", err)
				time.Sleep(3 * time.Second)
				continue
			}

			err = cw.CompressImg(job.MangaID, job.PageIdx)

			if err != nil {
				cw.db.MarkScanJobFailed(job.ID, err.Error())
				fmt.Println("job failed:", err)
			} else {
				cw.db.DeleteScan(job.ID)
			}

			// Explicitly sleep for 500ms after a job to throttle CPU
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (cw *CacheWorker) SubmitJob(mangaID int64, pageIdx int) {
	if err := cw.db.InsertScanJob(mangaID, pageIdx); err != nil {
		fmt.Println(err)
	}

	select {
	case cw.wakeUp <- struct{}{}:
	default:
	}
}

func (cw *CacheWorker) CompressImg(mangaID int64, pageIdx int) error {
	manga, err := cw.db.GetMangaByID(mangaID)
	if err != nil || manga == nil {
		return fmt.Errorf(MANGA_NOT_FOUND)
	}

	err = cw.ImgCompressWorker(manga, pageIdx)
	return err
}

func (cw *CacheWorker) ImgCompressWorker(m *db.MangaRow, pageIdx int) error {
	data, err := cw.compressPage(m, pageIdx)
	if err != nil {
		return err
	}

	err = cw.writeToFile(m, pageIdx, data)
	return err
}

func (cw *CacheWorker) compressPage(m *db.MangaRow, pageIdx int) ([]byte, error) {
	manga, err := cw.cbz.Open(m.Path)
	if err != nil {
		return nil, err
	}
	defer manga.Close()

	rc, err := manga.GetPageReader(pageIdx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if err := imgutil.Compress(rc, &buf, 1200, 70); err != nil {
		return nil, err
	}

	data := buf.Bytes()

	return data, err
}

func (cw *CacheWorker) writeToFile(m *db.MangaRow, pageIdx int, data []byte) error {
	mangaCacheDir := filepath.Join(cw.dataDir, "cache", strconv.FormatInt(m.ID, 10))
	finalFilePath := filepath.Join(mangaCacheDir, fmt.Sprintf("%d.jpg", pageIdx))

	if err := os.MkdirAll(mangaCacheDir, 0755); err != nil {
		return err
	}

	tempFilePath := finalFilePath + ".tmp"
	if err := os.WriteFile(tempFilePath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tempFilePath, finalFilePath); err != nil {
		os.Remove(tempFilePath)
		return err
	}

	return nil
}
