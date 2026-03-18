package main

import (
	"bytes"
	"cloud_disk/core/internal/config"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zeromicro/go-zero/core/conf"
)

var (
	configFile = flag.String("f", "etc/core-api.yaml", "配置文件路径")
	limit      = flag.Int("n", 10, "显示日志条数")
	level      = flag.String("level", "", "过滤日志级别 (ERROR, FATAL, PANIC)")
	traceID    = flag.String("trace", "", "过滤 trace_id")
)

func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化 ES 客户端
	cfg := elasticsearch.Config{
		Addresses: c.Elasticsearch.Addresses,
	}
	if c.Elasticsearch.Username != "" {
		cfg.Username = c.Elasticsearch.Username
		cfg.Password = c.Elasticsearch.Password
	}

	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("初始化 ES 客户端失败: %v", err)
	}

	// 构建查询
	query := buildQuery(*level, *traceID)

	// 执行搜索
	logs, err := searchLogs(esClient, c.Elasticsearch.IndexPrefix, query, *limit)
	if err != nil {
		log.Fatalf("搜索日志失败: %v", err)
	}

	// 打印结果
	fmt.Printf("\n找到 %d 条日志:\n", len(logs))
	fmt.Println("========================================")
	for i, logEntry := range logs {
		fmt.Printf("\n[%d] %s\n", i+1, formatLog(logEntry))
	}
}

// buildQuery 构建查询条件
func buildQuery(level, traceID string) map[string]interface{} {
	must := []map[string]interface{}{}

	if level != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"level": level,
			},
		})
	}

	if traceID != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"trace_id": traceID,
			},
		})
	}

	if len(must) == 0 {
		return map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
	}
}

// searchLogs 搜索日志
func searchLogs(client *elasticsearch.Client, indexPrefix string, query map[string]interface{}, size int) ([]map[string]interface{}, error) {
	ctx := context.Background()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("编码查询失败: %w", err)
	}

	res, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(fmt.Sprintf("%s-*", indexPrefix)),
		client.Search.WithBody(&buf),
		client.Search.WithSize(size),
		client.Search.WithSort("timestamp:desc"),
	)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES 返回错误: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	logs := make([]map[string]interface{}, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		logs = append(logs, source)
	}

	return logs, nil
}

// formatLog 格式化日志输出
func formatLog(logEntry map[string]interface{}) string {
	timestamp := logEntry["timestamp"]
	level := logEntry["level"]
	traceID := logEntry["trace_id"]
	message := logEntry["message"]
	method := logEntry["method"]
	path := logEntry["path"]

	result := fmt.Sprintf("时间: %v\n", timestamp)
	result += fmt.Sprintf("级别: %v\n", level)
	result += fmt.Sprintf("TraceID: %v\n", traceID)
	result += fmt.Sprintf("请求: %v %v\n", method, path)
	result += fmt.Sprintf("消息: %v\n", message)

	if stackTrace, ok := logEntry["stack_trace"]; ok && stackTrace != "" {
		result += fmt.Sprintf("堆栈:\n%v\n", stackTrace)
	}

	return result
}
