#!/usr/bin/env sh

SERVER_PORT=${SERVER_PORT:-80}
REDIRECT_SERVE_URL=${REDIRECT_SERVE_URL:-http://localhost}
MYSQL_SERVER_ADDR=${MYSQL_SERVER_ADDR:-localhost:3306}
MYSQL_SERVER_ROOT_PASSWORD=${MYSQL_SERVER_ROOT_PASSWORD:-}
REDIS_SERVER_ADDR=${REDIS_SERVER_ADDR:-localhost:6379}
REDIS_SERVER_ADMIN_PASSWORD=${REDIS_SERVER_ADMIN_PASSWORD:-}

./server \
  -server_port="${SERVER_PORT}" \
  -redirect_serve_url="${REDIRECT_SERVE_URL}" \
  -mysql_server_addr="${MYSQL_SERVER_ADDR}" \
  -mysql_server_root_password="${MYSQL_SERVER_ROOT_PASSWORD}" \
  -redis_server_addr="${REDIS_SERVER_ADDR}" \
  -redis_server_admin_password="${REDIS_SERVER_ADMIN_PASSWORD}"
