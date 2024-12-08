# Build stage
ARG ALPINE_VERSION=3.19

FROM docker.io/library/alpine:${ALPINE_VERSION} AS builder
ARG TARGETARCH
ARG TARGETOS
ARG TARGETVARIANT
COPY dist/shiori_${TARGETOS}_${TARGETARCH}${TARGETVARIANT}/shiori /usr/bin/shiori
RUN apk add --no-cache ca-certificates tzdata && \
    chmod +x /usr/bin/shiori && \
    rm -rf /tmp/*

# Server image
FROM scratch

ENV PORT=8080
ENV SHIORI_DIR=/shiori
WORKDIR ${SHIORI_DIR}

LABEL org.opencontainers.image.source="https://github.com/go-shiori/shiori"
LABEL maintainer="Felipe Martin <github@fmartingr.com>"

COPY --from=builder /tmp /tmp
COPY --from=builder /usr/bin/shiori /usr/bin/shiori
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE ${PORT}

ENTRYPOINT ["/usr/bin/shiori"]
CMD ["server"]
