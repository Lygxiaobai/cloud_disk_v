# Docker Deploy Guide

This project can now be deployed on Linux with Docker Compose.

## What Gets Started

- `mysql`
- `redis`
- `backend`
- `nginx`
- optional `rabbitmq` profile

The frontend is built into the `nginx` image and served as static files.
Requests to `/api/*` are proxied from `nginx` to the `backend` container.

## 1. Prepare the Server

Install Docker and the Compose plugin, then allow inbound traffic to your chosen HTTP port.

Recommended open ports:

- `22` for SSH
- `80` for HTTP
- `443` if you later add an HTTPS reverse proxy

## 2. Upload the Repo

```bash
cd /opt
git clone https://github.com/Lygxiaobai/cloud_disk_v.git my_cloud_disk
cd /opt/my_cloud_disk
```

## 3. Prepare the Docker Env File

```bash
cp .env.docker.example .env.docker
```

At minimum, update these values in `.env.docker`:

- `MYSQL_PASSWORD`
- `MYSQL_ROOT_PASSWORD`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`

If you need registration email verification, also fill:

- `MAIL_FROM`
- `MAIL_HOST`
- `MAIL_SERVER_NAME`
- `MAIL_USERNAME`
- `MAIL_PASSWORD`

If you need upload and preview, also fill:

- `OSS_REGION`
- `OSS_BUCKET`
- `OSS_ENDPOINT`
- `OSS_ROLE_ARN`
- `OSS_ACCESS_KEY_ID`
- `OSS_ACCESS_KEY_SECRET`

## 4. Start the Core Services

```bash
docker compose --env-file .env.docker up -d --build
```

If you also want RabbitMQ:

```bash
docker compose --env-file .env.docker --profile rabbitmq up -d --build
```

## 5. Check Status

```bash
docker compose --env-file .env.docker ps
docker compose --env-file .env.docker logs backend --tail=100
docker compose --env-file .env.docker logs nginx --tail=100
```

Open the site in your browser:

```text
http://your-server-ip
```

Or if you changed `HTTP_PORT`:

```text
http://your-server-ip:your-http-port
```

## 6. Stop or Restart

Stop:

```bash
docker compose --env-file .env.docker down
```

Restart after changes:

```bash
docker compose --env-file .env.docker up -d --build
```

## Notes

- The backend is not published directly to the host; only `nginx` exposes the site.
- MySQL and Redis data are persisted in Docker volumes.
- Backend logs are persisted in the `backend_logs` Docker volume.
- RabbitMQ is optional. Without it, the main site still runs, but async email and async log pipeline are degraded.
- HTTPS is not bundled here. The simplest production setup is to put Caddy, Nginx Proxy Manager, Traefik, or another host-level reverse proxy in front of this stack.

