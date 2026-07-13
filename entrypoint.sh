#!/bin/sh
set -e

PUID=${PUID:-1000}
PGID=${PGID:-1000}

# Ensure data dir is owned by the target user so bind-mount works with any host UID/GID
chown -R "${PUID}:${PGID}" /data

exec su-exec "${PUID}:${PGID}" /usr/bin/shiori "$@"
