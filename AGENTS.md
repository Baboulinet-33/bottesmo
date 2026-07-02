# Agent Instructions: Version Bumping

This project follows **Semantic Versioning** (`MAJOR.MINOR.PATCH`).

## Version source of truth

The canonical version is stored in `internal/version/version.go`:

```go
var Version = "0.1.0"
```

Always update this file when bumping the version. Do NOT rely on `package.json` — that version belongs to the Playwright test suite, not the application.

## Bump rules

| Change type | Bump | Example | Action |
|---|---|---|---|
| Bug fix, refactor, minor tweak | PATCH | `0.1.0` → `0.1.1` | Increment PATCH |
| New feature (non-breaking) | MINOR | `0.1.0` → `0.2.0` | Increment MINOR, reset PATCH to `0` |
| Breaking change | MAJOR | `0.1.0` → `1.0.0` | Increment MAJOR, reset MINOR and PATCH to `0` |

## Default rule

Every time you modify application code (server, handlers, game logic, dictionary, etc.), bump the **PATCH** version in `internal/version/version.go` — unless the change clearly qualifies as MINOR or MAJOR per the rules above.

## Example

After fixing a bug in `internal/handlers/game.go`:

1. Open `internal/version/version.go`
2. Change `Version = "0.1.0"` to `Version = "0.1.1"`
3. Save the file
