package main

import (
	"bytes"
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/eshttp"
	"cloud_disk/core/internal/rabbitmq"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/core-api.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	log.Println("========================================")
	log.Println("Elasticsearch 日志写入工作进程启动中...")
	log.Println("========================================")

	mq, err := rabbitmq.NewRabbitMQ(c.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("初始化 RabbitMQ 失败: %v", err)
	}
	defer mq.Close()

	if err := mq.DeclareExchange(c.RabbitMQ.LogExchange, "fanout"); err != nil {
		log.Fatalf("声明交换机失败: %v", err)
	}
	if err := mq.DeclareQueue(c.RabbitMQ.ESLogQueue); err != nil {
		log.Fatalf("声明队列失败: %v", err)
	}
	if err := mq.BindQueueToExchange(c.RabbitMQ.ESLogQueue, c.RabbitMQ.LogExchange, ""); err != nil {
		log.Fatalf("绑定队列失败: %v", err)
	}

	esClient, err := initElasticsearch(c)
	if err != nil {
		log.Fatalf("初始化 Elasticsearch 失败: %v", err)
	}

	info, err := esClient.Info(context.Background())
	if err != nil {
		log.Fatalf("连接 Elasticsearch 失败: %v", err)
	}
	if info.StatusCode >= 400 {
		log.Fatalf("连接 Elasticsearch 失败: %v", eshttp.ResponseError(info))
	}
	info.Body.Close()
	log.Println("✅ Elasticsearch 连接成功")

	if err := createIndexTemplate(esClient, c.Elasticsearch.IndexPrefix); err != nil {
		log.Printf("警告: 创建索引模板失败: %v (不影响日志写入)", err)
	}

	esLogHandler := func(logMsg *rabbitmq.LogMessage) error {
		return writeToElasticsearch(esClient, c.Elasticsearch.IndexPrefix, logMsg)
	}

	consumer := rabbitmq.NewLogConsumer(mq, c.RabbitMQ.ESLogQueue, esLogHandler)
	log.Printf("✅ ES 日志消费者已启动，监听队列: %s", c.RabbitMQ.ESLogQueue)
	if err := consumer.Start(); err != nil {
		log.Fatalf("消费者启动失败: %v", err)
	}
}

func initElasticsearch(c config.Config) (*eshttp.Client, error) {
	return eshttp.NewClient(c.Elasticsearch.Addresses, c.Elasticsearch.Username, c.Elasticsearch.Password)
}

func writeToElasticsearch(client *eshttp.Client, indexPrefix string, logMsg *rabbitmq.LogMessage) error {
	ctx := context.Background()
	indexName := fmt.Sprintf("%s-%s", indexPrefix, time.Now().Format("2006-01-02"))

	body, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("序列化日志失败: %w", err)
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		res, err := client.IndexDocument(ctx, indexName, bytes.NewReader(body))
		if err == nil {
			if res.StatusCode < 400 {
				res.Body.Close()
				log.Printf("✅ ES 日志写入成功: index=%s, trace_id=%s", indexName, logMsg.TraceID)
				return nil
			}
			err = eshttp.ResponseError(res)
		}

		log.Printf("写入 ES 失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	return fmt.Errorf("写入 ES 失败，已重试 %d 次", maxRetries)
}

func createIndexTemplate(client *eshttp.Client, indexPrefix string) error {
	ctx := context.Background()

	template := map[string]interface{}{
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
					"level":       map[string]interface{}{"type": "keyword"},
					"trace_id":    map[string]interface{}{"type": "keyword"},
					"user_id":     map[string]interface{}{"type": "keyword"},
					"method":      map[string]interface{}{"type": "keyword"},
					"path":        map[string]interface{}{"type": "keyword"},
					"message":     map[string]interface{}{"type": "text"},
					"stack_trace": map[string]interface{}{"type": "text"},
				},
			},
		},
	}

	body, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("序列化模板失败: %w", err)
	}

	templateName := fmt.Sprintf("%s-template", indexPrefix)
	res, err := client.Perform(ctx, "PUT", fmt.Sprintf("/_index_template/%s", templateName), bytes.NewReader(body), nil)
	if err != nil {
		return fmt.Errorf("执行请求失败: %w", err)
	}
	if res.StatusCode >= 400 {
		return eshttp.ResponseError(res)
	}
	res.Body.Close()

	log.Printf("✅ 索引模板创建成功: %s", templateName)
	return nil
}
