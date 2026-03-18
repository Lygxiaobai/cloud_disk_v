# Elasticsearch 日志系统实现说明

## 📋 已实现功能

### 1. ES 日志写入 Worker (`cmd/log_worker_es/main.go`)

**核心功能：**
- ✅ 从 RabbitMQ 消费日志消息
- ✅ 自动创建索引模板（优化日志存储）
- ✅ 按日期创建索引（`logs-2026-03-18`）
- ✅ 错误重试机制（最多 3 次，指数退避）
- ✅ 连接测试和健康检查

**索引模板配置：**
```yaml
settings:
  number_of_shards: 1
  number_of_replicas: 0
  refresh_interval: 5s

mappings:
  timestamp: date (yyyy-MM-dd HH:mm:ss)
  level: keyword
  trace_id: keyword
  user_id: keyword
  method: keyword
  path: keyword
  message: text
  stack_trace: text
```

### 2. ES 日志查询工具 (`cmd/query_es_logs/main.go`)

**支持的查询：**
```bash
# 查看最新 N 条日志
./bin/query_es_logs.exe -n 10

# 按级别过滤
./bin/query_es_logs.exe -level ERROR

# 按 trace_id 查询
./bin/query_es_logs.exe -trace abc123

# 组合查询
./bin/query_es_logs.exe -level ERROR -n 20
```

**输出格式：**
```
[1] 时间: 2026-03-18 15:30:45
级别: ERROR
TraceID: abc123
请求: POST /api/login
消息: 登录失败: 密码错误
堆栈: goroutine 1 [running]...
```

### 3. Docker 启动脚本 (`start_elasticsearch.sh`)

**功能：**
- 自动检测 ES 容器是否存在
- 创建或启动 ES 容器
- 配置单节点模式
- 禁用安全认证（开发环境）
- 限制内存使用（512MB）

### 4. 测试脚本 (`test_elasticsearch.sh`)

**测试内容：**
- ES 连接状态
- 索引模板检查
- 索引列表查看
- 日志查询测试
- 日志数量统计

## 🔧 配置说明

### core-api.yaml

```yaml
Elasticsearch:
  Addresses:
    - http://localhost:9200
  Username: ""          # 可选：ES 认证用户名
  Password: ""          # 可选：ES 认证密码
  IndexPrefix: logs     # 索引前缀
```

## 🚀 使用流程

### 1. 启动 Elasticsearch

```bash
bash start_elasticsearch.sh
```

### 2. 编译程序

```bash
cd core
go build -o bin/log_worker_es.exe cmd/log_worker_es/main.go
go build -o bin/query_es_logs.exe cmd/query_es_logs/main.go
```

### 3. 启动 ES Worker

```bash
./bin/log_worker_es.exe -f etc/core-api.yaml
```

**启动日志：**
```
========================================
Elasticsearch 日志写入工作进程启动中...
========================================
✓ Elasticsearch 连接成功
✓ 索引模板创建成功: logs-template
✓ ES 日志消费者已启动，监听队列: es_log_queue
```

### 4. 查询日志

```bash
# 查看最新日志
./bin/query_es_logs.exe -n 10

# 只看错误日志
./bin/query_es_logs.exe -level ERROR
```

## 📊 ES 操作详解

### 1. 索引模板创建

**作用：** 为所有 `logs-*` 索引预定义配置

**实现：**
```go
func createIndexTemplate(client *elasticsearch.Client, indexPrefix string) error {
    // 定义模板
    template := map[string]interface{}{
        "index_patterns": []string{fmt.Sprintf("%s-*", indexPrefix)},
        "template": map[string]interface{}{
            "settings": {...},
            "mappings": {...},
        },
    }

    // 发送 PUT 请求到 /_index_template/logs-template
    // ...
}
```

**验证：**
```bash
curl http://localhost:9200/_index_template/logs-template?pretty
```

### 2. 日志写入

**索引命名：** `logs-2026-03-18`（按日期自动创建）

**实现：**
```go
func writeToElasticsearch(client *elasticsearch.Client, indexPrefix string, logMsg *rabbitmq.LogMessage) error {
    // 构建索引名
    indexName := fmt.Sprintf("%s-%s", indexPrefix, time.Now().Format("2006-01-02"))

    // 序列化日志
    body, _ := json.Marshal(logMsg)

    // 写入 ES（带重试）
    for i := 0; i < 3; i++ {
        res, err := client.Index(indexName, bytes.NewReader(body))
        if err == nil && !res.IsError() {
            return nil
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
}
```

### 3. 日志查询

**查询构建：**
```go
// 按级别过滤
query := map[string]interface{}{
    "query": map[string]interface{}{
        "bool": map[string]interface{}{
            "must": []map[string]interface{}{
                {"term": map[string]interface{}{"level": "ERROR"}},
            },
        },
    },
}

// 执行搜索
res, _ := client.Search(
    client.Search.WithIndex("logs-*"),
    client.Search.WithBody(&buf),
    client.Search.WithSize(10),
    client.Search.WithSort("timestamp:desc"),
)
```

## 🔍 常用 ES 命令

### 查看索引

```bash
# 列出所有日志索引
curl http://localhost:9200/_cat/indices/logs-*?v

# 查看索引详情
curl http://localhost:9200/logs-2026-03-18?pretty
```

### 查询日志

```bash
# 查询所有日志
curl -X GET "http://localhost:9200/logs-*/_search?pretty&size=10&sort=timestamp:desc"

# 按级别查询
curl -X GET "http://localhost:9200/logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "term": {"level": "ERROR"}
  },
  "size": 10,
  "sort": [{"timestamp": "desc"}]
}
'

# 按 trace_id 查询
curl -X GET "http://localhost:9200/logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "term": {"trace_id": "abc123"}
  }
}
'
```

### 统计信息

```bash
# 统计日志总数
curl http://localhost:9200/logs-*/_count?pretty

# 按级别聚合
curl -X GET "http://localhost:9200/logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 0,
  "aggs": {
    "by_level": {
      "terms": {"field": "level"}
    }
  }
}
'
```

### 删除索引

```bash
# 删除指定日期的索引
curl -X DELETE http://localhost:9200/logs-2026-03-18

# 删除所有日志索引（谨慎！）
curl -X DELETE http://localhost:9200/logs-*
```

## 🎯 性能优化建议

### 1. 批量写入

当前是单条写入，可以改为批量：

```go
// 收集 100 条或 5 秒后批量写入
var buffer []LogMessage
if len(buffer) >= 100 || time.Since(lastFlush) > 5*time.Second {
    client.Bulk(...)
}
```

### 2. 索引生命周期管理（ILM）

自动删除旧日志：

```bash
curl -X PUT "http://localhost:9200/_ilm/policy/logs-policy?pretty" -H 'Content-Type: application/json' -d'
{
  "policy": {
    "phases": {
      "delete": {
        "min_age": "7d",
        "actions": {"delete": {}}
      }
    }
  }
}
'
```

### 3. 分片优化

根据日志量调整分片数：
- 小于 10GB/天：1 个分片
- 10-50GB/天：2-3 个分片
- 大于 50GB/天：5+ 个分片

## ✅ 总结

### 实现的核心功能

1. ✅ **ES 写入**：自动创建索引，按日期分割
2. ✅ **索引模板**：预定义字段类型和设置
3. ✅ **错误重试**：3 次重试，指数退避
4. ✅ **日志查询**：支持过滤、排序、分页
5. ✅ **Docker 集成**：一键启动 ES 容器
6. ✅ **测试工具**：完整的测试和验证脚本

### 技术亮点

- **按日期分割索引**：便于管理和删除旧数据
- **索引模板**：统一配置，自动应用
- **重试机制**：提高可靠性
- **查询工具**：方便开发调试

---

**文档版本：** v1.0
**创建时间：** 2026-03-18
