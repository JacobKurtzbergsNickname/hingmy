# CLAUDE.md

This file provides guidance for AI assistants working with the **hingmy** codebase.

## Project Overview

**hingmy** is a Scottish-themed CLI todo manager built in Go. It uses Cobra for the CLI framework, SQLite for storage, and SQLC for type-safe SQL query generation. The application has a fun personality — Scottish accent and humor are intentional throughout the UI strings.

## Repository Structure

```
hingmy/
├── cmd/                    # Cobra CLI command definitions
│   ├── root.go             # Root command: welcome animation + DB init + interactive mode
│   ├── interactive.go      # Full interactive TUI (RunInteractiveMode)
│   ├── create.go           # `hingmy create` — create a new todo
│   ├── read.go             # `hingmy read`   — list active todos
│   ├── update.go           # `hingmy update` — update a todo interactively
│   └── delete.go           # `hingmy delete` — delete a todo interactively
├── database/               # Database layer
│   ├── init.db.go          # DB file creation and connection setup
│   ├── migration.db.go     # Manual table creation / migration logic
│   ├── db_utils.go         # Shared DB utilities
│   ├── accessor.db.go      # Generic DB accessor pattern
│   ├── accessor.todos.go   # Todo-specific DB access methods
│   └── sqlc/               # Auto-generated code — DO NOT edit by hand
├── models/                 # SQL schema definitions (source of truth for tables)
│   ├── todos.sql
│   ├── tags.sql
│   ├── notes.sql
│   └── tag_entities.sql
├── queries/                # SQLC query definitions (source of truth for queries)
│   ├── todos.sql
│   ├── tags.sql
│   ├── notes.sql
│   ├── tag_entities.sql
│   ├── schema.sql
│   └── drop_tables.sql
├── go.mod                  # Go module: "hingmy", requires Go 1.21+
├── go.sum
├── sqlc.yaml               # SQLC configuration
└── README.md
```

## Development Commands

```bash
# Run the app
go run . <command>

# Build
go build -o hingmy

# Regenerate SQLC code after changing queries/ or models/
sqlc generate

# Tidy dependencies
go mod tidy
```

## Key Conventions

### CLI Commands
- All commands live in `cmd/`. Each file exports one Cobra command.
- Commands are registered in `cmd/root.go` via `rootCmd.AddCommand(...)`.
- Terminal output uses **pterm** for styled/colored output and animations. Keep the Scottish tone consistent with existing messages.

### Database Layer
- The database is SQLite stored at `~/.local/share/hingmy/database.db` by default.
- Override with the `DB_PATH` environment variable (place in a `.env` file).
- On startup (`root.go`), `CreateIfNotExists()` and `RunMigrations()` are called automatically — no manual setup required.
- All queries are defined in `queries/*.sql` and the Go code in `database/sqlc/` is **auto-generated** by SQLC. After changing any `.sql` query file, run `sqlc generate` to regenerate.
- The `database/accessor.todos.go` pattern wraps SQLC methods in a higher-level accessor. Follow this pattern when adding domain-specific logic.

### Schema Design
- All tables use `created_at` and `updated_at` timestamps.
- Soft deletion is implemented via `deleted_at` (nullable). Use soft-delete queries rather than hard deletes where possible.
- Tags relate to todos via a `tag_entities` join table (many-to-many).
- Notes have a foreign key to `todos`.

### SQLC
- Config is in `sqlc.yaml`. Generated output goes to `database/sqlc/`.
- Never edit files under `database/sqlc/` manually — they will be overwritten by `sqlc generate`.
- Query names follow the pattern `VerbNoun` (e.g., `CreateTodo`, `ListActiveTodos`, `SoftDeleteTag`).

## Current State & Known Incomplete Features

- `update` and `delete` CLI commands are fully implemented via the interactive TUI (`doUpdate`/`doDelete` in `cmd/interactive.go`).
- No test suite exists. When adding tests, use the standard Go `testing` package and place test files alongside the code they test (`*_test.go`).
- No linting configuration is present. Standard Go formatting (`gofmt`) is expected.

## Environment Configuration

Create a `.env` file in the project root (it is gitignored):

```env
DB_PATH=.local/share/hingmy/database.db
```

## Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/mattn/go-sqlite3` | SQLite driver (requires CGo) |
| `github.com/golang-migrate/migrate/v4` | Database migration support |
| `github.com/pterm/pterm` | Terminal UI, colors, animations |
| `github.com/joho/godotenv` | `.env` file loading |

## Tone & Style

The app personality is intentionally Scottish-accented and cheerful. When adding user-facing strings, match the existing tone (e.g., "Aye!", "Weel done!", "Och no!"). Keep it fun but not at the expense of clarity.
