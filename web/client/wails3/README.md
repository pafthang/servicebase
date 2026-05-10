# PocketPaw Wails 3 Migration

This folder contains the Wails 3 target for `web/client`.

## Goals

- Keep all existing UI functionality from the Tauri version.
- Migrate backend commands incrementally without breaking current flows.
- Preserve command names used in the frontend (`invoke("...")`) for compatibility.

## Current status

- Wails entrypoint scaffold is in place.
- Frontend now has a runtime adapter at `src/lib/native/runtime.ts`.
- `Invoke(command, args)` compatibility entrypoint is available in Go.
- Implemented baseline filesystem commands:
  - `fs_read_dir`, `fs_read_file_text`, `fs_write_file`, `fs_delete`, `fs_rename`
  - `fs_stat`, `fs_stat_extended`, `fs_create_dir`, `fs_exists`
  - `fs_get_default_dirs`, `fs_read_file_head`, `fs_read_file_base64`
  - `fs_copy_file`, `fs_copy_dir`, `fs_search_recursive`
- Implemented `platform` command.

## Not yet implemented

- Window orchestration (`toggle_side_panel`, quick ask window, attach mode, vibrancy).
- OAuth/login bridge commands.
- Updater/autostart/global shortcuts/tray parity.
- FS watch stream (`fs_watch` / `fs_unwatch`) and thumbnail pipeline.

## Next migration batch

1. Add events bridge parity (`EventsEmit` / `EventsOn`) in Wails and wire `fs_watch`.
2. Port window/sidepanel commands.
3. Port onboarding/backend install commands.
4. Port settings commands (`set_vibrancy_theme`) and updater/process actions.
