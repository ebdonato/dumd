# AGENTS.md

## Build & Run

- **Dev:** `wails dev`
- **Production build:** `wails build -clean -ldflags "-s -w"`
- Do **not** use `go build` — Wails orchestrates the Go compile + Vite frontend build + embedding.

## Architecture

- Go backend + vanilla JS frontend, glued by **Wails v2**.
- `main.go` — entry point, embeds `frontend/dist` via `go:embed`.
- `app.go` — all business logic (file I/O, Goldmark markdown→HTML, window control). Methods on `App` struct are auto-exposed to JS as Wails bindings.
- `frontend/` — vanilla HTML/CSS/JS frontend built with **Vite**. No React/Vue/Angular.
- `frontend/wailsjs/` — **auto-generated** by Wails. Never edit manually.
- `wails.json` — Wails project config (window size, build scripts).

## Key Behavior

- File path passed as CLI arg: `dumd readme.md`. Stored in `App.filePath`, read at runtime.
- Themes: CSS classes `theme-dark`, `theme-sepia` on `<html>`. Default light has no class.
- Theme & zoom persist via `localStorage` (`dumd-theme`, `dumd-zoom`).
- ESC key: closes the settings popover if open, otherwise calls `CloseApp()`.

## Testing

No tests exist yet. No test framework is configured.

## Constraints

- Binary must stay under **18 MB** (spec requirement).
- Idle RAM under **60 MB**.
- Frontend must remain vanilla JS — no framework bundles.
- Cross-platform targets: Windows x64, macOS (ARM + Intel), Linux x64.
