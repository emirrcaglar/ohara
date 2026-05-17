package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media"
	"ohara/src/internal/scanner"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
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

	DefaultUploadChunkSize = 8 << 20
	MaxUploadChunkSize     = 16 << 20
	MaxUploadSize          = 5 << 30
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

type ChunkedUploadInitRequest struct {
	FileName string `json:"filename"`
	Size     int64  `json:"size"`
	Profile  string `json:"profile"`
}

type ChunkedUploadMetadata struct {
	ID        string    `json:"id"`
	FileName  string    `json:"filename"`
	Size      int64     `json:"size"`
	Profile   string    `json:"profile"`
	ChunkSize int64     `json:"chunkSize"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *UploadHandler) HandleChunkedUploadInit(w http.ResponseWriter, r *http.Request) {
	var req ChunkedUploadInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileName, err := sanitizeUploadFileName(req.FileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Size <= 0 || req.Size > MaxUploadSize {
		http.Error(w, fmt.Sprintf("file size must be between 1 byte and %d bytes", MaxUploadSize), http.StatusBadRequest)
		return
	}
	if !isSupportedUploadFile(fileName) {
		http.Error(w, "Unsupported file type", http.StatusUnsupportedMediaType)
		return
	}

	uploadID, err := newUploadID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	meta := ChunkedUploadMetadata{
		ID:        uploadID,
		FileName:  fileName,
		Size:      req.Size,
		Profile:   req.Profile,
		ChunkSize: DefaultUploadChunkSize,
		CreatedAt: time.Now(),
	}

	chunksDir := chunkedUploadChunksDir(uploadID)
	if err := os.MkdirAll(chunksDir, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := writeUploadMetadata(meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] chunked init id=%s name=%s size_mb=%.2f chunk_mb=%.2f", uploadID, fileName, float64(req.Size)/(1024*1024), float64(meta.ChunkSize)/(1024*1024))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"uploadId":  uploadID,
		"chunkSize": meta.ChunkSize,
	})
}

func (h *UploadHandler) HandleChunkedUploadChunk(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}

	meta, err := readUploadMetadata(uploadID)
	if err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	chunkIndex, err := strconv.Atoi(r.Header.Get("X-Chunk-Index"))
	if err != nil || chunkIndex < 0 {
		http.Error(w, "invalid X-Chunk-Index", http.StatusBadRequest)
		return
	}
	if int64(chunkIndex) >= totalChunks(meta) {
		http.Error(w, "chunk index out of range", http.StatusBadRequest)
		return
	}

	limit := MaxUploadChunkSize + 1
	chunkPath := chunkedUploadChunkPath(uploadID, chunkIndex)
	tmpPath := chunkPath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(chunkPath), 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	written, copyErr := io.Copy(dst, io.LimitReader(r.Body, int64(limit)))
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, copyErr.Error(), http.StatusInternalServerError)
		return
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, closeErr.Error(), http.StatusInternalServerError)
		return
	}
	if written > MaxUploadChunkSize {
		_ = os.Remove(tmpPath)
		http.Error(w, "chunk too large", http.StatusRequestEntityTooLarge)
		return
	}
	if err := os.Rename(tmpPath, chunkPath); err != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *UploadHandler) HandleChunkedUploadComplete(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}

	meta, err := readUploadMetadata(uploadID)
	if err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	targetPath := defaultTargetPath(meta.FileName)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.assembleChunkedUpload(meta, targetPath); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] chunked complete id=%s name=%s path=%s", uploadID, meta.FileName, targetPath)
	}
	go func() {
		defer os.RemoveAll(chunkedUploadDir(uploadID))
		h.indexPath(targetPath)
	}()

	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"filename": meta.FileName,
		"indexing": "queued",
	})
}

func (h *UploadHandler) HandleChunkedUploadStatus(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}
	meta, err := readUploadMetadata(uploadID)
	if err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	uploaded := make([]int, 0)
	for i := int64(0); i < totalChunks(meta); i++ {
		if _, err := os.Stat(chunkedUploadChunkPath(uploadID, int(i))); err == nil {
			uploaded = append(uploaded, int(i))
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"uploadedChunks": uploaded,
		"complete":       int64(len(uploaded)) == totalChunks(meta),
	})
}

func (h *UploadHandler) HandleChunkedUploadCancel(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}
	if err := os.RemoveAll(chunkedUploadDir(uploadID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *UploadHandler) assembleChunkedUpload(meta ChunkedUploadMetadata, targetPath string) error {
	tmpTarget := targetPath + ".tmp"
	dst, err := os.Create(tmpTarget)
	if err != nil {
		return err
	}

	var assembled int64
	for i := int64(0); i < totalChunks(meta); i++ {
		chunkPath := chunkedUploadChunkPath(meta.ID, int(i))
		src, err := os.Open(chunkPath)
		if err != nil {
			_ = dst.Close()
			_ = os.Remove(tmpTarget)
			return fmt.Errorf("missing chunk %d", i)
		}
		n, copyErr := io.Copy(dst, src)
		closeErr := src.Close()
		if copyErr != nil {
			_ = dst.Close()
			_ = os.Remove(tmpTarget)
			return copyErr
		}
		if closeErr != nil {
			_ = dst.Close()
			_ = os.Remove(tmpTarget)
			return closeErr
		}
		assembled += n
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tmpTarget)
		return err
	}
	if assembled != meta.Size {
		_ = os.Remove(tmpTarget)
		return fmt.Errorf("assembled size mismatch: got %d want %d", assembled, meta.Size)
	}
	return os.Rename(tmpTarget, targetPath)
}

func (h *UploadHandler) indexPath(targetPath string) {
	if err := h.sc.Index(targetPath); err != nil && h.Log != nil {
		h.Log.Error("[upload] index failed path=%s err=%v", targetPath, err)
	}
}

func detectFileType(file string) FileExtension {
	return FileExtension(strings.ToLower(filepath.Ext(file)))
}

func isSupportedUploadFile(fileName string) bool {
	fileType := detectFileType(fileName)
	return fileType == FileExtensionCBZ || fileType.IsAudio()
}

func defaultTargetPath(fileName string) string {
	if detectFileType(fileName) == FileExtensionCBZ {
		return filepath.Join(media.DefaultMangaDir, fileName)
	}
	return filepath.Join(media.DefaultAudioDir, fileName)
}

func sanitizeUploadFileName(fileName string) (string, error) {
	base := filepath.Base(strings.TrimSpace(fileName))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "", fmt.Errorf("invalid filename")
	}
	if strings.Contains(base, "\x00") {
		return "", fmt.Errorf("invalid filename")
	}
	return base, nil
}

func newUploadID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isValidUploadID(uploadID string) bool {
	if len(uploadID) != 32 {
		return false
	}
	for _, r := range uploadID {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func chunkedUploadsRoot() string {
	return filepath.Join(media.DefaultStorageDir, "uploads")
}

func chunkedUploadDir(uploadID string) string {
	return filepath.Join(chunkedUploadsRoot(), uploadID)
}

func chunkedUploadChunksDir(uploadID string) string {
	return filepath.Join(chunkedUploadDir(uploadID), "chunks")
}

func chunkedUploadMetadataPath(uploadID string) string {
	return filepath.Join(chunkedUploadDir(uploadID), "meta.json")
}

func chunkedUploadChunkPath(uploadID string, chunkIndex int) string {
	return filepath.Join(chunkedUploadChunksDir(uploadID), fmt.Sprintf("%06d.part", chunkIndex))
}

func writeUploadMetadata(meta ChunkedUploadMetadata) error {
	f, err := os.Create(chunkedUploadMetadataPath(meta.ID))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(meta)
}

func readUploadMetadata(uploadID string) (ChunkedUploadMetadata, error) {
	var meta ChunkedUploadMetadata
	f, err := os.Open(chunkedUploadMetadataPath(uploadID))
	if err != nil {
		return meta, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return meta, err
	}
	return meta, nil
}

func totalChunks(meta ChunkedUploadMetadata) int64 {
	return (meta.Size + meta.ChunkSize - 1) / meta.ChunkSize
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
