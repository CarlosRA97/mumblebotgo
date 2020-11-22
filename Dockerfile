FROM golang AS builder

RUN apt-get update
RUN apt-get install -y libopus-dev gcc
WORKDIR /build

COPY . .
RUN go build

FROM debian:buster AS installed_dependencies
RUN apt-get update
RUN apt-get install -y ffmpeg
RUN curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl
RUN chmod a+rx /usr/bin/youtube-dl

FROM installed_dependencies
COPY --from=builder /build/mumblebot /usr/bin

CMD /usr/bin/mumblebot --server $MUMBLE_SERVER --username $MUMBLE_USERNAME --insecure