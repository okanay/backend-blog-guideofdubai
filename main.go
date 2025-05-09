package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	c "github.com/okanay/backend-blog-guideofdubai/configs"
	db "github.com/okanay/backend-blog-guideofdubai/database"
	"github.com/okanay/backend-blog-guideofdubai/handlers"
	AIHandler "github.com/okanay/backend-blog-guideofdubai/handlers/ai"
	BlogHandler "github.com/okanay/backend-blog-guideofdubai/handlers/blog"
	ImageHandler "github.com/okanay/backend-blog-guideofdubai/handlers/image"
	UserHandler "github.com/okanay/backend-blog-guideofdubai/handlers/user"
	"github.com/okanay/backend-blog-guideofdubai/middlewares"
	mw "github.com/okanay/backend-blog-guideofdubai/middlewares"
	AIRepository "github.com/okanay/backend-blog-guideofdubai/repositories/ai"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	ImageRepository "github.com/okanay/backend-blog-guideofdubai/repositories/image"
	R2Repository "github.com/okanay/backend-blog-guideofdubai/repositories/r2"
	TokenRepository "github.com/okanay/backend-blog-guideofdubai/repositories/token"
	UserRepository "github.com/okanay/backend-blog-guideofdubai/repositories/user"
	AIService "github.com/okanay/backend-blog-guideofdubai/services/ai"
	"github.com/okanay/backend-blog-guideofdubai/services/cache"
	"github.com/okanay/backend-blog-guideofdubai/types"
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

	ar := AIRepository.NewRepository(os.Getenv("OPENAI_API_KEY"))

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

	blogCache := cache.NewCache(30 * time.Minute)
	aiRateLimit := middlewares.NewAIRateLimitMiddleware(blogCache)
	ais := AIService.NewAIService(ar, br)

	// Handler Initialization
	mh := handlers.NewHandler()
	uh := UserHandler.NewHandler(ur, tr)
	bh := BlogHandler.NewHandler(br, blogCache)
	ih := ImageHandler.NewHandler(ir, r2)
	ah := AIHandler.NewHandler(ar, br, ais)

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

	blogAuth := auth.Group("/blog")
	{
		// Oluşturma işlemleri
		blogAuth.POST("", bh.CreateBlogPost)
		blogAuth.POST("/tag", bh.CreateBlogTag)
		blogAuth.POST("/category", bh.CreateBlogCategory)

		// Güncelleme işlemleri
		blogAuth.PATCH("", bh.UpdateBlogPost)
		blogAuth.PATCH("/status", bh.UpdateBlogStatus)

		// Featured işlemleri (YENİ)
		blogAuth.POST("/featured", bh.AddToFeatured)
		blogAuth.DELETE("/featured/:id", bh.RemoveFromFeatured)
		blogAuth.PATCH("/featured/ordering", bh.UpdateFeaturedOrdering)

		blogAuth.GET("/stats", bh.GetBlogStats)
		blogAuth.GET("/stats/:id", bh.GetBlogStatByID)
		// Silme işlemleri
		blogAuth.DELETE("/:id", bh.DeleteBlogByID)
	}

	// Blog Routes - Public Access
	blogPublic := router.Group("/blog")
	{
		blogPublic.GET("", bh.SelectBlogBySlugID)
		blogPublic.GET("/cards", bh.SelectBlogCards)
		blogPublic.GET("/:id", bh.SelectBlogByID)
		blogPublic.GET("/tags", bh.SelectAllTags)
		blogPublic.GET("/categories", bh.SelectAllCategories)
		blogPublic.GET("/recent", bh.SelectRecentPosts)
		blogPublic.GET("/featured", bh.GetFeaturedBlogs)
		blogPublic.GET("/related", bh.SelectRelatedPosts)
		blogPublic.GET("/sitemap", bh.SelectBlogSitemap)

		blogPublic.GET("/view", bh.TrackBlogView)
	}

	imageAuth := auth.Group("/images")
	{
		imageAuth.POST("/presign", ih.CreatePresignedURL)
		imageAuth.POST("/confirm", ih.ConfirmUpload)
		imageAuth.GET("", ih.GetUserImages)
		imageAuth.DELETE("/:id", ih.DeleteImage)
	}

	aiRoutes := auth.Group("/ai")
	aiRoutes.Use(aiRateLimit.RateLimit())
	{
		aiRoutes.POST("/translate", ah.TranslateBlogPostJSON)
		aiRoutes.POST("/generate-metadata", ah.GenerateBlogMetadata)
	}

	adminAuth := auth.Group("/admin")
	adminAuth.Use(mw.RequireRole("Admin"))
	{
		adminAuth.GET("/cache", func(c *gin.Context) {
			stats := blogCache.GetStats()
			c.JSON(200, gin.H{
				"success": true,
				"stats":   stats,
			})
		})

		adminAuth.DELETE("/cache", func(c *gin.Context) {
			// Blog cache'ini temizle, ama AI rate limit'lerini korur
			blogCache.ClearExceptPrefixes([]string{"ai_rate_limit:", "ai_rate_limit_minute:"})

			c.JSON(200, gin.H{
				"success": true,
				"message": "Blog cache başarıyla temizlendi",
			})
		})

		// AI Rate Limit cache temizleme endpoint'i (yeni)
		adminAuth.DELETE("/ai/rate-limits", func(c *gin.Context) {
			// Sadece AI rate limit cache'lerini temizle
			blogCache.ClearAIRateLimits()

			c.JSON(200, gin.H{
				"success": true,
				"message": "AI rate limit cache'leri başarıyla temizlendi",
			})
		})

		adminAuth.GET("/ai/rate-limits", func(c *gin.Context) {
			// Tüm rate limit kayıtlarını al
			rateLimits := []map[string]any{}

			// Cache'den "ai_rate_limit:" önekine sahip tüm anahtarları ara
			allRateLimits := blogCache.GetAllWithPrefix("ai_rate_limit:")

			for _, data := range allRateLimits {
				var rateInfo types.RateLimitInfo
				if err := json.Unmarshal(data, &rateInfo); err == nil {
					rateLimits = append(rateLimits, map[string]any{
						"userId":          rateInfo.UserID,
						"requestCount":    rateInfo.RequestCount,
						"tokensUsed":      rateInfo.TokensUsed,
						"firstRequest":    rateInfo.FirstRequest,
						"lastRequest":     rateInfo.LastRequest,
						"windowResetAt":   rateInfo.WindowResetAt,
						"requestsPerMin":  rateInfo.RequestsPerMin,
						"minuteStartedAt": rateInfo.MinuteStartedAt,
						"remaining": gin.H{
							"requests": configs.AI_RATE_LIMIT_MAX_REQUESTS - rateInfo.RequestCount,
							"tokens":   configs.AI_RATE_LIMIT_MAX_TOKENS - rateInfo.TokensUsed,
						},
					})
				}
			}

			c.JSON(200, gin.H{
				"success": true,
				"data":    rateLimits,
			})
		})
	}

	// Start Server
	err = router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
