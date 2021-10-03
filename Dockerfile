FROM golang:1.17 as builder

RUN mkdir /build
COPY . /build
RUN cd /build && go mod download && CGO_ENABLED=0 go build -o /build/citadel487-bot .


FROM ubuntu:focal

ENV DEBIAN_FRONTEND noninteractive

RUN mkdir -p /opt/app && \
    apt update && \
    apt install -y ca-certificates ffmpeg && \
    apt clean

COPY --from=builder /build/citadel487-bot /opt/app
COPY downloader487/downloader/dist/downloader487 /opt/app

ENTRYPOINT [ "/opt/app/citadel487-bot", "--downloader-path", "/opt/app/downloader487" ]
