package handler

import (
	"fmt"
	"io"
	"net/http"
	"ohara/src/internal/db"
	"ohara/src/internal/media"
	"ohara/src/internal/scanner"
	"os"
	"path/filepath"
	"slices"
)

type UploadHandler struct {
	DB *db.DB
	sc *scanner.Scanner
}

func NewUploadHandler(db *db.DB, sc *scanner.Scanner) *UploadHandler {
	return &UploadHandler{DB: db, sc: sc}
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
	fmt.Printf("HandleUpload called")

	req, err := parseUploadRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer req.File.Close()

	fileType := detectFileType(req.FileName)
	fmt.Printf("fileType: %s\n", fileType)
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
	fmt.Printf("saveCBZ called\n")
	destination := req.Destination
	if destination == "" {
		destination = media.DefaultMangaDir
	}

	fmt.Printf("destination: %s\n", destination)

	targetPath := filepath.Join(destination, req.FileName)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("targetPath: %s\n", targetPath)

	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	fmt.Printf("Uploading %s to %s\n", req.FileName, targetPath)

	buffer := make([]byte, 32*1024) // 32KB buffer

	bytes, err := io.CopyBuffer(dst, req.File, buffer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Uploaded %d KB\n", bytes/1024)
	fmt.Printf("buffer use count: %d\n", bytes/1024/32)

	h.sc.Index(targetPath)
}

func (h *UploadHandler) saveAudio(w http.ResponseWriter, req *UploadRequest) {
	destination := req.Destination
	if destination == "" {
		destination = media.DefaultAudioDir
	}

	targetPath := filepath.Join(destination, req.FileName)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Uploading %s to %s\n", req.FileName, targetPath)

	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	fmt.Printf("Created file at %s\n", targetPath)

	if _, err := io.Copy(dst, req.File); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sc.Index(targetPath)
}

func detectFileType(file string) FileExtension {
	return FileExtension(filepath.Ext(file))
}
