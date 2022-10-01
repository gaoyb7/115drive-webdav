FROM alpine:latest

ENTRYPOINT ["/usr/bin/115drive-webdav"]

COPY 115drive-webdav /usr/bin/115drive-webdav

RUN wget https://raw.githubusercontent.com/gaoyb7/115drive-webdav/main/115/libencode115.so -O /lib64/libencode115.so
