#!/usr/bin/env sh

SERVER_HOST=${SERVER_HOST:-"localhost"}
SERVER_PORT=${SERVER_PORT:-"16000"}
MYSQL_SERVER_ADDR=${MYSQL_SERVER_ADDR:-"localhost:3306"}
MYSQL_SERVER_ROOT_PASSWORD=${MYSQL_SERVER_ROOT_PASSWORD:-""}
REDIS_SERVER_ADDR=${REDIS_SERVER_ADDR:-"localhost:6379"}
REDIS_SERVER_ADMIN_PASSWORD=${REDIS_SERVER_ADMIN_PASSWORD:-""}

./server \
  -server_host="${SERVER_HOST}" \
  -server_port="${SERVER_PORT}" \
  -mysql_server_addr="${MYSQL_SERVER_ADDR}" \
  -mysql_server_root_password="${MYSQL_SERVER_ROOT_PASSWORD}" \
  -redis_server_addr="${REDIS_SERVER_ADDR}" \
  -redis_server_admin_password="${REDIS_SERVER_ADMIN_PASSWORD}"
