package types

import "time"

// ChatCompletionRequest OpenAI兼容的聊天完成请求
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	MaxTokens        *int                    `json:"max_tokens,omitempty"`
	Temperature      *float32                `json:"temperature,omitempty"`
	TopP             *float32                `json:"top_p,omitempty"`
	N                *int                    `json:"n,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Stop             []string                `json:"stop,omitempty"`
	PresencePenalty  *float32                `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int          `json:"logit_bias,omitempty"`
	User             string                  `json:"user,omitempty"`
	ResponseFormat   *ResponseFormat         `json:"response_format,omitempty"`
	Tools            []Tool                  `json:"tools,omitempty"`
	ToolChoice       interface{}             `json:"tool_choice,omitempty"`
	StreamOptions    *StreamOptions          `json:"stream_options,omitempty"`
	Logprobs         bool                    `json:"logprobs,omitempty"`
	TopLogprobs      *int                    `json:"top_logprobs,omitempty"`

	// 阿里云特有参数
	EnableThinking *bool `json:"enable_thinking,omitempty"` // 是否开启思考模式
	ThinkingBudget *int  `json:"thinking_budget,omitempty"` // 思考预算token数
}

// ChatCompletionMessage 聊天消息
type ChatCompletionMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`

	// DeepSeek特有字段
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

// ChatCompletionResponse OpenAI兼容的聊天完成响应
type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             *Usage                 `json:"usage,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
}

// ChatCompletionChoice 聊天完成选择
type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *ChatCompletionMessage `json:"message,omitempty"`
	Delta        *ChatCompletionMessage `json:"delta,omitempty"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	Logprobs     *LogprobsContent       `json:"logprobs,omitempty"`
}

// Usage 使用统计
type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	PromptCacheHitTokens    int                      `json:"prompt_cache_hit_tokens,omitempty"`
	PromptCacheMissTokens   int                      `json:"prompt_cache_miss_tokens,omitempty"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

// PromptTokensDetails 提示词token详情
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// CompletionTokensDetails 完成token详情
type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// ResponseFormat 响应格式
type ResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// StreamOptions 流选项
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// LogprobsContent 日志概率内容
type LogprobsContent struct {
	Content []TokenLogprob `json:"content,omitempty"`
}

// TokenLogprob token日志概率
type TokenLogprob struct {
	Token       string       `json:"token"`
	Logprob     float64      `json:"logprob"`
	Bytes       []int        `json:"bytes,omitempty"`
	TopLogprobs []TopLogprob `json:"top_logprobs,omitempty"`
}

// TopLogprob 顶部日志概率
type TopLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// ChatCompletionStreamResponse 流式响应
type ChatCompletionStreamResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             *Usage                 `json:"usage,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
}

// 角色常量
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// 响应格式常量
const (
	ResponseFormatText       = "text"
	ResponseFormatJSONObject = "json_object"
)

// 完成原因常量
const (
	FinishReasonStop          = "stop"
	FinishReasonLength        = "length"
	FinishReasonToolCalls     = "tool_calls"
	FinishReasonContentFilter = "content_filter"
	FinishReasonFunctionCall  = "function_call"
)

// 工具类型常量
const (
	ToolTypeFunction = "function"
)

// 辅助函数
func ToPtr[T any](v T) *T {
	return &v
}

// GetCurrentTimestamp 获取当前时间戳
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// 阿里云思考控制参数辅助函数

// WithEnableThinking 设置是否开启思考模式
func (r *ChatCompletionRequest) WithEnableThinking(enable bool) *ChatCompletionRequest {
	r.EnableThinking = &enable
	return r
}

// WithThinkingBudget 设置思考预算token数
func (r *ChatCompletionRequest) WithThinkingBudget(budget int) *ChatCompletionRequest {
	r.ThinkingBudget = &budget
	return r
}

// IsThinkingEnabled 检查是否开启了思考模式
func (r *ChatCompletionRequest) IsThinkingEnabled() bool {
	return r.EnableThinking != nil && *r.EnableThinking
}

// GetThinkingBudget 获取思考预算，如果未设置返回0
func (r *ChatCompletionRequest) GetThinkingBudget() int {
	if r.ThinkingBudget == nil {
		return 0
	}
	return *r.ThinkingBudget
}
