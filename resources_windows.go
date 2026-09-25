//go:build windows && !dev && !desktop

// Links the committed Windows resource object (icon, manifest, version info)
// into plain `go build` / `go install` binaries. Wails builds (tags
// `desktop`/`dev`) compile and link their own <name>-res.syso; linking both
// fails with "too many .rsrc sections".
package main

import _ "github.com/ebdonato/dumd/internal/winres"
