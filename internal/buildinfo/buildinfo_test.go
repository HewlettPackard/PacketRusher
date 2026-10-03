// SPDX-License-Identifier: Apache-2.0
package buildinfo

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestDevelopmentBuildUsesRevisionAndDirtyState(t *testing.T) {
	b := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}, Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
		{Key: "vcs.modified", Value: "true"}, {Key: "vcs.time", Value: "2026-10-03T00:00:00Z"},
	}}
	info := derive(b, "", "", "", "")
	if info.Version != "devel" || !info.Modified || info.BuildTime != "2026-10-03T00:00:00Z" || info.String() != "devel (0123456789ab) dirty" {
		t.Fatalf("unexpected metadata: %+v %s", info, info.String())
	}
}

func TestInjectedReleaseVersionWorksWithoutGitMetadata(t *testing.T) {
	info := derive(nil, "20261003", "abcdef0123456789abcdef0123456789abcdef01", "2026-10-03T00:00:00Z", "false")
	if info.String() != "20261003 (abcdef012345)" || info.Modified || info.BuildTime == "" {
		t.Fatalf("unexpected release: %+v", info)
	}
	unknown := derive(nil, "", "", "", "")
	if unknown.Version != "devel" || unknown.Revision != "unknown" || strings.Contains(unknown.String(), "1.0.1") {
		t.Fatalf("invented build identity: %+v", unknown)
	}
}

func TestModuleVersionsAndDependencyReplacements(t *testing.T) {
	b := &debug.BuildInfo{Main: debug.Module{Version: "v0.2.0"}, Deps: []*debug.Module{
		{Path: "github.com/free5gc/go-gtp5gnl", Version: "v1.6.2", Replace: &debug.Module{Version: "v1.6.3"}},
	}}
	info := derive(b, "", "", "", "")
	if info.Version != "v0.2.0" || info.GTP5GVersion != "v1.6.3" {
		t.Fatalf("wrong module provenance: %+v", info)
	}
}

func TestInjectedDirtyStateUsesTheSelectedCheckout(t *testing.T) {
	for _, modified := range []string{"true", "false"} {
		outer := &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "true"}}}
		info := derive(outer, "devel", "abcdef0123456789abcdef0123456789abcdef01", "", modified)
		if info.Modified != (modified == "true") {
			t.Fatalf("wrong checkout dirty state: %+v", info)
		}
	}
}
