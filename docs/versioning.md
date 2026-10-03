# Build identity and releases

`packetrusher --version` prints the version and source revision. For toolchain,
target architecture and linked GTP library information, use `packetrusher version`
or `packetrusher version --json`. These commands need no configuration or core
connection and do not start metrics or reserve report files.

Ordinary `go build ./cmd` builds report the VCS revision and dirty state embedded
by Go. Unversioned builds identify themselves as `devel`; builds without source
metadata report an unknown revision. They do not advertise a fixed release.
`make` uses `scripts/build.sh`, selecting an exact checked-out tag when available.
The build script accepts `PACKETRUSHER_VERSION`, `PACKETRUSHER_REVISION` and
`PACKETRUSHER_BUILD_TIME` for source archives or Docker builds without `.git`.

## Tagged release pipeline

The release workflow runs when a `v` version tag or an eight-digit date tag is
pushed. It builds Linux amd64/arm64 archives, verifies SHA-256 checksums, and
publishes a release plus a matching multi-platform GHCR image. Tags containing
a hyphen are marked as prereleases. Images use the exact tag and include source
revision/version OCI labels. No mutable `latest` image is published.

Each archive contains the CLI, example config, README, project license and the
patched NAS library's license/provenance. NAS replacement source remains in the
repository. The Docker image retains the host SCTP/TUN/gtp5g requirements of the
selected tunnel backend.

`scripts/release.sh` only packages a clean, checked-out commit matching the named
tag and revision. It disables implicit VCS metadata and injects that identity
explicitly. Archive order, ownership and timestamps derive from the source
commit so repeating a build with the same source, Go toolchain and architecture
produces the same archive and checksum.

```sh
PACKETRUSHER_VERSION=v0.2.0 GOARCH=amd64 ./scripts/release.sh
sha256sum --check dist/SHA256SUMS-amd64
```

For a source checkout without a published tag, build the development CLI with
`make`; release packaging intentionally requires an existing matching local tag.
The pipeline definitions are prepared locally and will run only after a tag is
pushed to a repository containing the workflow.
