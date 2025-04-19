package configs

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsConfig() gin.HandlerFunc {
	var origins = []string{"http://localhost:3000"}

	if gin.Mode() == gin.DebugMode {
		origins = append(origins, "http://localhost:3000")
	}

	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept", "Origin", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowOrigins:     origins,
		AllowCredentials: true,
		MaxAge:           60 * 24 * 30,
	})
}
