package configs

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsConfig() gin.HandlerFunc {
	var origins = []string{}

	if gin.Mode() == gin.DebugMode {
		origins = append(origins, "http://localhost:3000")
	}

	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowOrigins:     origins,
		AllowCredentials: true,
		MaxAge:           60 * 24 * 30,
	})
}
