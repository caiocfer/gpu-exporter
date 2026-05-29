package cache

import (
	"sync"
	"time"
)

type Entry struct {
	Name      string
	Cmdline   string
	ExpiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[uint32]*Entry
	ttl   time.Duration
}

func New(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[uint32]*Entry),
		ttl:   ttl,
	}
}

func (c *Cache) Get(pid uint32) (*Entry, bool) {
	c.mu.RLock()
	e, ok := c.items[pid]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.ExpiresAt) {
		c.mu.Lock()
		delete(c.items, pid)
		c.mu.Unlock()
		return nil, false
	}
	return e, true
}

func (c *Cache) Set(pid uint32, name, cmdline string) {
	c.mu.Lock()
	c.items[pid] = &Entry{
		Name:      name,
		Cmdline:   cmdline,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
