FROM alpine:latest

WORKDIR /gypsum

RUN wget -q -O- https://api.github.com/repos/yuudi/gypsum/releases/latest \
    | grep browser_download_url.*linux-x86_64 \
    | cut -f 4 -d "\"" \
    | wget -q -O- $(cat /dev/stdin) \
    | tar zxf - -C /usr/bin

ENTRYPOINT [ "/usr/bin/gypsum" ]
