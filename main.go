package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	c "github.com/okanay/backend-blog-guideofdubai/configs"
	db "github.com/okanay/backend-blog-guideofdubai/database"
	"github.com/okanay/backend-blog-guideofdubai/handlers"
	BlogHandler "github.com/okanay/backend-blog-guideofdubai/handlers/blog"
	ImageHandler "github.com/okanay/backend-blog-guideofdubai/handlers/image"
	UserHandler "github.com/okanay/backend-blog-guideofdubai/handlers/user"
	mw "github.com/okanay/backend-blog-guideofdubai/middlewares"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	ImageRepository "github.com/okanay/backend-blog-guideofdubai/repositories/image"
	R2Repository "github.com/okanay/backend-blog-guideofdubai/repositories/r2"
	TokenRepository "github.com/okanay/backend-blog-guideofdubai/repositories/token"
	UserRepository "github.com/okanay/backend-blog-guideofdubai/repositories/user"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
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

	ir := ImageRepository.NewRepository(sqlDB)
	r2 := R2Repository.NewRepository(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_ACCESS_KEY_SECRET"),
		os.Getenv("R2_BUCKET_NAME"),
		os.Getenv("R2_FOLDER_NAME"),
		os.Getenv("R2_PUBLIC_URL_BASE"),
		os.Getenv("R2_ENDPOINT"),
	)

	blogCache := cache.NewCache(20 * time.Minute)

	// Handler Initialization
	mh := handlers.NewHandler()
	uh := UserHandler.NewHandler(ur, tr)
	bh := BlogHandler.NewHandler(br, blogCache)
	ih := ImageHandler.NewHandler(ir, r2)

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
		blogPublic.GET("/tags", bh.SelectAllTags)
		blogPublic.GET("/categories", bh.SelectAllCategories)
		blogPublic.GET("/recent", bh.SelectRecentPosts)
		blogPublic.GET("/featured", bh.SelectFeaturedPosts)
		blogPublic.GET("/related", bh.SelectRelatedPosts)
		blogPublic.GET("/sitemap", bh.SelectBlogSitemap)
	}

	// Blog Routes - Authenticated Access
	blogAuth := auth.Group("/blog")
	{
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

	imageAuth := auth.Group("/images")
	{
		imageAuth.POST("/presign", ih.CreatePresignedURL)
		imageAuth.POST("/confirm", ih.ConfirmUpload)
		imageAuth.GET("", ih.GetUserImages)
		imageAuth.DELETE("/:id", ih.DeleteImage)
	}

	// Start Server
	err = router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
