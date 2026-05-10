# Cron Service

## Responsibility

Dynamic cron CRUD, scheduler runtime, webhook execution and execution analytics.

## Structure

- `service.go` — cron use-cases: create/update/delete/list/test/clone.
- `config.go` — cron defaults and config helpers.
- `apis/` — project API routes under `/crons`.
- `forms/` — reserved for submit-flow/form objects.
- `models/` — `Cron` and `CronExecution` persistence models.
- `queries/` — reserved for read-side/query helpers.
- `migrations/` — cron-owned DB schema updates.
- `runtime.go` — logger/runtime helpers.
- `scheduler.go` / `execution.go` — mutable scheduler and execution runtime.
- `store.go` — DAO helpers used by service/runtime.

## Dependencies

- `core.App`
- `tools/cron`
- `services/cron/models`

## Public flows

- Create/update cron jobs.
- Schedule, pause/resume and execute jobs.
- Persist execution history.
- Expose stats/metrics endpoints for cron executions.

## Migration notes

Cron owns the `cron_executions` table and schema completion for the legacy `crons` table. The older combined `crons + newsletter_settings` migration remains in `migrations/dbase` for history; new cron-owned changes live in `services/cron/migrations` and are registered via `migrations/dbase/cron_module.go`.
