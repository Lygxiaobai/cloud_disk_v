# 云盘项目 Web 前端分离方案

## 1. 目标

这份文档用于评估当前 `cloud_disk` 项目是否适合增加一个独立前端，并给出一个可以直接落地的 MVP 方案。

结论先说：**可行，而且现有后端已经足够支撑第一版前端上线**。
前端可以先按现有接口完成核心链路，再逐步补细节体验。

## 2. 当前项目现状

从代码结构看，这个项目目前是一个纯 Go 后端服务：

- 后端框架：`go-zero`
- 接口定义：`core/core.api`
- 启动入口：`core/core.go`
- 配置文件：`core/etc/core-api.yaml`
- 数据存储：MySQL + Redis
- 文件存储：阿里云 OSS

当前仓库里没有前端目录，也没有服务端模板渲染逻辑，所以做成“前后端分离”是自然且合适的方向。

## 3. 已有后端能力盘点

### 3.1 用户与认证

已具备：

- 登录：`POST /user/login`
- 注册验证码：`POST /mail/code/send/register`
- 注册：`POST /user/register`
- 刷新 Token：`PUT /refresh/token`
- 用户详情：`GET /user/detail`

认证方式：

- 受保护接口通过请求头 `Authorization` 读取 token
- 当前后端读取的是**原始 JWT 字符串**
- 不是常见的 `Bearer <token>` 解析方式
- 刷新接口通过请求体接收 `refresh_token`

这意味着前端接入时要直接传：

```http
Authorization: xxxxx.yyyyy.zzzzz
```

### 3.2 文件管理

已具备：

- 文件上传：`POST /file/upload`
- 分片上传：`POST /file/upload/multipart`
- 关联用户文件：`POST /user/repository/save`
- 文件列表：`GET /user/file/list`
- 创建文件夹：`PUT /user/folder/create`
- 重命名：`PUT /user/file/name/update`
- 移动文件：`PUT /user/file/move`
- 删除文件：`DELETE /user/file/delete`

这里要特别注意一件事：

**上传文件不是一步完成的，而是两步流程。**

1. 先调用上传接口，把文件放进公共存储池 `repository_pool`
2. 再调用 `POST /user/repository/save`，把这个文件挂到当前用户目录下

所以前端上传流程必须按这个顺序设计。

当前后端已经补了 `is_dir`，前端可以直接用它区分文件与文件夹。

### 3.3 分享能力

已具备：

- 创建分享：`POST /share/basic/create`
- 查看分享详情：`GET /share/file/detail`
- 保存分享文件到我的网盘：`POST /share/file/save`

这已经足够支持第一版“分享链接查看 + 保存到我的网盘”。

## 4. 可行性结论

### 4.1 可直接做出来的 MVP

基于当前后端，前端可以实现下面这些核心能力：

- 登录 / 注册
- 网盘首页文件列表
- 面包屑目录切换
- 创建文件夹
- 上传文件
- 文件重命名
- 文件移动
- 文件删除
- 生成分享链接
- 访问分享页
- 保存分享文件到个人网盘

### 4.2 当前就能做，但体验会受限的点

- 登录接口没有直接返回用户基础信息，前端首次进入页面仍需要自行初始化用户信息
- 移动文件时，如果前端要做“选择目标文件夹”的树形弹窗，后端目前没有单独的目录树接口
- 受保护接口发生 `401` 时，需要前端统一实现一次 `refresh_token` 自动续签和原请求重放

### 4.3 已补充与建议继续优化

目前已经补上的项：

1. 开发环境可通过 Nginx 反向代理把前端和后端收敛到同一域名
2. 文件列表已经支持 `is_dir`
3. 刷新接口已经统一为 `/refresh/token`
4. 文档中已经约定 axios `401 -> refresh_token -> 重试原请求` 策略

建议继续优化的项：

1. 登录接口可考虑直接返回 `identity`、`name`，减少一次用户信息初始化请求
2. 如果要做“移动到目录树”的体验，建议补一个目录树接口

如果暂时不继续改后端，前端依然能做，只是会多一些兼容代码。

## 5. 推荐前端技术方案

推荐技术栈：

- `Vue 3`
- `TypeScript`
- `Vite`
- `Pinia`
- `Vue Router`
- `Axios`
- `Element Plus`

选择这套的原因：

- 上手快，适合中后台和文件管理型页面
- 表格、弹窗、上传、树组件成熟
- 你这个项目是典型“业务台 + 文件管理台”场景，用 Vue 方案开发效率高

如果你更偏 React，也可以做，但在“尽快出第一版”这个目标下，我更建议 Vue 3。

## 6. 前端建议实现范围

建议先做一个独立目录：

```text
web/
  src/
    api/
    assets/
    components/
    composables/
    router/
    stores/
    types/
    utils/
    views/
      auth/
      disk/
      share/
```

建议页面如下：

- `/login` 登录页
- `/register` 注册页
- `/disk` 我的网盘主页
- `/share/:identity` 分享详情页

## 7. 页面与功能设计

### 7.1 登录页

包含：

- 用户名
- 密码
- 登录按钮
- 跳转注册入口

成功后：

- 保存 `token` 和 `refresh_token`
- 跳转到 `/disk`

### 7.2 注册页

包含：

- 用户名
- 邮箱
- 密码
- 邮箱验证码
- 发送验证码按钮
- 注册按钮

对应接口：

- 发送验证码：`POST /mail/code/send/register`
- 注册：`POST /user/register`

### 7.3 网盘主页

建议布局：

- 左侧：简单导航
- 顶部：当前用户、退出登录
- 主区域：面包屑 + 工具栏 + 文件表格

工具栏按钮：

- 上传文件
- 新建文件夹
- 刷新列表

表格操作：

- 下载/打开文件链接
- 重命名
- 移动
- 删除
- 分享

### 7.4 分享页

包含：

- 文件名
- 文件大小
- 文件后缀
- 下载链接
- 保存到我的网盘按钮

## 8. 关键接口对接说明

### 8.1 登录

```http
POST /user/login
Content-Type: application/json
```

请求体：

```json
{
  "name": "admin",
  "password": "123456"
}
```

响应体：

```json
{
  "token": "...",
  "refresh_token": "..."
}
```

### 8.2 刷新 Token

```http
PUT /refresh/token
Content-Type: application/json
```

请求体：

```json
{
  "refresh_token": "..."
}
```

响应体：

```json
{
  "token": "...",
  "refresh_token": "..."
}
```

前端约定：

- 业务接口只使用 `Authorization: <token>`
- 当业务接口返回 `401` 时，前端先不要立即退出登录
- 前端先调用 `/refresh/token`
- 若刷新成功，覆盖本地的 `token` 和 `refresh_token`，然后自动重试刚才失败的请求
- 若 `/refresh/token` 也返回 `401` 或刷新失败，才视为真正未登录

### 8.3 文件列表

```http
GET /user/file/list?id=0&page=1&size=10
Authorization: <token>
```

约定建议：

- `id=0` 作为根目录
- 直接使用后端返回的 `is_dir` 作为文件夹判断依据

### 8.4 上传文件

前端流程建议：

1. 选择文件
2. 以 `multipart/form-data` 调用 `POST /file/upload/multipart`
3. 拿到返回的 `identity`
4. 再调用 `POST /user/repository/save`
5. 刷新当前目录列表

第二步上传接口返回的是公共存储池文件标识，不代表已经出现在用户目录里。

### 8.5 新建文件夹

```http
PUT /user/folder/create
Authorization: <token>
Content-Type: application/json
```

```json
{
  "name": "文档",
  "parentId": 0
}
```

### 8.6 分享流程

生成分享：

```http
POST /share/basic/create
Authorization: <token>
```

查看分享：

```http
GET /share/file/detail?identity=<shareIdentity>
```

保存到我的网盘：

```http
POST /share/file/save
Authorization: <token>
```

## 9. 前端状态管理与 Axios 设计

建议拆成 3 个 store：

- `authStore`
  - 保存 `token`
  - 保存 `refreshToken`
  - 处理登录、退出、刷新 token
- `diskStore`
  - 当前目录 id
  - 面包屑路径
  - 当前文件列表
  - 分页参数
- `shareStore`
  - 当前分享详情
  - 保存分享文件动作

### 9.1 Axios 鉴权与自动续签设计

核心规则：

- 请求拦截器统一给受保护接口补 `Authorization: <token>`
- `/user/login`、`/user/register`、`/mail/code/send/register`、`/refresh/token` 不走自动续签逻辑
- 业务接口返回 `401` 时，优先尝试使用 `refresh_token` 调用 `/refresh/token`
- 刷新成功后，自动重放原请求
- 只有 `refresh_token` 也失效时，才真正清空登录态并跳转登录页

推荐实现细节：

- 维护一个全局 `isRefreshing` 标记，避免多个并发 `401` 同时触发多次刷新
- 维护一个等待队列，刷新期间后续失败请求先挂起，等刷新成功后统一重放
- 给原请求打一个 `_retry` 标记，避免单个请求无限循环重试

示例伪代码：

```ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const service = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

let isRefreshing = false
let pendingQueue: Array<(token: string | null) => void> = []

service.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  const token = authStore.token
  if (token && config.headers) {
    config.headers.Authorization = token
  }
  return config
})

service.interceptors.response.use(
  (response) => response,
  async (error) => {
    const authStore = useAuthStore()
    const originalRequest = error.config
    const status = error.response?.status
    const isRefreshApi = originalRequest?.url?.includes('/refresh/token')

    if (status !== 401 || isRefreshApi || originalRequest?._retry) {
      return Promise.reject(error)
    }

    originalRequest._retry = true

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        pendingQueue.push((newToken) => {
          if (!newToken) {
            reject(error)
            return
          }
          originalRequest.headers.Authorization = newToken
          resolve(service(originalRequest))
        })
      })
    }

    isRefreshing = true

    try {
      const { data } = await axios.put('/api/refresh/token', {
        refresh_token: authStore.refreshToken,
      })

      authStore.setTokens(data.token, data.refresh_token)
      pendingQueue.forEach((cb) => cb(data.token))
      pendingQueue = []
      originalRequest.headers.Authorization = data.token
      return service(originalRequest)
    } catch (refreshError) {
      const refreshStatus = axios.isAxiosError(refreshError)
        ? refreshError.response?.status
        : undefined

      pendingQueue.forEach((cb) => cb(null))
      pendingQueue = []

      // 只有 refresh_token 也失效时，才真正清空登录态。
      if (refreshStatus === 401 || refreshStatus === 403) {
        authStore.clearAuth()
      }

      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)
```

这一套的最终效果是：

- 普通 `token` 过期时，用户通常无感知
- `refresh_token` 仍有效时，前端自动续签并重试
- 只有 `refresh_token` 也过期，或刷新接口明确返回未授权时，才是真正的 `401`
- 如果刷新阶段只是网络异常、超时或 `5xx`，前端应提示错误，而不是立刻把用户踢下线

## 10. 前端到后端的访问方式

推荐开发和部署都走同一种路径：

- 浏览器访问前端站点，例如 `http://localhost`
- 前端所有接口都请求 `/api/...`
- Nginx 把 `/api/` 反向代理到 Go 后端 `http://127.0.0.1:8888`

这样浏览器侧不会出现跨域问题，部署方式也更接近生产环境。

### 10.1 仓库里的 Nginx 配置文件

当前仓库已经提供两份配置：

- [cloud-disk.conf](../deploy/nginx/cloud-disk.conf)
- [cloud-disk.dev.conf.example](../deploy/nginx/cloud-disk.dev.conf.example)

用途说明：

- `cloud-disk.conf`：用于正式部署。Nginx 直接托管前端打包后的 `web/dist`
- `cloud-disk.dev.conf.example`：用于本地开发。Nginx 把 `/` 转发到 Vite `127.0.0.1:5173`

### 10.2 开发环境怎么用

适用场景：

- 前端本地运行 Vite
- 后端本地运行 Go 服务
- 浏览器统一访问 Nginx

使用步骤：

1. 启动 Go 后端，确认监听 `127.0.0.1:8888`
2. 启动前端开发服务，确认监听 `127.0.0.1:5173`
3. 把 [cloud-disk.dev.conf.example](../deploy/nginx/cloud-disk.dev.conf.example) 复制到你的 Nginx `conf.d` 或主配置引用目录中
4. 根据你的实际环境，确认 `server_name`、端口、Vite 地址是否需要调整
5. 重载 Nginx
6. 浏览器访问 `http://localhost`

这时访问链路就是：

```text
浏览器 -> nginx -> vite
浏览器 -> /api/* -> nginx -> go-zero
```

前端请求示例：

```ts
axios.create({
  baseURL: '/api',
})
```

例如：

- 前端请求 `/api/user/login`
- Nginx 转发到 `http://127.0.0.1:8888/user/login`

### 10.3 正式部署怎么用

适用场景：

- 前端已经执行打包
- 产物输出到 `web/dist`
- Nginx 负责静态资源和接口反代

使用步骤：

1. 在前端项目中执行打包，生成 `web/dist`
2. 打开 [cloud-disk.conf](../deploy/nginx/cloud-disk.conf)
3. 把其中的 `root D:/Go_Project/cloud_disk/web/dist;` 改成你机器上的真实部署路径
4. 把该配置放进 Nginx `conf.d` 或主配置引用目录
5. 启动 Go 后端，确认监听 `127.0.0.1:8888`
6. 重载 Nginx
7. 浏览器访问 `http://localhost`

这时访问链路就是：

```text
浏览器 -> nginx -> web/dist
浏览器 -> /api/* -> nginx -> go-zero
```

`cloud-disk.conf` 里这一段：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8888/;
}
```

表示会把 `/api/` 前缀去掉再转发，所以：

- `/api/user/file/list` 会被转发成 `/user/file/list`
- `/api/refresh/token` 会被转发成 `/refresh/token`

### 10.4 Nginx 重载示例

如果你是在 Windows 上直接跑 Nginx，常见命令是：

```powershell
nginx.exe -t
nginx.exe -s reload
```

如果你是在 Linux 服务器上，常见命令是：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### 10.5 使用约定

为了让前端和 Nginx 配置保持一致，建议固定下面这条约定：

- 前端永远只请求 `/api/*`
- 不在前端代码里写死 `http://127.0.0.1:8888`
- 本地开发和正式部署都通过 Nginx 做统一入口

## 11. 与当前后端的适配风险

下面这些点不影响第一版上线，但建议尽早处理：

### 11.1 用户详情初始化

登录接口目前只返回 `token` 和 `refresh_token`。
前端登录后如果要展示用户昵称、邮箱，仍需要走一次用户信息初始化流程。

### 11.2 目录树接口缺失

如果后续要做“移动到某个目录”的树形选择器，最好补一个专门的目录树接口，不然前端要多次请求文件列表自己拼树。

### 11.3 验证码 Redis key 设计

注册验证码当前看起来是统一写在 Redis 的 `code` 这个 key。
如果多人同时注册，验证码会互相覆盖。

建议改为：

```text
register_code:<email>
```

### 11.4 需要约定“真正的 401”

前端实现时需要统一一个判断规则：

- 业务接口 `401` 不等于真正登录失效
- 应先尝试用 `refresh_token` 调用 `/refresh/token`
- 刷新成功则自动重试原请求
- 只有刷新接口也失败，才清空登录态并跳转登录页

这条规则一定要写进 axios 封装，否则前端每个页面可能会各自处理一套，后续很难维护。

## 12. 推荐开发顺序

### 第一阶段：先出能跑的前端骨架

- 初始化 `web/`
- 配置路由、请求封装、环境变量
- 做登录页、注册页、网盘主页壳子

### 第二阶段：打通文件管理主流程

- 文件列表
- 新建文件夹
- 上传文件
- 删除
- 重命名
- 移动

### 第三阶段：补分享能力和体验细节

- 创建分享
- 分享详情页
- 保存到我的网盘
- 空状态、加载态、错误提示

## 13. 我的建议

如果你的目标是：

- 先把项目做完整
- 能演示
- 能继续迭代

那这条路线是值得做的。

我建议采用下面的执行策略：

1. 先按现有后端接口实现前端 MVP
2. 首先把 axios 的鉴权、刷新、重试链路封装稳定
3. 等页面跑通后，再做预览、拖拽上传、目录树、批量操作这些增强功能

## 14. 下一步可以怎么做

如果你认可这份方案，下一步我可以直接开始搭前端项目，建议优先做：

1. `Vue 3 + Vite + TypeScript` 工程初始化
2. 登录 / 注册页面
3. axios 请求封装和自动刷新逻辑
4. 网盘文件列表页
5. 上传、新建文件夹、删除、重命名

如果你愿意，我下一步可以继续帮你把这个前端直接在仓库里搭起来。
