FROM alpine:latest

COPY /dist/cqhttp /usr/bin/gypsum

WORKDIR /gypsum

ENTRYPOINT [ "/usr/bin/gypsum" ]
