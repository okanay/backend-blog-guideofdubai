-- Önce tsvector sütunları ekle
ALTER TABLE blog_posts ADD COLUMN search_vector tsvector;
ALTER TABLE blog_content ADD COLUMN search_vector tsvector;

-- Mevcut verileri güncellemek için fonksiyon oluştur
CREATE OR REPLACE FUNCTION update_blog_search_vectors() RETURNS TRIGGER AS $$
BEGIN
  -- Blog posts tablosunda dil bilgisini dikkate alarak güncelleme yap
  IF TG_TABLE_NAME = 'blog_posts' THEN
    -- Tekli parametre versiyonunu kullan
    NEW.search_vector = setweight(to_tsvector(COALESCE(NEW.slug, '')), 'A');

  -- Blog content tablosunda başlık ve açıklamayı ekle
  ELSIF TG_TABLE_NAME = 'blog_content' THEN
    NEW.search_vector =
      setweight(to_tsvector(COALESCE(NEW.title, '')), 'A') ||
      setweight(to_tsvector(COALESCE(NEW.description, '')), 'B');
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger'ları oluştur
CREATE TRIGGER blog_posts_search_vector_update
BEFORE INSERT OR UPDATE ON blog_posts
FOR EACH ROW EXECUTE FUNCTION update_blog_search_vectors();

CREATE TRIGGER blog_content_search_vector_update
BEFORE INSERT OR UPDATE ON blog_content
FOR EACH ROW EXECUTE FUNCTION update_blog_search_vectors();

-- Mevcut verileri güncelle
UPDATE blog_posts SET search_vector = setweight(to_tsvector(COALESCE(slug, '')), 'A');

-- Her blog için content verisini güncelle
DO $$
DECLARE
  blog_record RECORD;
BEGIN
  FOR blog_record IN SELECT bp.id, bp.language FROM blog_posts bp
  LOOP
    UPDATE blog_content bc SET
      search_vector =
        setweight(to_tsvector(COALESCE(title, '')), 'A') ||
        setweight(to_tsvector(COALESCE(description, '')), 'B')
    WHERE bc.id = blog_record.id;
  END LOOP;
END;
$$;

-- İndeksleri oluştur
CREATE INDEX blog_posts_search_idx ON blog_posts USING gin(search_vector);
CREATE INDEX blog_content_search_idx ON blog_content USING gin(search_vector);
