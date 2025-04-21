package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	c "github.com/okanay/backend-blog-guideofdubai/configs"
	db "github.com/okanay/backend-blog-guideofdubai/database"
	"github.com/okanay/backend-blog-guideofdubai/handlers"
	BlogHandler "github.com/okanay/backend-blog-guideofdubai/handlers/blog"
	UserHandler "github.com/okanay/backend-blog-guideofdubai/handlers/user"
	mw "github.com/okanay/backend-blog-guideofdubai/middlewares"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
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
	br := BlogRepository.NewRepository(sqlDB)

	// Handler Initialization
	mh := handlers.NewHandler()
	uh := UserHandler.NewHandler(ur, tr)
	bh := BlogHandler.NewHandler(br)

	// Router ve Middleware Yapılandırması
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)
	router.MaxMultipartMemory = 10 << 20 // 10 MB

	// Kimlik doğrulama gerektiren rotalar için grup
	auth := router.Group("/auth")
	auth.Use(mw.AuthMiddleware(ur, tr))

	// Global Routes
	router.GET("/", mh.Index)
	router.NoRoute(mh.NotFound)

	// Authentication Routes (public)
	router.POST("/login", uh.Login)
	router.POST("/register", uh.Register)
	auth.GET("/logout", uh.Logout)
	auth.GET("/get-me", uh.GetMe)

	// Blog Routes - Public Access
	blogPublic := router.Group("/blog")
	{
		blogPublic.GET("", bh.SelectBlogByGroupID)
		blogPublic.GET("/cards", bh.SelectBlogCards)
		blogPublic.GET("/:id", bh.SelectBlogByID)
	}

	// Blog Routes - Authenticated Access
	blogAuth := auth.Group("/blog")
	{
		// Listeleme işlemleri
		blogAuth.GET("/tags", bh.SelectAllTags)
		blogAuth.GET("/categories", bh.SelectAllCategories)

		// Oluşturma işlemleri
		blogAuth.POST("", bh.CreateBlogPost)
		blogAuth.POST("/tag", bh.CreateBlogTag)
		blogAuth.POST("/category", bh.CreateBlogCategory)

		// Güncelleme işlemleri
		blogAuth.PATCH("", bh.UpdateBlogPost)
		blogAuth.PATCH("/status", bh.UpdateBlogStatus)

		// Silme işlemleri
		blogAuth.DELETE("/:id", bh.DeleteBlogByID)
	}

	// Start Server
	err = router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
