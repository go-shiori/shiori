# build stage
FROM golang:1.15.2-alpine AS builder

WORKDIR /src
RUN adduser \
    --disabled-password \    
    --uid 10001 shioriuser
RUN apk --update add \
	ca-certificates \
	musl-dev \
	gcc
COPY . .
RUN go generate ./... && CGO_ENABLED=1 go build -a -ldflags '-linkmode external -extldflags "-static" -s -w' .

# server image
FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /src/shiori /usr/local/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER shioriuser
ENV SHIORI_DIR /srv/shiori/
EXPOSE 8080
CMD ["/usr/local/bin/shiori", "serve"]
