package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
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
		runtime.WindowSetTitle(ctx, fmt.Sprintf("DuMD — %s", filepath.Base(a.filePath)))
	}
}

func (a *App) GetRenderedMarkdown() string {
	if a.filePath == "" {
		return `<div class="empty-state"><p>No file provided.</p><p>Usage: <code>dumd &lt;file.md&gt;</code></p></div>`
	}
	source, err := os.ReadFile(a.filePath)
	if err != nil {
		return fmt.Sprintf(`<div class="empty-state"><p>File not found</p><code>%s</code></div>`, a.filePath)
	}
	if len(bytes.TrimSpace(source)) == 0 {
		return `<div class="empty-state"><p>This file is empty.</p></div>`
	}
	var buf bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.Table))
	if err := md.Convert(source, &buf); err != nil {
		return fmt.Sprintf("<p>Error rendering Markdown: %s</p>", err.Error())
	}
	return buf.String()
}

func (a *App) CloseApp() {
	runtime.Quit(a.ctx)
}

func (a *App) OpenInBrowser(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

func (a *App) OpenLocalMarkdown(relPath string) error {
	relPath = filepath.FromSlash(relPath)

	if idx := strings.Index(relPath, "#"); idx != -1 {
		relPath = relPath[:idx]
	}

	resolved := filepath.Join(filepath.Dir(a.filePath), relPath)
	resolved = filepath.Clean(resolved)

	if filepath.Ext(resolved) != ".md" {
		return nil
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("File not found: %s", resolved)
	}
	if info.IsDir() {
		return fmt.Errorf("Path is a directory: %s", resolved)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Could not determine executable path: %s", err.Error())
	}

	cmd := exec.Command(exePath, resolved)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to open file: %s", err.Error())
	}

	return nil
}
