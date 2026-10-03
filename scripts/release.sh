#!/usr/bin/env bash
set -euo pipefail

build_arch=${GOARCH:-$(go env GOARCH)}
case "$build_arch" in amd64|arm64) ;; *) echo 'Release architecture must be amd64 or arm64' >&2; exit 2 ;; esac
export GOARCH="$build_arch" GOOS=linux CGO_ENABLED=0
export PACKETRUSHER_VERSION=${PACKETRUSHER_VERSION:?Set PACKETRUSHER_VERSION to the release tag}
source_revision=$(git rev-parse HEAD)
export PACKETRUSHER_REVISION=${PACKETRUSHER_REVISION:-$source_revision}
if [[ "$PACKETRUSHER_REVISION" != "$source_revision" ]]; then
    echo 'Release revision must match the checked-out source commit' >&2
    exit 2
fi
if [[ -n "$(git status --porcelain)" ]]; then
    echo 'Release builds require a clean source worktree' >&2
    exit 2
fi
if [[ "$(git rev-parse --verify "refs/tags/$PACKETRUSHER_VERSION^{commit}")" != "$source_revision" ]]; then
    echo 'Release tag must point to the checked-out source commit' >&2
    exit 2
fi
export PACKETRUSHER_BUILD_TIME=${PACKETRUSHER_BUILD_TIME:-$(git show -s --format=%cI HEAD)}
# Release metadata is explicit and independent of generated/untracked artifacts.
export GOFLAGS="${GOFLAGS:-} -buildvcs=false"
build_epoch=$(git show -s --format=%ct HEAD)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
./scripts/build.sh "$work/packetrusher"
mkdir -p "$work/config" "$work/licenses/free5gc-nas" dist
cp config/config.yml "$work/config/"
cp LICENSE README.md "$work/"
cp third_party/free5gc-nas/LICENSE third_party/free5gc-nas/UPSTREAM.json "$work/licenses/free5gc-nas/"
archive="packetrusher-${PACKETRUSHER_VERSION}-linux-${build_arch}.tar.gz"
tar --sort=name --mtime="@$build_epoch" --owner=0 --group=0 --numeric-owner -C "$work" -czf "dist/$archive" .
(cd dist && sha256sum "$archive" > "SHA256SUMS-$build_arch")
