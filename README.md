# DuMD

A minimal desktop Markdown viewer built with [Wails](https://wails.io).

## Installation

Download the binary for your platform from [Releases](https://github.com/anomalyco/dumd/releases).

## Usage

```
dumd <file.md>
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Esc` | Close app |
| `Space` / `Shift+Space` | Scroll down / up |
| `j` / `k` | Scroll down / up (vim) |
| `h` / `l` | Scroll left / right (vim) |
| `t` | Cycle theme (Light → Dark → Sepia) |
| `Ctrl+=` / `Ctrl+-` | Zoom in / out |

## Themes

- **Light** — Clean white background
- **Dark** — GitHub Dark inspired
- **Sepia** — Warm reading tone

## Build from Source

Requirements: Go ≥ 1.21, [Wails CLI](https://wails.io/docs/gettingstarted/installation), Node.js

```
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails build -clean -ldflags "-s -w"
```
