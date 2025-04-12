-- USER HELPERS
CREATE TYPE user_status AS ENUM ('Active', 'Suspended', 'Deleted');
CREATE TYPE role AS ENUM ('User', 'Editor', 'Admin');

-- USER TABLE
CREATE TABLE IF NOT EXISTS users (
  id bigint primary key generated always as identity,
  unique_id TEXT DEFAULT ('1' || substring(md5(random()::text) from 1 for 8)) UNIQUE,
  email TEXT UNIQUE NOT NULL,
  username TEXT NOT NULL UNIQUE,
  hashed_password TEXT NOT NULL,
  membership role DEFAULT 'User' NOT NULL,
  email_verified BOOLEAN DEFAULT FALSE,
  status user_status DEFAULT 'Active' NOT NULL,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  last_login TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- USER TABLE INDEXES
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

-- REFRESH TOKEN TABLE
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id bigint primary key generated always as identity,
  user_id bigint NOT NULL,
  user_email TEXT,
  user_username TEXT,
  token TEXT UNIQUE NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  expires_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days',
  created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  last_used_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  is_revoked BOOLEAN DEFAULT FALSE,
  revoked_reason TEXT,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- REFRESH TOKEN TABLE INDEXES
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id_is_revoked ON refresh_tokens (user_id, is_revoked);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_email ON refresh_tokens (user_email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_username ON refresh_tokens (user_username);
