# DuMD — Implementation Plan

> Detailed, step-by-step plan derived from the [Technical Specification](./spec.md).
> Each phase is self-contained and builds on the previous one. A phase is considered **done** when all of its acceptance criteria pass.

---

## Phase 0: Environment & Project Scaffolding

**Goal:** Set up the development environment, install tooling, and scaffold the Wails project so that a blank window can be opened.

### Tasks

| #   | Task                        | Details                                                                                                                                   |
| --- | --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| 0.1 | **Install prerequisites**   | Ensure Go ≥ 1.21 is installed. Ensure platform WebView runtime is available (WebView2 on Windows, WebKit on macOS/Linux).                 |
| 0.2 | **Install Wails CLI**       | `go install github.com/wailsapp/wails/v2/cmd/wails@latest`                                                                                |
| 0.3 | **Run Wails doctor**        | `wails doctor` — verify all system dependencies are satisfied.                                                                            |
| 0.4 | **Scaffold the project**    | `wails init -n dumd -t vanilla` inside the repo root. Move generated files to the repo root so the directory layout matches the spec.     |
| 0.5 | **Clean up boilerplate**    | Remove the default Wails demo content from `index.html`, `main.js`, and `style.css`. Leave empty shells so subsequent phases start clean. |
| 0.6 | **Add Goldmark dependency** | `go get github.com/yuin/goldmark`                                                                                                         |
| 0.7 | **Verify blank window**     | `wails dev` — confirm a blank window with native OS decorations opens and closes normally.                                                |

### Acceptance Criteria

- [ ] `wails doctor` reports no errors.
- [ ] `wails dev` launches a blank window with native close/minimize/maximize buttons (**FR03**).
- [ ] `go.mod` lists both `wails/v2` and `goldmark` as dependencies.

### Deliverables

```text
dumd/
├── main.go
├── app.go              (empty struct + lifecycle stubs)
├── wails.json
├── go.mod
├── go.sum
└── frontend/
    ├── index.html       (minimal shell)
    ├── src/
    │   ├── main.js      (empty)
    │   └── style.css    (empty)
    └── assets/
        └── icon.png     (placeholder)
```

---

## Phase 1: Backend — Markdown Parsing & File Loading

**Goal:** The Go backend reads a `.md` file path from CLI arguments, converts it to HTML with Goldmark, and exposes the result to the frontend via a Wails binding.

### Tasks

| #   | Task                                  | Details                                                                                                                                                                                                       |
| --- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1.1 | **Define the `App` struct**           | In `app.go`, define the `App` struct with a `ctx context.Context` field and a `filePath string` field.                                                                                                        |
| 1.2 | **Capture CLI arguments**             | In `main.go`, before `wails.Run(...)`, read `os.Args`. If `len(os.Args) >= 2`, store `os.Args[1]` as the Markdown file path and pass it to the `App` instance. If no argument is given, show a usage message. |
| 1.3 | **Implement `GetRenderedMarkdown()`** | Bound method on `App` that: reads the file at `filePath` with `os.ReadFile`, converts bytes to HTML via `goldmark.Convert`, and returns the HTML string. Return an error-wrapped HTML snippet on failure.     |
| 1.4 | **Implement `CloseApp()`**            | Bound method that calls `runtime.Quit(a.ctx)` to cleanly shut down the application. This will be invoked from JS on `ESC`.                                                                                    |
| 1.5 | **Configure `wails.json`**            | Set window title to the filename (e.g., `"DuMD — readme.md"`), default width `900`, height `700`, `"frameless": false` to keep native decorations.                                                            |

### Code Sketch — `app.go`

```go
package main

import (
  "bytes"
  "context"
  "fmt"
  "os"

  "github.com/wailsapp/wails/v2/pkg/runtime"
  "github.com/yuin/goldmark"
)

type App struct {
  ctx      context.Context
  filePath string
}

func NewApp(filePath string) *App {
  return &App{filePath: filePath}
}

func (a *App) startup(ctx context.Context) {
  a.ctx = ctx
  if a.filePath != "" {
    runtime.WindowSetTitle(ctx, fmt.Sprintf("DuMD — %s", a.filePath))
  }
}

func (a *App) GetRenderedMarkdown() string {
  if a.filePath == "" {
    return "<p>No file provided. Usage: <code>dumd &lt;file.md&gt;</code></p>"
  }
  source, err := os.ReadFile(a.filePath)
  if err != nil {
    return fmt.Sprintf("<p>Error reading file: %s</p>", err.Error())
  }
  var buf bytes.Buffer
  if err := goldmark.Convert(source, &buf); err != nil {
    return fmt.Sprintf("<p>Error rendering Markdown: %s</p>", err.Error())
  }
  return buf.String()
}

func (a *App) CloseApp() {
  runtime.Quit(a.ctx)
}
```

### Code Sketch — `main.go`

```go
package main

import (
  "embed"
  "os"

  "github.com/wailsapp/wails/v2"
  "github.com/wailsapp/wails/v2/pkg/options"
  "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
  filePath := ""
  if len(os.Args) >= 2 {
    filePath = os.Args[1]
  }

  app := NewApp(filePath)

  err := wails.Run(&options.App{
    Title:     "DuMD",
    Width:     900,
    Height:    700,
    AssetServer: &assetserver.Options{
      Assets: assets,
    },
    OnStartup: app.startup,
    Bind: []interface{}{
      app,
    },
  })
  if err != nil {
    println("Error:", err.Error())
  }
}
```

### Acceptance Criteria

- [ ] Running `wails dev -- readme.md` renders the file content (raw HTML) inside the webview (**FR01**, **FR02**).
- [ ] Running without arguments shows a friendly "No file provided" message.
- [ ] `CloseApp()` can be called from the browser console to quit the app.

---

## Phase 2: Frontend — HTML Structure & Markdown Styling

**Goal:** Build the static HTML shell, call the Go backend on load to inject rendered Markdown, and apply GitHub-Readme-quality typography.

### Tasks

| #   | Task                                 | Details                                                                                                                                                                                                                                        |
| --- | ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2.1 | **Create `index.html` structure**    | Minimal HTML5 document with: `<meta charset="UTF-8">`, `<meta name="viewport">`, link to `style.css`, a single `<article id="content"></article>` container, and a `<script>` tag for `main.js`.                                               |
| 2.2 | **Call backend on DOMContentLoaded** | In `main.js`, import the Wails runtime and call `window.go.main.App.GetRenderedMarkdown()`. Insert the returned HTML into `#content`.                                                                                                          |
| 2.3 | **Style Markdown elements**          | In `style.css`, create comprehensive styles for all Markdown output elements: `h1`–`h6`, `p`, `a`, `ul`/`ol`/`li`, `blockquote`, `code`/`pre`, `table`/`thead`/`tbody`/`tr`/`th`/`td`, `hr`, `img`. Aim for a look similar to GitHub's Readme. |
| 2.4 | **Centered content column**          | Set `#content` to `max-width: 750px; margin: 0 auto; padding: 2rem 1.5rem;` for comfortable centered reading regardless of window size.                                                                                                        |
| 2.5 | **System font stack**                | Define `font-family` using a system font stack: `'Segoe UI', -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Arial, sans-serif`. Set base `font-size: 16px` and `line-height: 1.7`.                                                       |
| 2.6 | **Code block styling**               | `pre` blocks: horizontal scroll with `overflow-x: auto`, padding, rounded corners, distinct background. Inline `code`: subtle background, slight padding, rounded corners.                                                                     |

### Code Sketch — `index.html`

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>DuMD</title>
    <link rel="stylesheet" href="/src/style.css" />
  </head>
  <body>
    <article id="content">
      <p>Loading…</p>
    </article>
    <script src="/src/main.js" type="module"></script>
  </body>
</html>
```

### Acceptance Criteria

- [ ] Markdown is rendered with styled headings, paragraphs, code blocks, tables, and links (**FR02**).
- [ ] Content is centered and never exceeds `750px` wide.
- [ ] The page is readable with no visible scrollbar jank; `pre` blocks scroll horizontally.

---

## Phase 3: Themes & Zoom

**Goal:** Implement the three reading themes (Light, Dark, Sepia) and font-size zoom control, both via CSS custom properties.

### Tasks

| #   | Task                               | Details                                                                                                                                                                                                                     |
| --- | ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 3.1 | **Define CSS custom properties**   | In `style.css`, declare the Light theme tokens on `:root` and override them on `body.theme-dark` and `body.theme-sepia`. Variables: `--bg-color`, `--text-color`, `--accent-color`, `--code-bg`, `--border-color`.          |
| 3.2 | **Apply variables throughout CSS** | Refactor all colors in the stylesheet to reference the CSS variables instead of hard-coded hex values.                                                                                                                      |
| 3.3 | **Theme cycling logic in JS**      | In `main.js`, create a `cycleTheme()` function that rotates `document.body.className` through `""` (light) → `"theme-dark"` → `"theme-sepia"` → `""`. Store current theme in `localStorage` for persistence.                |
| 3.4 | **Restore theme on load**          | On `DOMContentLoaded`, read `localStorage.getItem('dumd-theme')` and apply the saved class to `<body>`.                                                                                                                     |
| 3.5 | **Zoom control in JS**             | Create `zoomIn()` and `zoomOut()` functions. Maintain a `zoomLevel` variable (default `100`, min `80`, max `150`, step `10`). Apply by setting `document.body.style.fontSize = zoomLevel + '%'`. Persist in `localStorage`. |
| 3.6 | **Restore zoom on load**           | On `DOMContentLoaded`, read `localStorage.getItem('dumd-zoom')` and apply it.                                                                                                                                               |

### CSS Theme Tokens (from spec)

```css
/* Light Theme (Default) */
:root {
  --bg-color: #ffffff;
  --text-color: #24292f;
  --accent-color: #0969da;
  --code-bg: #f6f8fa;
  --border-color: #d0d7de;
}

/* Dark Theme */
body.theme-dark {
  --bg-color: #0d1117;
  --text-color: #c9d1d9;
  --accent-color: #58a6ff;
  --code-bg: #161b22;
  --border-color: #30363d;
}

/* Sepia Theme */
body.theme-sepia {
  --bg-color: #f4ecd8;
  --text-color: #433422;
  --accent-color: #925d23;
  --code-bg: #e9dec4;
  --border-color: #d3c2a0;
}
```

### Acceptance Criteria

- [ ] Three distinct themes render correctly and switch smoothly (**FR05**).
- [ ] Theme selection persists across app restarts.
- [ ] Zoom adjusts font size from 80% to 150% in 10% increments (**FR06**).
- [ ] Zoom level persists across app restarts.

---

## Phase 4: Keyboard Navigation & Floating Settings Menu

**Goal:** Wire up all keyboard shortcuts from the spec and build the floating gear-button popover for mouse users.

### Tasks

| #    | Task                                         | Details                                                                                                                                                                                                             |
| ---- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 4.1  | **Global `keydown` listener**                | In `main.js`, add `window.addEventListener('keydown', handler)`. Route keys to their respective actions. Ensure shortcuts are suppressed when focus is inside an input/textarea (edge-case guard).                  |
| 4.2  | **Implement `ESC` → Close**                  | On `Escape`, call `window.go.main.App.CloseApp()`.                                                                                                                                                                  |
| 4.3  | **Implement `Space` / `Shift+Space` scroll** | `Space` → `window.scrollBy({top: window.innerHeight * 0.8, behavior: 'smooth'})`. `Shift+Space` → same but negative. Call `e.preventDefault()` to avoid default spacebar scroll.                                    |
| 4.4  | **Implement `j`/`k`/`h`/`l` Vim scroll**     | `j` → scroll down 50px, `k` → up 50px, `h` → left 50px, `l` → right 50px. All with `behavior: 'auto'`.                                                                                                              |
| 4.5  | **Implement `Ctrl+`/`Ctrl-` zoom**           | Intercept `Ctrl+=` (zoom in) and `Ctrl+-` (zoom out). Call `zoomIn()` / `zoomOut()` from Phase 3. `e.preventDefault()` to block native browser zoom.                                                                |
| 4.6  | **Implement `t` → theme cycle**              | Call `cycleTheme()` from Phase 3.                                                                                                                                                                                   |
| 4.7  | **Build floating gear button (HTML/CSS)**    | Add a `<button id="settings-btn">` with a gear SVG icon, positioned `fixed` at `bottom: 1.5rem; right: 1.5rem`. Default `opacity: 0`, transitions to `opacity: 1` when the cursor enters a hotzone near the corner. |
| 4.8  | **Build settings popover (HTML/CSS)**        | A `<div id="settings-popover">` anchored above the gear button. Contains: a theme selector (3 labeled buttons/swatches) and a zoom slider or +/- buttons with a percentage label.                                   |
| 4.9  | **Popover toggle logic**                     | Click gear → toggle `popover.classList.toggle('open')`. Click outside popover → close it. Press `ESC` while popover is open → close popover (do **not** quit the app on the first ESC if the popover is open).      |
| 4.10 | **Gear button hover reveal**                 | Use JS `mousemove` on `document` to detect if cursor is within ~80px of the bottom-right corner. If so, add a `.visible` class to the gear button. Remove it after a short delay when the cursor leaves the zone.   |

### Key Handler Sketch

```javascript
window.addEventListener("keydown", (e) => {
  // Don't intercept if user is typing in an input
  if (e.target.tagName === "INPUT" || e.target.tagName === "TEXTAREA") return;

  switch (e.key) {
    case "Escape":
      if (isPopoverOpen()) {
        closePopover();
      } else {
        window.go.main.App.CloseApp();
      }
      break;
    case " ":
      e.preventDefault();
      const dir = e.shiftKey ? -0.8 : 0.8;
      window.scrollBy({ top: window.innerHeight * dir, behavior: "smooth" });
      break;
    case "j":
      window.scrollBy({ top: 50, behavior: "auto" });
      break;
    case "k":
      window.scrollBy({ top: -50, behavior: "auto" });
      break;
    case "h":
      window.scrollBy({ left: -50, behavior: "auto" });
      break;
    case "l":
      window.scrollBy({ left: 50, behavior: "auto" });
      break;
    case "t":
      cycleTheme();
      break;
    case "=":
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        zoomIn();
      }
      break;
    case "-":
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        zoomOut();
      }
      break;
  }
});
```

### Acceptance Criteria

- [ ] All 10 keyboard shortcuts from the spec table work correctly (**FR07**).
- [ ] `ESC` closes the popover first (if open) before quitting the app.
- [ ] The gear button is invisible by default and fades in when the cursor approaches (**FR04**).
- [ ] The popover allows changing theme and zoom via mouse (**FR04**, **FR05**, **FR06**).

---

## Phase 5: Polish, Edge Cases & UX Refinements

**Goal:** Handle edge cases, add smooth transitions, fine-tune the reading experience, and ensure cross-platform consistency.

### Tasks

| #   | Task                                     | Details                                                                                                                                                                    |
| --- | ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 5.1 | **Smooth theme transitions**             | Add `transition: background-color 0.3s ease, color 0.3s ease;` to `body` and key elements so theme switches feel fluid rather than jarring.                                |
| 5.2 | **Scroll position indicator (optional)** | Consider a subtle thin progress bar at the top of the viewport showing scroll percentage. Lightweight CSS + JS only.                                                       |
| 5.3 | **Empty/invalid file handling**          | If the file is empty, show a friendly "This file is empty" centered message. If the file doesn't exist, show "File not found: `<path>`".                                   |
| 5.4 | **Large file performance test**          | Test with a Markdown file ≥ 5,000 lines. Ensure rendering completes in < 500ms and scrolling stays at 60fps.                                                               |
| 5.5 | **Image rendering**                      | Ensure `<img>` tags from Markdown render properly. Set `img { max-width: 100%; height: auto; }` to prevent overflow.                                                       |
| 5.6 | **Link behavior**                        | External links (`http://`, `https://`) should open in the system default browser via Wails' `runtime.BrowserOpenURL()`. Prevent navigating away from the rendered content. |
| 5.7 | **Accessible contrast ratios**           | Verify all three themes meet WCAG AA contrast (≥ 4.5:1 for body text). Adjust hex values if needed.                                                                        |
| 5.8 | **Selection & copy styling**             | Style `::selection` for each theme so that selected text has a pleasant highlight color.                                                                                   |
| 5.9 | **Window title from filename**           | Extract just the basename from the file path for the window title (e.g., `DuMD — readme.md` not the full absolute path).                                                   |

### Acceptance Criteria

- [ ] Theme transitions are smooth (no flash of unstyled content).
- [ ] Missing / empty file paths produce user-friendly messages.
- [ ] Links open externally in the default browser, not inside the webview.
- [ ] All themes pass WCAG AA contrast.

---

## Phase 6: Production Build & Distribution

**Goal:** Produce optimized, single-file executables for each target platform and validate against the non-functional requirements.

### Tasks

| #   | Task                               | Details                                                                                                                                                            |
| --- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 6.1 | **Production build command**       | `wails build -clean -ldflags "-s -w"` — strips debug symbols for a smaller binary.                                                                                 |
| 6.2 | **Verify binary size**             | Confirm the output binary is **≤ 18 MB** (**NFR02**). If it exceeds, investigate and strip unnecessary assets.                                                     |
| 6.3 | **Verify memory consumption**      | Launch the production binary, open a moderately sized `.md` file, and verify idle RAM is **< 60 MB** using OS task manager (**NFR03**).                            |
| 6.4 | **Cross-compile targets**          | Build for all three target platforms (**NFR01**):                                                                                                                  |
|     |                                    | • Windows x64: `GOOS=windows GOARCH=amd64 wails build`                                                                                                             |
|     |                                    | • macOS arm64: `GOOS=darwin GOARCH=arm64 wails build`                                                                                                              |
|     |                                    | • Linux x64: `GOOS=linux GOARCH=amd64 wails build`                                                                                                                 |
| 6.5 | **Test on each platform**          | Smoke-test the binary on each OS: file loading, theme switching, keyboard shortcuts, zoom, and app close.                                                          |
| 6.6 | **Create app icon**                | Design or source a minimal icon for DuMD. Place it at `frontend/assets/icon.png` and configure it in `wails.json`.                                                 |
| 6.7 | **Write README.md**                | Document: what DuMD is, installation (download binary), usage (`dumd <file.md>`), keyboard shortcuts table, theme screenshots, and build-from-source instructions. |
| 6.8 | **No external runtime validation** | Confirm the binary runs on a clean machine with no Go, Node.js, or Python installed (**NFR05**).                                                                   |

### Acceptance Criteria

- [ ] Single binary launches and works without any runtime dependencies (**NFR04**, **NFR05**).
- [ ] Binary size ≤ 18 MB (**NFR02**).
- [ ] Idle memory < 60 MB (**NFR03**).
- [ ] Builds and runs on Windows x64, macOS (ARM + Intel), and Linux x64 (**NFR01**).

---

## Summary: Requirements Traceability

| Requirement | Covered In                                  |
| ----------- | ------------------------------------------- |
| **FR01**    | Phase 1 (task 1.2, 1.3)                     |
| **FR02**    | Phase 1 (task 1.3), Phase 2 (tasks 2.2–2.6) |
| **FR03**    | Phase 0 (task 0.7)                          |
| **FR04**    | Phase 4 (tasks 4.7–4.10)                    |
| **FR05**    | Phase 3 (tasks 3.1–3.4)                     |
| **FR06**    | Phase 3 (tasks 3.5–3.6)                     |
| **FR07**    | Phase 4 (tasks 4.1–4.6)                     |
| **NFR01**   | Phase 6 (task 6.4)                          |
| **NFR02**   | Phase 6 (task 6.2)                          |
| **NFR03**   | Phase 6 (task 6.3)                          |
| **NFR04**   | Phase 6 (task 6.1)                          |
| **NFR05**   | Phase 6 (task 6.8)                          |
