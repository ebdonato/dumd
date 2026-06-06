package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

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
		runtime.WindowSetTitle(ctx, fmt.Sprintf("DuMD — %s", filepath.Base(a.filePath)))
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
