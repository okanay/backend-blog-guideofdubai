package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTrueClientIP(c *gin.Context) string {
	// Debug için tüm başlıkları yazdır
	fmt.Println("-------- TÜM HTTP BAŞLIKLARI --------")
	for name, values := range c.Request.Header {
		fmt.Printf("%s: %v\n", name, values)
	}
	fmt.Println("-------- BAŞLIKLAR SONU --------")

	// Cloudflare'ın özel başlığını kontrol et
	cfIP := c.Request.Header.Get("CF-Connecting-IP")
	if cfIP != "" {
		fmt.Println("CF-Connecting-IP kullanıldı:", cfIP)
		return cfIP
	}

	// True-Client-IP kontrol et
	trueClientIP := c.Request.Header.Get("True-Client-IP")
	if trueClientIP != "" {
		fmt.Println("True-Client-IP kullanıldı:", trueClientIP)
		return trueClientIP
	}

	// X-Forwarded-For başlığını kontrol et
	forwardedFor := c.Request.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			// İlk IP'yi al (genellikle orijinal istemci IP'si)
			firstIP := strings.TrimSpace(ips[0])
			fmt.Println("X-Forwarded-For ilk IP kullanıldı:", firstIP)
			if firstIP != "" {
				return firstIP
			}
		}
	}

	// X-Real-IP başlığını kontrol et
	realIP := c.Request.Header.Get("X-Real-IP")
	if realIP != "" {
		fmt.Println("X-Real-IP kullanıldı:", realIP)
		return realIP
	}

	// Son çare
	clientIP := c.ClientIP()
	fmt.Println("ClientIP() kullanıldı:", clientIP)
	return clientIP
}
