FROM debian:latest

RUN apt-get update && \
    apt-get install -y wget && \
    rm -rf /var/lib/apt/lists/* && \
    wget https://raw.githubusercontent.com/gaoyb7/115drive-webdav/main/115/libencode115.so -O /usr/lib/libencode115.so

COPY 115drive-webdav /usr/bin/115drive-webdav

ENTRYPOINT ["/usr/bin/115drive-webdav"]
