# AGENTS.md

## Build & Run

- **Dev:** `wails dev`
- **Production build:** `wails build -clean -ldflags "-s -w"`
- Do **not** use `go build` for production — Wails orchestrates the Go compile + Vite frontend build + embedding. Plain `go build` / `go install` work only because `frontend/dist/` is committed (no icon, no `-H windowsgui`, no `-s -w`).
- First-time setup: run `npm install` inside `frontend/` before `wails dev`.
- Output binary lands in `build/bin/`.

## Architecture

- Go backend + vanilla JS frontend, glued by **Wails v2**.
- Go module name is `github.com/ebdonato/dumd` (required by remote `go install`).
- `main.go` — entry point, embeds `frontend/dist` via `go:embed`.
- `app.go` — all business logic (file I/O, Goldmark markdown→HTML, window control). Methods on `App` struct are auto-exposed to JS as Wails bindings.
- `console_windows.go` (`windows && !dev`) — per-monitor DPI, window icon from embedded `build/windows/icon.ico`, and self-detach so console-subsystem `go install` builds release the terminal.
- `console_stub.go` (`!windows || dev`) — no-op stubs.
- `frontend/src/main.js` — all frontend logic (keyboard handling, themes, zoom, IPC calls). Single file, no framework.
- `frontend/src/style.css` — all CSS including theme variables and typography.
- `frontend/wailsjs/` — **auto-generated** by Wails. Never edit.
- `frontend/dist/` — **auto-generated** by Vite build. Never edit. **Tracked in git** (needed by `go:embed` for `go install`). Rebuild with `npm run build` in `frontend/` and commit whenever `frontend/src/` changes. CI `dist-check.yml` fails if stale.
- No `vite.config.js` — Wails provides its own Vite defaults.
- Goldmark is configured with `extension.Table` (GFM tables).

## Key Behavior

- File path passed as CLI arg: `dumd readme.md`. Stored in `App.filePath`, read at runtime.
- If no arg is given, an empty-state welcome screen is shown.
- Themes: CSS classes `theme-dark`, `theme-sepia` on `<html>`. Default light has no class.
- Theme & zoom persist via `localStorage` (`dumd-theme`, `dumd-zoom`).
- `q` and `Esc` both quit the app (Esc closes settings popover first if open).
- Relative `.md` links spawn a new DuMD process (`OpenLocalMarkdown` uses `os/exec`).

## Testing

No tests exist yet. No test framework is configured.

## Install (`go install`)

- Remote install: `go install -tags production github.com/ebdonato/dumd@latest` (Linux: add `webkit2_41`).
- Requires the module path to stay `github.com/ebdonato/dumd` and `frontend/dist/` to be committed and current.
- `.github/workflows/dist-check.yml` enforces the dist freshness and that the `go build` compile path works.

## Constraints

- Binary must stay under **18 MB** (spec requirement).
- Idle RAM under **60 MB**.
- Frontend must remain vanilla JS — no framework bundles.
- Cross-platform targets: Windows x64, macOS (ARM + Intel), Linux x64.
