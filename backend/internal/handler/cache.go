package handler

import (
	"fmt"
	"ohara/src/internal/db"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MaxPageCacheSizeBytes int64 = 1 << 30 // 1 GB

type pageKey struct {
	mangaID int64
	page    int
}

type PageCache struct {
	dir     string
	db      *db.DB
	maxSize int64
}

func NewPageCache(dataDir string, database *db.DB) *PageCache {
	cacheDir := filepath.Join(dataDir, "cache")
	os.MkdirAll(cacheDir, 0o755)
	return &PageCache{dir: cacheDir, db: database, maxSize: MaxPageCacheSizeBytes}
}

func (c *PageCache) getPath(mangaID int64, page int) string {
	return filepath.Join(c.dir, fmt.Sprintf("%d", mangaID), fmt.Sprintf("%d.jpg", page))
}

func (c *PageCache) getLegacyPath(mangaID int64, page int) string {
	return filepath.Join(c.dir, fmt.Sprintf("%d_%d.jpg", mangaID, page))
}

func (c *PageCache) Get(mangaID int64, page int) ([]byte, bool) {
	for _, path := range []string{c.getPath(mangaID, page), c.getLegacyPath(mangaID, page)} {
		data, err := os.ReadFile(path)
		if err == nil {
			c.touch(path)
			return data, true
		}
	}
	return nil, false
}

func (c *PageCache) Set(mangaID int64, page int, data []byte) {
	finalPath := c.getPath(mangaID, page)
	tempPath := finalPath + ".tmp"

	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return
	}
	if err := os.WriteFile(tempPath, data, 0o644); err == nil {
		if os.Rename(tempPath, finalPath) == nil {
			_ = c.EnforceMaxSize()
		}
	}
}

func (c *PageCache) touch(path string) {
	now := time.Now()
	_ = os.Chtimes(path, now, now)
}

func (c *PageCache) EnforceMaxSize() error {
	entries, total, err := c.cacheEntries()
	if err != nil {
		return err
	}
	if total <= c.maxSize {
		return nil
	}

	progress := map[int64]int{}
	if c.db != nil {
		if progress, err = c.db.MaxMangaProgressByID(); err != nil {
			return err
		}
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
		if total <= c.maxSize {
			break
		}
		if err := os.Remove(entry.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		total -= entry.size
	}

	return nil
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

func (c *PageCache) cacheEntries() ([]cacheEntry, int64, error) {
	var entries []cacheEntry
	var total int64

	err := filepath.WalkDir(c.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".jpg" {
			return nil
		}

		mangaID, page, ok := c.parseCachePath(path)
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

func (c *PageCache) parseCachePath(path string) (int64, int, bool) {
	rel, err := filepath.Rel(c.dir, path)
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

func (c *PageCache) DeleteManga(mangaID int64) error {
	flatPatterns := []string{
		filepath.Join(c.dir, fmt.Sprintf("%d_*.jpg", mangaID)),
		filepath.Join(c.dir, fmt.Sprintf("%d_*.jpg.tmp", mangaID)),
	}

	for _, pattern := range flatPatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, path := range matches {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	mangaDir := filepath.Join(c.dir, fmt.Sprintf("%d", mangaID))
	if err := os.RemoveAll(mangaDir); err != nil {
		return err
	}

	return nil
}

type inflightCall struct {
	wg  sync.WaitGroup
	val []byte
	err error
}

type Inflight struct {
	mu    sync.Mutex
	calls map[pageKey]*inflightCall
}

func NewInflight() *Inflight {
	return &Inflight{calls: make(map[pageKey]*inflightCall)}
}

func (g *Inflight) Do(mangaID int64, page int, fn func() ([]byte, error)) ([]byte, error) {
	key := pageKey{mangaID, page}

	g.mu.Lock()
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &inflightCall{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return c.val, c.err
}
