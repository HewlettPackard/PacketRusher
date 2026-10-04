ARG GO_VERSION=1.26.2
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /src
# The build context has no .git: pass --build-arg VERSION=<tag or commit> to identify the build.
ARG VERSION
COPY . .
# Local module replacements must be present when Go resolves dependencies.
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=${VERSION}" -o /out/packetrusher ./cmd

FROM debian:bookworm-slim AS packetrusher
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
       ca-certificates iproute2 iputils-ping iperf3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /packetrusher
COPY --from=builder /out/packetrusher /packetrusher/packetrusher
COPY docker/config.yml /packetrusher/config/config.yml

STOPSIGNAL SIGINT
ENTRYPOINT ["/packetrusher/packetrusher"]
CMD ["--help"]
