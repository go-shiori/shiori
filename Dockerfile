FROM golang:1.10-alpine as builder

RUN apk update \
  && apk --no-cache add git build-base

WORKDIR /go/src/github.com/RadhiFadlillah/shiori
COPY . .
RUN bin/setup
RUN bin/build

FROM alpine:latest

ENV ENV_SHIORI_DB /srv/shiori.db

RUN apk --no-cache add dumb-init ca-certificates
COPY --from=builder /go/src/github.com/RadhiFadlillah/shiori/shiori_linux_amd64 /usr/local/bin/shiori

WORKDIR /srv/
RUN touch shiori.db

EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve"]

