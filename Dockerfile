FROM alpine:latest

COPY ./dist/gypsum_linux_amd64/gypsum /usr/bin/gypsum

WORKDIR /gypsum

ENTRYPOINT [ "/usr/bin/gypsum" ]
