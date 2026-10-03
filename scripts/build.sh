#!/usr/bin/env bash
set -euo pipefail

output=${1:-packetrusher}
package=my5G-RANTester/internal/buildinfo
build_version=${PACKETRUSHER_VERSION:-}
build_revision=${PACKETRUSHER_REVISION:-}
build_time=${PACKETRUSHER_BUILD_TIME:-}
build_modified=false
if [[ -n "$(git status --porcelain --untracked-files=all 2>/dev/null || true)" ]]; then
    build_modified=true
fi
if [[ -z "$build_version" ]]; then
    build_version=$(git describe --tags --exact-match 2>/dev/null || true)
fi
if [[ -z "$build_revision" ]]; then
    build_revision=$(git rev-parse --verify HEAD 2>/dev/null || true)
fi
if [[ -z "$build_time" ]]; then
    build_time=$(git show -s --format=%cI HEAD 2>/dev/null || true)
fi
build_version=${build_version:-devel}
build_revision=${build_revision:-unknown}
if [[ ! "$build_version" =~ ^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$ ]]; then
    echo 'PACKETRUSHER_VERSION must be a simple tag or version identifier' >&2
    exit 2
fi
if [[ "$build_revision" != unknown && ! "$build_revision" =~ ^[0-9a-fA-F]{40,64}$ ]]; then
    echo 'PACKETRUSHER_REVISION must be a full source commit hash' >&2
    exit 2
fi
if [[ -n "$build_time" && ! "$build_time" =~ ^[0-9T:+Z.-]+$ ]]; then
    echo 'PACKETRUSHER_BUILD_TIME must be an ISO-8601 timestamp' >&2
    exit 2
fi
flags="-s -w -buildid= -X $package.Version=$build_version -X $package.Revision=$build_revision -X $package.Modified=$build_modified"
if [[ -n "$build_time" ]]; then flags+=" -X $package.BuildTime=$build_time"; fi
# The explicit checkout identity also handles worktrees nested in another repo.
go build -buildvcs=false -trimpath -ldflags "$flags" -o "$output" ./cmd
