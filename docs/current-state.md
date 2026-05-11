# Servicebase — Current Project State

_Last updated: 2026-05-11_

## Vision

Servicebase is evolving from a PocketBase fork into a full platform runtime:

- PocketBase-derived core runtime
- Custom platform APIs
- Realtime-first architecture
- Collections Studio
- Agent/runtime orchestration
- Internal platform UX

The long-term direction is closer to:

- Supabase Studio
- Convex dashboard
- LangGraph Studio
- Retool-like internal platform

rather than a direct PocketBase admin clone.

---

# Architecture

## Backend

Backend is based on a heavily modified PocketBase core.

Current direction:

```txt
servicebase
 ├── core (PocketBase-derived runtime)
 ├── platform APIs
 ├── collections studio APIs
 ├── realtime layer
 ├── auth/session layer
 ├── agent runtime
 ├── MCP/runtime integrations
 └── internal platform services
```

### Current backend modules

Implemented and registered:

- backup
- collections
- cron
- files
- health
- logs
- realtime
- settings
- updater

Collections API is now connected into the main router.

---

## Frontend

Frontend originated from PocketPaw.

We are NOT trying to preserve full PocketPaw backend compatibility.

Instead:

- frontend UX/components are reused
- backend contracts are adapted to Servicebase
- compatibility layers are minimized
- architecture is being simplified

Main frontend direction:

- app shell UX
- platform-oriented navigation
- realtime interactions
- collections-first workflow
- admin + runtime unified UI

---

# Auth Migration

## Goal

Replace PocketPaw OAuth/bootstrap assumptions with native Servicebase auth.

### Current direction

- email/password auth
- JWT-based auth
- simplified token lifecycle
- backend-native auth APIs
- remove unnecessary OAuth complexity

### Current status

Implemented foundation:

- auth bootstrap
- token loading
- auth header wiring
- authenticated API requests
- admin access middleware

Still planned:

- session APIs
- logout endpoint
- token regeneration
- refresh handling
- auth realtime synchronization

---

# Collections Studio

## Goal

Recreate the PocketBase collections experience using the Servicebase app shell and architecture.

Reference inspirations:

- PocketBase admin
- pocket-admin
- Supabase Studio

But visually and architecturally integrated into Servicebase.

---

## Implemented

### Frontend

Added:

- `/collections`
- `/collections/[collection]`
- collections API client
- collection schema viewer
- records table
- sidebar integration
- collection search/filtering
- collection type badges

### Backend

Added/connected:

- `GET /api/collections`
- `GET /api/collections/:collection`
- `GET /api/collections/:collection/records`

Collections API is now registered in `app/routes.go`.

---

# Current Frontend/Backend Gap

Frontend still contains many PocketPaw-era APIs that do not yet exist in backend.

## Major missing backend areas

### Chat/runtime

- `/chat`
- `/chat/stream`
- `/chat/stop`

### Sessions

- `/sessions`
- `/sessions/history`
- `/sessions/search`
- `/sessions/export`

### Skills

- `/skills/*`

### Memory

- `/memory/*`

### Identity

- `/identity`

### MCP

- `/mcp/*`

### Backends

- `/backends/*`

### Metrics

- `/metrics/*`

### Kits

- `/kits/*`

### Mission Control

- `/mission-control/*`

### Deep Work

- `/deep-work/*`

---

# Current Priorities

## Priority 1

Auth + Collections foundation.

This is the base platform layer.

## Priority 2

Collections Studio improvements:

- record CRUD
- collection CRUD
- schema editor
- rules editor
- realtime updates
- relation explorer

## Priority 3

Runtime/session layer:

- chat
- sessions
- memory
- skills
- MCP

## Priority 4

Advanced platform modules:

- mission control
- deep work
- kits
- runtime orchestration

---

# Important Architectural Decisions

## We are NOT building:

- a direct PocketBase clone
- a thin PocketPaw compatibility wrapper
- a generic admin dashboard

## We ARE building:

- a platform runtime
- a collections-first internal platform
- a realtime-oriented developer environment
- a unified runtime + admin experience

---

# Current Working Branch

Primary active branch:

```txt
collections-studio-foundation
```

Main contains:

- auth foundation
- initial backend cleanup

Collections Studio work is currently being developed incrementally on top.
