# Team Service

## Responsibility

Team management, membership grants and access-control helpers.

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

- Team create/update/delete.
- Membership grant/revoke.

## Migration notes

Domain migrations live in `services/team/migrations/` and are registered via `services/team/migrations.Register`.
