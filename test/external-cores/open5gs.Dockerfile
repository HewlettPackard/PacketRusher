# Source and required fallback branches are pinned, including the upstream
# prometheus wrap that otherwise follows a moving branch.
FROM ubuntu:24.04@sha256:a853f94d226358a79c740cfc7bce0c289748f3fe3488d921d038ccd752c61b60
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    python3 python3-setuptools python3-wheel ninja-build build-essential flex bison git cmake meson \
    libsctp-dev libgnutls28-dev libgcrypt-dev libssl-dev libmongoc-dev libbson-dev libyaml-dev \
    libmicrohttpd-dev libcurl4-gnutls-dev libnghttp2-dev libtins-dev libtalloc-dev libidn-dev \
    iproute2 tcpdump ca-certificates netbase pkg-config \
    && rm -rf /var/lib/apt/lists/*
RUN git init /src && cd /src && git remote add origin https://github.com/open5gs/open5gs.git \
    && git fetch --depth=1 origin 157f611a530e292e40ec50f9d23f0ef5d4fcd6a6 \
    && git checkout --detach FETCH_HEAD \
    && test "$(git rev-parse HEAD)" = 157f611a530e292e40ec50f9d23f0ef5d4fcd6a6 \
    && sed -i 's/^revision =.*/revision = 14725af3ba0edbf9ff61c4e3239ed42464423b2e/' subprojects/freeDiameter.wrap \
    && sed -i 's/^revision =.*/revision = a58ba25bf87a9b1b7c6be4e6f4c62047d620f402/' subprojects/prometheus-client-c.wrap \
    && meson setup build --prefix=/opt/open5gs --buildtype=release \
    && ninja -C build -j2 install
COPY test/external-cores/*.py /probe/
STOPSIGNAL SIGTERM
