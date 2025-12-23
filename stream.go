package mimo_go

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ChatCompletionStreamResponse 流式响应包
type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}

type ChatCompletionStreamChoice struct {
	Index        int                             `json:"index"`
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason string                          `json:"finish_reason"`
}

// ChatCompletionStreamChoiceDelta 增量内容
type ChatCompletionStreamChoiceDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`

	// ReasoningContent 是 MiMo 流式响应的特有字段 (Thinking Mode)
	ReasoningContent string `json:"reasoning_content,omitempty"`

	// ToolCalls 流式返回的工具调用片段
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ChatCompletionStream 管理流式连接
type ChatCompletionStream struct {
	reader   *bufio.Reader
	response *http.Response
}

// CreateChatCompletionStream 发起流式请求
func (c *Client) CreateChatCompletionStream(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionStream, error) {
	request.Stream = true
	endpoint := c.config.BaseURL + "/chat/completions"

	req, err := c.requestBuilder.Build(ctx, http.MethodPost, endpoint, request)
	if err != nil {
		return nil, err
	}

	// 设置 SSE 必需的 Headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close() // 只有错误时才立即关闭
		return nil, c.handleErrorResp(resp)
	}

	return &ChatCompletionStream{
		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}, nil
}

// Recv 阻塞读取下一个响应块
// 返回 io.EOF 表示流结束
func (stream *ChatCompletionStream) Recv() (response ChatCompletionStreamResponse, err error) {
	for {
		// 读取一行
		line, err := stream.reader.ReadBytes('\n')
		if err != nil {
			return response, err
		}

		line = bytes.TrimSpace(line)
		// 跳过空行和非数据行（如 keep-alive 注释）
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))

		// 检查结束标记 [DONE]
		if string(data) == "[DONE]" {
			return response, io.EOF
		}

		// 反序列化
		if err := json.Unmarshal(data, &response); err != nil {
			return response, fmt.Errorf("unmarshal error: %w, data: %s", err, string(data))
		}

		return response, nil
	}
}

// Close 关闭流
func (stream *ChatCompletionStream) Close() error {
	return stream.response.Body.Close()
}
