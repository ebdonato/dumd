package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

const localFilePrefix = "/__mdlocal__/"

var imgSrcRe = regexp.MustCompile(`(<img\b[^>]*?\ssrc\s*=\s*["'])([^"']+)(["'])`)

var imageMIME = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".ico":  "image/x-icon",
	".avif": "image/avif",
}

func isAbsoluteOrSpecialURL(src string) bool {
	lower := strings.ToLower(src)
	for _, p := range []string{"http://", "https://", "data:", "blob:", "wails:", "about:", "//"} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return strings.HasPrefix(src, "#")
}

func rewriteImageSrcs(html, mdDir string) string {
	return imgSrcRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := imgSrcRe.FindStringSubmatch(match)
		src := parts[2]
		if isAbsoluteOrSpecialURL(src) {
			return match
		}
		rel := strings.TrimPrefix(src, "./")
		rel = strings.TrimPrefix(rel, "/")
		if decoded, err := url.PathUnescape(rel); err == nil {
			rel = decoded
		}
		abs := filepath.Clean(filepath.Join(mdDir, filepath.FromSlash(rel)))
		u := url.URL{Path: localFilePrefix + filepath.ToSlash(abs)}
		return parts[1] + u.String() + parts[3]
	})
}

type LocalFileHandler struct {
	app *App
}

func NewLocalFileHandler(app *App) *LocalFileHandler {
	return &LocalFileHandler{app: app}
}

func (h *LocalFileHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := req.URL.Path
	if !strings.HasPrefix(path, localFilePrefix) || h.app.filePath == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	rel := strings.TrimPrefix(path, localFilePrefix)
	abs := filepath.FromSlash(rel)
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(filepath.Dir(h.app.filePath), abs)
	}
	abs = filepath.Clean(abs)
	ext := strings.ToLower(filepath.Ext(abs))
	mimeType, ok := imageMIME[ext]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	f, err := os.Open(abs)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	res.Header().Set("Content-Type", mimeType)
	http.ServeContent(res, req, info.Name(), info.ModTime(), f)
}

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
		return `<div class="empty-state"><img class="empty-state-logo" src="/src/assets/logo.svg" alt="DuMD" /><p>No file provided.</p><p>Usage: <code>dumd &lt;file.md&gt;</code></p></div>`
	}
	source, err := os.ReadFile(a.filePath)
	if err != nil {
		return fmt.Sprintf(`<div class="empty-state"><img class="empty-state-logo" src="/src/assets/logo.svg" alt="DuMD" /><p>File not found</p><code>%s</code></div>`, a.filePath)
	}
	if len(bytes.TrimSpace(source)) == 0 {
		return `<div class="empty-state"><img class="empty-state-logo" src="/src/assets/logo.svg" alt="DuMD" /><p>This file is empty.</p></div>`
	}
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	if err := md.Convert(source, &buf); err != nil {
		return fmt.Sprintf("<p>Error rendering Markdown: %s</p>", err.Error())
	}
	mdDir, err := filepath.Abs(filepath.Dir(a.filePath))
	if err != nil {
		mdDir = filepath.Dir(a.filePath)
	}
	return rewriteImageSrcs(buf.String(), mdDir)
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
