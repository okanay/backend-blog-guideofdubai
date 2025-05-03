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
    -- Her blog-dil kombinasyonu unique olmalı
    UNIQUE (blog_id, language),
    -- Her dilde pozisyonlar unique olmalı
    UNIQUE (language, position)
);

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
