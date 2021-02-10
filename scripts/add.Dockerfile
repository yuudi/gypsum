FROM alpine:latest

ADD ./gypsum /usr/bin/gypsum

WORKDIR /gypsum

ENTRYPOINT [ "/usr/bin/gypsum" ]
