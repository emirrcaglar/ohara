package worker

import (
	"bytes"
	"database/sql"
	"fmt"
	"ohara/src/internal/db"
	"ohara/src/internal/media/manga"
	"ohara/src/internal/utils/imgutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxPageCacheSizeBytes int64 = 1 << 30

type CacheWorker struct {
	dataDir string
	db      *db.DB
	wakeUp  chan struct{}
}

func NewCacheWorker(dataDir string, database *db.DB) *CacheWorker {
	return &CacheWorker{
		dataDir: dataDir,
		db:      database,
		wakeUp:  make(chan struct{}, 1),
	}
}

func (cw *CacheWorker) Start() {
	go func() {
		for {
			job, err := cw.db.ClaimNextCacheJob()
			if err != nil {
				if err == sql.ErrNoRows {
					<-cw.wakeUp
					continue
				}
				fmt.Println("DB Claim Error:", err)
				time.Sleep(3 * time.Second)
				continue
			}

			hasProgress, err := cw.db.HasMangaProgress(job.MangaID)
			if err != nil {
				cw.db.MarkCacheJobFailed(job.ID, err.Error())
				fmt.Println("job failed:", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if hasProgress {
				cw.db.DeleteCacheJob(job.ID)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			err = cw.cachePage(job.MangaID, job.PageIndex)

			if err != nil {
				cw.db.MarkCacheJobFailed(job.ID, err.Error())
				fmt.Println("job failed:", err)
			} else {
				cw.db.DeleteCacheJob(job.ID)
			}

			// Explicitly sleep for 500ms after a job to throttle CPU
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (cw *CacheWorker) EnqueuePage(mangaID int64, pageIndex int) error {
	if err := cw.db.InsertCacheJob(mangaID, pageIndex); err != nil {
		return fmt.Errorf("enqueue manga %d page %d: %w", mangaID, pageIndex, err)
	}

	select {
	case cw.wakeUp <- struct{}{}:
	default:
	}
	return nil
}

func (cw *CacheWorker) cachePage(mangaID int64, pageIndex int) error {
	record, err := cw.db.GetMangaByID(mangaID)
	if err != nil {
		return fmt.Errorf("load manga %d for page cache: %w", mangaID, err)
	}
	if record == nil {
		return fmt.Errorf("manga not found: %d", mangaID)
	}
	data, err := cw.compressPage(record, pageIndex)
	if err != nil {
		return err
	}

	return cw.writePage(record, pageIndex, data)
}

func (cw *CacheWorker) compressPage(m *db.MangaRow, pageIndex int) ([]byte, error) {
	archive, err := manga.Open(m.Path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	rc, err := archive.OpenPage(pageIndex)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if err := imgutil.Compress(rc, &buf, 1200, 70); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (cw *CacheWorker) writePage(m *db.MangaRow, pageIndex int, data []byte) error {
	mangaCacheDir := filepath.Join(cw.dataDir, "cache", strconv.FormatInt(m.ID, 10))
	finalFilePath := filepath.Join(mangaCacheDir, fmt.Sprintf("%d.jpg", pageIndex))

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

	return cw.enforceMaxCacheSize()
}

type cacheEntry struct {
	path    string
	size    int64
	modTime time.Time
	mangaID int64
	page    int
}

func (e cacheEntry) isRead(progress map[int64]int) bool {
	page, ok := progress[e.mangaID]
	return ok && e.page <= page
}

func (cw *CacheWorker) enforceMaxCacheSize() error {
	entries, total, err := cw.cacheEntries()
	if err != nil {
		return err
	}
	if total <= maxPageCacheSizeBytes {
		return nil
	}

	progress, err := cw.db.MaxMangaProgressByID()
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		iRead := entries[i].isRead(progress)
		jRead := entries[j].isRead(progress)
		if iRead != jRead {
			return !iRead
		}
		return entries[i].modTime.Before(entries[j].modTime)
	})

	for _, entry := range entries {
		if total <= maxPageCacheSizeBytes {
			break
		}
		if err := os.Remove(entry.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		total -= entry.size
	}

	return nil
}

func (cw *CacheWorker) cacheEntries() ([]cacheEntry, int64, error) {
	cacheDir := filepath.Join(cw.dataDir, "cache")
	var entries []cacheEntry
	var total int64

	err := filepath.WalkDir(cacheDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".jpg" {
			return nil
		}

		mangaID, page, ok := parseCachePath(cacheDir, path)
		if !ok {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, cacheEntry{path: path, size: info.Size(), modTime: info.ModTime(), mangaID: mangaID, page: page})
		total += info.Size()
		return nil
	})
	return entries, total, err
}

func parseCachePath(cacheDir, path string) (int64, int, bool) {
	rel, err := filepath.Rel(cacheDir, path)
	if err != nil {
		return 0, 0, false
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) == 2 {
		mangaID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		pageStr := strings.TrimSuffix(parts[1], ".jpg")
		page, err := strconv.Atoi(pageStr)
		return mangaID, page, err == nil
	}
	if len(parts) == 1 {
		name := strings.TrimSuffix(parts[0], ".jpg")
		legacyParts := strings.Split(name, "_")
		if len(legacyParts) != 2 {
			return 0, 0, false
		}
		mangaID, err := strconv.ParseInt(legacyParts[0], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		page, err := strconv.Atoi(legacyParts[1])
		return mangaID, page, err == nil
	}
	return 0, 0, false
}
