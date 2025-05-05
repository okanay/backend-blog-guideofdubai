package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTrueClientIP(c *gin.Context) string {
	// Öncelikle X-Real-IP başlığını kontrol et (genellikle Nginx tarafından ayarlanır)
	ip := c.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// X-Forwarded-For başlığını kontrol et ve varsa son elemanı al
	// Son eleman genellikle en yakın istemci IP'sidir (proxy zincirinde)
	forwardedFor := c.Request.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			// Proxy zincirinde en sondaki IP'yi al (en yakın istemci)
			lastIP := strings.TrimSpace(ips[len(ips)-1])
			if lastIP != "" {
				return lastIP
			}
		}
	}

	// Hiçbir başlık bulunamadıysa, Gin'in varsayılan ClientIP metodunu kullan
	return c.ClientIP()
}
