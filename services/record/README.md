# Record Service

## Responsibility

Record listing/lookups/mutations and record-scoped helper flows.

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

- Find/list records.
- Record upsert/delete flows.

## Migration notes

Domain migrations live in `services/record/migrations/` and are registered via `services/record/migrations.Register`.
