#!/bin/sh
# entrypoint.redis.sh — Redis startup with optional port, password, and ACL user from env.
set -e

REDIS_PORT="${REDIS_PORT:-6379}"
CONF="/tmp/redis.conf"

echo "→ generating redis.conf (port=${REDIS_PORT})"

cat > "${CONF}" <<EOF
port ${REDIS_PORT}
dir /data
EOF

# Require a password when REDIS_PASSWORD is set.
if [ -n "${REDIS_PASSWORD}" ]; then
  echo "requirepass ${REDIS_PASSWORD}" >> "${CONF}"
fi

# Add an ACL entry when both username and password are provided.
# This allows the app to authenticate with AUTH <username> <password>.
if [ -n "${REDIS_USERNAME}" ] && [ -n "${REDIS_PASSWORD}" ]; then
  echo "aclfile /tmp/redis.acl" >> "${CONF}"
  cat > /tmp/redis.acl <<EOACL
user default off
user ${REDIS_USERNAME} on >${REDIS_PASSWORD} ~* &* +@all
EOACL
  echo "→ ACL user '${REDIS_USERNAME}' configured"
fi

echo "→ starting redis-server on port ${REDIS_PORT}"
exec redis-server "${CONF}"
