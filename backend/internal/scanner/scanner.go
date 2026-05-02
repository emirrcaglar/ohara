package scanner

import (
	"fmt"
	"path/filepath"

	"ohara/src/internal/db"
	"ohara/src/internal/media/audio"
	"ohara/src/internal/media/cbz"
)

type Scanner struct {
	db       *db.DB
	scanDir  string
	scanType ScanType

	CBZService cbz.ICBZService
}

type ScannerOption func(*Scanner)

type ScanType string

var (
	ScanTypeAll   ScanType = "all"
	ScanTypeManga ScanType = "manga"
	ScanTypeAudio ScanType = "audio"
)

func WithScanDir(scanDir string) ScannerOption {
	return func(s *Scanner) {
		s.scanDir = scanDir
	}
}

func WithScanType(scanType ScanType) ScannerOption {
	return func(s *Scanner) {
		s.scanType = scanType
	}
}

func NewScanner(db *db.DB, cbzService cbz.ICBZService, opts ...ScannerOption) *Scanner {
	scanner := &Scanner{
		db:         db,
		CBZService: cbzService,
		scanDir:    ".",
		scanType:   ScanTypeAll,
	}

	for _, opt := range opts {
		opt(scanner)
	}

	return scanner
}

func (s *Scanner) Index(targetPath string) error {
	fileType := filepath.Ext(targetPath)

	switch fileType {
	case ".cbz":
		s.indexManga(targetPath)

	default:
		return fmt.Errorf("unsupported file type: %s", fileType)
	}

	return nil
}

func (s *Scanner) Run() (int, error) {
	added := 0

	if s.scanType == ScanTypeManga || s.scanType == ScanTypeAll {
		n, err := s.scanManga()
		if err != nil {
			return added, err
		}
		added += n
	}

	if s.scanType == ScanTypeAudio || s.scanType == ScanTypeAll {
		n, err := s.scanAudio()
		if err != nil {
			return added, err
		}
		added += n
	}

	return added, nil
}

func (s *Scanner) scanManga() (int, error) {
	matches, err := filepath.Glob(filepath.Join(s.scanDir, "*.cbz"))
	if err != nil {
		return 0, err
	}

	indexed, err := s.db.IndexedPaths()
	if err != nil {
		return 0, err
	}

	added := 0
	for _, path := range matches {
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("scanner: skipping %s: %v", path, err)
			continue
		}
		if _, exists := indexed[absPath]; exists {
			continue
		}
		if err := s.indexManga(absPath); err == nil {
			added++
		}
	}
	return added, nil
}

var audioExts = []string{"*.mp3", "*.flac", "*.ogg", "*.m4a", "*.wav", "*.aac"}

func (s *Scanner) scanAudio() (int, error) {
	indexed, err := s.db.IndexedAudioPaths()
	if err != nil {
		return 0, err
	}

	added := 0
	for _, ext := range audioExts {
		matches, err := filepath.Glob(filepath.Join(s.scanDir, ext))
		if err != nil {
			return added, err
		}
		for _, path := range matches {
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Printf("scanner: skipping %s: %v\n", path, err)
				continue
			}
			if _, exists := indexed[absPath]; exists {
				continue
			}
			if err := s.indexAudio(absPath); err != nil {
				fmt.Printf("scanner: %v\n", err)
			} else {
				added++
			}
		}
	}
	return added, nil
}

func (s *Scanner) indexManga(absPath string) error {
	fmt.Printf("indexer: indexing %s\n", absPath)
	manga, err := s.CBZService.Open(absPath)
	if err != nil {
		return fmt.Errorf("indexer: skipping %s: %v", absPath, err)
	}
	s.CBZService.Close()

	if err := s.db.InsertManga(absPath, manga.Title, manga.PageCount); err != nil {
		return fmt.Errorf("indexer: failed to insert %s: %v", absPath, err)
	}

	fmt.Printf("indexer: indexed %q (%d pages)\n", manga.Title, manga.PageCount)

	return nil
}

func (s *Scanner) indexAudio(absPath string) error {
	audio, err := audio.Open(absPath)
	if err != nil {
		return fmt.Errorf("indexer: skipping %s: %v", absPath, err)
	}

	if err := s.db.InsertAudio(audio); err != nil {
		return fmt.Errorf("indexer: failed to insert %s: %v", absPath, err)
	}

	fmt.Printf("indexer: indexed %s (%s)\n", audio.Title, audio.FilePath)

	return nil
}
