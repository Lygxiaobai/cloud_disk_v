# Ubuntu One-Click Deploy

Script:
- `deploy/scripts/ubuntu-oneclick-deploy.sh`

What it does:
- installs Ubuntu runtime dependencies
- installs the Go version declared in `go.mod` by default
- installs Node.js
- configures MySQL, Redis, Nginx, and optional RabbitMQ
- writes `.env`
- builds backend and frontend
- writes systemd and Nginx config
- verifies MySQL, Redis, backend HTTP, frontend HTTP, and Nginx API proxy

## Quick Start

```bash
cd /opt
git clone <your-repo> my_cloud_disk
cd /opt/my_cloud_disk
chmod +x deploy/scripts/ubuntu-oneclick-deploy.sh
```

Minimal deployment without mail and without upload:

```bash
sudo APP_DIR=/opt/my_cloud_disk \
APP_USER=$USER \
MYSQL_PASSWORD='ReplaceWithStrongPassword' \
bash deploy/scripts/ubuntu-oneclick-deploy.sh
```

Full deployment with SMTP and OSS:

```bash
sudo DOMAIN=your-domain.com \
SERVER_PUBLIC_HOST=your-domain.com \
APP_DIR=/opt/my_cloud_disk \
APP_USER=$USER \
MYSQL_PASSWORD='ReplaceWithStrongPassword' \
MAIL_FROM='noreply@example.com' \
MAIL_HOST='smtp.example.com:465' \
MAIL_SERVER_NAME='smtp.example.com' \
MAIL_USERNAME='noreply@example.com' \
MAIL_PASSWORD='ReplaceWithSmtpAppPassword' \
OSS_REGION='cn-hangzhou' \
OSS_BUCKET='your-bucket' \
OSS_ENDPOINT='oss-cn-hangzhou.aliyuncs.com' \
OSS_ROLE_ARN='acs:ram::<account-id>:role/<role-name>' \
OSS_ACCESS_KEY_ID='your-ak' \
OSS_ACCESS_KEY_SECRET='your-sk' \
bash deploy/scripts/ubuntu-oneclick-deploy.sh
```

## Important Variables

- `APP_DIR`: repo root on the VM
- `APP_USER`: Linux user that builds and runs the app
- `APP_GROUP`: Linux group for the app user
- `DOMAIN`: Nginx `server_name`
- `SERVER_PUBLIC_HOST`: printed external host, useful when `DOMAIN` is `_`
- `MYSQL_HOST`: MySQL host, default `127.0.0.1`
- `MYSQL_PORT`: MySQL port, default `3306`
- `MYSQL_DB`: MySQL database name, default `cloud_disk`
- `MYSQL_USER`: MySQL application user, default `cloud_disk`
- `MYSQL_PASSWORD`: MySQL application password
- `MYSQL_ROOT_PASSWORD`: required only if root login needs a password
- `REDIS_ADDR`: Redis address, default `127.0.0.1:6379`
- `REDIS_PASSWORD`: Redis password
- `MAIL_FROM`: sender email
- `MAIL_HOST`: SMTP host and port, for example `smtp.example.com:465`
- `MAIL_SERVER_NAME`: SMTP TLS server name, for example `smtp.example.com`
- `MAIL_USERNAME`: SMTP account
- `MAIL_PASSWORD`: SMTP app password or auth token
- `OSS_REGION`: OSS region, for example `cn-hangzhou`
- `OSS_BUCKET`: OSS bucket name
- `OSS_ENDPOINT`: OSS endpoint, for example `oss-cn-hangzhou.aliyuncs.com`
- `OSS_ROLE_ARN`: RAM role ARN used for STS
- `OSS_EXTERNAL_ID`: optional external ID for the RAM role
- `OSS_ACCESS_KEY_ID`: OSS access key ID
- `OSS_ACCESS_KEY_SECRET`: OSS access key secret
- `INSTALL_RABBITMQ=1`: install and start RabbitMQ
- `ENABLE_HTTPS=1`: install certbot and request certificates after Nginx is ready
- `ENABLE_UFW=0`: skip automatic UFW changes
- `GO_VERSION`: optional manual override; default is read from `go.mod`

## Notes

- SMTP is optional. If it is not configured, registration email verification will fail but the site can still run.
- OSS is optional. If it is not configured, upload and preview will fail but the site can still run.
- The script changes ownership of the entire repo directory to `APP_USER:APP_GROUP`.
- Before `ENABLE_HTTPS=1`, make sure the domain already resolves to the VM public IP and TCP `80` is open.
