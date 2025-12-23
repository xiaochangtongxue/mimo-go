package mimo_go

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DefaultBaseURL 定义了小米MiMo API的官方入口地址
const DefaultBaseURL = "https://api.xiaomimimo.com/v1"

// ClientConfig 封装了客户端的配置参数
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// DefaultConfig 生成默认配置
func DefaultConfig(apiKey string) ClientConfig {
	return ClientConfig{
		BaseURL:    DefaultBaseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// Client 是与MiMo API交互的主入口
type Client struct {
	config         ClientConfig
	requestBuilder requestBuilder // 用于封装请求构建逻辑，处理特殊的Header
}

// NewClient 创建一个新的客户端实例
func NewClient(apiKey string) *Client {
	return NewClientWithConfig(DefaultConfig(apiKey))
}

// NewClientWithConfig 允许使用自定义配置创建客户端
func NewClientWithConfig(config ClientConfig) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	return &Client{
		config:         config,
		requestBuilder: newRequestBuilder(config.APIKey),
	}
}

// requestBuilder 负责构建带有正确鉴权信息的请求
// MiMo API 使用 "api-key" header 而不是标准的 "Authorization: Bearer"
type requestBuilder struct {
	apiKey string
}

func newRequestBuilder(apiKey string) requestBuilder {
	return requestBuilder{
		apiKey: apiKey,
	}
}

// Build 构造带有正确 Header 的 HTTP 请求
func (b *requestBuilder) Build(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		reqBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("json marshal error: %w", err)
		}
		bodyReader = bytes.NewBuffer(reqBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置通用 Header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 关键：设置 MiMo 专属鉴权头
	req.Header.Set("api-key", b.apiKey)

	return req, nil
}

// handleErrorResp 处理非 200 响应，提取错误信息
func (c *Client) handleErrorResp(resp *http.Response) error {
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	// 尝试解析错误体，如果解析失败则返回状态码错误
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("api error: status_code=%d", resp.StatusCode)
	}
	return fmt.Errorf("api error: status_code=%d type=%s code=%s message=%s",
		resp.StatusCode, errResp.Error.Type, errResp.Error.Code, errResp.Error.Message)
}
