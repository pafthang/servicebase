# File Service

## Responsibility

File domain helpers: token issuance, protected file download resolution, thumbnail generation coordination, quota checks, sanitization, zip/export flows and file-owned schema.

## Structure

- `service.go` — token and auth-record resolution service.
- `zip.go` — zip/export stream builder.
- `quota.go` — user storage quota helpers.
- `sanitization.go` — filename/path sanitization helpers.
- `store.go` — persistence helpers for file/folder/quota records.
- `apis/` — core `/files` endpoints and project file/export endpoints.
- `models/` — file/folder/quota models.
- `forms/` — reserved for file submit-flow forms.
- `queries/` — reserved for file read-side query helpers.
- `migrations/` — file schema registration.
- `runtime.go` — logger/runtime helpers.

## Public flows

- Issue file access tokens.
- Resolve auth records from protected file tokens.
- Serve original files and generated thumbnails.
- Stream selected folders/files as ZIP.
- Track file/folder/quota persistence.

## Migrations

File-owned schema lives in `services/file/migrations` and is registered by `migrations/dbase/file_module.go`.
