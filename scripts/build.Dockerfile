FROM node:14 AS web_builder

WORKDIR /build

RUN set -ex \
    && git clone https://github.com/yuudi/gypsum-web.git . --depth=1 \
    && yarn install \
    && yarn build

FROM golang:1.16rc1 AS builder

COPY --from=web_builder /build/dist/web /tmp/web

WORKDIR /build

RUN set -ex \
    && git clone https://github.com/yuudi/gypsum.git . --depth=1 \
    && mv /tmp/web ./gypsum/web \
    && export GO111MODULE=on \
    && export CGO_ENABLED=0 \
    && go generate ./... \
    && go build -trimpath -tags=jsoniter -ldflags="-s -w" -o dist/gypsum .

FROM alpine:latest

COPY --from=builder /build/dist/gypsum /usr/bin/gypsum

WORKDIR /gypsum

ENTRYPOINT [ "/usr/bin/gypsum" ]
