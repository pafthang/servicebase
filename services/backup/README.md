# Backup Service

## Responsibility

Backup listing, creation, upload, restore, download authorization and deletion orchestration.

## Structure

- `service.go` — service descriptor and constructor.
- `config.go` — small service config/defaults boundary.
- `forms/` — create/upload validation and submit flows.
- `apis/` — `/api/backups` route binding and HTTP handlers.
- `models/` — backup read-side DTOs returned by the API.
- `queries/` — reserved for future backup metadata queries.
- `migrations/` — explicit backup migration pack boundary. No DB schema is registered yet.
- `runtime.go` — filesystem/runtime operations and active backup state checks.

## Dependencies

- `core.App` for backup filesystem, app backup/restore operations and app store state.
- `services/base/forms` for submit interceptors.
- `services/file` + `services/team` for download token and admin-team authorization checks.
- `tools/filesystem` and uploaded-file helpers.

## Public flows

- `GET /api/backups` — list backup archives.
- `POST /api/backups` — create a new backup archive.
- `POST /api/backups/upload` — upload a backup archive.
- `GET /api/backups/:key?token=...` — download an archive after file-token authorization.
- `POST /api/backups/:key/restore` — start asynchronous restore.
- `DELETE /api/backups/:key` — delete an archive when it is not actively used.

## Migration notes

Backups currently live in the configured backups filesystem, so the service does not need backup-owned tables. `migrations/register.go` intentionally registers no migrations for now, but keeps the domain migration boundary ready for future backup metadata/audit tables.
