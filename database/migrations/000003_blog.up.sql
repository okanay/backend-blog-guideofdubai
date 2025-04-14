-- BLOG STATUS TYPE
CREATE TYPE blog_status AS ENUM ('draft', 'published', 'archived', 'deleted');

-- MAIN TABLE: BLOG POSTS
CREATE TABLE IF NOT EXISTS blog_posts (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    last_editor_id UUID REFERENCES users (id),
    group_id TEXT NOT NULL,
    slug TEXT NOT NULL,
    language TEXT NOT NULL,
    featured BOOLEAN DEFAULT FALSE,
    status blog_status DEFAULT 'draft' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    published_at TIMESTAMPTZ,
    UNIQUE (slug, language)
);

-- METADATA TABLE
CREATE TABLE IF NOT EXISTS blog_metadata (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    blog_id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    image TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);

-- CONTENT TABLE
CREATE TABLE IF NOT EXISTS blog_content (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    blog_id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    read_time INTEGER DEFAULT 0,
    html TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);

-- STATISTICS TABLE
CREATE TABLE IF NOT EXISTS blog_stats (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    blog_id UUID NOT NULL REFERENCES blog_posts (id) ON DELETE CASCADE,
    views INTEGER DEFAULT 0,
    likes INTEGER DEFAULT 0,
    shares INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    last_viewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);

-- CATEGORIES TABLE
CREATE TABLE IF NOT EXISTS categories (
    name TEXT NOT NULL,
    value TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    PRIMARY KEY (value)
);

-- TAGS TABLE
CREATE TABLE IF NOT EXISTS tags (
    name TEXT NOT NULL,
    value TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    PRIMARY KEY (value)
);

-- BLOG-CATEGORY RELATIONSHIP TABLE
CREATE TABLE IF NOT EXISTS blog_categories (
    blog_id UUID REFERENCES blog_posts (id) ON DELETE CASCADE,
    category_value TEXT REFERENCES categories (value) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, category_value)
);

-- BLOG-TAG RELATIONSHIP TABLE
CREATE TABLE IF NOT EXISTS blog_tags (
    blog_id UUID REFERENCES blog_posts (id) ON DELETE CASCADE,
    tag_value TEXT REFERENCES tags (value) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, tag_value)
);

CREATE INDEX idx_blog_posts_group_id ON blog_posts (group_id);

CREATE INDEX idx_blog_posts_user_id ON blog_posts (user_id);

CREATE INDEX idx_blog_posts_status ON blog_posts (status);

CREATE INDEX idx_blog_posts_language ON blog_posts (language);

CREATE INDEX idx_blog_metadata_blog_id ON blog_metadata (blog_id);

CREATE INDEX idx_blog_content_blog_id ON blog_content (blog_id);

CREATE INDEX idx_blog_stats_blog_id ON blog_stats (blog_id);

CREATE INDEX idx_blog_stats_views ON blog_stats (views);

CREATE INDEX idx_blog_stats_likes ON blog_stats (likes);

CREATE INDEX idx_categories_name ON categories (name);

CREATE INDEX idx_tags_name ON tags (name);
