ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_fkey;
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_user_id_fkey;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_user_id_fkey;
ALTER TABLE model_configs DROP CONSTRAINT IF EXISTS model_configs_user_id_fkey;
ALTER TABLE tts_configs DROP CONSTRAINT IF EXISTS tts_configs_user_id_fkey;
ALTER TABLE voice_categories DROP CONSTRAINT IF EXISTS voice_categories_user_id_fkey;
ALTER TABLE voices DROP CONSTRAINT IF EXISTS voices_user_id_fkey;
ALTER TABLE prompts DROP CONSTRAINT IF EXISTS prompts_user_id_fkey;
ALTER TABLE prompt_selections DROP CONSTRAINT IF EXISTS prompt_selections_user_id_fkey;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_user_id_fkey;
ALTER TABLE usage_events DROP CONSTRAINT IF EXISTS usage_events_user_id_fkey;

DROP TABLE IF EXISTS users;
