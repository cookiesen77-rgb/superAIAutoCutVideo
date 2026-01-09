# Cloud API Contract (Draft)

This is the target contract for the Electron + cloud backend migration.

## Auth

- `POST /api/auth/register`
  - `{ "email": "user@example.com", "password": "..." }`
  - returns `{ access_token, refresh_token, expires_in, token_type }`
- `POST /api/auth/login`
  - same body/response as register
- `POST /api/auth/refresh`
  - `{ "refresh_token": "..." }`
  - returns new token pair
- `POST /api/auth/logout`
  - `{ "refresh_token": "..." }`
- `GET /api/auth/me`
  - `Authorization: Bearer <access_token>`

## Admin

- `GET /api/admin/users`
- `GET /api/admin/users/{userId}`
- `PATCH /api/admin/users/{userId}` `{ "email"?, "is_admin"?, "status"?: "active|disabled" }`
- `POST /api/admin/users/{userId}/reset-password` `{ "password": "..." }`
- `GET /api/admin/users/{userId}/usage`
- `POST /api/admin/users/{userId}/plan` `{ "plan_id": "free|..." }`

## Uploads (S3 presign)

- `POST /api/uploads/init`
  - `Authorization: Bearer <access_token>`
  - `{ "filename": "video.mp4", "content_type": "video/mp4" }`
  - returns `{ upload_url, file_url, expires_in }`

## Jobs

- `POST /api/jobs`
  - `Authorization: Bearer <access_token>`
  - `{ "type": "merge|asr|tts|script|normalize", "input_url": "https://...", "payload": { ... } }`
  - returns job object with `status=queued` and `progress`
- `GET /api/jobs`
- `GET /api/jobs/{jobID}`
  - returns `{ id, type, status, progress, input_url, output_url?, result?, error? }`

Job payload examples:
- `script`: `{ "project_id": "...", "subtitle_url": "...", "drama_name": "...", "plot_analysis": "..." }`
- `tts`: `{ "provider": "aws_polly", "config_id": "aws_polly_default", "voice_id": "Zhiyu", "text": "..." }`

## Models (Settings)

- `GET /api/models/video-analysis/configs`
- `PUT /api/models/video-analysis/configs/{configId}`
- `POST /api/models/video-analysis/configs/{configId}/activate`
- `POST /api/models/video-analysis/test/{configId}`
- `GET /api/models/content-generation/configs`
- `PUT /api/models/content-generation/configs/{configId}`
- `POST /api/models/content-generation/test/{configId}`

Notes:
- `config` fields: `provider`, `api_key`, `base_url`, `model_name`, `extra_params`.
- OpenAI-compatible endpoints are supported via `base_url` (e.g. OpenAI, OpenRouter, Qwen, DeepSeek).
- Gemini uses `provider=gemini` and `base_url=https://generativelanguage.googleapis.com/v1beta`.

## TTS (Settings)

- `GET /api/tts/engines`
- `GET /api/tts/voices?provider=edge_tts`
- `GET /api/tts/configs`
- `PATCH /api/tts/configs/{configId}`
- `POST /api/tts/configs/{configId}/activate`
- `POST /api/tts/configs/{configId}/test`
- `POST /api/tts/voices/{voiceId}/preview`
- `GET /api/tts/emotions`
- `GET /api/tts/index-tts/status`
- `POST /api/tts/index-tts/preload`
- `POST /api/tts/index-tts/test`

Notes:
- `provider=aws_polly` 支持云端语音合成，需要 `secret_id/secret_key`（AWS Access Key/Secret Key）。

## Voices (Custom)

- `GET /api/voices/categories`
- `POST /api/voices/categories`
- `PATCH /api/voices/categories/{categoryId}`
- `DELETE /api/voices/categories/{categoryId}`
- `GET /api/voices?category=...`
- `GET /api/voices/{voiceId}`
- `POST /api/voices/upload`
- `POST /api/voices/record`
- `PATCH /api/voices/{voiceId}`
- `DELETE /api/voices/{voiceId}`
- `GET /api/voices/{voiceId}/audio`
- `POST /api/voices/{voiceId}/preview`

## Prompts

- `GET /api/prompts/categories`
- `GET /api/prompts?category=...`
- `POST /api/prompts`
- `GET /api/prompts/{id}`
- `PUT /api/prompts/{id}`
- `DELETE /api/prompts/{id}`
- `POST /api/prompts/{id}/render-preview`
- `POST /api/prompts/validate`
- `POST /api/prompts/projects/{projectId}/prompts/select`
- `GET /api/prompts/projects/{projectId}/prompts/selection`

## Compatibility Notes

Legacy `/api/projects/*` endpoints remain for backward compatibility with existing UI flows.
Legacy `/api/video/*` and `/api/task/*` endpoints return stub task status for older clients.

## Billing

- `GET /api/billing/plan`
- `GET /api/billing/usage`

Usage categories:
- `llm_chars`: 文案/脚本生成使用的字符量
- `tts_chars`: 语音合成使用的字符量
