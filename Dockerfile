FROM golang:alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /src
COPY . .
RUN go build

# server image
FROM golang:alpine
COPY --from=builder /src/shiori /usr/local/bin/
ENV SHIORI_DIR /srv/shiori/
EXPOSE 8080
RUN adduser --disabled-password --shell /bin/ash --gecos "User" shiori && mkdir /srv/shiori && chown -R shiori:shiori /srv/shiori
USER shiori
CMD ["/usr/local/bin/shiori", "serve"]
