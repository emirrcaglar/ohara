package handler

import (
	"io"
	"net/http"
	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media"
	"ohara/src/internal/scanner"
	"os"
	"path/filepath"
	"slices"
)

type UploadHandler struct {
	DB  *db.DB
	sc  *scanner.Scanner
	Log *logger.Logger
}

func NewUploadHandler(db *db.DB, sc *scanner.Scanner, log *logger.Logger) *UploadHandler {
	return &UploadHandler{DB: db, sc: sc, Log: log}
}

type FileExtension string

const (
	FileExtensionCBZ  FileExtension = ".cbz"
	FileExtensionMP3  FileExtension = ".mp3"
	FileExtensionFLAC FileExtension = ".flac"
	FileExtensionOGG  FileExtension = ".ogg"
	FileExtensionM4A  FileExtension = ".m4a"
	FileExtensionWAV  FileExtension = ".wav"
	FileExtensionAAC  FileExtension = ".aac"
	FileExtensionMP4  FileExtension = ".mp4"
)

// AudioExtensions lists every audio format the server accepts.
var AudioExtensions = []FileExtension{
	FileExtensionMP3,
	FileExtensionFLAC,
	FileExtensionOGG,
	FileExtensionM4A,
	FileExtensionWAV,
	FileExtensionAAC,
}

// IsAudio reports whether the extension belongs to a supported audio format.
func (f FileExtension) IsAudio() bool {
	return slices.Contains(AudioExtensions, f)
}

type UploadRequest struct {
	File        io.ReadCloser
	FileName    string
	Destination string
	Profile     string
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if h.Log != nil {
		h.Log.Info("[upload] request received")
	}

	req, err := parseUploadRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer req.File.Close()

	fileType := detectFileType(req.FileName)
	if h.Log != nil {
		h.Log.Info("[upload] detected file type=%s name=%s", fileType, req.FileName)
	}
	switch {
	case fileType == FileExtensionCBZ:
		h.saveCBZ(w, req)
	case fileType.IsAudio():
		h.saveAudio(w, req)
	default:
		http.Error(w, "Unsupported file type", http.StatusUnsupportedMediaType)
	}

}

func parseUploadRequest(r *http.Request) (*UploadRequest, error) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return nil, err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}

	return &UploadRequest{
		File:        file,
		FileName:    header.Filename,
		Destination: r.FormValue("destination"),
		Profile:     r.FormValue("profile"),
	}, nil
}

func (h *UploadHandler) saveCBZ(w http.ResponseWriter, req *UploadRequest) {
	if h.Log != nil {
		h.Log.Info("[upload] handling cbz name=%s", req.FileName)
	}
	destination := req.Destination
	if destination == "" {
		destination = media.DefaultMangaDir
	}

	targetPath := filepath.Join(destination, req.FileName)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if h.Log != nil {
		h.Log.Info("[upload] storing file name=%s destination=%s path=%s", req.FileName, destination, targetPath)
	}

	buffer := make([]byte, 32*1024) // 32KB buffer

	bytes, err := io.CopyBuffer(dst, req.File, buffer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] cbz upload complete name=%s size_kb=%d", req.FileName, bytes/1024)
	}

	if err := h.sc.Index(targetPath); err != nil && h.Log != nil {
		h.Log.Error("[upload] index failed path=%s err=%v", targetPath, err)
	}
}

func (h *UploadHandler) saveAudio(w http.ResponseWriter, req *UploadRequest) {
	if h.Log != nil {
		h.Log.Info("[upload] handling audio name=%s", req.FileName)
	}
	destination := req.Destination
	if destination == "" {
		destination = media.DefaultAudioDir
	}

	targetPath := filepath.Join(destination, req.FileName)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] storing file name=%s destination=%s path=%s", req.FileName, destination, targetPath)
	}

	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, req.File); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] audio upload complete name=%s path=%s", req.FileName, targetPath)
	}
	if err := h.sc.Index(targetPath); err != nil && h.Log != nil {
		h.Log.Error("[upload] index failed path=%s err=%v", targetPath, err)
	}
}

func detectFileType(file string) FileExtension {
	return FileExtension(filepath.Ext(file))
}
