package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type pageKey struct {
	mangaID int64
	page    int
}

type PageCache struct {
	dir string
}

func NewPageCache(dataDir string) *PageCache {
	cacheDir := filepath.Join(dataDir, "cache")
	os.MkdirAll(cacheDir, 0o755)
	return &PageCache{dir: cacheDir}
}

func (c *PageCache) getPath(mangaID int64, page int) string {
	return filepath.Join(c.dir, fmt.Sprintf("%d_%d.jpg", mangaID, page))
}

func (c *PageCache) Get(mangaID int64, page int) ([]byte, bool) {
	data, err := os.ReadFile(c.getPath(mangaID, page))
	if err != nil {
		return nil, false
	}
	return data, true
}

func (c *PageCache) Set(mangaID int64, page int, data []byte) {
	finalPath := c.getPath(mangaID, page)
	tempPath := finalPath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0o644); err == nil {
		os.Rename(tempPath, finalPath)
	}
}

type inflightCall struct {
	wg  sync.WaitGroup
	val []byte
	err error
}

type Inflight struct {
	mu    sync.Mutex
	calls map[pageKey]*inflightCall
}

func NewInflight() *Inflight {
	return &Inflight{calls: make(map[pageKey]*inflightCall)}
}

func (g *Inflight) Do(mangaID int64, page int, fn func() ([]byte, error)) ([]byte, error) {
	key := pageKey{mangaID, page}

	g.mu.Lock()
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &inflightCall{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return c.val, c.err
}
