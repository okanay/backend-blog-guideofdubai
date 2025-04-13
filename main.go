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
	mw "github.com/okanay/backend-blog-guideofdubai/middlewares"
	TokenRepository "github.com/okanay/backend-blog-guideofdubai/repositories/token"
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
	tr := TokenRepository.NewRepository(sqlDB)

	// Handler Initialization
	mh := handlers.NewHandler()
	uh := UserHandler.NewHandler(ur, tr)

	// Router Initialize
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)

	// Router Configuration
	router.MaxMultipartMemory = 10 << 20 // MB : 10 MB

	auth := router.Group("/auth")
	auth.Use(mw.AuthMiddleware(ur, tr))

	// Global Routes
	router.GET("/", mh.Index)
	router.NoRoute(mh.NotFound)

	// User Routes
	router.POST("/login", uh.Login)
	router.POST("/register", uh.Register)

	// Start Server
	err = router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
