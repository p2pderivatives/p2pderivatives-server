FROM golang:1.14-alpine

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
RUN make install
RUN make gen-proto
RUN make server

CMD ["make", "run-server-local"]
