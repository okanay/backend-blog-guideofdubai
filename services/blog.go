// services/blog_cache.go
package cache

import (
	"encoding/json"
	"fmt"

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

// GetBlogByGroupIDAndLang blog'u group ID ve dile göre cache'den getirir
func (s *BlogCacheService) GetBlogByGroupIDAndLang(groupID, lang string) (*types.BlogPostView, bool) {
	cacheKey := fmt.Sprintf("blog_id:%s:%s", groupID, lang)
	return s.getBlogFromCache(cacheKey)
}

// SaveBlogByGroupIDAndLang blog'u group ID ve dile göre cache'e kaydeder
func (s *BlogCacheService) SaveBlogByGroupIDAndLang(groupID, lang string, blog *types.BlogPostView) error {
	cacheKey := fmt.Sprintf("blog_id:%s:%s", groupID, lang)
	return s.saveBlogToCache(cacheKey, blog)
}

// GetBlogCards blog kartlarını sorgu parametrelerine göre cache'den getirir
func (s *BlogCacheService) GetBlogCards(queryOptions types.BlogCardQueryOptions) ([]types.BlogPostCardView, bool) {
	cacheKey := fmt.Sprintf("blog_cards:%v", queryOptions)

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

// SaveBlogCards blog kartlarını sorgu parametrelerine göre cache'e kaydeder
func (s *BlogCacheService) SaveBlogCards(queryOptions types.BlogCardQueryOptions, blogs []types.BlogPostCardView) error {
	cacheKey := fmt.Sprintf("blog_cards:%v", queryOptions)

	jsonData, err := json.Marshal(blogs)
	if err != nil {
		return err
	}

	s.cache.Set(cacheKey, jsonData)
	return nil
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
