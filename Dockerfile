FROM golang:alpine AS builder

RUN apk upgrade && apk add --upgrade opus-dev gcc musl-dev git
WORKDIR /build

COPY . .
# RUN go get -d -v -u
RUN go build

FROM alpine:3.18.3

RUN apk update && apk add --upgrade yt-dlp opus ffmpeg

COPY --from=builder /build/mumblebot /usr/bin
COPY entrypoint.sh /usr/bin

RUN chmod +x /usr/bin/entrypoint.sh

ENV MUMBLE_OPTIONS="--insecure"
ENV AUTO_UPDATE_YDLP=1

CMD entrypoint.sh