# Services

This directory mixes a few different service styles today:

- simple app-bound services that wrap `core.App`
- stateful runtime services with background workers or caches
- integration services with external clients
- compatibility layers that still rely on package-level runtime state

To make new code and gradual refactors more consistent, the preferred baseline is:

1. Put shared app access behind `services/base`.
2. Expose a package-level `Descriptor` using `base.Descriptor`.
3. Keep service config in a dedicated `config.go` or `config/` package.
4. Prefer `app.Settings()` as the primary source of persisted configuration.
5. Use env vars as fallback/bootstrap overrides, not as the only long-term source.
6. Keep constructors explicit about dependencies.
7. Limit package-global runtime state to compatibility helpers and adapters.

Recommended service shape:

```go
package example

import (
    "github.com/pafthang/servicebase/core"
    servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
    Name:    "example",
    Purpose: "Short statement of responsibility.",
    Dependencies: []string{
        "core.App",
        "optional collaborators",
    },
    RuntimeState: []string{
        "List mutable state only when it exists",
    },
    Operations: []string{
        "List",
        "FindByID",
        "Delete",
    },
}

type Service struct {
    servicebase.Service
}

func New(app core.App) *Service {
    return &Service{
        Service: servicebase.NewWithApp(app),
    }
}
```

Recommended config shape:

```go
package example

type Config struct {
    Enabled bool
    BaseURL string
    Timeout time.Duration
}

func DefaultConfig() Config {
    return Config{
        Enabled: false,
        BaseURL: "https://example.com",
        Timeout: 10 * time.Second,
    }
}

func ConfigFromSettings(app core.App) Config {
    cfg := DefaultConfig()

    if app == nil || app.Settings() == nil {
        return cfg
    }

    // Map from app settings here.
    return cfg
}
```

Field guidance:

- `Purpose`: one sentence, focused on business responsibility
- `Dependencies`: durable collaborators the service relies on
- `RuntimeState`: caches, workers, background loops, package globals, mutexes
- `Operations`: the main public entrypoints worth discovering quickly

Config guidance:

- Prefer `settings -> env -> default` precedence for long-lived service config
- Use `services/base/config` for small shared parsing helpers
- Keep secrets in settings only when they already participate in redact/validate flows
- If a service has no durable user-managed config yet, keep env fallback and add settings later

This is intentionally small. The goal is a stable shape for discovery and
incremental cleanup, not a new framework around every service.
