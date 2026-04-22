# Linux VM 部署指南

本文基于当前仓库实际结构整理，目标是在 Linux VM 上部署本项目，并让外部用户通过公网访问。

适用架构：

- 前端：Vue + Vite，构建产物位于 `web/dist`
- 后端：Go + go-zero，监听 `0.0.0.0:8888`
- 反向代理：Nginx
- 必需依赖：MySQL、Redis
- 可选依赖：RabbitMQ、Elasticsearch
- 对象存储：阿里云 OSS

推荐部署方式：

- Nginx 对外监听 `80/443`
- Nginx 直接托管前端静态文件
- `/api/` 反向代理到 Go 后端 `127.0.0.1:8888`
- Go 后端只监听内网或本机，不直接暴露到公网

## 1. 准备一台 Linux VM

推荐系统：

- Ubuntu 22.04/24.04
- Debian 12
- CentOS Stream 9 / Rocky Linux 9

需要保证：

- VM 有公网 IP，或者已绑定公网负载均衡
- 安全组/防火墙已放行 `80`、`443`
- 如果你暂时不用 HTTPS，至少先放行 `80`

如果云厂商有安全组，请放行：

- `22`：SSH
- `80`：HTTP
- `443`：HTTPS

如果使用 `ufw`：

```bash
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status
```

## 2. 安装基础软件

Ubuntu / Debian：

```bash
sudo apt update
sudo apt install -y git curl wget unzip nginx redis-server mysql-server
```

安装 Go：

```bash
cd /tmp
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

安装 Node.js 20：

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
node -v
npm -v
```

## 3. 部署代码目录

建议统一放到：

```text
/opt/my_cloud_disk
```

执行：

```bash
sudo mkdir -p /opt/my_cloud_disk
sudo chown -R $USER:$USER /opt/my_cloud_disk
git clone <你的仓库地址> /opt/my_cloud_disk
cd /opt/my_cloud_disk
```

如果代码已经在本地 Windows 上，也可以直接上传整个仓库到 VM。

## 4. 初始化 MySQL

登录 MySQL：

```bash
sudo mysql
```

创建数据库和账号：

```sql
CREATE DATABASE cloud_disk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'cloud_disk'@'127.0.0.1' IDENTIFIED BY '请改成强密码';
GRANT ALL PRIVILEGES ON cloud_disk.* TO 'cloud_disk'@'127.0.0.1';
FLUSH PRIVILEGES;
```

说明：

- 当前项目启动时会调用 `xorm Sync2`
- 也就是核心表会在应用启动时自动补齐
- `deploy/sql/migrations` 下的 SQL 是补充迁移，建议按需执行

如果你是全新库，先启动应用即可自动建核心表。

## 5. 启动 Redis

```bash
sudo systemctl enable redis-server
sudo systemctl start redis-server
sudo systemctl status redis-server
```

## 6. 准备后端环境变量

项目后端配置文件是：

- `core/etc/core-api.yaml`

其中大量配置通过环境变量注入，所以你需要创建：

- `/opt/my_cloud_disk/.env`

可以先参考仓库根目录的 `.env.example`。

示例：

```bash
cat > /opt/my_cloud_disk/.env <<'EOF'
JWT_ACCESS_SECRET=请替换成openssl生成的32字节随机串
JWT_REFRESH_SECRET=请替换成另一条不同的随机串

MYSQL_DSN=cloud_disk:请改成强密码@tcp(127.0.0.1:3306)/cloud_disk?charset=utf8mb4&parseTime=True&loc=Local

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/

MAIL_FROM=你的邮箱
MAIL_USERNAME=你的邮箱
MAIL_PASSWORD=你的SMTP授权码

OSS_ACCESS_KEY_ID=你的阿里云AK
OSS_ACCESS_KEY_SECRET=你的阿里云SK
EOF
```

生产环境建议生成 JWT 密钥：

```bash
openssl rand -hex 32
openssl rand -hex 32
```

注意：

- 上传、预览依赖阿里云 OSS
- 如果 OSS 不配置，文件上传相关能力会失败
- 注册验证码依赖 SMTP
- RabbitMQ 当前不是强依赖，不启动也能跑主链路，但异步邮件/日志会降级

## 7. 根据你的域名或公网 IP 修改后端配置

编辑：

- `core/etc/core-api.yaml`

重点检查：

- `Host: 0.0.0.0`
- `Port: 8888`
- `CORS.AllowedOrigins`

如果你走 Nginx 同域名转发，前后端同源访问，CORS 基本不是必须项；但为了保险，建议加入你的正式域名：

```yaml
CORS:
  AllowedOrigins:
    - http://你的域名
    - https://你的域名
    - http://你的公网IP
```

如果你没有域名，也可以先填公网 IP。

## 8. 编译后端

```bash
cd /opt/my_cloud_disk
mkdir -p bin
go build -o bin/cloud-disk ./core
```

手动启动验证：

```bash
cd /opt/my_cloud_disk
set -a
source .env
set +a
./bin/cloud-disk -f core/etc/core-api.yaml
```

看到类似输出表示成功：

```text
Starting server at 0.0.0.0:8888...
```

另开一个终端测试：

```bash
curl http://127.0.0.1:8888/user/detail
```

这个接口通常需要认证，返回业务错误也没关系；只要能连通，说明服务已起来。

## 9. 构建前端

```bash
cd /opt/my_cloud_disk/web
npm install
npm run build
```

产物会生成到：

```text
/opt/my_cloud_disk/web/dist
```

前端代码当前使用相对地址 `/api` 调后端，所以很适合用 Nginx 同域代理。

## 10. 配置 Nginx

仓库里已经提供了 Linux 示例：

- `deploy/nginx/cloud-disk.linux.conf.example`

复制到 Nginx：

```bash
sudo cp /opt/my_cloud_disk/deploy/nginx/cloud-disk.linux.conf.example /etc/nginx/sites-available/cloud-disk
sudo ln -sf /etc/nginx/sites-available/cloud-disk /etc/nginx/sites-enabled/cloud-disk
sudo rm -f /etc/nginx/sites-enabled/default
```

然后编辑：

```bash
sudo vim /etc/nginx/sites-available/cloud-disk
```

至少改这两项：

- `server_name`
- `root`

例如：

```nginx
server_name your-domain.com;
root /opt/my_cloud_disk/web/dist;
```

检查配置：

```bash
sudo nginx -t
```

重载：

```bash
sudo systemctl enable nginx
sudo systemctl restart nginx
sudo systemctl status nginx
```

## 11. 使用 systemd 托管后端

仓库里已经提供了示例：

- `deploy/systemd/cloud-disk-backend.service`

复制：

```bash
sudo cp /opt/my_cloud_disk/deploy/systemd/cloud-disk-backend.service /etc/systemd/system/cloud-disk-backend.service
```

如果你的部署目录不是 `/opt/my_cloud_disk`，先改掉 service 文件里的路径。

说明：

- 当前示例没有强制指定 `User`/`Group`
- 这样你第一次部署时更容易直接跑通
- 跑通后再改成专门的部署用户会更稳妥

加载并启动：

```bash
sudo systemctl daemon-reload
sudo systemctl enable cloud-disk-backend
sudo systemctl start cloud-disk-backend
sudo systemctl status cloud-disk-backend
```

查看日志：

```bash
journalctl -u cloud-disk-backend -f
```

## 12. 验证外部访问

如果你已经做完上面的步骤，外部访问链路应该是：

```text
浏览器 -> 公网IP/域名 -> Nginx:80/443 -> 前端静态文件
浏览器 -> /api/* -> Nginx -> 127.0.0.1:8888
```

验证方法：

1. 浏览器访问 `http://你的公网IP`
2. 能打开登录页，说明前端和 Nginx 正常
3. 打开浏览器开发者工具，登录时观察 `/api/user/login`
4. 如果返回 200 或业务错误，说明反向代理和后端链路正常

也可以在本机外部执行：

```bash
curl http://你的公网IP
curl http://你的公网IP/api/user/detail
```

## 13. 配置 HTTPS

如果你有域名，建议直接上 HTTPS。

安装 certbot：

```bash
sudo apt install -y certbot python3-certbot-nginx
```

申请证书：

```bash
sudo certbot --nginx -d your-domain.com
```

自动续期测试：

```bash
sudo certbot renew --dry-run
```

注意：

- 申请证书前，域名必须先解析到这台 VM 公网 IP
- `80` 端口必须能被公网访问

## 14. 可选能力

### 14.1 RabbitMQ

如果你要启用异步邮件/异步日志，可安装 RabbitMQ：

```bash
sudo apt install -y rabbitmq-server
sudo systemctl enable rabbitmq-server
sudo systemctl start rabbitmq-server
sudo systemctl status rabbitmq-server
```

不启用 RabbitMQ 的影响：

- 主站可运行
- 上传、列表、预览等主链路仍可工作
- 邮件验证码和异步日志链路会降级或失效

### 14.2 Elasticsearch

Elasticsearch 仅在你需要 ES 日志链路时再部署，不是主站上线必需。

## 15. 常见问题

### 页面能打开，但接口 502

排查顺序：

1. `systemctl status cloud-disk-backend`
2. `journalctl -u cloud-disk-backend -n 200`
3. `curl http://127.0.0.1:8888/user/detail`
4. `sudo nginx -t`
5. `sudo tail -f /var/log/nginx/error.log`

### 接口报数据库连接失败

检查：

- `.env` 里的 `MYSQL_DSN`
- MySQL 是否在 `127.0.0.1:3306`
- 数据库名是否为 `cloud_disk`

### 上传失败

检查：

- OSS AK/SK 是否正确
- `Bucket`、`Region`、`RoleArn` 是否有效
- Nginx `client_max_body_size` 是否足够大

### 注册验证码发不出去

检查：

- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- SMTP 是否使用授权码而不是邮箱登录密码

### 页面能访问，但外部访问不了

检查：

1. 云平台安全组是否放行 `80/443`
2. Linux 防火墙是否放行 `80/443`
3. Nginx 是否正常监听：

```bash
sudo ss -lntp | grep -E ':80|:443|:8888'
```

## 16. 推荐上线顺序

1. 先在 VM 本机完成 MySQL、Redis、Go、Node、Nginx 安装
2. 编译后端并手动跑通
3. 构建前端并配置 Nginx
4. 在 VM 本机用 `curl http://127.0.0.1` 验证
5. 再放开安全组和防火墙
6. 最后切域名和 HTTPS

## 17. 当前项目最小可上线组合

最小必需：

- MySQL
- Redis
- Go 后端
- Nginx
- 前端构建产物
- OSS 配置

建议但非必需：

- RabbitMQ
- SMTP
- HTTPS
