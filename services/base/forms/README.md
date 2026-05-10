# Base Forms

This package is the canonical entrypoint for platform-level PocketBase forms.

It currently re-exports the legacy top-level `forms` package so modules can
migrate without a large atomic rename. New service code should prefer importing:

`github.com/pocketbase/pocketbase/services/base/forms`

over:

`github.com/pocketbase/pocketbase/forms`

The legacy package remains as a compatibility layer until the migration is
complete.
