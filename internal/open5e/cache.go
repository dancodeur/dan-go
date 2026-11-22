package open5e

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// Cache garde les réponses de l'API en mémoire pendant une durée limitée.
// Le mutex protège la map, car le cache est utilisé par plusieurs goroutines.
type Cache struct {
	mutex   sync.Mutex
	entries map[string]cacheEntry
	ttl     time.Duration
	now     func() time.Time // remplaçable dans les tests
	stop    chan struct{}
}

// NewCache crée un cache et lance la goroutine qui supprime
// les entrées expirées toutes les reapInterval. Appeler Close à la fin.
func NewCache(ttl, reapInterval time.Duration) *Cache {
	cache := newCache(ttl, time.Now)
	go cache.reapLoop(reapInterval)
	return cache
}

// newCache crée un cache sans goroutine de nettoyage.
func newCache(ttl time.Duration, now func() time.Time) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		now:     now,
		stop:    make(chan struct{}),
	}
}

// Get renvoie la valeur gardée pour cette clé, si elle n'a pas expiré.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, found := c.entries[key]
	if !found {
		return nil, false
	}
	if !c.now().Before(entry.expiresAt) {
		delete(c.entries, key)
		return nil, false
	}
	return entry.data, true
}

// Set garde une valeur pendant la durée de vie du cache.
func (c *Cache) Set(key string, data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = cacheEntry{data: data, expiresAt: c.now().Add(c.ttl)}
}

// Close arrête la goroutine de nettoyage. À appeler une seule fois.
func (c *Cache) Close() {
	close(c.stop)
}

// reapLoop supprime régulièrement les entrées expirées jusqu'à Close.
func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache) removeExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := c.now()
	for key, entry := range c.entries {
		if !now.Before(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}
