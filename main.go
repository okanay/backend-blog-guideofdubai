package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	c "github.com/okanay/go-websocket-backend/configs"
	db "github.com/okanay/go-websocket-backend/database"
	"github.com/okanay/go-websocket-backend/handlers"
	BlogHandler "github.com/okanay/go-websocket-backend/handlers/blog"
	UserHandler "github.com/okanay/go-websocket-backend/handlers/user"
	BlogRepository "github.com/okanay/go-websocket-backend/repositories/blog"
	UserRepository "github.com/okanay/go-websocket-backend/repositories/user"
)

func main() {
	// Environment Variables and Database Connection
	if err := godotenv.Load(".env"); err != nil {
		// Geliştirme ortamında .env dosyası olmayabilir, bu yüzden bu hata ihmal edilebilir
		log.Println("Warning: .env file not loaded, using environment variables")
	}

	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error connecting to database")
		return
	}
	defer sqlDB.Close()

	// Repository Initialization
	ur := UserRepository.NewRepository(sqlDB)
	br := BlogRepository.NewRepository(sqlDB)

	// Handler Initialization
	mainHandler := handlers.NewHandler()
	uh := UserHandler.NewHandler(ur)
	bh := BlogHandler.NewHandler(br)

	// Router Initialize
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)

	// Router Configuration
	router.MaxMultipartMemory = 10 << 20 // MB : 10 MB

	// Global Routes
	router.GET("/", mainHandler.Index)
	router.NoRoute(mainHandler.NotFound)

	// Socket Routes
	router.GET("/blog", bh.BlogPageIndex)
	router.GET("/user", uh.UserPageIndex)

	// Start Server
	err = router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
