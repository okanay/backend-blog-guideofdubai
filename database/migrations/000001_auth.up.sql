-- EXTENSIONS
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- USER HELPERS
CREATE TYPE user_status AS ENUM ('Active', 'Suspended', 'Deleted');

CREATE TYPE role AS ENUM ('User', 'Editor', 'Admin');

-- USER TABLE
CREATE TABLE IF NOT EXISTS users (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    role role DEFAULT 'User' NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    status user_status DEFAULT 'Active' NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    last_login TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- USER TABLE INDEXES
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);

-- REFRESH TOKEN TABLE
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID NOT NULL,
    user_email TEXT,
    user_username TEXT,
    token TEXT UNIQUE NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days',
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    last_used_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_reason TEXT,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- REFRESH TOKEN TABLE INDEXES
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);
