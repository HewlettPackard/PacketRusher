// SPDX-License-Identifier: Apache-2.0
package buildinfo

import (
	"runtime"
	"runtime/debug"
	"strconv"
)

// Release builds inject these values with -ldflags. Ordinary go builds use VCS
// metadata embedded by the Go toolchain, without requiring git at runtime.
var Version, Revision, BuildTime, Modified string

type Info struct {
	Version      string `json:"version"`
	Revision     string `json:"revision"`
	Modified     bool   `json:"modified"`
	BuildTime    string `json:"build_time,omitempty"`
	GoVersion    string `json:"go_version"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	GTP5GVersion string `json:"gtp5gnl_version,omitempty"`
}

func Current() Info {
	b, _ := debug.ReadBuildInfo()
	return derive(b, Version, Revision, BuildTime, Modified)
}

func derive(b *debug.BuildInfo, version, revision, built, modified string) Info {
	info := Info{Version: version, Revision: revision, BuildTime: built, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	if b != nil {
		if b.GoVersion != "" {
			info.GoVersion = b.GoVersion
		}
		if info.Version == "" && b.Main.Version != "" && b.Main.Version != "(devel)" {
			info.Version = b.Main.Version
		}
		for _, setting := range b.Settings {
			switch setting.Key {
			case "vcs.revision":
				if info.Revision == "" {
					info.Revision = setting.Value
				}
			case "vcs.modified":
				info.Modified, _ = strconv.ParseBool(setting.Value)
			case "vcs.time":
				if info.BuildTime == "" {
					info.BuildTime = setting.Value
				}
			}
		}
		for _, dep := range b.Deps {
			if dep.Path == "github.com/free5gc/go-gtp5gnl" {
				info.GTP5GVersion = dep.Version
				if dep.Replace != nil && dep.Replace.Version != "" {
					info.GTP5GVersion = dep.Replace.Version
				}
			}
		}
	}
	if modified != "" {
		info.Modified, _ = strconv.ParseBool(modified)
	}
	if info.Revision == "" {
		info.Revision = "unknown"
	}
	if info.Version == "" {
		info.Version = "devel"
	}
	return info
}

func (info Info) String() string {
	version := info.Version
	if info.Revision != "unknown" && info.Revision != "" {
		revision := info.Revision
		if len(revision) > 12 {
			revision = revision[:12]
		}
		version += " (" + revision + ")"
	}
	if info.Modified {
		version += " dirty"
	}
	return version
}
