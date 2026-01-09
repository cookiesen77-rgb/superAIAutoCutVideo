# Cloud API (Go)

Go-based API service for the cloud-only Electron app.

## Environment

Required:
- `DATABASE_URL` (local Postgres DSN for projects/jobs/tts/billing)
- `USER_DATABASE_URL` (Supabase Postgres DSN for users table only)
- `JWT_SECRET`
- `S3_BUCKET`

Optional:
- `SERVER_ADDR` (default `:8080`)
- `AWS_REGION` (default `us-east-1`)
- `ACCESS_TOKEN_TTL` (default `15m`)
- `REFRESH_TOKEN_TTL` (default `720h`)
- `CORS_ALLOWED_ORIGINS` (comma-separated, default `http://localhost:1420,null`)

## Run

```bash
cd server
go run ./cmd/api
```

## Bootstrap Admin (Supabase users DB)

```bash
cd server
USER_DATABASE_URL="..." ADMIN_EMAIL="3327390662@qq.com" ADMIN_PASSWORD="314394" \
  go run ./cmd/bootstrap-admin
```

## Migrations

Local DB: apply `server/migrations/001_init.sql` through `server/migrations/006_local_user_refs.sql`.

Supabase users DB: apply `server/migrations/supabase_users.sql`.
