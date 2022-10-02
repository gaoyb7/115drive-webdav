FROM alpine:latest

COPY 115drive-webdav /usr/bin/115drive-webdav

ENTRYPOINT ["/usr/bin/115drive-webdav"]
