#!/bin/sh
# entrypoint.server.sh — App container entrypoint.
# Sets default env vars, runs envsubst to populate config/config.toml from the
# template, then exec's the CMD unchanged (transparent wrapper).
set -e

# ── Defaults ──────────────────────────────────────────────────────────────────
export POSTGRES_PORT="${POSTGRES_PORT:-5432}"
export POSTGRES_DB="${POSTGRES_DB:-db}"
export POSTGRES_CONNECTION_STRING="${POSTGRES_CONNECTION_STRING:-}"
export POSTGRES_USER="${POSTGRES_USER:-}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"

export REDIS_SOCKET="${REDIS_SOCKET:-}"
export REDIS_HOST="${REDIS_HOST:-localhost}"
export REDIS_PORT="${REDIS_PORT:-6379}"
export REDIS_USERNAME="${REDIS_USERNAME:-}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# ── Config generation ─────────────────────────────────────────────────────────
TMPL="/etc/mindforge/config.toml.tmpl"
OUT="/etc/mindforge/config.toml"

if [ -f "${TMPL}" ]; then
  echo "→ generating ${OUT} from ${TMPL}"
  mkdir -p "$(dirname "${OUT}")"
  envsubst < "${TMPL}" > "${OUT}"
else
  echo "⚠  config template not found at ${TMPL}, skipping"
fi

exec "$@"
