#---------------------------------------------------------------------
# BUILDER IMAGE
#---------------------------------------------------------------------
ARG BASE_IMAGE=ubuntu:focal
FROM $BASE_IMAGE as builder

RUN apt-get update && apt install wget -y

RUN wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz && tar -xvf go1.14.4.linux-amd64.tar.gz && mv go /usr/local
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

COPY . /my5G-RANTester

RUN cd /my5G-RANTester \
    && go mod download \
    && cd cmd && go build -o app

#---------------------------------------------------------------------
# TARGET IMAGE
#---------------------------------------------------------------------
ARG BASE_IMAGE=ubuntu:focal
FROM $BASE_IMAGE AS my5grantester

RUN apt update && apt install iproute2 iputils-ping iperf3 -y

WORKDIR /usr/local
COPY --from=builder /usr/local/go .
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /my5G-RANTester/config/
COPY --from=builder /my5G-RANTester/docker/config.yml .

WORKDIR /my5G-RANTester/cmd
COPY --from=builder /my5G-RANTester/cmd/app .
COPY --from=builder /my5G-RANTester/docker/entrypoint.sh .

ENTRYPOINT ["/my5G-RANTester/cmd/entrypoint.sh"]
