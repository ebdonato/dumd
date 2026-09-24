// Command resgen compiles a committed Windows resource (.syso) for the DuMD
// binary so that `go install github.com/ebdonato/dumd@latest` embeds the icon,
// the application manifest and the version info — the same resources `wails
// build` injects through its temporary `<name>-res.syso`.
//
// It mirrors Wails' build/pkg/packager.go compileResources: templates in
// build/windows/{info.json,wails.exe.manifest} are resolved with the project
// data from wails.json, then bundled with build/windows/icon.ico into
// <name>-res.syso at the repository root. The name intentionally matches
// Wails' ephemeral file: two distinct .syso files would create two .rsrc
// sections and break the Windows linker ("too many .rsrc sections"). Wails
// simply overwrites this file with identical content during `wails dev` /
// `wails build`, and deletes it when finished (restore it with
// `go run ./tools/resgen` or `git restore -- <name>-res.syso`).
//
// Regenerate and re-commit <name>-res.syso whenever wails.json info or
// build/windows/{icon.ico,info.json,wails.exe.manifest} change. CI
// (dist-check.yml) enforces freshness. Additionally, after any local
// `wails dev` or `wails build` the file is deleted by the Wails CLI;
// regenerate or `git restore` it before committing.
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
	Name           string   `json:"name"`
	OutputFilename string   `json:"outputfilename"`
	Info           appInfo  `json:"info"`
	InfoDefaults   appInfo  `json:"-"`
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

func main() {
	archFlag := "amd64"
	if len(os.Args) > 1 {
		archFlag = os.Args[1]
	}
	arch, ok := archs[archFlag]
	if !ok {
		fmt.Fprintf(os.Stderr, "resgen: unsupported arch %q (want amd64|arm64|386)\n", archFlag)
		os.Exit(2)
	}

	project, err := loadProject("wails.json")
	if err != nil {
		fatal(err)
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
		fatal(err)
	}
	defer iconFile.Close()
	ico, err := winres.LoadICO(iconFile)
	if err != nil {
		fatal(fmt.Errorf("couldn't load icon from icon.ico: %w", err))
	}
	if err := rs.SetIcon(winres.RT_ICON, ico); err != nil {
		fatal(err)
	}

	// Manifest
	manifest, err := resolveTemplate(filepath.Join("build", "windows", "wails.exe.manifest"), data)
	if err != nil {
		fatal(err)
	}
	xmlData, err := winres.AppManifestFromXML(manifest)
	if err != nil {
		fatal(err)
	}
	rs.SetManifest(xmlData)

	// Version info
	versionInfo, err := resolveTemplate(filepath.Join("build", "windows", "info.json"), data)
	if err != nil {
		fatal(err)
	}
	if len(versionInfo) != 0 {
		var v version.Info
		if err := v.UnmarshalJSON(versionInfo); err != nil {
			fatal(err)
		}
		rs.SetVersionInfo(v)
	}

	// Must exactly match Wails' ephemeral filename (see compileResources in
	// build/pkg/packager.go) so only one .rsrc section reaches the linker.
	target := strings.ReplaceAll(project.Name, " ", "_") + "-res.syso"
	fout, err := os.Create(target)
	if err != nil {
		fatal(err)
	}
	defer fout.Close()
	if err := rs.WriteObject(fout, arch); err != nil {
		fatal(err)
	}
	fmt.Printf("resgen: wrote %s (%s)\n", target, archFlag)
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
