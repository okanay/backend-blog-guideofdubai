-- BLOG STATUS TYPE
CREATE TYPE blog_status AS ENUM ('draft', 'published', 'archived', 'deleted');

-- MAIN TABLE: BLOG POSTS (featured alanı kaldırıldı)
CREATE TABLE IF NOT EXISTS blog_posts (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    group_id TEXT NOT NULL,
    slug TEXT NOT NULL,
    language TEXT NOT NULL,
    status blog_status DEFAULT 'draft' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    published_at TIMESTAMPTZ DEFAULT NULL,
    UNIQUE (slug, language)
);

-- METADATA TABLE
CREATE TABLE IF NOT EXISTS blog_metadata (
    id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    image TEXT
);

-- CONTENT TABLE
CREATE TABLE IF NOT EXISTS blog_content (
    id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    image TEXT,
    read_time INTEGER DEFAULT 0,
    html TEXT NOT NULL,
    json TEXT NOT NULL
);

-- STATISTICS TABLE
CREATE TABLE IF NOT EXISTS blog_stats (
    id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE PRIMARY KEY,
    views INTEGER DEFAULT 0,
    likes INTEGER DEFAULT 0,
    shares INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    last_viewed_at TIMESTAMPTZ
);

-- CATEGORIES TABLE
CREATE TABLE IF NOT EXISTS categories (
    name TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    PRIMARY KEY (value)
);

-- TAGS TABLE
CREATE TABLE IF NOT EXISTS tags (
    name TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    PRIMARY KEY (value)
);

-- BLOG-CATEGORY RELATIONSHIP TABLE
CREATE TABLE IF NOT EXISTS blog_categories (
    blog_id UUID REFERENCES blog_posts (id) ON DELETE CASCADE,
    category_name TEXT REFERENCES categories (name) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, category_name)
);

-- BLOG-TAG RELATIONSHIP TABLE
CREATE TABLE IF NOT EXISTS blog_tags (
    blog_id UUID REFERENCES blog_posts (id) ON DELETE CASCADE,
    tag_name TEXT REFERENCES tags (name) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, tag_name)
);

-- FEATURED BLOGS TABLE (YENİ)
CREATE TABLE IF NOT EXISTS blog_featured (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    blog_id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE,
    language TEXT NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    CONSTRAINT blog_featured_language_position_key UNIQUE (language, position) DEFERRABLE INITIALLY IMMEDIATE
);

CREATE OR REPLACE FUNCTION sync_featured_blog_language()
RETURNS TRIGGER AS $$
DECLARE
    current_position INTEGER;
    target_position INTEGER;
    max_pos INTEGER;
    conflict_exists INTEGER;
BEGIN
    IF NEW.language IS DISTINCT FROM OLD.language THEN

        SELECT bf.position
        INTO current_position
        FROM blog_featured bf
        WHERE bf.blog_id = NEW.id;

        IF current_position IS NULL THEN
            RAISE NOTICE 'Blog post % is not featured, skipping position update.', NEW.id;
            RETURN NEW;
        END IF;

        SELECT 1
        INTO conflict_exists
        FROM blog_featured bf
        WHERE bf.language = NEW.language
          AND bf.position = current_position
          AND bf.blog_id != NEW.id;

        IF conflict_exists IS NOT NULL THEN
            RAISE NOTICE 'Position % is already taken in language %. Finding new position for blog post %.', current_position, NEW.language, NEW.id;
            SELECT COALESCE(MAX(bf.position), 0)
            INTO max_pos
            FROM blog_featured bf
            WHERE bf.language = NEW.language;

            target_position := max_pos + 1;
            RAISE NOTICE 'Assigning new position % to blog post % in language %.', target_position, NEW.id, NEW.language;

        ELSE
             RAISE NOTICE 'Position % is available in language %. Keeping original position for blog post %.', current_position, NEW.language, NEW.id;
            target_position := current_position;
        END IF;

        UPDATE blog_featured
        SET
            language = NEW.language,
            position = target_position,
            updated_at = NOW()
        WHERE
            blog_id = NEW.id;

    END IF;

    RETURN NEW;

END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_featured_blog_language
AFTER UPDATE OF language ON blog_posts
FOR EACH ROW
EXECUTE FUNCTION sync_featured_blog_language();

-- Yeni birleşik indeksler (slug ve group_id sorguları için)
CREATE INDEX IF NOT EXISTS idx_blog_posts_slug_language ON blog_posts (slug, language);

CREATE INDEX IF NOT EXISTS idx_blog_posts_group_id_language ON blog_posts (group_id, language);

-- Temel filtreleme indeksleri
CREATE INDEX IF NOT EXISTS idx_blog_posts_user_id ON blog_posts (user_id);

CREATE INDEX IF NOT EXISTS idx_blog_posts_status ON blog_posts (status);

-- İlişkisel indeksler (Foreign Key)
CREATE INDEX IF NOT EXISTS idx_blog_metadata_id ON blog_metadata (id);

CREATE INDEX IF NOT EXISTS idx_blog_content_id ON blog_content (id);

CREATE INDEX IF NOT EXISTS idx_blog_stats_id ON blog_stats (id);

-- İstatistik sorguları için indeksler
CREATE INDEX IF NOT EXISTS idx_blog_stats_views ON blog_stats (views);

CREATE INDEX IF NOT EXISTS idx_blog_stats_likes ON blog_stats (likes);

-- Kategori ve etiketler için indeksler
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories (name);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags (name);

-- Featured blog indeksleri (YENİ)
CREATE INDEX IF NOT EXISTS idx_blog_featured_blog_id ON blog_featured (blog_id);

CREATE INDEX IF NOT EXISTS idx_blog_featured_language_position ON blog_featured (language, position);
