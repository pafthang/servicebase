# Health Service

## Responsibility

Lightweight runtime health checks for the public/admin API surface.

## Structure

- `service.go` — health check orchestration.
- `runtime.go` — mutable health runtime placeholder.
- `config.go` — service configuration defaults.
- `apis/` — `/health` HTTP bindings.
- `forms/` — forms placeholder.
- `queries/` — read-side helper placeholder.
- `migrations/` — health-owned migration registration hook.

## Dependencies

- `core.App`

## Public flows

- `HEAD /health`
- `GET /health`

## Migration notes

Health is runtime-only right now. `services/health/migrations` intentionally registers no schema migrations yet, but exists as the local home for future health persistence.
