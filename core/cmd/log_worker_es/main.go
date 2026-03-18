package main

import (
	"bytes"
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/rabbitmq"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/core-api.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	log.Println("========================================")
	log.Println("Elasticsearch 日志写入工作进程启动中...")
	log.Println("========================================")

	// 2. 初始化 RabbitMQ
	mq, err := rabbitmq.NewRabbitMQ(c.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("初始化 RabbitMQ 失败: %v", err)
	}
	defer mq.Close()

	// 3. 声明交换机（fanout 类型）
	err = mq.DeclareExchange(c.RabbitMQ.LogExchange, "fanout")
	if err != nil {
		log.Fatalf("声明交换机失败: %v", err)
	}

	// 4. 声明 ES 日志队列
	err = mq.DeclareQueue(c.RabbitMQ.ESLogQueue)
	if err != nil {
		log.Fatalf("声明队列失败: %v", err)
	}

	// 5. 绑定队列到交换机
	err = mq.BindQueueToExchange(c.RabbitMQ.ESLogQueue, c.RabbitMQ.LogExchange, "")
	if err != nil {
		log.Fatalf("绑定队列失败: %v", err)
	}

	// 6. 初始化 Elasticsearch 客户端
	esClient, err := initElasticsearch(c)
	if err != nil {
		log.Fatalf("初始化 Elasticsearch 失败: %v", err)
	}

	// 测试 ES 连接
	info, err := esClient.Info()
	if err != nil {
		log.Fatalf("连接 Elasticsearch 失败: %v", err)
	}
	defer info.Body.Close()
	log.Println("✓ Elasticsearch 连接成功")

	// 创建索引模板
	if err := createIndexTemplate(esClient, c.Elasticsearch.IndexPrefix); err != nil {
		log.Printf("警告: 创建索引模板失败: %v (不影响日志写入)", err)
	}

	// 7. 定义 Elasticsearch 写入处理函数
	esLogHandler := func(logMsg *rabbitmq.LogMessage) error {
		return writeToElasticsearch(esClient, c.Elasticsearch.IndexPrefix, logMsg)
	}

	// 8. 创建消费者
	consumer := rabbitmq.NewLogConsumer(mq, c.RabbitMQ.ESLogQueue, esLogHandler)

	// 9. 启动消费者（阻塞运行）
	log.Println("✓ ES 日志消费者已启动，监听队列:", c.RabbitMQ.ESLogQueue)
	err = consumer.Start()
	if err != nil {
		log.Fatalf("消费者启动失败: %v", err)
	}
}

// initElasticsearch 初始化 Elasticsearch 客户端
func initElasticsearch(c config.Config) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: c.Elasticsearch.Addresses,
	}

	// 如果配置了用户名和密码
	if c.Elasticsearch.Username != "" {
		cfg.Username = c.Elasticsearch.Username
		cfg.Password = c.Elasticsearch.Password
	}

	return elasticsearch.NewClient(cfg)
}

// writeToElasticsearch 写入日志到 Elasticsearch（带重试机制）
func writeToElasticsearch(client *elasticsearch.Client, indexPrefix string, logMsg *rabbitmq.LogMessage) error {
	ctx := context.Background()

	// 1. 构建索引名称（按日期：logs-2026-03-18）
	indexName := fmt.Sprintf("%s-%s", indexPrefix, time.Now().Format("2006-01-02"))

	// 2. 序列化日志消息
	body, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("序列化日志失败: %w", err)
	}

	// 3. 重试写入 Elasticsearch（最多3次）
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		res, err := client.Index(
			indexName,
			bytes.NewReader(body),
			//Context 就是给请求加一个"保险"，防止请求无限等待卡死程序。
			client.Index.WithContext(ctx),
		)

		if err == nil {
			defer res.Body.Close()

			// 检查响应
			if !res.IsError() {
				log.Printf("✓ ES 日志写入成功: index=%s, trace_id=%s", indexName, logMsg.TraceID)
				return nil
			}

			log.Printf("ES 返回错误 (尝试 %d/%d): %s", i+1, maxRetries, res.String())
		} else {
			log.Printf("写入 ES 失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
		}

		// 如果不是最后一次重试，等待后重试
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	return fmt.Errorf("写入 ES 失败，已重试 %d 次", maxRetries)
}

// createIndexTemplate 创建索引模板（优化日志存储）  存储在es中，等到有匹配的index就会自动使用模板
func createIndexTemplate(client *elasticsearch.Client, indexPrefix string) error {
	ctx := context.Background()

	// 定义索引模板
	template := map[string]interface{}{
		//匹配所有的logs前缀的index
		"index_patterns": []string{fmt.Sprintf("%s-*", indexPrefix)},
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 0,
				"refresh_interval":   "5s",
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"timestamp": map[string]interface{}{
						"type":   "date",
						"format": "yyyy-MM-dd HH:mm:ss",
					},
					"level": map[string]interface{}{
						"type": "keyword",
					},
					"trace_id": map[string]interface{}{
						"type": "keyword",
					},
					"user_id": map[string]interface{}{
						"type": "keyword",
					},
					"method": map[string]interface{}{
						"type": "keyword",
					},
					"path": map[string]interface{}{
						"type": "keyword",
					},
					"message": map[string]interface{}{
						"type": "text",
					},
					"stack_trace": map[string]interface{}{
						"type": "text",
					},
				},
			},
		},
	}

	body, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("序列化模板失败: %w", err)
	}

	// 构建完整的URL
	templateName := fmt.Sprintf("%s-template", indexPrefix)
	url := fmt.Sprintf("/_index_template/%s", templateName)

	// 使用 Perform 方法发送请求
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	//// ES 客户端内部会拼接成:
	//  PUT http://localhost:9200/_index_template/logs-template
	res, err := client.Perform(req)
	if err != nil {
		return fmt.Errorf("执行请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("ES 返回错误: %s", res.Status)
	}

	log.Printf("✓ 索引模板创建成功: %s", templateName)
	return nil
}
