# Updater Service

## Responsibility

Infra service for self-update command orchestration and GitHub release update execution.

## Structure

- `service.go` — descriptor, service construction and command binding entrypoint.
- `config.go` — GitHub repository/update client configuration.
- `runtime.go` — update command, release download, extraction and executable replacement flow.
- `release.go` — GitHub release/asset DTOs and asset selection helpers.
- `forms/` — reserved for future API/submit DTOs.
- `apis/` — reserved HTTP binding hook; updater is currently CLI-only.
- `queries/` — reserved read-side helpers; no database query layer yet.
- `migrations/` — module migration registration hook; no schema yet.

## Dependencies

- `core.App`
- `cobra`
- GitHub releases API
- `tools/archive`

## Public flows

- Bind the `update` CLI command.
- Fetch latest GitHub release metadata.
- Download and extract the matching release archive.
- Replace the current executable.
- Optionally create a `pb_data` backup before update finalization.

## Migration notes

Updater is infra/CLI tooling and currently owns no database tables. `services/updater/migrations` exists only as the canonical module registration point.
