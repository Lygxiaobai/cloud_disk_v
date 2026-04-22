#!/usr/bin/env bash

set -euo pipefail

log() {
  printf '[INFO] %s\n' "$*"
}

warn() {
  printf '[WARN] %s\n' "$*" >&2
}

fatal() {
  printf '[ERROR] %s\n' "$*" >&2
  exit 1
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    fatal "Please run this script with sudo or as root."
  fi
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

random_hex() {
  openssl rand -hex 32
}

ensure_line_in_file() {
  local file="$1"
  local line="$2"
  grep -Fqx "$line" "$file" 2>/dev/null || printf '%s\n' "$line" >>"$file"
}

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP_DIR="${APP_DIR:-$CURRENT_DIR}"
APP_USER="${APP_USER:-${SUDO_USER:-root}}"
APP_GROUP="${APP_GROUP:-$APP_USER}"
DOMAIN="${DOMAIN:-_}"
BACKEND_PORT="${BACKEND_PORT:-8888}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-}"
MYSQL_DB="${MYSQL_DB:-cloud_disk}"
MYSQL_USER="${MYSQL_USER:-cloud_disk}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-cloud_disk_ChangeMe_123456}"
REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
SERVER_PUBLIC_HOST="${SERVER_PUBLIC_HOST:-$DOMAIN}"
INSTALL_RABBITMQ="${INSTALL_RABBITMQ:-0}"
ENABLE_UFW="${ENABLE_UFW:-1}"
ENABLE_HTTPS="${ENABLE_HTTPS:-0}"
MAIL_FROM="${MAIL_FROM:-}"
MAIL_USERNAME="${MAIL_USERNAME:-}"
MAIL_PASSWORD="${MAIL_PASSWORD:-}"
OSS_ACCESS_KEY_ID="${OSS_ACCESS_KEY_ID:-}"
OSS_ACCESS_KEY_SECRET="${OSS_ACCESS_KEY_SECRET:-}"
RABBITMQ_URL="${RABBITMQ_URL:-amqp://guest:guest@127.0.0.1:5672/}"
JWT_ACCESS_SECRET="${JWT_ACCESS_SECRET:-$(random_hex)}"
JWT_REFRESH_SECRET="${JWT_REFRESH_SECRET:-$(random_hex)}"
GO_VERSION="${GO_VERSION:-1.24.2}"
NODE_MAJOR="${NODE_MAJOR:-20}"
SYSTEMD_SERVICE_NAME="${SYSTEMD_SERVICE_NAME:-cloud-disk-backend}"
ENV_FILE="${ENV_FILE:-$APP_DIR/.env}"
NGINX_SITE_PATH="${NGINX_SITE_PATH:-/etc/nginx/sites-available/cloud-disk}"
NGINX_SITE_LINK="${NGINX_SITE_LINK:-/etc/nginx/sites-enabled/cloud-disk}"
SYSTEMD_SERVICE_PATH="${SYSTEMD_SERVICE_PATH:-/etc/systemd/system/${SYSTEMD_SERVICE_NAME}.service}"

require_root

[[ -d "$APP_DIR" ]] || fatal "APP_DIR does not exist: $APP_DIR"
[[ -f "$APP_DIR/go.mod" ]] || fatal "APP_DIR does not look like the repo root: $APP_DIR"
[[ -f "$APP_DIR/core/etc/core-api.yaml" ]] || fatal "Missing backend config file under APP_DIR."
[[ -f "$APP_DIR/web/package.json" ]] || fatal "Missing frontend package.json under APP_DIR."

if [[ "$JWT_ACCESS_SECRET" == "$JWT_REFRESH_SECRET" ]]; then
  fatal "JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be different."
fi

if ! id "$APP_USER" >/dev/null 2>&1; then
  fatal "APP_USER does not exist: $APP_USER"
fi

install_base_packages() {
  log "Installing base packages"
  export DEBIAN_FRONTEND=noninteractive
  apt-get update
  apt-get install -y ca-certificates curl wget gnupg lsb-release software-properties-common unzip git nginx redis-server mysql-server
}

install_go() {
  local current_go=""
  if command_exists go; then
    current_go="$(go version | awk '{print $3}' | sed 's/^go//')"
  fi

  if [[ "$current_go" == "$GO_VERSION" ]]; then
    log "Go ${GO_VERSION} already installed"
    return
  fi

  log "Installing Go ${GO_VERSION}"
  cd /tmp
  wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O "go${GO_VERSION}.linux-amd64.tar.gz"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
  rm -f "go${GO_VERSION}.linux-amd64.tar.gz"
  ensure_line_in_file /etc/profile 'export PATH=$PATH:/usr/local/go/bin'
  export PATH="$PATH:/usr/local/go/bin"
}

install_node() {
  local current_major=""
  if command_exists node; then
    current_major="$(node -v | sed 's/^v//' | cut -d. -f1)"
  fi

  if [[ "$current_major" == "$NODE_MAJOR" ]]; then
    log "Node.js ${NODE_MAJOR} already installed"
    return
  fi

  log "Installing Node.js ${NODE_MAJOR}"
  curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash -
  apt-get install -y nodejs
}

install_optional_packages() {
  if [[ "$INSTALL_RABBITMQ" == "1" ]]; then
    log "Installing RabbitMQ"
    apt-get install -y rabbitmq-server
    systemctl enable rabbitmq-server
    systemctl restart rabbitmq-server
  else
    warn "Skipping RabbitMQ installation. Async email/log pipeline will be degraded."
  fi

  if [[ "$ENABLE_HTTPS" == "1" ]]; then
    log "Installing certbot"
    apt-get install -y certbot python3-certbot-nginx
  fi
}

ensure_app_permissions() {
  log "Ensuring application directory ownership for ${APP_USER}:${APP_GROUP}"
  chown -R "$APP_USER:$APP_GROUP" "$APP_DIR"
}

setup_services() {
  log "Enabling MySQL, Redis and Nginx"
  systemctl enable mysql
  systemctl restart mysql
  systemctl enable redis-server
  systemctl restart redis-server
  systemctl enable nginx
}

setup_firewall() {
  if [[ "$ENABLE_UFW" != "1" ]]; then
    warn "Skipping ufw configuration"
    return
  fi

  if ! command_exists ufw; then
    log "Installing ufw"
    apt-get install -y ufw
  fi

  log "Configuring ufw"
  ufw allow 22/tcp || true
  ufw allow 80/tcp || true
  ufw allow 443/tcp || true
  ufw --force enable || true
}

configure_mysql() {
  log "Configuring MySQL database and application user"

  local sql_file
  sql_file="$(mktemp)"
  cat >"$sql_file" <<SQL
CREATE DATABASE IF NOT EXISTS \`${MYSQL_DB}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '${MYSQL_USER}'@'127.0.0.1' IDENTIFIED BY '${MYSQL_PASSWORD}';
ALTER USER '${MYSQL_USER}'@'127.0.0.1' IDENTIFIED BY '${MYSQL_PASSWORD}';
GRANT ALL PRIVILEGES ON \`${MYSQL_DB}\`.* TO '${MYSQL_USER}'@'127.0.0.1';
FLUSH PRIVILEGES;
SQL

  if [[ -n "$MYSQL_ROOT_PASSWORD" ]]; then
    mysql -uroot "-p${MYSQL_ROOT_PASSWORD}" <"$sql_file"
  else
    mysql <"$sql_file"
  fi
  rm -f "$sql_file"
}

write_env_file() {
  log "Writing environment file to $ENV_FILE"
  cat >"$ENV_FILE" <<EOF
JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}

MYSQL_DSN=${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(127.0.0.1:3306)/${MYSQL_DB}?charset=utf8mb4&parseTime=True&loc=Local

REDIS_ADDR=${REDIS_ADDR}
REDIS_PASSWORD=${REDIS_PASSWORD}

RABBITMQ_URL=${RABBITMQ_URL}

MAIL_FROM=${MAIL_FROM}
MAIL_USERNAME=${MAIL_USERNAME}
MAIL_PASSWORD=${MAIL_PASSWORD}

OSS_ACCESS_KEY_ID=${OSS_ACCESS_KEY_ID}
OSS_ACCESS_KEY_SECRET=${OSS_ACCESS_KEY_SECRET}
EOF
  chmod 600 "$ENV_FILE"
  chown "$APP_USER:$APP_GROUP" "$ENV_FILE"
}

build_backend() {
  log "Building backend binary"
  mkdir -p "$APP_DIR/bin" "$APP_DIR/logs"
  chown -R "$APP_USER:$APP_GROUP" "$APP_DIR/bin" "$APP_DIR/logs"

  sudo -u "$APP_USER" env PATH="/usr/local/go/bin:${PATH}" bash -lc "
    set -euo pipefail
    cd '$APP_DIR'
    /usr/local/go/bin/go build -o '$APP_DIR/bin/cloud-disk' ./core
  "
}

build_frontend() {
  log "Installing frontend dependencies and building"
  sudo -u "$APP_USER" bash -lc "
    set -euo pipefail
    cd '$APP_DIR/web'
    npm install
    npm run build
  "
}

write_nginx_site() {
  log "Writing Nginx site config"
  cat >"$NGINX_SITE_PATH" <<EOF
server {
    listen 80;
    server_name ${DOMAIN};

    client_max_body_size 1024m;

    root ${APP_DIR}/web/dist;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:${BACKEND_PORT}/;
        proxy_http_version 1.1;

        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_connect_timeout 60s;
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;
    }
}
EOF
  ln -sf "$NGINX_SITE_PATH" "$NGINX_SITE_LINK"
  rm -f /etc/nginx/sites-enabled/default
  nginx -t
  systemctl restart nginx
}

write_systemd_service() {
  log "Writing systemd service"
  cat >"$SYSTEMD_SERVICE_PATH" <<EOF
[Unit]
Description=Cloud Disk Backend
After=network.target mysql.service redis-server.service
Wants=network.target

[Service]
Type=simple
User=${APP_USER}
Group=${APP_GROUP}
WorkingDirectory=${APP_DIR}
EnvironmentFile=${ENV_FILE}
ExecStart=${APP_DIR}/bin/cloud-disk -f ${APP_DIR}/core/etc/core-api.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable "$SYSTEMD_SERVICE_NAME"
  systemctl restart "$SYSTEMD_SERVICE_NAME"
}

setup_https() {
  if [[ "$ENABLE_HTTPS" != "1" ]]; then
    warn "Skipping HTTPS setup"
    return
  fi

  if [[ "$DOMAIN" == "_" || -z "$DOMAIN" ]]; then
    warn "ENABLE_HTTPS=1 but DOMAIN is empty or default '_'; skipping certbot"
    return
  fi

  log "Requesting HTTPS certificate for ${DOMAIN}"
  certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos -m "${MAIL_FROM:-admin@${DOMAIN}}"
}

print_summary() {
  cat <<EOF

Deployment complete.

App directory:      ${APP_DIR}
Backend service:    ${SYSTEMD_SERVICE_NAME}
Nginx site:         ${NGINX_SITE_PATH}
Env file:           ${ENV_FILE}
Server name:        ${DOMAIN}
External URL:       http://${SERVER_PUBLIC_HOST}

Check commands:
  systemctl status ${SYSTEMD_SERVICE_NAME}
  journalctl -u ${SYSTEMD_SERVICE_NAME} -f
  systemctl status nginx
  curl http://127.0.0.1:${BACKEND_PORT}/user/detail

Important:
  1. Ensure your cloud security group allows TCP 80/443.
  2. Upload and preview require valid OSS credentials.
  3. Registration email requires valid SMTP credentials.
EOF
}

main() {
  install_base_packages
  install_go
  install_node
  install_optional_packages
  ensure_app_permissions
  setup_services
  setup_firewall
  configure_mysql
  write_env_file
  build_backend
  build_frontend
  write_nginx_site
  write_systemd_service
  setup_https
  print_summary
}

main "$@"
