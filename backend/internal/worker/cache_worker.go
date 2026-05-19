package worker

import (
	"bytes"
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
	dataDir  string
	db       *db.DB
	cbz      cbz.CBZService
	jobQueue chan CompressJob
}

type CompressJob struct {
	MangaID int64
	PageIdx int
}

func NewCacheWorker(dataDir string, db *db.DB, cbz cbz.CBZService, jobQueue chan CompressJob) *CacheWorker {
	return &CacheWorker{
		dataDir:  dataDir,
		db:       db,
		cbz:      cbz,
		jobQueue: jobQueue,
	}
}

func (cw *CacheWorker) Start() {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for job := range cw.jobQueue {
			<-ticker.C
			err := cw.CompressImg(job)
			if err != nil {
				if err == fmt.Errorf(MANGA_NOT_FOUND) {
					continue
				}
				fmt.Println(err)
			}
		}
	}()
}

func (cw *CacheWorker) SubmitJob(mangaID int64, pageIdx int) {
	job := CompressJob{
		MangaID: mangaID,
		PageIdx: pageIdx,
	}
	cw.jobQueue <- job
}

func (cw *CacheWorker) CompressImg(job CompressJob) error {
	manga, err := cw.db.GetMangaByID(job.MangaID)
	if err != nil || manga == nil {
		return fmt.Errorf(MANGA_NOT_FOUND)
	}

	err = cw.ImgCompressWorker(manga, job.PageIdx)
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
