FROM golang AS builder

RUN apt-get update
RUN apt-get install -y libopus-dev gcc
WORKDIR /build

COPY . .
RUN GOOS=linux GOARCH=arm GOARM=7 go build

FROM debian:buster AS installedDependencies
RUN apt-get update
RUN apt-get install -y youtube-dl ffmpeg

FROM installedDependencies
COPY --from=builder /build/mumblebot /usr/bin

CMD /usr/bin/mumblebot --server $MUMBLE_SERVER --username $MUMBLE_USERNAME --insecure