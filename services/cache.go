// cache/cache.go
package cache

import (
	"fmt"
	"sync"
	"time"
)

// ---- Veri Yapıları ----

// Cache genel amaçlı bir önbellek yapısı
type Cache struct {
	mu              sync.RWMutex
	data            map[string]cacheItem
	ttl             time.Duration
	stopCleanup     chan struct{}
	lastCleanupTime time.Time     // Son temizleme zamanı
	cleanupInterval time.Duration // Temizleme aralığı
	startTime       time.Time     // Cache başlatma zamanı
}

// cacheItem önbellekteki bir veriyi ve metadata'sını temsil eder
type cacheItem struct {
	value    []byte
	cachedAt time.Time
	ttl      time.Duration // Opsiyonel TTL
}

// ---- Temel İşlemler ----

// NewCache yeni bir önbellek oluşturur
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		data:            make(map[string]cacheItem),
		ttl:             ttl,
		stopCleanup:     make(chan struct{}),
		cleanupInterval: 30 * time.Minute, // Varsayılan temizleme aralığı
		startTime:       time.Now(),
	}

	// Otomatik temizleme goroutine'ini başlat
	go cache.startCleanupRoutine()

	return cache
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

// SetWithTTL özel TTL ile bir değeri önbelleğe kaydeder
func (c *Cache) SetWithTTL(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:    value,
		cachedAt: time.Now(),
		ttl:      ttl,
	}
}

// Get bir anahtara karşılık gelen değeri önbellekten döndürür
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Özel TTL kontrolü
	if item.ttl > 0 && time.Since(item.cachedAt) > item.ttl {
		// TTL dolmuş, veriyi sil ve false dön
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, false
	}

	// Genel TTL kontrolü
	if time.Since(item.cachedAt) > c.ttl {
		return nil, false
	}

	return item.value, true
}

// GetAllWithPrefix belirli bir önekle başlayan tüm anahtarları ve değerlerini döndürür
func (c *Cache) GetAllWithPrefix(prefix string) map[string][]byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]byte)
	for key, item := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			// TTL kontrolü
			if item.ttl > 0 && time.Since(item.cachedAt) > item.ttl {
				continue // Süresi dolmuş
			}

			if time.Since(item.cachedAt) > c.ttl {
				continue // Genel TTL süresi dolmuş
			}

			result[key] = item.value
		}
	}

	return result
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

func (c *Cache) ClearAIRateLimits() {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := "ai_rate_limit:"
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// ClearExceptPrefixes belirli öneklerle başlayan anahtarlar dışındaki tüm anahtarları temizler
func (c *Cache) ClearExceptPrefixes(prefixes []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keysToDelete := make([]string, 0)
	for key := range c.data {
		keep := false
		for _, prefix := range prefixes {
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				keep = true
				break
			}
		}
		if !keep {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.data, key)
	}
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

// ---- Temizleme İşlemleri ----

// startCleanupRoutine temizleme rutinini başlatır
func (c *Cache) startCleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.lastCleanupTime = time.Now()
			c.mu.Unlock()
			c.cleanupExpiredItems()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpiredItems süresi dolmuş cache item'larını temizler
func (c *Cache) cleanupExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.data {
		// Özel TTL kontrolü
		if item.ttl > 0 && now.Sub(item.cachedAt) > item.ttl {
			delete(c.data, key)
			continue
		}

		// Genel TTL kontrolü
		if now.Sub(item.cachedAt) > c.ttl {
			delete(c.data, key)
		}
	}
}

// Stop cleanup goroutine'ini durdurur (graceful shutdown için)
func (c *Cache) Stop() {
	close(c.stopCleanup)
}

// ---- İstatistik ve Bilgi Fonksiyonları ----

// GetStats önbellek istatistiklerini döndürür
func (c *Cache) GetStats() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	totalItems := len(c.data)
	var totalSize int64
	var expiredItems int
	var activeItems int
	var itemsWithCustomTTL int
	var oldestItemAge time.Duration
	var newestItemAge time.Duration
	var avgItemAge time.Duration
	var totalAge time.Duration

	// Item detayları için
	itemDetails := make([]map[string]any, 0)

	for key, item := range c.data {
		itemSize := int64(len(item.value))
		totalSize += itemSize
		itemAge := now.Sub(item.cachedAt)
		totalAge += itemAge

		// En eski ve en yeni item'ları bul
		if oldestItemAge == 0 || itemAge > oldestItemAge {
			oldestItemAge = itemAge
		}
		if newestItemAge == 0 || itemAge < newestItemAge {
			newestItemAge = itemAge
		}

		// Özel TTL'e sahip item'ları say
		if item.ttl > 0 {
			itemsWithCustomTTL++
		}

		// Expired item kontrolü
		isExpired := false
		var expiresIn time.Duration

		if item.ttl > 0 {
			expiresIn = item.ttl - itemAge
			if itemAge > item.ttl {
				expiredItems++
				isExpired = true
			}
		} else {
			expiresIn = c.ttl - itemAge
			if itemAge > c.ttl {
				expiredItems++
				isExpired = true
			}
		}

		if !isExpired {
			activeItems++
		}

		// Item detayını ekle (opsiyonel, büyük cache'lerde performans için kapatılabilir)
		itemDetail := map[string]any{
			"key":            key,
			"size":           itemSize,
			"sizeHuman":      formatBytes(itemSize),
			"age":            itemAge.String(),
			"ageHuman":       formatDuration(itemAge),
			"isExpired":      isExpired,
			"expiresIn":      expiresIn.String(),
			"expiresInHuman": formatDuration(expiresIn),
			"hasCustomTTL":   item.ttl > 0,
		}

		if item.ttl > 0 {
			itemDetail["customTTL"] = item.ttl.String()
		}

		itemDetails = append(itemDetails, itemDetail)
	}

	// Ortalama yaş hesapla
	if totalItems > 0 {
		avgItemAge = totalAge / time.Duration(totalItems)
	}

	// Sonraki temizleme zamanını hesapla
	var nextCleanupIn time.Duration
	if !c.lastCleanupTime.IsZero() {
		timeSinceLastCleanup := now.Sub(c.lastCleanupTime)
		nextCleanupIn = c.cleanupInterval - timeSinceLastCleanup
		if nextCleanupIn < 0 {
			nextCleanupIn = 0
		}
	} else {
		nextCleanupIn = c.cleanupInterval - now.Sub(c.startTime)
	}

	// Cache uptime
	uptime := now.Sub(c.startTime)

	return map[string]any{
		"summary": map[string]any{
			"totalItems":         totalItems,
			"activeItems":        activeItems,
			"expiredItems":       expiredItems,
			"itemsWithCustomTTL": itemsWithCustomTTL,
			"totalSize":          totalSize,
			"totalSizeHuman":     formatBytes(totalSize),
			"avgItemSize":        totalSize / int64(max(totalItems, 1)),
			"avgItemSizeHuman":   formatBytes(totalSize / int64(max(totalItems, 1))),
		},
		"timing": map[string]any{
			"defaultTTL":         c.ttl.String(),
			"cleanupInterval":    c.cleanupInterval.String(),
			"nextCleanupIn":      nextCleanupIn.String(),
			"nextCleanupInHuman": formatDuration(nextCleanupIn),
			"lastCleanupTime":    c.lastCleanupTime.Format(time.RFC3339),
			"uptime":             uptime.String(),
			"uptimeHuman":        formatDuration(uptime),
		},
		"itemAge": map[string]any{
			"oldest":       oldestItemAge.String(),
			"oldestHuman":  formatDuration(oldestItemAge),
			"newest":       newestItemAge.String(),
			"newestHuman":  formatDuration(newestItemAge),
			"average":      avgItemAge.String(),
			"averageHuman": formatDuration(avgItemAge),
		},
		"health": map[string]any{
			"status":           getHealthStatus(float64(expiredItems) / float64(max(totalItems, 1))),
			"expirationRate":   fmt.Sprintf("%.2f%%", float64(expiredItems)/float64(max(totalItems, 1))*100),
			"memoryEfficiency": getMemoryEfficiency(totalSize),
		},
		"details": itemDetails, // Büyük cache'lerde bu kısım kapatılabilir
	}
}

// ---- Yardımcı Fonksiyonlar ----

// formatDuration süreyi okunaklı formata çevirir
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "Süresi dolmuş"
	}

	totalSeconds := int(d.Seconds())
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%d gün", days)
	case hours > 0:
		return fmt.Sprintf("%d saat", hours)
	case minutes > 0:
		return fmt.Sprintf("%d dakika", minutes)
	default:
		return fmt.Sprintf("%d saniye", seconds)
	}
}

// getHealthStatus cache sağlık durumunu belirler
func getHealthStatus(expirationRate float64) string {
	switch {
	case expirationRate < 0.1:
		return "Sağlıklı"
	case expirationRate < 0.3:
		return "Normal"
	case expirationRate < 0.5:
		return "Uyarı"
	default:
		return "Kritik"
	}
}

// getMemoryEfficiency memory kullanım verimliliğini değerlendirir
func getMemoryEfficiency(totalSize int64) string {
	switch {
	case totalSize < 1<<20: // 1 MB
		return "Mükemmel"
	case totalSize < 10<<20: // 10 MB
		return "İyi"
	case totalSize < 100<<20: // 100 MB
		return "Normal"
	case totalSize < 500<<20: // 500 MB
		return "Zayıf"
	default:
		return "Kritik"
	}
}

// formatBytes byte cinsinden değeri okunaklı formata çevirir
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d Byte", bytes)
	}
}

// max helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
