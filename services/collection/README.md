# Collection Service

## Responsibility

Collection listing/lookup and admin collection mutation flows, plus colocated transport helpers.

## Structure

- `service.go`
- `config.go`
- `forms/`
- `apis/`
- `queries/`
- `migrations/`
- `runtime.go`

## Dependencies

- `core.App`
- `services/base` (shared forms/migrations helpers)

## Public flows

- List collections.
- Find collection by name/ID.
- Upsert/delete/import collections.

## Migration notes

Domain migrations live in `services/collection/migrations/` and are registered via `services/collection/migrations.Register`.
