# Technical Specification: DuMD - Ultra-Minimalist Markdown Viewer

This specification describes the requirements, architecture, and implementation details for **DuMD**, an ultra-minimalist, portable, keyboard-driven desktop Markdown viewer focused on high readability and lightness.

## Product Overview

**DuMD** is a desktop application whose sole purpose is to render local Markdown files in an elegant and clean way. It is designed for users who value minimalism, keyboard efficiency (Vim-like shortcuts / speed reading), and ease of distribution (single binary, no complex installers).

### Key Differentiators

- **Single Executable:** A single compiled file containing the Go engine and the embedded Webview interface.
- **Keyboard-First:** Navigation control inspired by Vim and smooth scrolling via spacebar.
- **Visual Minimalism:** Interface stripped of complex menus, using only the native window frame and a discreet floating button for basic customization.

## Functional Requirements (FR)

| ID       | Requirement Name           | Description                                                                                                                                                  |
| :------- | :------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **FR01** | **File Loading**           | The application must open and read a Markdown file (`.md`) passed as a command-line argument (e.g., `dumd readme.md`).                                       |
| **FR02** | **Markdown Rendering**     | The file content must be converted into semantic HTML5 and styled with highly readable typography.                                                           |
| **FR03** | **Native Window Controls** | The window must retain the default decorations of the host operating system (close, minimize, and maximize buttons).                                         |
| **FR04** | **Quick Settings Menu**    | A floating panel or discreet button (e.g., bottom-right corner) must allow zoom changes and visual theme switching.                                          |
| **FR05** | **Theme Support**          | The system must support 3 reading themes: **Light** (clean paper), **Dark** (high contrast and eye-friendly at night), and **Sepia** (warm reading comfort). |
| **FR06** | **Zoom Control**           | The user must be able to adjust the font size (e.g., 80% to 150%) through the menu or shortcuts.                                                             |
| **FR07** | **Keyboard Navigation**    | Implement a complete key mapping for scrolling and app control.                                                                                              |

## Non-Functional Requirements (NFR)

| ID        | Non-Functional Requirement         | Performance Target / Constraint                                                                                                             |
| :-------- | :--------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------ |
| **NFR01** | **Portability**                    | The application must compile natively for Windows (x64), macOS (Apple Silicon/Intel), and Linux (x64).                                      |
| **NFR02** | **Executable Size**                | The final compiled and optimized binary size must not exceed **18 MB**.                                                                     |
| **NFR03** | **Memory Consumption**             | Idle RAM consumption must be below **60 MB**.                                                                                               |
| **NFR04** | **Single-File Distribution**       | All frontend assets (HTML, CSS, JS) and additional resources must be embedded in the Go binary using the native `embed` library (Go 1.16+). |
| **NFR05** | **No External Runtime Dependency** | The end user must not need to install Node.js, Python, or extra libraries on the system to run the app.                                     |

## Technology Architecture

The application will use a hybrid model of a **Go Backend** and an **HTML/CSS/JS Frontend** encapsulated by a native WebView from the operating system.

```text
+-------------------------------------------------------------+
|                        WINDOW FRAME                         |
|  +-------------------------------------------------------+  |
|  | [Wails Webview] HTML / CSS / JS (Vanilla)             |  |
|  |                                                       |  |
|  | - Keyboard event capture (Vim / Space / ESC)          |  |
|  | - Styled HTML rendering                               |  |
|  | - Theme management (Light, Dark, Sepia)               |  |
|  | - Dynamic zoom adjustment                             |  |
|  +-------------------------------------------------------+  |
|                           | (Bidirectional IPC Communication)
|                           v                                 |
|  +-------------------------------------------------------+  |
|  | [Go Backend]                                          |  |
|  |                                                       |  |
|  | - .md file reading (I/O)                              |  |
|  | - Markdown -> HTML parser (Goldmark)                  |  |
|  | - System calls (Close window on ESC)                  |  |
|  +-------------------------------------------------------+  |
+-------------------------------------------------------------+
```

### Software Components

1. **Main Framework (Wails v2):** Connects the Go backend to the frontend natively and optimally using native Web engines:
   - Windows: _Microsoft Edge WebView2 (Chromium)_
   - macOS: _WebKit (WKWebView)_
   - Linux: _WebKitGTK_
2. **Markdown Parser (Goldmark):** An extremely fast parser written in pure Go that follows the CommonMark specification.
3. **Frontend (HTML5/CSS3/Vanilla JS):** The frontend will not use React, Vue, or Angular to keep the loading weight minimal. CSS will define the responsive design, fluid typography styles, and theme variables.

## Keyboard Shortcuts Mapping

Keyboard control is a top priority for the application's usability.

| Key / Combination  | Action Performed                       | Implementation Mechanism                                                         |
| :----------------- | :------------------------------------- | :------------------------------------------------------------------------------- |
| `Escape` (ESC)     | Closes the application immediately     | Captured in JS -> Calls `runtime.WindowClose(ctx)` in Go.                        |
| `Space`            | Scrolls down (80% of visible height)   | JS event `window.scrollBy({top: window.innerHeight * 0.8, behavior: 'smooth'})`  |
| `Shift + Space`    | Scrolls up (80% of visible height)     | JS event `window.scrollBy({top: -window.innerHeight * 0.8, behavior: 'smooth'})` |
| `j`                | Scrolls down slightly (line by line)   | JS event `window.scrollBy({top: 50, behavior: 'auto'})`                          |
| `k`                | Scrolls up slightly (line by line)     | JS event `window.scrollBy({top: -50, behavior: 'auto'})`                         |
| `h`                | Scrolls left (useful for code blocks)  | JS event `window.scrollBy({left: -50, behavior: 'auto'})`                        |
| `l`                | Scrolls right (useful for code blocks) | JS event `window.scrollBy({left: 50, behavior: 'auto'})`                         |
| `Ctrl +` / `Cmd +` | Increases Zoom (overall font size)     | JS event modifying `document.body.style.fontSize` property                       |
| `Ctrl -` / `Cmd -` | Decreases Zoom (overall font size)     | JS event modifying `document.body.style.fontSize` property                       |
| `t`                | Cycles through Themes                  | JS rotates body class: `light` -> `dark` -> `sepia` -> `light`.                  |

## Interface Design (UI) and Themes

The design must follow the concept of _"zero distraction"_.

- **No extra margins:** The Markdown content will occupy the center of the window with a maximum readable width of `750px` (or `70ch`), regardless of the window size.
- **Typography:** Use system fonts to avoid web requests or the weight of embedded font files (e.g., Inter, Segoe UI, San Francisco, or Roboto).
- **Floating Button (Menu):** A small gear icon in the bottom-right corner that only appears when the mouse moves nearby. When clicked, it reveals a simple dropdown menu (Pop-over) with theme selectors and zoom level controls.

### CSS Theme Variables (Design Tokens)

```css
/* Light Theme (Default) */
:root {
  --bg-color: #ffffff;
  --text-color: #24292f;
  --accent-color: #0969da;
  --code-bg: #f6f8fa;
  --border-color: #d0d7de;
}
```

```css
/* Dark Theme */
body.theme-dark {
  --bg-color: #0d1117;
  --text-color: #c9d1d9;
  --accent-color: #58a6ff;
  --code-bg: #161b22;
  --border-color: #30363d;
}
```

```css
/* Sepia Theme */
body.theme-sepia {
  --bg-color: #f4ecd8;
  --text-color: #433422;
  --accent-color: #925d23;
  --code-bg: #e9dec4;
  --border-color: #d3c2a0;
}
```

## **7\. Project Structure (Directory Tree)**

The following tree follows the standard structure required by the **Wails** framework:

```text
dumd/
├── main.go             # Go entry point and Wails setup
├── app.go              # Business logic (Go Bindings: read MD, close app)
├── wails.json          # Wails build and window configuration
├── go.mod              # Dependency modules (Wails, Goldmark)
├── go.sum
└── frontend/           # Frontend resources (Embedded via go:embed)
    ├── index.html      # Base interface structure
    ├── src/
    │   ├── main.js     # Keyboard navigation logic, themes, and Go calls
    │   └── style.css   # Markdown styles (similar to GitHub Readme) and themes
    └── assets/
        └── icon.png    # Official executable icon
```

## Implementation and Build Strategy

To compile and generate the single distribution file, the development workflow should follow the steps below:

### Step 1: Project Initialization

```shell
# Install the Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Create the project using the Vanilla JS template
wails init -n dumd -t vanilla
```

### Step 2: Backend Parser Implementation (app.go)

1. Import the `github.com/yuin/goldmark` package.
2. Add support for reading system arguments (`os.Args`) during the app lifecycle initialization to detect the `.md` file to be opened.
3. Expose a Go function to JS called `GetRenderedMarkdown()` that returns the HTML string generated by Goldmark.

### Step 3: Window Event Listening in JS

1. In the `frontend/src/main.js` file, trigger the Markdown reading by calling the Go backend.
2. Update `document.getElementById('content').innerHTML` with the response.
3. Set up global keyboard listeners using `window.addEventListener('keydown')` to capture keyboard navigation and call the exposed native routines (such as closing the window on ESC).

### Step 4: Production Build

To compile everything and generate the final optimized executable with reduced size, run the command corresponding to your operating system:

```shell
# Build for the current system with production optimizations
wails build -clean -ldflags "-s \-w"
```

_The `-ldflags "-s \-w"` parameter removes debug information from the Go binary, reducing the final executable by approximately 30% to 40% of its original size._
