# 本地启动与联调说明

这份文档用于下次本地测试时，快速启动：

- Go 后端
- Vue 前端
- Nginx 反向代理

启动完成后，推荐统一通过：

```text
http://localhost
```

来访问前端和接口。

## 1. 启动前检查

先确认下面几个依赖已经准备好：

- MySQL 已启动，端口 `3306`
- Redis 已启动，端口 `6379`
- Node.js 已安装
- Go 已安装
- Nginx 已安装

可以用下面命令简单检查：

```powershell
Test-NetConnection 127.0.0.1 -Port 3306
Test-NetConnection 127.0.0.1 -Port 6379
node -v
go version
```

如果 PowerShell 直接执行 `npm` 被脚本策略拦住，请使用：

```powershell
npm.cmd
```

## 2. 启动 Go 后端

在第一个终端中执行：

```powershell
cd d:\Go_Project\cloud_disk
go run ./core -f core/etc/core-api.yaml
```

如果启动成功，会看到类似输出：

```text
Starting server at 0.0.0.0:8888...
```

后端访问地址：

```text
http://127.0.0.1:8888
```

## 3. 启动 Vue 前端

第一次启动前，如果依赖还没装：

```powershell
cd d:\Go_Project\my_cloud_disk\web
npm.cmd install
```

正常启动前端开发服务：

```powershe
npm.cmd run dev -- --host 127.0.0.1 --port 5173
```

如果启动成功，会看到类似输出：

```text
Local: http://127.0.0.1:5173/
```

前端开发地址：

```text
http://127.0.0.1:5173
```

## 4. 启动 Nginx

本项目本地联调时，推荐让 Nginx 作为统一入口：

- `/` 转发到前端开发服务 `127.0.0.1:5173`
- `/api/` 转发到 Go 后端 `127.0.0.1:8888`

### 4.1 使用仓库中的开发配置

开发配置文件在这里：

- [cloud-disk.dev.conf.example](d:\Go_Project\cloud_disk\deploy\nginx\cloud-disk.dev.conf.example)

Nginx 安装路径：

```text
C:\Users\38624\AppData\Local\Microsoft\WinGet\Packages\nginxinc.nginx_Microsoft.Winget.Source_8wekyb3d8bbwe\nginx-1.29.6
```

### 4.2 生成可运行的 nginx.conf

为了方便下次测试，先创建运行目录：

```powershell
cd d:\Go_Project\cloud_disk
New-Item -ItemType Directory -Force .runlogs\nginx\logs | Out-Null
New-Item -ItemType Directory -Force .runlogs\nginx\temp\client_body_temp | Out-Null
New-Item -ItemType Directory -Force .runlogs\nginx\temp\proxy_temp | Out-Null
New-Item -ItemType Directory -Force .runlogs\nginx\temp\fastcgi_temp | Out-Null
New-Item -ItemType Directory -Force .runlogs\nginx\temp\uwsgi_temp | Out-Null
New-Item -ItemType Directory -Force .runlogs\nginx\temp\scgi_temp | Out-Null
```

然后在：

- `d:\Go_Project\my_cloud_disk\.runlogs\nginx\nginx.conf`

写入下面内容：

```nginx
worker_processes  1;
error_log  logs/error.log info;
pid        logs/nginx.pid;

events {
    worker_connections  1024;
}

http {
    include       C:/Users/38624/AppData/Local/Microsoft/WinGet/Packages/nginxinc.nginx_Microsoft.Winget.Source_8wekyb3d8bbwe/nginx-1.29.6/conf/mime.types;
    default_type  application/octet-stream;
    sendfile      on;
    keepalive_timeout 65;

    include D:/Go_Project/my_cloud_disk/deploy/nginx/cloud-disk.dev.conf.example;
}
```

### 4.3 检查配置

```powershell
& 'C:\Users\38624\AppData\Local\Microsoft\WinGet\Packages\nginxinc.nginx_Microsoft.Winget.Source_8wekyb3d8bbwe\nginx-1.29.6\nginx.exe' `
  -p 'd:\Go_Project\my_cloud_disk\.runlogs\nginx\' `
  -c 'd:\Go_Project\my_cloud_disk\.runlogs\nginx\nginx.conf' `
  -t
```

如果成功，会看到：

```text
nginx: configuration file ... test is successful
```

### 4.4 启动 Nginx

在第三个终端执行：

```powershell
& 'C:\Users\38624\AppData\Local\Microsoft\WinGet\Packages\nginxinc.nginx_Microsoft.Winget.Source_8wekyb3d8bbwe\nginx-1.29.6\nginx.exe' `
  -p 'd:\Go_Project\cloud_disk\.runlogs\nginx\' `
  -c 'd:\Go_Project\cloud_disk\.runlogs\nginx\nginx.conf'
```

启动后访问：

```text
http://localhost
```

## 5. 联调验证

### 5.1 页面访问

浏览器打开：

```text
http://localhost/login
```

### 5.2 测试账号

当前联调可用测试账号：

```text
用户名: frontend_demo_user
密码:   demo123456
```

### 5.3 接口验证

也可以手动测试：

```powershell
$body = @{ name = 'frontend_demo_user'; password = 'demo123456' } | ConvertTo-Json
Invoke-RestMethod -Uri 'http://localhost/api/user/login' -Method Post -ContentType 'application/json' -Body $body
```

拿到 `token` 后测试文件列表：

```powershell
$body = @{ name = 'frontend_demo_user'; password = 'demo123456' } | ConvertTo-Json
$login = Invoke-RestMethod -Uri 'http://localhost/api/user/login' -Method Post -ContentType 'application/json' -Body $body
$headers = @{ Authorization = $login.token }
Invoke-RestMethod -Uri 'http://localhost/api/user/file/list?id=0&page=1&size=20' -Method Get -Headers $headers
```

## 6. 停止服务

### 6.1 停止前端

直接在前端终端按：

```text
Ctrl + C
```

### 6.2 停止后端

直接在后端终端按：

```text
Ctrl + C
```

### 6.3 停止 Nginx

```powershell
& 'C:\Users\38624\AppData\Local\Microsoft\WinGet\Packages\nginxinc.nginx_Microsoft.Winget.Source_8wekyb3d8bbwe\nginx-1.29.6\nginx.exe' `
  -p 'd:\Go_Project\cloud_disk\.runlogs\nginx\' `
  -c 'd:\Go_Project\cloud_disk\.runlogs\nginx\nginx.conf' `
  -s stop
```

## 7. 常见问题

### 7.1 `npm` 被 PowerShell 拦住

请改用：

```powershell
npm.cmd run dev
```

### 7.2 `localhost` 打不开

先确认三个端口是否正常：

```powershell
Test-NetConnection 127.0.0.1 -Port 8888
Test-NetConnection 127.0.0.1 -Port 5173
Test-NetConnection 127.0.0.1 -Port 80
```

### 7.3 Nginx 报 temp 目录不存在

重新执行第 `4.2` 节中的目录创建命令。

### 7.4 前端页面能开，接口失败

先确认后端是否启动成功：

```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/login' -Method Post -ContentType 'application/json' -Body (@{ name = 'frontend_demo_user'; password = 'demo123456' } | ConvertTo-Json)
```

如果后端能通，而 `http://localhost/api/...` 不通，再检查 Nginx 配置和启动状态。