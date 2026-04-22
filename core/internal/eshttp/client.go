package eshttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client 是一个非常轻量的 Elasticsearch HTTP 客户端。
// 它只封装当前项目命令行工具真正需要的少量接口，
// 目的是替换体积很大的 go-elasticsearch/v8，降低 Windows 下的编译压力。
type Client struct {
	baseURL    string
	httpClient *http.Client
	password   string
	username   string
}

func NewClient(addresses []string, username string, password string) (*Client, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("elasticsearch address is empty")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(addresses[0]), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("elasticsearch address is empty")
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		username: username,
		password: password,
	}, nil
}

// Info 用于验证 ES 连接是否可用。
func (c *Client) Info(ctx context.Context) (*http.Response, error) {
	return c.Perform(ctx, http.MethodGet, "/", nil, nil)
}

// IndexDocument 向指定索引写入一条文档。
func (c *Client) IndexDocument(ctx context.Context, index string, body io.Reader) (*http.Response, error) {
	return c.Perform(ctx, http.MethodPost, fmt.Sprintf("/%s/_doc", index), body, nil)
}

// Search 在指定索引模式上执行查询。
func (c *Client) Search(ctx context.Context, indexPattern string, body io.Reader, params map[string]string) (*http.Response, error) {
	return c.Perform(ctx, http.MethodPost, fmt.Sprintf("/%s/_search", indexPattern), body, params)
}

// Perform 是最底层的 HTTP 请求发送方法。
func (c *Client) Perform(ctx context.Context, method string, path string, body io.Reader, params map[string]string) (*http.Response, error) {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		query := url.Values{}
		for key, value := range params {
			if value == "" {
				continue
			}
			query.Set(key, value)
		}
		if encoded := query.Encode(); encoded != "" {
			fullURL += "?" + encoded
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	return c.httpClient.Do(req)
}

// ResponseError 把 ES 返回的 HTTP 错误包装成更容易读的错误信息。
func ResponseError(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("elasticsearch response is nil")
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if len(data) == 0 {
		return fmt.Errorf("elasticsearch returned %s", resp.Status)
	}
	return fmt.Errorf("elasticsearch returned %s: %s", resp.Status, bytes.TrimSpace(data))
}
