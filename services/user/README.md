# User Service

## Responsibility

User authentication workflows, auth-provider discovery and user-scoped external auth helpers.

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

- User signup/login/logout.
- Password reset and email verification.
- External auth provider binding.

## Migration notes

Domain migrations live in `services/user/migrations/` and are registered via `services/user/migrations.Register`.
