#!/bin/sh
# entrypoint.postgre.sh — PostgreSQL init script (runs via docker-entrypoint-initdb.d)
# Creates the application database if it doesn't already exist.
set -e

DB="${POSTGRES_DB:-db}"

echo "→ ensuring database '${DB}' exists"
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" <<-EOSQL
  SELECT 'CREATE DATABASE "${DB}"'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB}')
  \gexec
EOSQL
echo "→ database '${DB}' ready"
