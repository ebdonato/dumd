// Command resgen compiles the committed Windows resource objects (.syso) for
// the DuMD binary so that `go install github.com/ebdonato/dumd@latest` embeds
// the icon, the application manifest and the version info — the same
// resources `wails build` injects through its own ephemeral
// `<name>-res.syso`.
//
// It mirrors Wails' build/pkg/packager.go compileResources: templates in
// build/windows/{info.json,wails.exe.manifest} are resolved with the project
// data from wails.json, then bundled with build/windows/icon.ico into
// internal/winres/dumd_windows_<arch>.syso.
//
// The output lives in its own package and filename, deliberately distinct
// from Wails' ephemeral <name>-res.syso: the GOARCH-suffixed name keeps the
// Go toolchain from linking it on non-Windows targets, and the package is
// only pulled in by resources_windows.go, which excludes Wails' own build
// tags (dev, desktop) so the two resource objects never reach the linker
// together — two would fail with "too many .rsrc sections".
//
// Regenerate and re-commit the internal/winres/*.syso files by hand whenever
// wails.json info or build/windows/{icon.ico,info.json,wails.exe.manifest}
// change; CI (dist-check.yml) enforces freshness if that step is forgotten.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

// projectInfo is a trimmed copy of Wails' internal/project Project for the
// fields that templates may reference.
type projectInfo struct {
	Name           string  `json:"name"`
	OutputFilename string  `json:"outputfilename"`
	Info           appInfo `json:"info"`
	InfoDefaults   appInfo `json:"-"`
}

type appInfo struct {
	CompanyName    string `json:"companyName"`
	ProductName    string `json:"productName"`
	ProductVersion string `json:"productVersion"`
	Copyright      string `json:"copyright"`
	Comments       string `json:"comments"`
}

type assetData struct {
	Name           string
	Info           appInfo
	OutputFilename string
}

var archs = map[string]winres.Arch{
	"amd64": winres.ArchAMD64,
	"arm64": winres.ArchARM64,
	"386":   winres.ArchI386,
}

// defaultArchs is the set generated when no arch argument is given: every
// Windows architecture DuMD ships or supports via `go install`.
var defaultArchs = []string{"amd64", "arm64"}

func main() {
	archFlags := defaultArchs
	if len(os.Args) > 1 {
		archFlags = os.Args[1:]
	}
	for _, archFlag := range archFlags {
		arch, ok := archs[archFlag]
		if !ok {
			fmt.Fprintf(os.Stderr, "resgen: unsupported arch %q (want amd64|arm64|386)\n", archFlag)
			os.Exit(2)
		}
		if err := generate(archFlag, arch); err != nil {
			fatal(err)
		}
	}
}

func generate(archFlag string, arch winres.Arch) error {
	project, err := loadProject("wails.json")
	if err != nil {
		return err
	}
	project.setDefaults()

	data := &assetData{
		Name:           project.Name,
		Info:           project.Info,
		OutputFilename: project.OutputFilename,
	}

	rs := winres.ResourceSet{}

	// Icon
	icoPath := filepath.Join("build", "windows", "icon.ico")
	iconFile, err := os.Open(icoPath)
	if err != nil {
		return err
	}
	defer iconFile.Close()
	ico, err := winres.LoadICO(iconFile)
	if err != nil {
		return fmt.Errorf("couldn't load icon from icon.ico: %w", err)
	}
	if err := rs.SetIcon(winres.RT_ICON, ico); err != nil {
		return err
	}

	// Manifest
	manifest, err := resolveTemplate(filepath.Join("build", "windows", "wails.exe.manifest"), data)
	if err != nil {
		return err
	}
	xmlData, err := winres.AppManifestFromXML(manifest)
	if err != nil {
		return err
	}
	rs.SetManifest(xmlData)

	// Version info
	versionInfo, err := resolveTemplate(filepath.Join("build", "windows", "info.json"), data)
	if err != nil {
		return err
	}
	if len(versionInfo) != 0 {
		var v version.Info
		if err := v.UnmarshalJSON(versionInfo); err != nil {
			return err
		}
		rs.SetVersionInfo(v)
	}

	// Deliberately distinct from Wails' ephemeral <name>-res.syso — see the
	// package doc comment above.
	name := strings.ReplaceAll(strings.ToLower(project.Name), " ", "_")
	targetDir := filepath.Join("internal", "winres")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(targetDir, fmt.Sprintf("%s_windows_%s.syso", name, archFlag))
	fout, err := os.Create(target)
	if err != nil {
		return err
	}
	defer fout.Close()
	if err := rs.WriteObject(fout, arch); err != nil {
		return err
	}
	fmt.Printf("resgen: wrote %s\n", target)
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "resgen:", err)
	os.Exit(1)
}

func loadProject(path string) (*projectInfo, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p := &projectInfo{}
	if err := json.Unmarshal(raw, p); err != nil {
		return nil, err
	}
	return p, nil
}

// setDefaults mirrors Wails' project.setDefaults for the fields templates
// may reference. OutputFilename is always forced to the Windows semantics
// (".exe" suffix) so the generated syso is host-OS independent.
func (p *projectInfo) setDefaults() {
	if p.Name == "" {
		p.Name = "wailsapp"
	}
	if p.OutputFilename == "" {
		p.OutputFilename = p.Name
	}
	if !strings.HasSuffix(p.OutputFilename, ".exe") {
		p.OutputFilename += ".exe"
	}
	if p.Info.CompanyName == "" {
		p.Info.CompanyName = p.Name
	}
	if p.Info.ProductName == "" {
		p.Info.ProductName = p.Name
	}
	if p.Info.ProductVersion == "" {
		p.Info.ProductVersion = "1.0.0"
	}
}

func resolveTemplate(path string, data *assetData) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Normalize line endings so the generated syso is identical
	// regardless of the host's core.autocrlf setting.
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	tmpl, err := template.New("").Parse(string(content))
	if err != nil {
		return nil, err
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, err
	}
	return []byte(out.String()), nil
}
