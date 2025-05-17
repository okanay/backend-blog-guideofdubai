package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// Proje modülleri

	c "github.com/okanay/backend-blog-guideofdubai/configs"
	db "github.com/okanay/backend-blog-guideofdubai/database"
	"github.com/okanay/backend-blog-guideofdubai/handlers"
	AdminHandler "github.com/okanay/backend-blog-guideofdubai/handlers/admin"
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
)

// Uygulama bileşenlerini gruplamak için yapılar
type Repositories struct {
	User  *UserRepository.Repository
	Token *TokenRepository.Repository
	Blog  *BlogRepository.Repository
	AI    *AIRepository.Repository
	Image *ImageRepository.Repository
	R2    *R2Repository.Repository
}

type Services struct {
	BlogCache   *cache.Cache
	AIRateLimit *middlewares.AIRateLimitMiddleware
	AI          *AIService.AIService
}

type Handlers struct {
	Main  *handlers.Handler
	User  *UserHandler.Handler
	Blog  *BlogHandler.Handler
	Image *ImageHandler.Handler
	AI    *AIHandler.Handler
	Admin *AdminHandler.Handler
}

func main() {
	// 1. Çevresel Değişkenleri Yükle
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("[ENV]: .env dosyası yüklenemedi")
		return
	}

	// 2. Veritabanı Bağlantısı Kur
	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Veritabanına bağlanırken hata: %v", err)
		return
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	// 3. Repository Katmanını Başlat
	r := initRepositories(sqlDB)

	// 4. Servis Katmanını Başlat
	s := initServices(r)

	// 5. Handler Katmanını Başlat
	h := initHandlers(r, s)

	// 6. Router ve Middleware Yapılandırması
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)
	router.MaxMultipartMemory = 10 << 20 // 10 MB

	// Kimlik doğrulama gerektiren rotalar için grup
	auth := router.Group("/auth")
	auth.Use(mw.AuthMiddleware(r.User, r.Token))

	// Global Routes
	router.GET("/", h.Main.Index)
	router.NoRoute(h.Main.NotFound)

	// Authentication Routes (public)
	router.POST("/login", h.User.Login)
	router.POST("/register", h.User.Register)
	auth.GET("/logout", h.User.Logout)
	auth.GET("/get-me", h.User.GetMe)

	// Blog Routes - Auth Required
	blogAuth := auth.Group("/blog")
	{
		// Oluşturma işlemleri
		blogAuth.POST("", h.Blog.CreateBlogPost)
		blogAuth.POST("/tag", h.Blog.CreateBlogTag)
		blogAuth.POST("/category", h.Blog.CreateBlogCategory)

		// Güncelleme işlemleri
		blogAuth.PATCH("", h.Blog.UpdateBlogPost)
		blogAuth.PATCH("/status", h.Blog.UpdateBlogStatus)

		// Featured işlemleri
		blogAuth.POST("/featured", h.Blog.AddToFeatured)
		blogAuth.DELETE("/featured/:id", h.Blog.RemoveFromFeatured)
		blogAuth.PATCH("/featured/ordering", h.Blog.UpdateFeaturedOrdering)

		// İstatistik işlemleri
		blogAuth.GET("/stats", h.Blog.GetBlogStats)
		blogAuth.GET("/stats/:id", h.Blog.GetBlogStatByID)

		// Silme işlemleri
		blogAuth.DELETE("/:id", h.Blog.DeleteBlogByID)
	}

	// Blog Routes - Public Access
	blogPublic := router.Group("/blog")
	{
		blogPublic.GET("", h.Blog.SelectBlogBySlugID)
		blogPublic.GET("/cards", h.Blog.SelectBlogCards)
		blogPublic.GET("/:id", h.Blog.SelectBlogByID)
		blogPublic.GET("/tags", h.Blog.SelectAllTags)
		blogPublic.GET("/categories", h.Blog.SelectAllCategories)
		blogPublic.GET("/recent", h.Blog.SelectRecentPosts)
		blogPublic.GET("/featured", h.Blog.GetFeaturedBlogs)
		blogPublic.GET("/most-viewed", h.Blog.SelectMostViewedPosts)
		blogPublic.GET("/related", h.Blog.SelectRelatedPosts)
		blogPublic.GET("/sitemap", h.Blog.SelectBlogSitemap)
		blogPublic.GET("/view", h.Blog.TrackBlogView)
	}

	// Image Routes
	imageAuth := auth.Group("/images")
	{
		imageAuth.POST("/presign", h.Image.CreatePresignedURL)
		imageAuth.POST("/confirm", h.Image.ConfirmUpload)
		imageAuth.GET("", h.Image.GetUserImages)
		imageAuth.DELETE("/:id", h.Image.DeleteImage)
	}

	// AI Routes
	aiRoutes := auth.Group("/ai")
	aiRoutes.Use(s.AIRateLimit.RateLimit())
	{
		aiRoutes.POST("/translate", h.AI.TranslateBlogPostJSON)
		aiRoutes.POST("/generate-metadata", h.AI.GenerateBlogMetadata)
	}

	// Admin Routes
	adminAuth := auth.Group("/admin")
	adminAuth.Use(mw.RequireRole("Admin"))
	adminCache := adminAuth.Group("/cache")
	{
		adminCache.GET("", h.Admin.GetCacheStats)
		adminCache.DELETE("", h.Admin.ClearAllCache)
		adminCache.GET("/items", h.Admin.GetCacheItems)
		adminCache.DELETE("/prefix", h.Admin.ClearCacheWithPrefix)
		adminCache.GET("/rate-limits", h.Admin.GetAIRateLimits)
		adminCache.DELETE("/rate-limits", h.Admin.ClearAIRateLimits)
		adminCache.DELETE("/rate-limits/:userId", h.Admin.ResetUserRateLimit)
	}

	// 7. Sunucuyu Başlat
	port := os.Getenv("PORT")
	log.Printf("[SERVER]: %s portu üzerinde dinleniyor...", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[SERVER]: Sunucu başlatılırken hata: %v", err)
	}
}

// Repository'lerin başlatılması
func initRepositories(sqlDB *sql.DB) Repositories {
	return Repositories{
		User:  UserRepository.NewRepository(sqlDB),
		Token: TokenRepository.NewRepository(sqlDB),
		Blog:  BlogRepository.NewRepository(sqlDB),
		AI:    AIRepository.NewRepository(os.Getenv("OPENAI_API_KEY")),
		Image: ImageRepository.NewRepository(sqlDB),
		R2: R2Repository.NewRepository(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_ACCESS_KEY_SECRET"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_FOLDER_NAME"),
			os.Getenv("R2_PUBLIC_URL_BASE"),
			os.Getenv("R2_ENDPOINT"),
		),
	}
}

// Servislerin başlatılması
func initServices(repos Repositories) Services {
	// Cache ve servis oluştur
	blogCache := cache.NewCache(30 * time.Minute)

	return Services{
		BlogCache:   blogCache,
		AIRateLimit: middlewares.NewAIRateLimitMiddleware(blogCache),
		AI:          AIService.NewAIService(repos.AI, repos.Blog),
	}
}

// Handler'ların başlatılması
func initHandlers(repos Repositories, services Services) Handlers {
	return Handlers{
		Main:  handlers.NewHandler(),
		User:  UserHandler.NewHandler(repos.User, repos.Token),
		Blog:  BlogHandler.NewHandler(repos.Blog, services.BlogCache),
		Image: ImageHandler.NewHandler(repos.Image, repos.R2),
		AI:    AIHandler.NewHandler(repos.AI, repos.Blog, services.AI),
		Admin: AdminHandler.NewHandler(repos.Blog, services.BlogCache),
	}
}
