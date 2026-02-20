#!/bin/sh
set -e

# Fix ownership of mounted volume directories at startup.
# Volumes are created root-owned by Docker; this ensures appuser can write.
mkdir -p /app/app/log /app/filedata
chown -R appuser:appuser /app/app/log /app/filedata

# Drop privileges and exec the app (gosu preserves signals correctly).
exec gosu appuser /app/main
