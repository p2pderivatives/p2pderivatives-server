FROM golang:1.13.5-alpine

EXPOSE 8080

RUN apk update
RUN apk add make gcc libc-dev git wget unzip protobuf-dev

ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV GOBIN=/root/go/bin
ENV PATH=$PATH:$GOBIN:$GOPATH:$GOROOT/bin

RUN mkdir -p /opt/p2pserver
ADD . / /opt/p2pserver/

WORKDIR /opt/p2pserver
RUN make setup

CMD ["make", "run-server-local"]
