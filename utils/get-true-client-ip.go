package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTrueClientIP(c *gin.Context) string {
	// Cloudflare'ın özel başlığını kontrol et
	cfIP := c.Request.Header.Get("CF-Connecting-IP")
	if cfIP != "" {
		return cfIP
	}

	// True-Client-IP kontrol et
	trueClientIP := c.Request.Header.Get("True-Client-IP")
	if trueClientIP != "" {
		return trueClientIP
	}

	// X-Forwarded-For başlığını kontrol et
	forwardedFor := c.Request.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			// İlk IP'yi al (genellikle orijinal istemci IP'si)
			firstIP := strings.TrimSpace(ips[0])
			if firstIP != "" {
				return firstIP
			}
		}
	}

	// X-Real-IP başlığını kontrol et
	realIP := c.Request.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Son çare
	clientIP := c.ClientIP()
	return clientIP
}
