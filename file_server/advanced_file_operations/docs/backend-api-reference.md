# Backend API Reference & LLM Notes

This document explains how the Go backend exposes its HTTP API, the authentication expectations for each route group, and the required request payloads. It is intended both for human developers and for downstream language models so they can generate correct client code without guessing the contract.

## Router Topology

All HTTP routes are mounted beneath `/api` by `api.RegisterRoutes`:

- `/api/process` – optional authentication (`MaybeAuth`) for limited processing operations.
- `/api/metadata` – optional authentication (`MaybeAuth`) for metadata extraction.
- `/api/*` – everything else requires authentication via `AuthMiddleware.RequireAuth`.
- `/api/public` – currently unused placeholder for future public endpoints.

The orchestrator (`operations.NewOrchestrator`) wires feature-specific handlers (bulk processing, rename utilities, metadata write/remove, optimizer) into the authenticated group. Even if an individual handler allows anonymous users internally, the route-level middleware still enforces auth.

### Authentication Helpers

The middleware categorizes endpoints into three buckets:

| Bucket | Behavior | Typical Use |
| --- | --- | --- |
| Public | No auth header required. | `/health`, future `/api/public/*`. |
| MaybeAuth | Token is added when available, but anonymous requests are accepted. | `/api/process/limited`, `/api/metadata/extract`. |
| RequireAuth | Missing/invalid token yields `401`. | `/api/bulk/*`, `/api/processing/*`, `/api/metadata-*`, `/api/optimizer/*`, `/api/renamer/*`, `/api/profile`. |

Front-end clients must mirror this behavior (see `api/apiService.js`).

## Endpoint Catalogue

### Limited Processing (`/api/process/limited`)

- **Method:** `POST`
- **Auth:** Optional (MaybeAuth)
- **Content-Type:** `multipart/form-data`
- **Fields:**
  - `files`: One or more file parts.
  - Optional flags: `renameFiles`, `removeMetadata`, `optimizeFiles` (`"true"` to enable).
  - Optional `pattern` string and `maxFileSize` integer (bytes). Parsed by `limited_file_processing` service.
- **Response:**
  ```json
  {
    "message": "Files processed successfully (limited mode)",
    "results": [
      {
        "filename": "example.jpg",
        "newName": "...",            // when renaming enabled
        "contentType": "image/jpeg",
        "action": "renamed_metadata_removed_optimized",
        "success": true,
        "content": "<base64>"
      }
    ]
  }
  ```
- **Notes:** Designed for small synchronous jobs; failures abort the entire request.

### Bulk Processing (`/api/bulk/*`)

| Path | Method | Auth | Description |
| --- | --- | --- | --- |
| `/api/bulk/process` | `POST` | Required | Submit an asynchronous job. |
| `/api/bulk/user/stats` | `GET` | Required | Summary stats for the authenticated user. |
| `/api/bulk/history` | `GET` | Required | Recent jobs (default limit 10, `?limit=` up to 100). |
| `/api/bulk/download/{jobID}` | `GET` | Required + paying plan | Streams a ZIP archive. |
| `/api/bulk/download/{jobID}/{filename}` | `GET` | Required | Streams a single processed file. |

#### `/api/bulk/process`

- **Content-Type:** `multipart/form-data`
- **Fields:**
  - `files`: Multiple file parts.
  - Optional Boolean flags `renameFiles`, `removeMetadata`, `optimizeFiles` (`"true"`).
  - Optional `pattern` string (rename pattern ID).
  - Optional `maxFileSize` string/number (bytes). Parsed to `ProcessingOptions.MaxFileSize`.
- **Response Shape:**
  ```json
  {
    "results": [
      {
        "success": true,
        "action": "optimized",
        "data": {
          "newName": "image-optimized.jpg",
          "encodedContent": "<base64>",
          "contentType": "image/jpeg"
        }
      }
    ],
    "jobId": "bulk_job_123",
    "downloadUrl": "/api/bulk/download/bulk_job_123" // present for paying users
  }
  ```
- **Important:** The handler re-reads each `multipart.File` completely (`io.ReadAll`) and compares byte counts with `header.Size`. The client must not send truncated streams.

#### Job Monitoring (`/api/processing/*`)

| Path | Method | Auth | Description |
| --- | --- | --- | --- |
| `/api/processing/stats` | `GET` | Required | System-level stats. |
| `/api/processing/jobs/active` | `GET` | Required | Active jobs for the user. |
| `/api/processing/jobs/recent` | `GET` | Required | Recently finished jobs. |
| `/api/processing/bulk/status/{jobID}` | `GET` | Required | Per-job status + progress estimate. |

### Metadata Features

| Feature | Paths | Auth | Notes |
| --- | --- | --- | --- |
| Extraction | `POST /api/metadata/extract` | MaybeAuth | Multipart `files` (single file expected). Returns `{ success, filename, metadata }`. |
| Removal | `POST /api/metadata/images`, `POST /api/metadata/pdf` | Required | Multipart `files`. Returns Base64 encoded processed content. |
| Renaming Patterns | `GET /api/metadata-rename/patterns` | Required | Returns array of pattern descriptors (`id`, `name`, `description`, `example`). |
| Writer Fields | `GET /api/metadata-write/fields` | Required | Returns array of writable metadata field descriptors. |

### Optimizer (`/api/optimizer/*`)

| Path | Method | Auth | Description |
| --- | --- | --- | --- |
| `/api/optimizer/images` | `POST` | Required | Multipart image file. Returns optimized content (Base64).
| `/api/optimizer/pdf` | `POST` | Required | Multipart PDF file. |
| `/api/optimizer/files` | `POST` | Required | Multipart generic files; allows `text/plain`, `application/json`, etc. |

All optimizer endpoints validate content type against the per-user config and ensure post-processing data is non-empty.

### Rename Options (`/api/renamer/*`)

| Path | Method | Auth | Response |
| --- | --- | --- | --- |
| `/api/renamer/patterns` | `GET` | Required | `{ "patterns": ["default", "timestamp", ...] }`
| `/api/renamer/namers` | `GET` | Required | `{ "namers": ["basic", "timestamp", "hash"] }`

> **Note:** Although the handler itself does not enforce auth, the router mounts it under the authenticated group. Clients must send a valid bearer token.

## Request Construction Checklist (for LLMs)

1. **Always set `Content-Type: multipart/form-data`** when uploading files. Let the browser/HTTP library manage the boundary—do not override it manually.
2. **Use Boolean strings** (`"true"`) for checkbox-style options (`renameFiles`, `removeMetadata`, `optimizeFiles`). Absent or any other value defaults to `false`.
3. **Include `pattern` only when renaming is enabled.** Backend defaults to `"default"` when omitted.
4. **Respect file size limits:**
   - Limited mode uses `common.MaxPayload10MB` and per-user config.
   - Bulk mode allows larger payloads (`32MB` per request) but individual files must satisfy `ProcessingOptions.MaxFileSize` (default 50MB).
5. **Authenticate bulk and feature endpoints** by attaching a Firebase bearer token. Without it the middleware will reject the request before hitting handlers.
6. **Do not attempt to create users explicitly.** The auth middleware calls `user.Service.EnsureUserExists` on every authenticated request; clients just need to supply a valid token.
7. **When downloading results**, send `Accept: application/zip` (bulk) or accept the default. The response is binary; configure the client (`responseType: 'blob'`) accordingly.
8. **Job polling cadence:** backend is optimized for ~5s poll intervals; aggressive polling risks rate limiting by `security.UserRateLimitingMiddleware`.

## Troubleshooting 404s

- `POST /api/process/bulk` will return `404` because the server only exposes `/api/bulk/process` via the orchestrator. Update clients to the `/api/bulk` prefix.
- If CORS preflight succeeds (`OPTIONS 200`) but the main request fails, verify auth headers and endpoint spelling against this document.

## Related Files

- `operations/bulk_file_processing/handler.go` – multipart parsing and job orchestration.
- `operations/monitoring_and_statistics/handler.go` – job status endpoints.
- `operations/features/*/handler.go` – per-feature route definitions.
- `api/api.go` – top-level router with middleware wiring.
- Front-end constants: `src/api/endpoints.js`, `src/api/apiService.js`.

Keep this document updated whenever endpoints move or request bodies change so automated tooling (including LLMs) stays accurate.
