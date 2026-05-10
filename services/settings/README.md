# Settings Service

## Responsibility

App settings storage/reads, validation, admin APIs and settings-related test flows.

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
- `services/base` (shared migrations registration helpers)

## Public flows

- Upsert settings.
- Validate settings (email/filesystem tests).

## Migration notes

Domain migrations live in `services/settings/migrations/` and are registered via `services/settings/migrations.Register`.
