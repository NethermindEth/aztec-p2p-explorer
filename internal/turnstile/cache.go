package turnstile

import (
	"sync"
	"time"
)

type CachedResult struct {
	Valid       bool
	Hostname    string
	Action      string
	ChallengeTS time.Time
	IP          string
	UA          string
	ExpiresAt   time.Time
}

type Cache interface {
	Get(token string) (CachedResult, bool)
	Set(token string, r CachedResult)
}

type memoryCache struct {
	mu sync.RWMutex
	m  map[string]CachedResult
}

func NewMemoryCache() Cache { return &memoryCache{m: make(map[string]CachedResult)} }

func (c *memoryCache) Get(k string) (CachedResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.m[k]
	if !ok || time.Now().After(r.ExpiresAt) {
		return CachedResult{}, false
	}
	return r, true
}

func (c *memoryCache) Set(k string, r CachedResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = r
}
