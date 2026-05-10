# Realtime Service

## Responsibility

Realtime subscriptions, SSE connect/subscribe API, auth-model sync and record broadcast helpers.

## Structure

- `service.go` — module descriptor and service constructor.
- `config.go` — realtime service defaults.
- `apis/` — realtime `/realtime` connect/subscribe endpoints and event hook binding.
- `forms/` — realtime subscribe form and validation.
- `queries/` — read-side query helpers.
- `migrations/` — realtime migration registration hook.
- `runtime.go` — auth model and subscription runtime helpers.

## Dependencies

- `core.App`
- `subscriptions` broker
- record access/search helpers

## Public flows

- Open realtime SSE connections.
- Update client subscriptions.
- Broadcast record create/update/delete events.
- Sync or unregister clients when auth records change.

## Migration notes

Realtime is currently broker/event driven and owns no persistent schema.
