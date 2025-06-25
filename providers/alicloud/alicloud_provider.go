package alicloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yu1ec/go-anyllm/providers"
	"github.com/yu1ec/go-anyllm/types"
)

func init() {
	// 注册阿里云服务商创建函数
	providers.RegisterAliCloudProvider(func(config providers.ProviderConfig) (providers.Provider, error) {
		return NewAliCloudProvider(config)
	})
}

// AliCloudProvider 阿里云服务商实现
type AliCloudProvider struct {
	config     *AliCloudConfig
	httpClient *http.Client
}

// AliCloudConfig 阿里云配置
type AliCloudConfig struct {
	APIKey       string // DashScope API Key
	BaseURL      string
	Timeout      int
	ExtraHeaders map[string]string
}

// GetAPIKey 实现ProviderConfig接口
func (c *AliCloudConfig) GetAPIKey() string {
	return c.APIKey
}

// GetBaseURL 实现ProviderConfig接口
func (c *AliCloudConfig) GetBaseURL() string {
	if c.BaseURL == "" {
		return "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	return c.BaseURL
}

// GetTimeout 实现ProviderConfig接口
func (c *AliCloudConfig) GetTimeout() int {
	if c.Timeout == 0 {
		return 120
	}
	return c.Timeout
}

// GetExtraHeaders 实现ProviderConfig接口
func (c *AliCloudConfig) GetExtraHeaders() map[string]string {
	return c.ExtraHeaders
}

// NewAliCloudProvider 创建阿里云服务商
func NewAliCloudProvider(config providers.ProviderConfig) (*AliCloudProvider, error) {
	aliConfig, ok := config.(*AliCloudConfig)
	if !ok {
		// 尝试从通用配置创建
		aliConfig = &AliCloudConfig{
			APIKey:       config.GetAPIKey(),
			BaseURL:      config.GetBaseURL(),
			Timeout:      config.GetTimeout(),
			ExtraHeaders: config.GetExtraHeaders(),
		}
	}

	provider := &AliCloudProvider{
		config: aliConfig,
		httpClient: &http.Client{
			Timeout: time.Duration(aliConfig.GetTimeout()) * time.Second,
		},
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// GetName 实现Provider接口
func (p *AliCloudProvider) GetName() string {
	return "alicloud"
}

// GetBaseURL 实现Provider接口
func (p *AliCloudProvider) GetBaseURL() string {
	return p.config.GetBaseURL()
}

// ValidateConfig 实现Provider接口
func (p *AliCloudProvider) ValidateConfig() error {
	if p.config.GetAPIKey() == "" {
		return fmt.Errorf("alicloud: API key is required")
	}
	return nil
}

// SetupHeaders 实现Provider接口
func (p *AliCloudProvider) SetupHeaders(headers map[string]string) {
	headers["Authorization"] = "Bearer " + p.config.GetAPIKey()
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	// 添加额外的头部
	for k, v := range p.config.GetExtraHeaders() {
		headers[k] = v
	}
}

// CreateChatCompletion 实现Provider接口
func (p *AliCloudProvider) CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	// 兼容模式下直接使用OpenAI格式
	req.Stream = false

	// 发送请求
	respBody, err := p.doOpenAIRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	// 读取响应
	body, err := io.ReadAll(respBody)
	if err != nil {
		return nil, err
	}

	// 直接解析为OpenAI响应格式
	var resp types.ChatCompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateChatCompletionStream 实现Provider接口
func (p *AliCloudProvider) CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (io.ReadCloser, error) {
	// 兼容模式下直接使用OpenAI格式
	req.Stream = true

	// 发送请求
	return p.doOpenAIRequest(ctx, req)
}

// doOpenAIRequest 发送兼容OpenAI格式的HTTP请求
func (p *AliCloudProvider) doOpenAIRequest(ctx context.Context, req *types.ChatCompletionRequest) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/chat/completions", p.GetBaseURL())

	// 序列化请求体
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	headers := make(map[string]string)
	p.SetupHeaders(headers)
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("alicloud: HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// AliCloudRequest 阿里云请求格式
type AliCloudRequest struct {
	Model      string             `json:"model"`
	Input      AliCloudInput      `json:"input"`
	Parameters AliCloudParameters `json:"parameters"`
}

// AliCloudInput 阿里云输入格式
type AliCloudInput struct {
	Messages []AliCloudMessage `json:"messages"`
}

// AliCloudMessage 阿里云消息格式
type AliCloudMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AliCloudParameters 阿里云参数格式
type AliCloudParameters struct {
	MaxTokens         int      `json:"max_tokens,omitempty"`
	Temperature       float32  `json:"temperature,omitempty"`
	TopP              float32  `json:"top_p,omitempty"`
	IncrementalOutput bool     `json:"incremental_output,omitempty"`
	Stop              []string `json:"stop,omitempty"`

	// 思考控制参数
	EnableThinking *bool `json:"enable_thinking,omitempty"` // 是否开启思考模式
	ThinkingBudget *int  `json:"thinking_budget,omitempty"` // 思考预算token数
}

// AliCloudResponse 阿里云响应格式
type AliCloudResponse struct {
	Output    AliCloudOutput `json:"output"`
	Usage     AliCloudUsage  `json:"usage"`
	RequestId string         `json:"request_id"`
}

// AliCloudOutput 阿里云输出格式
type AliCloudOutput struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

// AliCloudUsage 阿里云使用统计
type AliCloudUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// convertToAliCloudRequest 转换为阿里云请求格式
func (p *AliCloudProvider) convertToAliCloudRequest(req *types.ChatCompletionRequest) *AliCloudRequest {
	aliReq := &AliCloudRequest{
		Model: req.Model,
		Input: AliCloudInput{},
		Parameters: AliCloudParameters{
			Stop: req.Stop,
		},
	}

	// 设置模型映射
	if req.Model == "" {
		aliReq.Model = "qwen-turbo" // 默认模型
	}

	// 转换消息
	for _, msg := range req.Messages {
		aliMsg := AliCloudMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		aliReq.Input.Messages = append(aliReq.Input.Messages, aliMsg)
	}

	// 转换参数
	if req.MaxTokens != nil {
		aliReq.Parameters.MaxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		aliReq.Parameters.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		aliReq.Parameters.TopP = *req.TopP
	}

	// 处理阿里云特有的思考控制参数
	if req.EnableThinking != nil {
		aliReq.Parameters.EnableThinking = req.EnableThinking
	}
	if req.ThinkingBudget != nil {
		aliReq.Parameters.ThinkingBudget = req.ThinkingBudget
	}

	return aliReq
}

// convertToOpenAIResponse 转换为OpenAI响应格式
func (p *AliCloudProvider) convertToOpenAIResponse(resp *AliCloudResponse, model string) *types.ChatCompletionResponse {
	openaiResp := &types.ChatCompletionResponse{
		ID:      resp.RequestId,
		Object:  "chat.completion",
		Created: types.GetCurrentTimestamp(),
		Model:   model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Message: &types.ChatCompletionMessage{
					Role:    types.RoleAssistant,
					Content: resp.Output.Text,
				},
				FinishReason: p.convertFinishReason(resp.Output.FinishReason),
			},
		},
		Usage: &types.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	return openaiResp
}

// convertFinishReason 转换完成原因
func (p *AliCloudProvider) convertFinishReason(reason string) string {
	switch reason {
	case "stop":
		return types.FinishReasonStop
	case "length":
		return types.FinishReasonLength
	default:
		return reason
	}
}
