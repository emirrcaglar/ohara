package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ohara/src/internal/logger"
)

type DiskCache struct {
	Dir string
	Log *logger.Logger
}

func NewDiskCache(dataDir string, log *logger.Logger) *DiskCache {
	cacheDir := filepath.Join(dataDir, "cache")
	os.MkdirAll(cacheDir, 0o755)
	return &DiskCache{Dir: cacheDir, Log: log}
}

func (c *DiskCache) cachePath(mangaID int64, page int) string {
	return filepath.Join(c.Dir, fmt.Sprintf("%d_%d.jpg", mangaID, page))
}

func (c *DiskCache) Get(mangaID int64, page int) ([]byte, bool) {
	data, err := os.ReadFile(c.cachePath(mangaID, page))
	if err != nil {
		return nil, false
	}

	// update time for cleanup worker (LRU)
	currentTime := time.Now()
	os.Chtimes(c.cachePath(mangaID, page), currentTime, currentTime)

	if c.Log != nil {
		c.Log.Info("[cache] hit manga=%d page=%d", mangaID, page)
	}
	return data, true
}

func (c *DiskCache) Set(mangaID int64, page int, data []byte) {
	// write to tmp because rename is atomic, write is not
	tempPath := c.cachePath(mangaID, page) + ".tmp"
	os.WriteFile(tempPath, data, 0o644)
	os.Rename(tempPath, c.cachePath(mangaID, page))
	if c.Log != nil {
		c.Log.Info("[cache] stored manga=%d page=%d", mangaID, page)
	}
}
