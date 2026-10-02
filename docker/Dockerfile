ARG GO_VERSION=1.26.2
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /src
COPY . .
# Local module replacements must be present when Go resolves dependencies.
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -o /out/packetrusher ./cmd

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
