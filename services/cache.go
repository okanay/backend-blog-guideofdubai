// cache/cache.go
package cache

import (
	"sync"
	"time"
)

// Cache genel amaçlı bir önbellek yapısı
type Cache struct {
	mu   sync.RWMutex
	data map[string]cacheItem
	ttl  time.Duration
}

// cacheItem önbellekteki bir veriyi ve metadata'sını temsil eder
type cacheItem struct {
	value    []byte
	cachedAt time.Time
	ttl      time.Duration // Opsiyonel TTL
}

// NewCache yeni bir Cache instance'ı oluşturur
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]cacheItem),
		ttl:  ttl,
	}
}

// Set verilen anahtarla bir değeri önbelleğe kaydeder
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:    value,
		cachedAt: time.Now(),
	}
}

func (c *Cache) SetWithTTL(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:    value,
		cachedAt: time.Now(),
		ttl:      ttl, // Bu alan cacheItem struct'ına eklenmeli
	}
}

// Get bir anahtara karşılık gelen değeri önbellekten döndürür
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Özel TTL kontrolü
	if item.ttl > 0 && time.Since(item.cachedAt) > item.ttl {
		// TTL dolmuş, veriyi sil ve false dön
		delete(c.data, key)
		return nil, false
	}

	// Genel TTL kontrolü
	if time.Since(item.cachedAt) > c.ttl {
		return nil, false
	}

	return item.value, true
}

// Delete bir anahtarı önbellekten siler
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear tüm önbelleği temizler
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem)
}

// ClearPrefix belirli bir önekle başlayan tüm anahtarları temizler
func (c *Cache) ClearPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}
