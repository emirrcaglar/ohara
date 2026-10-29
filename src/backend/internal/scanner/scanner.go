package scanner

import (
	"fmt"
	"path/filepath"
	"strings"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/audio"
	"ohara/src/internal/media/manga"
)

type Scanner struct {
	db       *db.DB
	scanDir  string
	scanType ScanType
	log      *logger.Logger
}

type ScannerOption func(*Scanner)

type ScanType string

var (
	ScanTypeAll   ScanType = "all"
	ScanTypeManga ScanType = "manga"
	ScanTypeAudio ScanType = "audio"
	ScanTypeVideo ScanType = "video"
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

func NewScanner(database *db.DB, log *logger.Logger, opts ...ScannerOption) *Scanner {
	scanner := &Scanner{
		db:       database,
		scanDir:  ".",
		scanType: ScanTypeAll,
		log:      log,
	}

	for _, opt := range opts {
		opt(scanner)
	}

	return scanner
}

func (s *Scanner) Index(targetPath string) error {
	return s.IndexWithFolder(targetPath, nil)
}

func (s *Scanner) IndexWithFolder(targetPath string, folderID *int64) error {
	fileType := strings.ToLower(filepath.Ext(targetPath))

	switch fileType {
	case ".cbz":
		return s.indexManga(targetPath, folderID)
	case ".mp3", ".flac", ".ogg", ".m4a", ".wav", ".aac":
		return s.indexAudio(targetPath, folderID)
	case ".mp4", ".mkv", ".webm", ".mov", ".avi", ".m4v":
		return s.indexVideo(targetPath, folderID)
	default:
		return fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func (s *Scanner) Run() (int, error) {
	added := 0
	if s.log != nil {
		s.log.Info("[scanner] starting scan dir=%s type=%s", s.scanDir, s.scanType)
	}

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

	if s.scanType == ScanTypeVideo || s.scanType == ScanTypeAll {
		n, err := s.scanVideo()
		if err != nil {
			return added, err
		}
		added += n
	}

	if s.log != nil {
		s.log.Info("[scanner] scan complete dir=%s type=%s added=%d", s.scanDir, s.scanType, added)
	}

	return added, nil
}

func (s *Scanner) scanManga() (int, error) {
	matches, err := filepath.Glob(filepath.Join(s.scanDir, "*.cbz"))
	if err != nil {
		return 0, err
	}

	indexed, err := s.db.IndexedMangaPaths()
	if err != nil {
		return 0, err
	}

	added := 0
	for _, path := range matches {
		absPath, err := filepath.Abs(path)
		if err != nil {
			if s.log != nil {
				s.log.Warn("[scanner] skipping manga path=%s err=%v", path, err)
			}
			continue
		}
		if _, exists := indexed[absPath]; exists {
			continue
		}
		if err := s.indexManga(absPath, nil); err == nil {
			added++
		} else if s.log != nil {
			s.log.Error("[scanner] manga index failed path=%s err=%v", absPath, err)
		}
	}
	return added, nil
}

var audioExts = []string{"*.mp3", "*.flac", "*.ogg", "*.m4a", "*.wav", "*.aac"}
var videoExts = []string{"*.mp4", "*.mkv", "*.webm", "*.mov", "*.avi", "*.m4v"}

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
				if s.log != nil {
					s.log.Warn("[scanner] skipping audio path=%s err=%v", path, err)
				}
				continue
			}
			if _, exists := indexed[absPath]; exists {
				continue
			}
			if err := s.indexAudio(absPath, nil); err != nil {
				if s.log != nil {
					s.log.Error("[scanner] audio index failed path=%s err=%v", absPath, err)
				}
			} else {
				added++
			}
		}
	}
	return added, nil
}

func (s *Scanner) indexManga(absPath string, folderID *int64) error {
	archive, err := manga.Open(absPath)
	if err != nil {
		return fmt.Errorf("indexer: skipping %s: %v", absPath, err)
	}
	defer archive.Close()

	seriesTitle, metadata := manga.ParseFilename(absPath)
	_, err = s.db.InsertMangaWithMetadata(absPath, archive.Title, archive.PageCount, folderID, seriesTitle, &metadata, manga.ParserVersion)
	if err != nil {
		return fmt.Errorf("indexer: failed to insert %s: %v", absPath, err)
	}
	return nil
}

func (s *Scanner) indexAudio(absPath string, folderID *int64) error {
	track, err := audio.Open(absPath)
	if err != nil {
		return fmt.Errorf("indexer: skipping %s: %v", absPath, err)
	}

	if err := s.db.InsertAudioWithFolder(track, folderID); err != nil {
		return fmt.Errorf("indexer: failed to insert %s: %v", absPath, err)
	}

	return nil
}

func (s *Scanner) scanVideo() (int, error) {
	indexed, err := s.db.IndexedVideoPaths()
	if err != nil {
		return 0, err
	}

	added := 0
	for _, ext := range videoExts {
		matches, err := filepath.Glob(filepath.Join(s.scanDir, ext))
		if err != nil {
			return added, err
		}
		for _, path := range matches {
			absPath, err := filepath.Abs(path)
			if err != nil {
				if s.log != nil {
					s.log.Warn("[scanner] skipping video path=%s err=%v", path, err)
				}
				continue
			}
			if _, exists := indexed[absPath]; exists {
				continue
			}
			if err := s.indexVideo(absPath, nil); err != nil {
				if s.log != nil {
					s.log.Error("[scanner] video index failed path=%s err=%v", absPath, err)
				}
			} else {
				added++
			}
		}
	}
	return added, nil
}

func (s *Scanner) indexVideo(absPath string, folderID *int64) error {
	title := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
	if err := s.db.InsertVideoWithFolder(absPath, title, 0, folderID); err != nil {
		return fmt.Errorf("indexer: failed to insert %s: %v", absPath, err)
	}
	return nil
}
