FROM golang:1.16-alpine3.12 as dev
RUN apk update
RUN apk add make gcc libc-dev git wget unzip protobuf-dev openssl

ENV GO111MODULE=on

WORKDIR /opt/p2pderivatives-server

COPY go.mod go.sum ./
RUN go mod download

COPY ./tools ./tools
COPY Makefile .
RUN make install-tools

COPY . .

RUN make gen

FROM dev as build
RUN make bin

FROM alpine as prod

WORKDIR /p2pdserver

# runtime dependencies
RUN apk update
RUN apk add libstdc++

RUN mkdir -p /config
COPY ./test/config/default.release.yml /config/default.yml
VOLUME [ "/config" ]

COPY --from=build /opt/p2pderivatives-server/bin/ /p2pdserver/

ENTRYPOINT [ "/p2pdserver/server" ]
CMD [ "-config", "/config", "-appname", "p2pdserver", "-e", "default", "-migrate" ]
