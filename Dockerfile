FROM golang AS builder

RUN apt-get update
RUN apt-get install -y libopus-dev gcc
WORKDIR /build

COPY . .
RUN GOOS=linux GOARCH=arm GOARM=7 go build

FROM debian:buster
RUN apt-get update
RUN apt-get install -y youtube-dl ffmpeg

COPY --from=builder /build/mumblebot /usr/bin

ENV server ""
ENV username "MusicBot"
ENV password ""
ENV cert ""


CMD [ "/usr/bin/mumblebot", "--server ${server}", "--username ${username}", "--insecure" ]