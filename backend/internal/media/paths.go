package media

import "path/filepath"

const DefaultStorageDir = ".storage"

var (
	DefaultMangaDir = filepath.Join(DefaultStorageDir, "manga")
	DefaultAudioDir = filepath.Join(DefaultStorageDir, "audio")
)
