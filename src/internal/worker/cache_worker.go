package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// interval: cache dir polling interval
// maxSizeMB: max size for cache dir
func StartCacheCleaner(cacheDir string, maxSizeMB int64, interval time.Duration) {
	maxSizeBytes := maxSizeMB * 1024 * 1024

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			if err := enforceCacheSize(cacheDir, maxSizeBytes); err != nil {
				fmt.Printf("[worker] Cache cleanup failed: %v", err)
			}
		}
	}()
}

func enforceCacheSize(cacheDir string, maxBytes int64) error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	var totalSize int64
	type cacheFile struct {
		path    string
		size    int64
		modTime time.Time
	}
	var files []cacheFile

	// get total dir size
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
		files = append(files, cacheFile{
			path:    filepath.Join(cacheDir, entry.Name()),
			size:    info.Size(),
			modTime: info.ModTime(),
		})
	}

	if totalSize < maxBytes {
		return nil
	}

	// sort files by modTime
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// delete least used files (modTime < )
	// reduce dir size by 20%
	targetSize := int64(float64(maxBytes) * 0.8)
	deletedCount := 0
	for _, f := range files {
		if totalSize <= targetSize {
			break
		}
		if err := os.Remove(f.path); err == nil {
			totalSize -= f.size
			deletedCount++
		}
	}

	if deletedCount > 0 {
		fmt.Printf("[worker] Cache cleaned: removed %d old files\n", deletedCount)
	}

	return nil
}
