FROM alpine:latest

ENTRYPOINT ["/115drive-webdav"]

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY 115drive-webdav /
