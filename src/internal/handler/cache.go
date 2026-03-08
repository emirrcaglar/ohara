package handler

import "sync"

type pageKey struct {
	mangaID int64
	page    int
}

type PageCache struct {
	mu    sync.RWMutex
	items map[pageKey][]byte
}

func NewPageCache() *PageCache {
	return &PageCache{items: make(map[pageKey][]byte)}
}

func (c *PageCache) Get(mangaID int64, page int) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.items[pageKey{mangaID, page}]
	return data, ok
}

func (c *PageCache) Set(mangaID int64, page int, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[pageKey{mangaID, page}] = data
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
