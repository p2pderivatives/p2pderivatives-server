FROM golang:1.14-alpine as dev
RUN apk update
RUN apk add make gcc libc-dev git wget unzip protobuf-dev openssl

ENV GO111MODULE=on

WORKDIR /opt/p2pserver

COPY go.mod go.sum ./
RUN go mod download

COPY ./tools ./tools
COPY Makefile .
RUN make install-tools

COPY . /opt/p2pserver/

FROM dev as build
RUN make gen
RUN make server

EXPOSE 8080

CMD ["make", "run-server-local"]
