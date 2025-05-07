package types

import (
	"time"
)

// RateLimitInfo bir kullanıcının rate limit bilgilerini tutar
type RateLimitInfo struct {
	UserID          string    `json:"userId"`
	RequestCount    int       `json:"requestCount"`    // Toplam istek sayısı
	TokensUsed      int       `json:"tokensUsed"`      // Kullanılan toplam token sayısı
	RequestsPerMin  int       `json:"requestsPerMin"`  // Dakika başına istek sayısı
	FirstRequest    time.Time `json:"firstRequest"`    // Pencere içindeki ilk istek zamanı
	LastRequest     time.Time `json:"lastRequest"`     // Son istek zamanı
	WindowResetAt   time.Time `json:"windowResetAt"`   // Zaman penceresinin sıfırlanma zamanı
	MinuteStartedAt time.Time `json:"minuteStartedAt"` // Dakika başlangıç zamanı
}

// RateLimitResponse bir rate limit yanıtı
type RateLimitResponse struct {
	Allowed         bool      `json:"allowed"`         // İstek izin veriliyor mu?
	Remaining       int       `json:"remaining"`       // Kalan istek sayısı
	RemainingTokens int       `json:"remainingTokens"` // Kalan token sayısı
	ResetAt         time.Time `json:"resetAt"`         // Sıfırlanma zamanı
	RetryAfter      int       `json:"retryAfter"`      // Tavsiye edilen yeniden deneme süresi (saniye)
}
