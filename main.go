package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	c "github.com/okanay/backend-blog-guideofdubai/configs"
	db "github.com/okanay/backend-blog-guideofdubai/database"
	"github.com/okanay/backend-blog-guideofdubai/handlers"
	UserHandler "github.com/okanay/backend-blog-guideofdubai/handlers/user"
	"github.com/okanay/backend-blog-guideofdubai/middlewares"
	UserRepository "github.com/okanay/backend-blog-guideofdubai/repositories/user"
)

func main() {
	// Environment Variables and Database Connection
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("[ENV]: .env file not loaded")
		return
	}

	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Error connecting to database")
		return
	}
	defer sqlDB.Close()

	// Repository Initialization
	ur := UserRepository.NewRepository(sqlDB)

	// Handler Initialization
	mainHandler := handlers.NewHandler()
	uh := UserHandler.NewHandler(ur)

	// Router Initialize
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)

	// Router Configuration
	router.MaxMultipartMemory = 10 << 20 // MB : 10 MB

	auth := router.Group("/auth")
	auth.Use(middlewares.AuthMiddleware(ur))

	// Global Routes
	router.GET("/", mainHandler.Index)
	router.NoRoute(mainHandler.NotFound)

	// Socket Routes
	router.POST("/user/register", uh.CreateNewUser)

	// Start Server
	err = router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
