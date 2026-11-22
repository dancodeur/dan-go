package open5e

import (
	"testing"
	"time"
)

// size donne le nombre d'entrées (utile seulement pour les tests).
func (c *Cache) size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.entries)
}

func TestCacheEntryExpires(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	cache := newCache(5*time.Minute, func() time.Time { return now })
	cache.Set("key", []byte("value"))

	data, found := cache.Get("key")
	if !found || string(data) != "value" {
		t.Fatalf("fresh entry: got %q, %v", data, found)
	}

	now = now.Add(4 * time.Minute)
	if _, found := cache.Get("key"); !found {
		t.Error("entry should still be valid after 4 minutes")
	}

	now = now.Add(2 * time.Minute)
	if _, found := cache.Get("key"); found {
		t.Error("entry should have expired after 6 minutes")
	}
}

func TestRemoveExpiredKeepsFreshEntries(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	cache := newCache(5*time.Minute, func() time.Time { return now })

	cache.Set("old", []byte("1"))
	now = now.Add(4 * time.Minute)
	cache.Set("recent", []byte("2"))
	now = now.Add(2 * time.Minute) // "old" a 6 minutes, "recent" en a 2

	cache.removeExpired()

	if cache.size() != 1 {
		t.Fatalf("cache size = %d, want 1", cache.size())
	}
	if _, found := cache.Get("recent"); !found {
		t.Error("the recent entry should have been kept")
	}
}

func TestReaperRemovesExpiredEntries(t *testing.T) {
	cache := NewCache(20*time.Millisecond, 10*time.Millisecond)
	defer cache.Close()
	cache.Set("key", []byte("value"))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cache.size() == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the reaper goroutine never removed the expired entry")
}
