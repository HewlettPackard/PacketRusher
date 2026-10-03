ARG GO_VERSION=1.26.2
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /src
ARG PACKETRUSHER_VERSION=devel
ARG PACKETRUSHER_REVISION=unknown
ARG PACKETRUSHER_BUILD_TIME
COPY . .
# Local module replacements must be present when Go resolves dependencies.
RUN go mod download
RUN CGO_ENABLED=0 ./scripts/build.sh /out/packetrusher

FROM debian:bookworm-slim AS packetrusher
ARG PACKETRUSHER_VERSION=devel
ARG PACKETRUSHER_REVISION=unknown
ARG PACKETRUSHER_BUILD_TIME
LABEL org.opencontainers.image.source="https://github.com/HewlettPackard/PacketRusher" \
      org.opencontainers.image.version=$PACKETRUSHER_VERSION \
      org.opencontainers.image.revision=$PACKETRUSHER_REVISION \
      org.opencontainers.image.created=$PACKETRUSHER_BUILD_TIME
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
       ca-certificates iproute2 iputils-ping iperf3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /packetrusher
COPY --from=builder /out/packetrusher /packetrusher/packetrusher
COPY docker/config.yml /packetrusher/config/config.yml
COPY LICENSE /packetrusher/LICENSE
COPY third_party/free5gc-nas/LICENSE third_party/free5gc-nas/UPSTREAM.json /packetrusher/licenses/free5gc-nas/

STOPSIGNAL SIGINT
ENTRYPOINT ["/packetrusher/packetrusher"]
CMD ["--help"]
