package mimo_go

import (
	"context"
	"encoding/json"
	"net/http"
)

// 常用角色定义
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleTool      = "tool"
)

// ToolType 工具类型
type ToolType string

const ToolTypeFunction ToolType = "function"

// Tool 定义模型可调用的工具
type Tool struct {
	Type     ToolType           `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 定义函数的具体元数据
type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters"` // JSON Schema 对象
}

// ToolCall 表示模型生成的工具调用请求
type ToolCall struct {
	ID       string           `json:"id"`
	Type     ToolType         `json:"type"`
	Function FunctionCallData `json:"function"`
}

// FunctionCallData 具体的函数调用参数
type FunctionCallData struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON 格式的字符串参数
}

// ChatCompletionMessage 表示对话历史中的单条消息
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
	// ReasoningContent 存储模型生成的思维链内容 (Thinking Mode)
	ReasoningContent string `json:"reasoning_content,omitempty"`

	// ToolCalls 模型生成的工具调用列表 (Assistant 消息)
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID 关联的工具调用ID (Tool 消息)
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ThinkingConfig 定义思考模式的具体配置
type ThinkingConfig struct {
	Type string `json:"type"` // "enabled" 或 "disabled"
}

// 预定义的思考模式配置
var (
	ThinkingEnabled  = &ThinkingConfig{Type: "enabled"}
	ThinkingDisabled = &ThinkingConfig{Type: "disabled"}
)

// ChatCompletionRequest 定义了调用 /chat/completions 接口的请求体
type ChatCompletionRequest struct {
	Model               string                  `json:"model"`
	Messages            []ChatCompletionMessage `json:"messages"`
	MaxTokens           int                     `json:"max_tokens,omitempty"`
	MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
	Temperature         float32                 `json:"temperature,omitempty"`
	TopP                float32                 `json:"top_p,omitempty"`
	Stream              bool                    `json:"stream,omitempty"`
	Stop                []string                `json:"stop,omitempty"`
	FrequencyPenalty    float32                 `json:"frequency_penalty,omitempty"`
	PresencePenalty     float32                 `json:"presence_penalty,omitempty"`

	// Thinking 配置模型的深度思考模式
	Thinking *ThinkingConfig `json:"thinking,omitempty"`

	// Tools 模型可用的工具列表
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice 控制模型是否/如何调用工具 ("none", "auto", "required" 或 指定工具)
	ToolChoice interface{} `json:"tool_choice,omitempty"`
}

// ChatCompletionResponse API 响应体
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

// ChatCompletionChoice 单个生成选项
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"` // 复用 Message 结构
	FinishReason string                `json:"finish_reason"`
}

// Usage Token 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CreateChatCompletion 发起非流式对话请求
func (c *Client) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (response ChatCompletionResponse, err error) {
	// 强制设置 Stream 为 false
	request.Stream = false

	endpoint := c.config.BaseURL + "/chat/completions"
	req, err := c.requestBuilder.Build(ctx, http.MethodPost, endpoint, request)
	if err != nil {
		return response, err
	}

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return response, c.handleErrorResp(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}
