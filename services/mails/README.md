# Mails Service

## Responsibility

Canonical home for application mail flows and Gmail mailbox integration: OAuth callback, token storage, label initialization, sync orchestration, message persistence and record-facing transactional emails.

## Structure

- `service.go` — descriptor and service-level contract.
- `config.go` — Google/Gmail configuration resolution.
- `client.go`, `label.go`, `message.go` — Gmail client, labels and message sync flows.
- `record.go` — PocketBase record-facing mail flows.
- `templates/` — HTML mail templates/layouts.
- `models/` — mail sync/message persistence models.
- `apis/` — project mail HTTP bindings.
- `queries/` — reusable DAO/query helpers.
- `forms/` — request/submit DTOs.
- `migrations/` — module-owned mail schema.
- `runtime.go` — placeholder for future background sync runtime/caches.

## Dependencies

- `core.App`
- `services/base/models` for OAuth tokens
- `tools/mailer`, `tools/tokens`
- Google OAuth/Gmail API client packages

## Public flows

- Compose and send record-facing emails.
- Render HTML templates/layouts.
- Start Gmail OAuth flow and persist Gmail token.
- Initialize labels and sync Gmail messages.
- Expose sync status and inactive syncs for re-auth flows.

## Migration notes

Mail sync and mail message tables are registered from `services/mails/migrations` through `migrations/dbase/mails_module.go`.
