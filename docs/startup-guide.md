# 错误日志系统启动指南

## 方式一：手动启动（推荐）

### 1. 启动 RabbitMQ

打开命令提示符（CMD）或 PowerShell，执行：

```cmd
cd D:\rabbitmq\rabbitmq-server-windows-4.1.5\rabbitmq_server-4.1.5\sbin
rabbitmq-server.bat
```

**启动成功标志**：
- 看到 "completed with X plugins" 的提示
- 端口 5672（AMQP）和 15672（管理界面）开始监听

**验证**：
- 访问管理界面：http://localhost:15672
- 用户名：guest
- 密码：guest

### 2. 启动 Elasticsearch（可选）

打开新的命令提示符，执行：

```cmd
cd D:\elasticsearch\elasticsearch-9.3.1-windows-x86_64\elasticsearch-9.3.1\bin
elasticsearch.bat
```

**注意**：
- 首次启动会生成密码，请记录下来
- 如果端口 9200 被占用，可以修改配置文件 `config/elasticsearch.yml`
- 或者在配置文件中修改端口为 9201

**验证**：
```bash
curl http://localhost:9200
```

### 3. 启动应用

```bash
cd D:\Go_Project\my_cloud_disk\core
go run core.go
```

---

## 方式二：使用 Docker（如果已安装）

### 1. 启动 RabbitMQ

```bash
docker run -d --name rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management
```

### 2. 启动 Elasticsearch

```bash
docker run -d --name elasticsearch \
  -p 9200:9200 -p 9300:9300 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.15.0
```

---

## 测试错误日志系统

### 测试 1：只使用文件日志（不需要 ES）

1. 只启动 RabbitMQ
2. 启动应用
3. 触发一个错误（如访问不存在的接口）
4. 查看日志文件：`tail -f logs/error.log`

### 测试 2：完整测试（RabbitMQ + ES）

1. 启动 RabbitMQ 和 Elasticsearch
2. 启动应用
3. 触发错误
4. 查看文件日志：`tail -f logs/error.log`
5. 查看 ES 日志：
   ```bash
   curl -X GET "localhost:9200/error_logs/_search?pretty"
   ```

---

## 常见问题

### Q1：RabbitMQ 启动失败

**可能原因**：
- Erlang 未安装或版本不匹配
- 端口被占用

**解决方案**：
1. 检查 Erlang 是否安装：`erl -version`
2. 检查端口：`netstat -ano | findstr "5672"`
3. 如果端口被占用，修改配置文件或停止占用端口的程序

### Q2：Elasticsearch 启动失败

**可能原因**：
- 端口 9200 被占用（如 Cpolar）
- 内存不足

**解决方案**：
1. 修改端口：编辑 `config/elasticsearch.yml`
   ```yaml
   http.port: 9201
   ```
2. 调整内存：编辑 `config/jvm.options`
   ```
   -Xms512m
   -Xmx512m
   ```

### Q3：应用启动时提示 MQ 连接失败

**解决方案**：
- 这是警告，不会阻止应用启动
- 确保 RabbitMQ 已启动
- 检查配置文件中的连接信息

---

## 监控和管理

### RabbitMQ 管理界面

访问：http://localhost:15672

功能：
- 查看队列状态
- 查看消息数量
- 查看消费者状态
- 手动发送/接收消息

### Elasticsearch 查询

```bash
# 查看所有索引
curl http://localhost:9200/_cat/indices?v

# 查看错误日志
curl -X GET "localhost:9200/error_logs/_search?pretty"

# 按时间查询
curl -X GET "localhost:9200/error_logs/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "range": {
      "timestamp": {
        "gte": "2026-03-15T00:00:00"
      }
    }
  }
}
'

# 按用户查询
curl -X GET "localhost:9200/error_logs/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "user_id": "user-001"
    }
  }
}
'
```

---

## 停止服务

### 停止 RabbitMQ

在 RabbitMQ 的命令提示符窗口按 `Ctrl+C`

或者：
```cmd
cd D:\rabbitmq\rabbitmq-server-windows-4.1.5\rabbitmq_server-4.1.5\sbin
rabbitmqctl.bat stop
```

### 停止 Elasticsearch

在 Elasticsearch 的命令提示符窗口按 `Ctrl+C`

### 停止应用

在应用的命令提示符窗口按 `Ctrl+C`

---

## 下一步

1. 手动启动 RabbitMQ
2. 启动应用并测试
3. 查看日志文件确认系统正常工作
4. （可选）启动 Elasticsearch 进行完整测试

