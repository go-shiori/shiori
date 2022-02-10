# build stage
FROM ghcr.io/ghcri/golang:1.17-alpine3.15 AS builder
RUN apk add --no-cache build-base
WORKDIR /src
COPY . .
RUN go build

# server image
LABEL org.opencontainers.image.source https://github.com/go-shiori/shiori
FROM ghcr.io/ghcri/alpine:3.15
COPY --from=builder /src/shiori /usr/local/bin/
ENV SHIORI_DIR /srv/shiori/
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/shiori"]
CMD ["serve"]
