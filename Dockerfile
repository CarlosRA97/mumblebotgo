FROM golang AS builder

RUN apt update
RUN apt install -y libopus-dev gcc
WORKDIR /build

COPY . .
RUN GOOS=linux GOARCH=arm GOARM=7 go build
RUN ls

FROM debian:buster
RUN apt update
RUN apt install -y youtube-dl ffmpeg

COPY --from=builder /build/mumblebot /usr/bin

ENV server ""
ENV username "MusicBot"
ENV password ""
ENV cert ""


CMD [ "/usr/bin/mumblebot", "--server ${server}", "--username ${username}", "--insecure" ]