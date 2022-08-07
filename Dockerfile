FROM alpine:latest

ENTRYPOINT ["/usr/bin/115drive-webdav"]

COPY 115drive-webdav /usr/bin/115drive-webdav
