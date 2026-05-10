# Migrate Service

## Responsibility

Migration CLI command orchestration, migration file generation and collection automigration hooks. This is an infra/tooling service, not a business-domain feature module.

## Structure

- `service.go` — service descriptor, factory and command binding
- `config.go` — migrate dir/language/automigrate config
- `command.go` — CLI command implementation
- `templates.go` — generated migration templates
- `runtime.go` — migration runtime helpers
- `apis/` — reserved, no default HTTP API
- `forms/` — reserved validation DTO package
- `queries/` — reserved read-side helper package
- `migrations/` — intentionally empty for this tooling service

## Dependencies

- `core.App`
- `cobra`
- shared `migrations/dbase` registry

## Public flows

- Bind the `migrate` CLI command.
- Generate JS/Go migration stubs.
- Generate collection snapshots.
- Wire automigration hooks when enabled.

## Migration notes

The migrate service should not own product schema migrations. Domain migrations live in `services/<service>/migrations` and are registered by `migrations/dbase/*_module.go`.
