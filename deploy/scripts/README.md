# Ubuntu 一键部署脚本

脚本文件：

- `deploy/scripts/ubuntu-oneclick-deploy.sh`

用途：

- 安装 Ubuntu 运行依赖
- 初始化 MySQL / Redis / Nginx
- 编译 Go 后端
- 构建 Vue 前端
- 生成 `.env`
- 写入 Nginx 配置
- 写入 systemd 服务

## 用法

先把仓库上传到 Ubuntu 服务器，例如：

```bash
cd /opt
git clone <your-repo> my_cloud_disk
cd /opt/my_cloud_disk
chmod +x deploy/scripts/ubuntu-oneclick-deploy.sh
```

最小可执行示例：

```bash
sudo DOMAIN=your-domain.com \
APP_DIR=/opt/my_cloud_disk \
APP_USER=$USER \
MYSQL_PASSWORD='YourStrongMysqlPassword' \
OSS_ACCESS_KEY_ID='your-ak' \
OSS_ACCESS_KEY_SECRET='your-sk' \
bash deploy/scripts/ubuntu-oneclick-deploy.sh
```

如果你先不配域名，也可以先用公网 IP：

```bash
sudo DOMAIN=你的公网IP \
SERVER_PUBLIC_HOST=你的公网IP \
APP_DIR=/opt/my_cloud_disk \
APP_USER=$USER \
MYSQL_PASSWORD='YourStrongMysqlPassword' \
OSS_ACCESS_KEY_ID='your-ak' \
OSS_ACCESS_KEY_SECRET='your-sk' \
bash deploy/scripts/ubuntu-oneclick-deploy.sh
```

## 常用参数

- `APP_DIR`：项目目录，默认脚本自动推导当前仓库根目录
- `APP_USER`：运行应用的 Linux 用户，默认取 `SUDO_USER`
- `DOMAIN`：Nginx `server_name`
- `SERVER_PUBLIC_HOST`：对外访问地址，默认跟 `DOMAIN` 一致
- `MYSQL_DB`：默认 `cloud_disk`
- `MYSQL_USER`：默认 `cloud_disk`
- `MYSQL_PASSWORD`：应用数据库密码
- `MYSQL_ROOT_PASSWORD`：如果你的 MySQL root 需要密码登录，就传这个
- `OSS_ACCESS_KEY_ID`：阿里云 OSS AK
- `OSS_ACCESS_KEY_SECRET`：阿里云 OSS SK
- `MAIL_FROM`：SMTP 发件人
- `MAIL_USERNAME`：SMTP 用户名
- `MAIL_PASSWORD`：SMTP 授权码
- `INSTALL_RABBITMQ=1`：需要 RabbitMQ 时开启
- `ENABLE_HTTPS=1`：需要 certbot 自动申请 HTTPS 时开启
- `ENABLE_UFW=0`：不想让脚本动防火墙时关闭

## 注意

- 脚本会把整个项目目录 `chown` 给 `APP_USER`
- 上传和预览依赖 OSS；不配置 OSS，主站能起来，但文件相关能力不可用
- `ENABLE_HTTPS=1` 前，必须保证域名已解析到服务器公网 IP，并且 `80` 端口已开放
