package indexer

import (
	"log"
	"path/filepath"

	"ohara/src/internal/db"
	cbzReader "ohara/src/internal/media/cbz"
)

func Run(database *db.DB, mangaDir string) (int, error) {
	matches, err := filepath.Glob(filepath.Join(mangaDir, "*.cbz"))
	if err != nil {
		return 0, err
	}

	indexed, err := database.IndexedPaths()
	if err != nil {
		return 0, err
	}

	added := 0
	for _, path := range matches {
		abs, err := filepath.Abs(path)
		if err != nil {
			log.Printf("indexer: skipping %s: %v", path, err)
			continue
		}

		if _, exists := indexed[abs]; exists {
			continue
		}

		manga, err := cbzReader.Open(abs)
		if err != nil {
			log.Printf("indexer: skipping %s: %v", abs, err)
			continue
		}
		manga.Close()

		if err := database.InsertManga(abs, manga.Title, manga.PageCount); err != nil {
			log.Printf("indexer: failed to insert %s: %v", abs, err)
			continue
		}

		log.Printf("indexer: indexed %q (%d pages)", manga.Title, manga.PageCount)
		added++
	}

	return added, nil
}
