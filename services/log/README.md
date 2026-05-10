# Log Service

## Responsibility

Canonical home for application and project log flows: system log read-side, project log ingest, batching, retention cleanup and aggregate querying.

## Structure

- `service.go` — orchestration and public service methods.
- `config.go` — logging runtime configuration.
- `runtime.go` / `worker.go` / `store.go` / `jobs.go` — mutable runtime, batching and cleanup internals.
- `models/` — `_logs`, `logging_projects` and app `logs` persistence models.
- `apis/` — module-owned HTTP bindings for admin logs and project log ingest.
- `queries/` — read-side query helpers.
- `forms/` — request/submit DTOs.
- `migrations/` — module-owned app/project logging schema.

## Dependencies

- `core.App`
- `services/base/models` for developer tokens
- `tools/search` for system log querying

## Public flows

- Query PocketBase internal request logs.
- Create logging projects.
- Ingest single or batched app logs with project API keys.
- Flush buffered logs to storage.
- Cleanup project logs by retention window.

## Migration notes

- PocketBase internal `_logs` remains core-owned by `migrations/logs`.
- Product/app logging tables are registered from `services/log/migrations` through `migrations/dbase/log_module.go`.
