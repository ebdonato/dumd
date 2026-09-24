package main

import (
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustURLPath(prefix, abs string) string {
	u := url.URL{Path: prefix + filepath.ToSlash(abs)}
	return u.String()
}

func TestRewriteImageSrcs(t *testing.T) {
	mdDir := t.TempDir()

	cases := []struct {
		name string
		in   string
		want string
		skip bool
	}{
		{"relative", `<img src="img/logo.png" alt="">`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "img", "logo.png")), false},
		{"dot-slash", `<img src="./img/logo.png" alt="">`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "img", "logo.png")), false},
		{"root-relative", `<img src="/img/logo.png" alt="">`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "img", "logo.png")), false},
		{"parent-relative", `<img src="../shared/pic.jpg" alt="">`, mustURLPath(localFilePrefix, filepath.Join(filepath.Dir(mdDir), "shared", "pic.jpg")), false},
		{"single-quotes", `<img src='docs/logo.svg'>`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "docs", "logo.svg")), false},
		{"attr-before-src", `<img class="x" src="a.png">`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "a.png")), false},
		{"percent-20", `<img src="my%20image.png">`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "my image.png")), false},
		{"http", `<img src="https://x.com/a.png">`, "", true},
		{"http-insecure", `<img src="http://x.com/a.png">`, "", true},
		{"data-uri", `<img src="data:image/png;base64,AAA">`, "", true},
		{"protocol-relative", `<img src="//x.com/a.png">`, "", true},
		{"data-src-not-matched", `<img data-src="a.png" src="b.png">`, mustURLPath(localFilePrefix, filepath.Join(mdDir, "b.png")), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rewriteImageSrcs(tc.in, mdDir)
			if tc.skip {
				if strings.Contains(got, localFilePrefix) {
					t.Fatalf("expected no rewrite, got %q", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Fatalf("want %q in %q", tc.want, got)
			}
		})
	}
}

func TestLocalFileHandler(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "logo.png")
	if err := os.WriteFile(img, []byte("PNGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "img")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(sub, "pic.jpg")
	if err := os.WriteFile(nested, []byte("JPGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := NewLocalFileHandler(&App{filePath: filepath.Join(dir, "readme.md")})

	serve := func(path string) (int, string) {
		req := httptest.NewRequest("GET", "http://wails.localhost"+path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}

	if code, body := serve(localFilePrefix + filepath.ToSlash(img)); code != 200 || body != "PNGDATA" {
		t.Fatalf("absolute image: code=%d body=%q", code, body)
	}
	if code, body := serve(localFilePrefix + "img/pic.jpg"); code != 200 || body != "JPGDATA" {
		t.Fatalf("relative image: code=%d body=%q", code, body)
	}
	if code, _ := serve(localFilePrefix + "nope.png"); code != 404 {
		t.Fatalf("missing file: code=%d", code)
	}
	if code, _ := serve(localFilePrefix + filepath.ToSlash(filepath.Join(dir, "notes.txt"))); code != 404 {
		t.Fatalf("non-image ext: code=%d", code)
	}
	if code, _ := serve("/assets/index.css"); code != 404 {
		t.Fatalf("unprefixed path: code=%d", code)
	}
	req := httptest.NewRequest("POST", "http://wails.localhost"+localFilePrefix+"logo.png", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Fatalf("POST: code=%d", rec.Code)
	}
}

func TestGetRenderedMarkdownRewritesReadmeImages(t *testing.T) {
	readme, err := filepath.Abs("README.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(readme); err != nil {
		t.Skip("README.md not present")
	}
	html := NewApp(readme).GetRenderedMarkdown()
	for _, frag := range []string{"docs/logo.svg", "docs/screenshots/light.png"} {
		abs := filepath.Join(filepath.Dir(readme), filepath.FromSlash(frag))
		want := mustURLPath(localFilePrefix, abs)
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in rendered HTML", want)
		}
	}
	if strings.Contains(html, `src="docs/`) {
		t.Fatal("left a bare relative src behind")
	}
}
