# RabbitMQ 邮件异步发送 - 快速使用指南

## 🚀 快速开始

### 1. 启动 RabbitMQ

```bash
# 使用 Docker 启动
docker start rabbitmq

# 或者新建容器
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

### 2. 启动主服务

```bash
cd D:/Go_Project/my_cloud_disk/core
go run core.go -f etc/core-api.yaml
```

### 3. 启动消费者服务

```bash
cd D:/Go_Project/my_cloud_disk/core
./email_worker.exe -f etc/core-api.yaml
```

### 4. 测试发送验证码

```bash
curl -X POST http://localhost:8888/mail/code/send/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

---

## 📋 常见问题 FAQ

### Q1: 消费者服务启动失败，提示连接 RabbitMQ 失败？

**原因：** RabbitMQ 服务未启动

**解决方案：**
```bash
# 检查 RabbitMQ 是否运行
docker ps | grep rabbitmq

# 如果没有运行，启动它
docker start rabbitmq

# 检查端口是否监听
netstat -an | grep 5672
```

---

### Q2: 主服务启动成功，但消费者没有收到消息？

**原因：** 消费者服务未启动或队列名称不匹配

**解决方案：**
```bash
# 1. 检查消费者是否运行
ps aux | grep email_worker

# 2. 检查 RabbitMQ 管理界面
# 访问 http://localhost:15672
# 用户名: guest, 密码: guest
# 查看 Queues 页面，确认 email_queue 有消费者

# 3. 检查配置文件中的队列名称是否一致
cat etc/core-api.yaml | grep EmailQueue
```

---

### Q3: 验证码发送成功，但用户收不到邮件？

**原因：** 邮件发送失败（SMTP 配置问题）

**解决方案：**
```bash
# 1. 查看消费者日志
# 日志中会显示邮件发送失败的原因

# 2. 检查 SMTP 配置
# 文件: core/internal/helper/helper.go
# 确认邮箱地址、密码、SMTP 服务器正确

# 3. 检查邮箱授权码
# 163 邮箱需要使用授权码，不是登录密码
# 在 core/internal/define/define.go 中设置 MailPassword
```

---

### Q4: Redis 中找不到验证码？

**原因：** Redis key 格式错误或验证码已过期

**解决方案：**
```bash
# 1. 检查 Redis key 格式
redis-cli KEYS "code:*"

# 2. 查看具体的验证码
redis-cli GET "code:your_email@example.com"

# 3. 检查过期时间
redis-cli TTL "code:your_email@example.com"

# 4. 如果返回 -2，说明 key 不存在或已过期
# 检查 define.CodeExpireTime 配置
```

---

### Q5: 如何查看 RabbitMQ 队列状态？

**方法 1：Web 管理界面**
```
访问: http://localhost:15672
用户名: guest
密码: guest
点击 Queues 标签页
```

**方法 2：命令行**
```bash
# 查看所有队列
docker exec rabbitmq rabbitmqctl list_queues

# 查看队列详情
curl -u guest:guest http://localhost:15672/api/queues/%2F/email_queue
```

---

### Q6: 如何启动多个消费者实例？

**方法：**
```bash
# 终端 1
./email_worker.exe -f etc/core-api.yaml

# 终端 2
./email_worker.exe -f etc/core-api.yaml

# 终端 3
./email_worker.exe -f etc/core-api.yaml
```

**验证：**
```bash
# 查看消费者数量
curl -u guest:guest http://localhost:15672/api/queues/%2F/email_queue | grep consumers
```

---

### Q7: 消息堆积在队列中，消费者不处理？

**原因：** 消费者处理失败或卡住

**解决方案：**
```bash
# 1. 查看消费者日志，找到错误原因

# 2. 重启消费者服务
# 先停止
ps aux | grep email_worker
kill -9 <PID>

# 再启动
./email_worker.exe -f etc/core-api.yaml

# 3. 如果消息格式错误，清空队列
docker exec rabbitmq rabbitmqctl purge_queue email_queue
```

---

### Q8: 如何测试邮件发送功能？

**测试脚本：**
```bash
# 使用提供的测试脚本
bash test_email_performance.sh

# 或者手动测试
curl -X POST http://localhost:8888/mail/code/send/register \
  -H "Content-Type: application/json" \
  -d '{"email":"your_email@example.com"}'
```

**验证步骤：**
1. 检查接口响应（应该返回验证码）
2. 检查 Redis（验证码应该存储）
3. 检查消费者日志（应该显示发送成功）
4. 检查邮箱（应该收到邮件）

---

### Q9: 如何修改邮件模板？

**位置：** `core/internal/helper/helper.go`

**修改：**
```go
func MailCodeSend(userEmail string, code string) error {
    e := email.NewEmail()
    e.From = "Your Name <your_email@163.com>"
    e.To = []string{userEmail}
    e.Subject = "验证码"

    // 修改邮件内容
    e.HTML = []byte(`
        <div style="padding: 20px;">
            <h2>您的验证码</h2>
            <p>验证码：<strong style="font-size: 24px; color: #007bff;">` + code + `</strong></p>
            <p>有效期：5分钟</p>
        </div>
    `)

    return e.SendWithTLS(...)
}
```

---

### Q10: 如何监控邮件发送成功率？

**方法 1：查看 RabbitMQ 统计**
```bash
# 查看消息统计
curl -u guest:guest http://localhost:15672/api/queues/%2F/email_queue | jq '.message_stats'

# 输出示例：
# {
#   "publish": 100,    # 发布的消息数
#   "ack": 95,         # 确认的消息数
#   "nack": 5          # 拒绝的消息数
# }
# 成功率 = ack / publish = 95%
```

**方法 2：添加日志统计**

在 `email_consumer.go` 中添加：
```go
var (
    totalCount   int64
    successCount int64
    failCount    int64
)

func emailHandler(email string, code string) error {
    atomic.AddInt64(&totalCount, 1)

    err := helper.MailCodeSend(email, code)
    if err != nil {
        atomic.AddInt64(&failCount, 1)
        return err
    }

    atomic.AddInt64(&successCount, 1)

    // 每100次打印统计
    if totalCount%100 == 0 {
        successRate := float64(successCount) / float64(totalCount) * 100
        log.Printf("邮件发送统计: 总数=%d, 成功=%d, 失败=%d, 成功率=%.2f%%",
            totalCount, successCount, failCount, successRate)
    }

    return nil
}
```

---

## 🔧 配置说明

### RabbitMQ 配置

**文件：** `core/etc/core-api.yaml`

```yaml
RabbitMQ:
  URL: amqp://guest:guest@localhost:5672/
  EmailQueue: email_queue
```

**参数说明：**
- `URL`: RabbitMQ 连接地址
  - 格式：`amqp://用户名:密码@地址:端口/虚拟主机`
  - 默认端口：5672
  - 默认用户名/密码：guest/guest

- `EmailQueue`: 邮件队列名称
  - 可以自定义，但主服务和消费者必须一致

### 邮件配置

**文件：** `core/internal/define/define.go`

```go
const (
    MailPassword = "your_smtp_password"  // SMTP 授权码
    CodeExpireTime = 300                 // 验证码过期时间（秒）
    CodeLen = 6                          // 验证码长度
)
```

**文件：** `core/internal/helper/helper.go`

```go
func MailCodeSend(userEmail string, code string) error {
    e := email.NewEmail()
    e.From = "Your Name <your_email@163.com>"  // 发件人
    // ...
    return e.SendWithTLS(
        "smtp.163.com:465",                    // SMTP 服务器
        smtp.PlainAuth("", "your_email@163.com", define.MailPassword, "smtp.163.com"),
        &tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"},
    )
}
```

---

## 📊 性能调优

### 1. 增加消费者数量

```bash
# 启动多个消费者实例（建议 2-4 个）
./email_worker.exe -f etc/core-api.yaml &
./email_worker.exe -f etc/core-api.yaml &
./email_worker.exe -f etc/core-api.yaml &
```

### 2. 调整队列参数

在 `rabbitmq.go` 中修改：
```go
_, err = r.channel.QueueDeclare(
    queueName,
    true,   // durable: 持久化
    false,  // autoDelete: 不自动删除
    false,  // exclusive: 不独占
    false,  // noWait: 等待确认
    amqp.Table{
        "x-max-length": 10000,  // 队列最大长度
        "x-message-ttl": 3600000, // 消息过期时间（1小时）
    },
)
```

### 3. 批量确认

在 `email_consumer.go` 中修改：
```go
// 设置预取数量
r.channel.Qos(
    10,    // prefetchCount: 每次预取10条消息
    0,     // prefetchSize: 0表示不限制
    false, // global: false表示仅应用于当前channel
)
```

---

## 🛠️ 故障排查

### 检查清单

- [ ] RabbitMQ 服务是否运行？
- [ ] Redis 服务是否运行？
- [ ] MySQL 服务是否运行？
- [ ] 主服务是否启动成功？
- [ ] 消费者服务是否启动成功？
- [ ] 配置文件路径是否正确？
- [ ] 队列名称是否一致？
- [ ] SMTP 配置是否正确？
- [ ] 网络是否畅通？

### 日志位置

- **主服务日志：** 控制台输出 + `./logs/error.log`
- **消费者日志：** 控制台输出
- **RabbitMQ 日志：** Docker 容器日志

```bash
# 查看主服务日志
tail -f ./logs/error.log

# 查看 RabbitMQ 日志
docker logs -f rabbitmq
```

---

## 📞 技术支持

如有问题，请检查：
1. 本文档的 FAQ 部分
2. 完整测试报告：`RABBITMQ_EMAIL_OPTIMIZATION_REPORT.md`
3. RabbitMQ 官方文档：https://www.rabbitmq.com/documentation.html

---

**文档版本：** v1.0
**更新时间：** 2026-03-18
