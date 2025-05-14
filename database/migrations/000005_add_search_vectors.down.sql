-- Trigger'ları kaldır
DROP TRIGGER IF EXISTS blog_posts_search_vector_update ON blog_posts;

DROP TRIGGER IF EXISTS blog_content_search_vector_update ON blog_content;

-- Fonksiyonu kaldır
DROP FUNCTION IF EXISTS update_blog_search_vectors ();

-- İndeksleri kaldır
DROP INDEX IF EXISTS blog_posts_search_idx;

DROP INDEX IF EXISTS blog_content_search_idx;

-- Sütunları kaldır
ALTER TABLE blog_posts
DROP COLUMN IF EXISTS search_vector;

ALTER TABLE blog_content
DROP COLUMN IF EXISTS search_vector;
