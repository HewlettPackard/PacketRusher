#---------------------------------------------------------------------
# BUILDER IMAGE
#---------------------------------------------------------------------
ARG BASE_IMAGE=ubuntu:focal
FROM $BASE_IMAGE as builder

RUN apt-get update && apt install wget -y

RUN wget https://go.dev/dl/go1.21.3.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

COPY . /packetrusher

RUN cd /packetrusher \
    && go mod download \
    && cd cmd && go build -o app

#---------------------------------------------------------------------
# TARGET IMAGE
#---------------------------------------------------------------------
ARG BASE_IMAGE=ubuntu:focal
FROM $BASE_IMAGE AS packetrusher

RUN apt update && apt install iproute2 iputils-ping iperf3 -y

WORKDIR /usr/local
COPY --from=builder /usr/local/go .
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /packetrusher/config/
COPY --from=builder /packetrusher/docker/config.yml .

WORKDIR /packetrusher/cmd
COPY --from=builder /packetrusher/cmd/app .
COPY --from=builder /packetrusher/docker/entrypoint.sh .

ENTRYPOINT ["/packetrusher/cmd/entrypoint.sh"]
