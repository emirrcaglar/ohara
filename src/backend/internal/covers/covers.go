// Package covers maintains disposable card thumbnails independently of reader pages.
package covers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"time"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/manga"
)

const (
	maxWidth        = 480
	maxHeight       = 720
	maxEncodedBytes = 32 << 20
	maxPixels       = 40_000_000
)

type Cache struct {
	dir string
}

func New(dataDir string) *Cache {
	return &Cache{dir: filepath.Join(dataDir, "cache", "manga-covers")}
}

// key changes when an archive is replaced, and when thumbnail settings change.
func (c *Cache) key(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("v1:%s:%d:%d", path, info.Size(), info.ModTime().UnixNano())))), nil
}

// Open only looks for a finished thumbnail; it never decodes or compresses images.
func (c *Cache) Open(path string) (*os.File, error) {
	key, err := c.key(path)
	if err != nil {
		return nil, err
	}
	return os.Open(filepath.Join(c.dir, key+".jpg"))
}

// Generate is called by the single background worker. Publication is atomic so
// concurrent HTTP requests see either the original or a complete JPEG.
func (c *Cache) Generate(path string) error {
	key, err := c.key(path)
	if err != nil {
		return err
	}
	for _, ext := range []string{".jpg", ".original"} {
		if _, err := os.Stat(filepath.Join(c.dir, key+ext)); err == nil {
			return nil
		}
	}
	archive, err := manga.Open(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	page, err := archive.OpenPage(0)
	if err != nil {
		return err
	}
	defer page.Close()
	raw, err := io.ReadAll(io.LimitReader(page, maxEncodedBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > maxEncodedBytes {
		return fmt.Errorf("cover exceeds encoded size limit")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > maxPixels {
		return fmt.Errorf("cover exceeds pixel limit")
	}
	source, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	width, height := config.Width, config.Height
	if width > maxWidth {
		height = max(1, height*maxWidth/width)
		width = maxWidth
	}
	if height > maxHeight {
		width = max(1, width*maxHeight/height)
		height = maxHeight
	}
	thumbnail := image.NewRGBA(image.Rect(0, 0, width, height))
	// JPEG has no alpha channel. Composite transparent PNG/WebP covers on white.
	draw.Draw(thumbnail, thumbnail.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(thumbnail, thumbnail.Bounds(), source, source.Bounds(), draw.Over, nil)
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, thumbnail, &jpeg.Options{Quality: 75}); err != nil {
		return err
	}
	if current, err := c.key(path); err != nil || current != key {
		return fmt.Errorf("archive changed during cover generation")
	}
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}
	if encoded.Len() >= len(raw) {
		// Remember covers that are already smaller, including across restarts.
		return os.WriteFile(filepath.Join(c.dir, key+".original"), nil, 0644)
	}
	file, err := os.CreateTemp(c.dir, "cover-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(encoded.Bytes()); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(c.dir, key+".jpg"))
}

// Run backfills existing manga at startup and discovers new imports every 30s.
// A single goroutine bounds decode memory and keeps compression off HTTP paths.
func (c *Cache) Run(ctx context.Context, database *db.DB, log *logger.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		if err := c.scan(ctx, database, log); err != nil && log != nil {
			log.Error("[covers] library scan failed err=%v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (c *Cache) scan(ctx context.Context, database *db.DB, log *logger.Logger) error {
	paths, err := database.IndexedMangaPaths()
	if err != nil {
		return err
	}
	for path := range paths {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := c.Generate(path); err != nil && log != nil {
			log.Warn("[covers] generation failed path=%s err=%v", path, err)
		}
	}
	return nil
}
