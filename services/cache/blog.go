// services/blog_cache.go
package cache

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// BlogCacheService blog verilerinin cache'lenmesi için kullanılacak yapı
type BlogCacheService struct {
	cache *Cache
}

// NewBlogCacheService yeni bir BlogCacheService oluşturur
func NewBlogCacheService(cache *Cache) *BlogCacheService {
	return &BlogCacheService{
		cache: cache,
	}
}

// GetBlogByID blog'u ID'ye göre cache'den getirir
func (s *BlogCacheService) GetBlogByID(blogID uuid.UUID) (*types.BlogPostView, bool) {
	cacheKey := fmt.Sprintf("blog_id:%s", blogID.String())
	return s.getBlogFromCache(cacheKey)
}

// SaveBlogByID blog'u ID'ye göre cache'e kaydeder
func (s *BlogCacheService) SaveBlogByID(blogID uuid.UUID, blog *types.BlogPostView) error {
	cacheKey := fmt.Sprintf("blog_id:%s", blogID.String())
	return s.saveBlogToCache(cacheKey, blog)
}

// GetBlogAndAlternativesBySlug blog ve alternatiflerini slug'a göre cache'den getirir
func (s *BlogCacheService) GetBlogAndAlternativesBySlug(slug string) (*types.BlogPostView, []*types.BlogPostView, bool) {
	cacheKey := fmt.Sprintf("blog_slug:%s", slug)

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, nil, false
	}

	type cachedBlogGroup struct {
		MainBlog     *types.BlogPostView   `json:"mainBlog"`
		Alternatives []*types.BlogPostView `json:"alternatives"`
	}

	var blogGroup cachedBlogGroup
	if err := json.Unmarshal(cachedData, &blogGroup); err != nil {
		return nil, nil, false
	}

	return blogGroup.MainBlog, blogGroup.Alternatives, true
}

// SaveBlogAndAlternativesBySlug blog ve alternatiflerini slug'a göre cache'e kaydeder
func (s *BlogCacheService) SaveBlogAndAlternativesBySlug(slug string, mainBlog *types.BlogPostView, alternatives []*types.BlogPostView) error {
	cacheKey := fmt.Sprintf("blog_slug:%s", slug)

	type cachedBlogGroup struct {
		MainBlog     *types.BlogPostView   `json:"mainBlog"`
		Alternatives []*types.BlogPostView `json:"alternatives"`
	}

	blogGroup := cachedBlogGroup{
		MainBlog:     mainBlog,
		Alternatives: alternatives,
	}

	jsonData, err := json.Marshal(blogGroup)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)

	// Ayrıca GroupID cache'i oluştur, eğer başka biri groupID ile sorgulama yaparsa diye
	if mainBlog != nil && mainBlog.GroupID != "" {
		groupCacheKey := fmt.Sprintf("blog_group:%s", mainBlog.GroupID)
		s.cache.Set(groupCacheKey, jsonData)

		// Ana blog için bireysel cache de tutalım
		mainBlogID, err := uuid.Parse(mainBlog.ID)
		if err == nil {
			s.SaveBlogByID(mainBlogID, mainBlog)
		}
	}

	// Alternatif bloglar için de bireysel cache tutalım
	for _, alt := range alternatives {
		altID, err := uuid.Parse(alt.ID)
		if err == nil {
			s.SaveBlogByID(altID, alt)

			// Alternatif slug'lar için de cache tutalım
			if alt.Slug != "" && alt.Slug != slug {
				altSlugCacheKey := fmt.Sprintf("blog_slug:%s", alt.Slug)
				// Alternatif cache'te ana blog olarak bu alternatifi, diğer alternatifler olarak da tüm listeyi koy
				var altBlogGroup cachedBlogGroup
				altBlogGroup.MainBlog = alt

				// Ana blog dahil tüm diğer içerikleri alternatif olarak ekle
				var otherBlogs []*types.BlogPostView
				if mainBlog != nil && mainBlog.ID != alt.ID {
					otherBlogs = append(otherBlogs, mainBlog)
				}

				for _, otherAlt := range alternatives {
					if otherAlt.ID != alt.ID {
						otherBlogs = append(otherBlogs, otherAlt)
					}
				}

				altBlogGroup.Alternatives = otherBlogs
				altJsonData, _ := json.Marshal(altBlogGroup)
				s.cache.Set(altSlugCacheKey, altJsonData)
			}
		}
	}

	return nil
}

func (s *BlogCacheService) GetBlogCards(queryOptions types.BlogCardQueryOptions) ([]types.BlogPostCardView, bool) {
	cacheKey := s.GenerateBlogCardsCacheKey(queryOptions)
	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}
	var blogs []types.BlogPostCardView
	if err := json.Unmarshal(cachedData, &blogs); err != nil {
		return nil, false
	}
	return blogs, true
}

func (s *BlogCacheService) SaveBlogCards(queryOptions types.BlogCardQueryOptions, blogs []types.BlogPostCardView) error {
	cacheKey := s.GenerateBlogCardsCacheKey(queryOptions)
	jsonData, err := json.Marshal(blogs)
	if err != nil {
		return err
	}
	s.cache.Set(cacheKey, jsonData)
	return nil
}

// GenerateBlogCardsCacheKey sorgu seçeneklerinden belirleyici bir cache key oluşturur
func (s *BlogCacheService) GenerateBlogCardsCacheKey(options types.BlogCardQueryOptions) string {
	// Ana key prefixini oluştur
	prefix := "blog_cards"

	// Filtreleme kriterlerini bir hash olarak ekle
	h := sha256.New()

	// ID varsa ekle
	if options.ID != uuid.Nil {
		h.Write([]byte(options.ID.String()))
	}

	// ID'ler slice'ı varsa ekle
	if len(options.IDs) > 0 {
		for _, id := range options.IDs {
			h.Write([]byte(id.String()))
		}
	}

	// Diğer string türündeki filtreleri ekle
	h.Write([]byte(options.Title))
	h.Write([]byte(options.Language))
	h.Write([]byte(options.CategoryValue))
	h.Write([]byte(options.TagValue))
	h.Write([]byte(string(options.Status)))
	h.Write([]byte(options.SortBy))
	h.Write([]byte(string(options.SortDirection)))

	// Boolean değerleri ekle
	if options.Featured {
		h.Write([]byte("featured:true"))
	}

	// Sayısal değerleri ekle
	h.Write([]byte(strconv.Itoa(options.Limit)))
	h.Write([]byte(strconv.Itoa(options.Offset)))

	// Tarih filtreleri varsa ekle
	if options.StartDate != nil {
		h.Write([]byte(options.StartDate.Format(time.RFC3339)))
	}
	if options.EndDate != nil {
		h.Write([]byte(options.EndDate.Format(time.RFC3339)))
	}

	// Hash'i base64 ile kodla
	hashSum := h.Sum(nil)
	hashBase64 := base64.StdEncoding.EncodeToString(hashSum)

	// Sonuç: prefix:hash
	return fmt.Sprintf("%s:%s", prefix, hashBase64)
}

// GetFeaturedPostsByLanguage belirli bir dil için featured postları cache'den getirir
func (s *BlogCacheService) GetFeaturedPostsByLanguage(language string) ([]types.BlogPostCardView, bool) {
	cacheKey := fmt.Sprintf("featured_posts_%s", language)

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var blogs []types.BlogPostCardView
	if err := json.Unmarshal(cachedData, &blogs); err != nil {
		return nil, false
	}

	return blogs, true
}

// SaveFeaturedPostsByLanguage belirli bir dil için featured postları cache'e kaydeder
func (s *BlogCacheService) SaveFeaturedPostsByLanguage(language string, blogs []types.BlogPostCardView) error {
	cacheKey := fmt.Sprintf("featured_posts_%s", language)

	jsonData, err := json.Marshal(blogs)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}

// InvalidateFeaturedPosts belirli bir dilin featured posts cache'ini temizler
func (s *BlogCacheService) InvalidateFeaturedPosts(language string) {
	cacheKey := fmt.Sprintf("featured_posts_%s", language)
	s.cache.Delete(cacheKey)
}

// InvalidateAllFeaturedPosts tüm dillerdeki featured posts cache'ini temizler
func (s *BlogCacheService) InvalidateAllFeaturedPosts() {
	// Featured posts cache'lerini temizle
	s.cache.ClearPrefix("featured_posts_")
}

// GetFeaturedPosts öne çıkan blog yazılarını cache'den getirir
func (s *BlogCacheService) GetFeaturedPosts() ([]types.BlogPostCardView, bool) {
	cacheKey := "featured_posts"

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var blogs []types.BlogPostCardView
	if err := json.Unmarshal(cachedData, &blogs); err != nil {
		return nil, false
	}

	return blogs, true
}

// SaveFeaturedPosts öne çıkan blog yazılarını cache'e kaydeder
func (s *BlogCacheService) SaveFeaturedPosts(blogs []types.BlogPostCardView) error {
	cacheKey := "featured_posts"

	jsonData, err := json.Marshal(blogs)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}

// GetRecentPosts son eklenen blog yazılarını cache'den getirir
func (s *BlogCacheService) GetRecentPosts() ([]types.BlogPostCardView, bool) {
	cacheKey := "recent_posts"

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var blogs []types.BlogPostCardView
	if err := json.Unmarshal(cachedData, &blogs); err != nil {
		return nil, false
	}

	return blogs, true
}

// SaveRecentPosts son eklenen blog yazılarını cache'e kaydeder
func (s *BlogCacheService) SaveRecentPosts(blogs []types.BlogPostCardView) error {
	cacheKey := "recent_posts"

	jsonData, err := json.Marshal(blogs)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}

// InvalidateAllBlogs tüm blog cache'lerini temizler
func (s *BlogCacheService) InvalidateAllBlogs() {
	s.cache.Clear()
}

// InvalidateBlogByID belirli bir blog ID'ye ait cache'i temizler
func (s *BlogCacheService) InvalidateBlogByID(blogID uuid.UUID) {
	cacheKey := fmt.Sprintf("blog_id:%s", blogID.String())
	s.cache.Delete(cacheKey)
}

func (s *BlogCacheService) GetRelatedPosts(blogID uuid.UUID, categories []string, tags []string, language string) ([]types.BlogPostCardView, bool) {
	cacheKey := fmt.Sprintf("related_posts:%s:%s:%s:%s",
		blogID.String(), language, strings.Join(categories, "_"), strings.Join(tags, "_"))

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var posts []types.BlogPostCardView
	if err := json.Unmarshal(cachedData, &posts); err != nil {
		return nil, false
	}

	return posts, true
}

// SaveRelatedPosts ilgili blog yazılarını cache'e kaydeder
func (s *BlogCacheService) SaveRelatedPosts(blogID uuid.UUID, categories []string, tags []string, language string, posts []types.BlogPostCardView) error {
	cacheKey := fmt.Sprintf("related_posts:%s:%s:%s:%s",
		blogID.String(), language, strings.Join(categories, "_"), strings.Join(tags, "_"))

	jsonData, err := json.Marshal(posts)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}

// GetSitemap sitemap verilerini cache'den getirir
func (s *BlogCacheService) GetSitemap() ([]map[string]any, bool) {
	cacheKey := "sitemap"

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var sitemap []map[string]any
	if err := json.Unmarshal(cachedData, &sitemap); err != nil {
		return nil, false
	}

	return sitemap, true
}

// SaveSitemap sitemap verilerini cache'e kaydeder
func (s *BlogCacheService) SaveSitemap(sitemap []gin.H) error {
	cacheKey := "sitemap"

	jsonData, err := json.Marshal(sitemap)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}

// GetMostViewedPosts en çok görüntülenen blog yazılarını cache'den getirir
func (s *BlogCacheService) GetMostViewedPosts(language string, period string) ([]types.BlogPostCardView, bool) {
	cacheKey := fmt.Sprintf("most_viewed_posts:%s:%s", language, period)

	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var blogs []types.BlogPostCardView
	if err := json.Unmarshal(cachedData, &blogs); err != nil {
		return nil, false
	}

	return blogs, true
}

// SaveMostViewedPosts en çok görüntülenen blog yazılarını cache'e kaydeder
func (s *BlogCacheService) SaveMostViewedPosts(language string, period string, blogs []types.BlogPostCardView) error {
	cacheKey := fmt.Sprintf("most_viewed_posts:%s:%s", language, period)

	jsonData, err := json.Marshal(blogs)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}

// İsteğe bağlı: Cache'i geçersiz kılma fonksiyonu
func (s *BlogCacheService) InvalidateMostViewedPosts() {
	s.cache.ClearPrefix("most_viewed_posts:")
}

// Bir görüntüleme eklendiğinde popüler postları güncelleme (opsiyonel çağrılabilir)
func (s *BlogCacheService) InvalidateMostViewedPostsOnView() {
	s.cache.ClearPrefix("most_viewed_posts:")
}

// Helper metotlar
func (s *BlogCacheService) getBlogFromCache(cacheKey string) (*types.BlogPostView, bool) {
	cachedData, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	var blog types.BlogPostView
	if err := json.Unmarshal(cachedData, &blog); err != nil {
		return nil, false
	}

	return &blog, true
}

func (s *BlogCacheService) saveBlogToCache(cacheKey string, blog *types.BlogPostView) error {
	jsonData, err := json.Marshal(blog)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
}
