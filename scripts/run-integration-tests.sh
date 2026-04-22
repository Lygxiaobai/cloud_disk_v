#!/usr/bin/env bash
set -euo pipefail

MYSQL_PORT="${MYSQL_PORT:-3307}"
REDIS_PORT="${REDIS_PORT:-6380}"

cleanup() {
  docker compose -f deploy/docker-compose.integration.yml down -v >/dev/null 2>&1 || true
}

trap cleanup EXIT

docker compose -f deploy/docker-compose.integration.yml up -d

ready=0
for _ in $(seq 1 30); do
  if docker exec cloud-disk-mysql-test mysqladmin ping -h 127.0.0.1 -p123456 --silent >/dev/null 2>&1 && \
     docker exec cloud-disk-redis-test redis-cli ping | grep -q PONG; then
    ready=1
    break
  fi
  sleep 5
done

if [ "$ready" -ne 1 ]; then
  echo "integration dependencies did not become healthy" >&2
  exit 1
fi

export RUN_INTEGRATION_TESTS=1
export MYSQL_DSN_TEST="root:123456@tcp(127.0.0.1:${MYSQL_PORT})/cloud_disk?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR_TEST="127.0.0.1:${REDIS_PORT}"

go test ./core/internal/test -count=1
