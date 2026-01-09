CREATE TABLE IF NOT EXISTS model_configs (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  id TEXT NOT NULL,
  type TEXT NOT NULL,
  config JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, id, type)
);

CREATE INDEX IF NOT EXISTS idx_model_configs_user_type ON model_configs(user_id, type);

CREATE TABLE IF NOT EXISTS tts_configs (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  id TEXT NOT NULL,
  provider TEXT NOT NULL,
  config JSONB NOT NULL,
  active BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, id)
);

CREATE INDEX IF NOT EXISTS idx_tts_configs_user_id ON tts_configs(user_id);

CREATE TABLE IF NOT EXISTS voice_categories (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  id TEXT NOT NULL,
  name TEXT NOT NULL,
  icon TEXT NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, id)
);

CREATE TABLE IF NOT EXISTS voices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT NOT NULL,
  is_preset BOOLEAN NOT NULL DEFAULT FALSE,
  audio_url TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  duration DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voices_user_id ON voices(user_id);

CREATE TABLE IF NOT EXISTS prompts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT NOT NULL,
  origin TEXT NOT NULL DEFAULT 'user',
  template TEXT NOT NULL,
  variables TEXT[] NOT NULL DEFAULT '{}',
  system_prompt TEXT,
  tags TEXT[] NOT NULL DEFAULT '{}',
  version TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prompts_user_id ON prompts(user_id);

CREATE TABLE IF NOT EXISTS prompt_selections (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  project_id TEXT NOT NULL,
  feature_key TEXT NOT NULL,
  selection JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, project_id, feature_key)
);
