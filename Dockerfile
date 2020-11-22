FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
RUN apk add opus-dev gcc musl-dev
WORKDIR /build

COPY . .
RUN go build
RUN ls

FROM qmcgaw/youtube-dl-alpine:latest

COPY --from=builder /build/mumblebot /usr/bin
ENV server ""
ENV username "MusicBot"
ENV password ""
ENV cert ""


CMD [ "/usr/bin/mumblebot", "--server ${server}", "--username ${username}", "--insecure" ]