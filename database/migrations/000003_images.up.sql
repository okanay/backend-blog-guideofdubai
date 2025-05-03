-- IMAGES TABLE
CREATE TABLE IF NOT EXISTS images (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    url TEXT NOT NULL,
    filename TEXT NOT NULL,
    alt_text TEXT,
    file_type TEXT NOT NULL,
    size_in_bytes INTEGER NOT NULL,
    width INTEGER,
    height INTEGER,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    UNIQUE (url)
);

-- UPLOAD SIGNATURES TABLE
CREATE TABLE IF NOT EXISTS upload_signatures (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    image_id UUID REFERENCES images (id) ON DELETE CASCADE,
    presigned_url TEXT NOT NULL,
    upload_url TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);
