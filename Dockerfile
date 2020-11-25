FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
RUN apk add opus-dev gcc musl-dev
WORKDIR /build

COPY . .
# RUN go get -d -v -u
RUN go build

FROM carlosra97/youtube-dl-alpine:latest

COPY --from=builder /build/mumblebot /usr/bin

ENV MUMBLE_OPTIONS="--insecure"

CMD /usr/bin/mumblebot --server $MUMBLE_SERVER --username $MUMBLE_USERNAME $MUMBLE_OPTIONS