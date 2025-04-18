-- İlişki tablolarını kaldır (önce foreign key'lere bağlı olan tabloları kaldırmalıyız)
DROP TABLE IF EXISTS blog_categories;

DROP TABLE IF EXISTS blog_tags;

-- İndeksleri kaldır
DROP INDEX IF EXISTS idx_blog_posts_slug_language;

DROP INDEX IF EXISTS idx_blog_posts_group_id_language;

DROP INDEX IF EXISTS idx_blog_posts_user_id;

DROP INDEX IF EXISTS idx_blog_posts_status;

DROP INDEX IF EXISTS idx_blog_metadata_id;

DROP INDEX IF EXISTS idx_blog_content_id;

DROP INDEX IF EXISTS idx_blog_stats_id;

DROP INDEX IF EXISTS idx_blog_stats_views;

DROP INDEX IF EXISTS idx_blog_stats_likes;

DROP INDEX IF EXISTS idx_categories_name;

DROP INDEX IF EXISTS idx_tags_name;

-- Alt tabloları kaldır
DROP TABLE IF EXISTS blog_metadata;

DROP TABLE IF EXISTS blog_content;

DROP TABLE IF EXISTS blog_stats;

-- Kategori ve etiket tablolarını kaldır
DROP TABLE IF EXISTS categories;

DROP TABLE IF EXISTS tags;

-- Ana tabloyu kaldır
DROP TABLE IF EXISTS blog_posts;

-- En son enum tipini kaldır (hiçbir tablo tarafından referans edilmediğinden emin olduğumuzda)
DROP TYPE IF EXISTS blog_status;
