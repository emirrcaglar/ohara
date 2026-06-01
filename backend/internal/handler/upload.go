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
	FileExtensionMKV  FileExtension = ".mkv"
	FileExtensionWEBM FileExtension = ".webm"
	FileExtensionMOV  FileExtension = ".mov"
	FileExtensionAVI  FileExtension = ".avi"
	FileExtensionM4V  FileExtension = ".m4v"

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

// VideoExtensions lists every video format the server accepts.
var VideoExtensions = []FileExtension{
	FileExtensionMP4,
	FileExtensionMKV,
	FileExtensionWEBM,
	FileExtensionMOV,
	FileExtensionAVI,
	FileExtensionM4V,
}

// IsAudio reports whether the extension belongs to a supported audio format.
func (f FileExtension) IsAudio() bool {
	return slices.Contains(AudioExtensions, f)
}

// IsVideo reports whether the extension belongs to a supported video format.
func (f FileExtension) IsVideo() bool {
	return slices.Contains(VideoExtensions, f)
}

type ChunkedUploadInitRequest struct {
	FileName     string `json:"filename"`
	Size         int64  `json:"size"`
	Profile      string `json:"profile"`
	LastModified int64  `json:"lastModified"`
}

type ChunkedUploadMetadata struct {
	ID        string    `json:"id"`
	FileName  string    `json:"filename"`
	Size      int64     `json:"size"`
	Profile   string    `json:"profile"`
	ChunkSize int64     `json:"chunkSize"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *UploadHandler) HandleUploadsList(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessions, err := h.DB.ListPendingUploadSessions(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]map[string]any, 0, len(sessions))
	for _, pending := range sessions {
		session := pending.UploadSession
		if err := h.reconcileUploadChunks(&session); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		uploaded, err := h.DB.ListUploadChunkIndexes(session.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, map[string]any{
			"uploadId":       session.ID,
			"filename":       session.Filename,
			"size":           session.Size,
			"chunkSize":      session.ChunkSize,
			"totalChunks":    session.TotalChunks,
			"uploadedChunks": uploaded,
			"uploadedCount":  len(uploaded),
			"status":         session.Status,
			"complete":       int64(len(uploaded)) == session.TotalChunks,
			"createdAt":      session.CreatedAt,
			"updatedAt":      session.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"uploads": items})
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

	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resumed := false
	session, err := h.DB.FindResumableUploadSession(user.ID, fileName, req.Size, req.LastModified, req.Profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil {
		uploadID, err := newUploadID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session = &db.UploadSession{
			ID:           uploadID,
			UserID:       user.ID,
			Filename:     fileName,
			Size:         req.Size,
			Profile:      req.Profile,
			LastModified: req.LastModified,
			ChunkSize:    DefaultUploadChunkSize,
			TotalChunks:  totalChunksForSize(req.Size, DefaultUploadChunkSize),
			Status:       db.UploadStatusActive,
			TargetPath:   defaultTargetPath(fileName),
		}
		if err := h.DB.CreateUploadSession(*session); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		resumed = true
		if err := h.reconcileUploadChunks(session); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := h.DB.UpdateUploadSessionStatus(session.ID, db.UploadStatusActive); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Status = db.UploadStatusActive
	}

	chunksDir := chunkedUploadChunksDir(session.ID)
	if err := os.MkdirAll(chunksDir, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uploaded, err := h.DB.ListUploadChunkIndexes(session.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] chunked init id=%s name=%s resumed=%t size_mb=%.2f chunk_mb=%.2f", session.ID, fileName, resumed, float64(req.Size)/(1024*1024), float64(session.ChunkSize)/(1024*1024))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"uploadId":       session.ID,
		"chunkSize":      session.ChunkSize,
		"totalChunks":    session.TotalChunks,
		"uploadedChunks": uploaded,
		"status":         session.Status,
		"resumed":        resumed,
	})
}

func (h *UploadHandler) HandleChunkedUploadChunk(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}

	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	session, err := h.DB.GetUploadSession(uploadID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil || session.Status == db.UploadStatusCancelled || session.Status == db.UploadStatusComplete {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	chunkIndex, err := strconv.Atoi(r.Header.Get("X-Chunk-Index"))
	if err != nil || chunkIndex < 0 {
		http.Error(w, "invalid X-Chunk-Index", http.StatusBadRequest)
		return
	}
	if int64(chunkIndex) >= session.TotalChunks {
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
		_ = h.DB.FailUploadSession(uploadID, copyErr)
		http.Error(w, copyErr.Error(), http.StatusInternalServerError)
		return
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		_ = h.DB.FailUploadSession(uploadID, closeErr)
		http.Error(w, closeErr.Error(), http.StatusInternalServerError)
		return
	}
	if written > MaxUploadChunkSize {
		_ = os.Remove(tmpPath)
		err := fmt.Errorf("chunk too large")
		_ = h.DB.FailUploadSession(uploadID, err)
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	if err := os.Rename(tmpPath, chunkPath); err != nil {
		_ = os.Remove(tmpPath)
		_ = h.DB.FailUploadSession(uploadID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.DB.UpsertUploadChunk(uploadID, chunkIndex, written, chunkPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.DB.UpdateUploadSessionStatus(uploadID, db.UploadStatusActive); err != nil {
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

	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	session, err := h.DB.GetUploadSession(uploadID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil || session.Status == db.UploadStatusCancelled {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}
	if err := h.reconcileUploadChunks(session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uploaded, err := h.DB.ListUploadChunkIndexes(uploadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if int64(len(uploaded)) != session.TotalChunks {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":          "missing chunks",
			"uploadedChunks": uploaded,
			"missingChunks":  missingChunkIndexes(uploaded, session.TotalChunks),
		})
		return
	}

	targetPath := session.TargetPath
	if targetPath == "" {
		targetPath = defaultTargetPath(session.Filename)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.DB.UpdateUploadSessionStatus(uploadID, db.UploadStatusAssembling); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.assembleChunkedUpload(session, targetPath); err != nil {
		_ = h.DB.FailUploadSession(uploadID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.DB.CompleteUploadSession(uploadID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[upload] chunked complete id=%s name=%s path=%s", uploadID, session.Filename, targetPath)
	}
	go func() {
		defer os.RemoveAll(chunkedUploadDir(uploadID))
		defer h.DB.DeleteUploadChunks(uploadID)
		h.indexPath(targetPath)
	}()

	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"filename": session.Filename,
		"indexing": "queued",
	})
}

func (h *UploadHandler) HandleChunkedUploadStatus(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	session, err := h.DB.GetUploadSession(uploadID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}
	if err := h.reconcileUploadChunks(session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uploaded, err := h.DB.ListUploadChunkIndexes(uploadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"uploadId":       session.ID,
		"filename":       session.Filename,
		"size":           session.Size,
		"chunkSize":      session.ChunkSize,
		"totalChunks":    session.TotalChunks,
		"uploadedChunks": uploaded,
		"uploadedCount":  len(uploaded),
		"status":         session.Status,
		"complete":       int64(len(uploaded)) == session.TotalChunks,
		"missingChunks":  missingChunkIndexes(uploaded, session.TotalChunks),
	})
}

func (h *UploadHandler) HandleChunkedUploadPause(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	session, err := h.DB.GetUploadSession(uploadID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil || session.Status == db.UploadStatusCancelled || session.Status == db.UploadStatusComplete {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}
	if err := h.DB.UpdateUploadSessionStatus(uploadID, db.UploadStatusPaused); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": db.UploadStatusPaused})
}

func (h *UploadHandler) HandleChunkedUploadCancel(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	if !isValidUploadID(uploadID) {
		http.Error(w, "invalid upload id", http.StatusBadRequest)
		return
	}
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	session, err := h.DB.GetUploadSession(uploadID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}
	if err := h.DB.CancelUploadSession(uploadID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.DB.DeleteUploadChunks(uploadID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.RemoveAll(chunkedUploadDir(uploadID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *UploadHandler) assembleChunkedUpload(session *db.UploadSession, targetPath string) error {
	tmpTarget := targetPath + ".tmp"
	dst, err := os.Create(tmpTarget)
	if err != nil {
		return err
	}

	var assembled int64
	for i := int64(0); i < session.TotalChunks; i++ {
		chunkPath := chunkedUploadChunkPath(session.ID, int(i))
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
	if assembled != session.Size {
		_ = os.Remove(tmpTarget)
		return fmt.Errorf("assembled size mismatch: got %d want %d", assembled, session.Size)
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
	return fileType == FileExtensionCBZ || fileType.IsAudio() || fileType.IsVideo()
}

func defaultTargetPath(fileName string) string {
	fileType := detectFileType(fileName)
	if fileType == FileExtensionCBZ {
		return filepath.Join(media.DefaultMangaDir, fileName)
	}
	if fileType.IsVideo() {
		return filepath.Join(media.DefaultVideoDir, fileName)
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
	return totalChunksForSize(meta.Size, meta.ChunkSize)
}

func totalChunksForSize(size, chunkSize int64) int64 {
	return (size + chunkSize - 1) / chunkSize
}

func (h *UploadHandler) reconcileUploadChunks(session *db.UploadSession) error {
	for i := int64(0); i < session.TotalChunks; i++ {
		chunkPath := chunkedUploadChunkPath(session.ID, int(i))
		info, err := os.Stat(chunkPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.IsDir() {
			continue
		}
		if err := h.DB.UpsertUploadChunk(session.ID, int(i), info.Size(), chunkPath); err != nil {
			return err
		}
	}
	return nil
}

func missingChunkIndexes(uploaded []int, total int64) []int {
	uploadedSet := make(map[int]struct{}, len(uploaded))
	for _, index := range uploaded {
		uploadedSet[index] = struct{}{}
	}

	missing := make([]int, 0)
	for i := int64(0); i < total; i++ {
		if _, ok := uploadedSet[int(i)]; !ok {
			missing = append(missing, int(i))
		}
	}
	return missing
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
