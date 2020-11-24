FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
RUN apk add opus-dev gcc musl-dev
WORKDIR /build

COPY . .
RUN go build

FROM qmcgaw/youtube-dl-alpine:latest

COPY --from=builder /build/mumblebot /usr/bin

CMD /usr/bin/mumblebot --server $MUMBLE_SERVER --username $MUMBLE_USERNAME --insecure