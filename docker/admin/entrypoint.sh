#!/bin/sh
set -eu

cd /app/build

bun ./index.js &
bun_pid=$!

trap 'kill -TERM "$bun_pid" 2>/dev/null || true; wait "$bun_pid" 2>/dev/null || true' EXIT INT TERM

exec nginx -g 'daemon off;'
