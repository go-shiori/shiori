FROM golang:alpine AS builder
ENV LDFLAGS="-s -W"
RUN apk add --no-cache build-base upx
WORKDIR /src
COPY . .
RUN go build
RUN upx --lzma /src/shiori
# server image
FROM alpine:3.12
# Make us secure every build if base image missing security fixes
RUN apk update && apk upgrade
COPY --from=builder /src/shiori /usr/local/bin/
ENV SHIORI_DIR /srv/shiori/
EXPOSE 8080
RUN adduser --disabled-password --shell /bin/ash --gecos "User" shiori && mkdir /srv/shiori && chown -R shiori:shiori /srv/shiori
USER shiori
CMD ["/usr/local/bin/shiori", "serve"]
