package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yu1ec/go-anyllm/internal"
	"github.com/yu1ec/go-anyllm/providers"
	"github.com/yu1ec/go-anyllm/request"
	"github.com/yu1ec/go-anyllm/response"
	"github.com/yu1ec/go-anyllm/types"
)

func init() {
	// 注册DeepSeek服务商创建函数
	providers.RegisterDeepSeekProvider(func(config providers.ProviderConfig) (providers.Provider, error) {
		return NewDeepSeekProvider(config)
	})
}

// DeepSeekProvider DeepSeek服务商实现
type DeepSeekProvider struct {
	config     *DeepSeekConfig
	httpClient *http.Client
}

// DeepSeekConfig DeepSeek配置
type DeepSeekConfig struct {
	APIKey       string
	BaseURL      string
	Timeout      int
	ExtraHeaders map[string]string
}

// GetAPIKey 实现ProviderConfig接口
func (c *DeepSeekConfig) GetAPIKey() string {
	return c.APIKey
}

// GetBaseURL 实现ProviderConfig接口
func (c *DeepSeekConfig) GetBaseURL() string {
	if c.BaseURL == "" {
		return internal.BASE_URL
	}
	return c.BaseURL
}

// GetTimeout 实现ProviderConfig接口
func (c *DeepSeekConfig) GetTimeout() int {
	if c.Timeout == 0 {
		return 120
	}
	return c.Timeout
}

// GetExtraHeaders 实现ProviderConfig接口
func (c *DeepSeekConfig) GetExtraHeaders() map[string]string {
	return c.ExtraHeaders
}

// NewDeepSeekProvider 创建DeepSeek服务商
func NewDeepSeekProvider(config providers.ProviderConfig) (*DeepSeekProvider, error) {
	deepseekConfig, ok := config.(*DeepSeekConfig)
	if !ok {
		// 尝试从通用配置创建
		deepseekConfig = &DeepSeekConfig{
			APIKey:       config.GetAPIKey(),
			BaseURL:      config.GetBaseURL(),
			Timeout:      config.GetTimeout(),
			ExtraHeaders: config.GetExtraHeaders(),
		}
	}

	provider := &DeepSeekProvider{
		config: deepseekConfig,
		httpClient: &http.Client{
			Timeout: time.Duration(deepseekConfig.GetTimeout()) * time.Second,
		},
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// GetName 实现Provider接口
func (p *DeepSeekProvider) GetName() string {
	return "deepseek"
}

// GetBaseURL 实现Provider接口
func (p *DeepSeekProvider) GetBaseURL() string {
	return p.config.GetBaseURL()
}

// ValidateConfig 实现Provider接口
func (p *DeepSeekProvider) ValidateConfig() error {
	if p.config.GetAPIKey() == "" {
		return fmt.Errorf("deepseek: API key is required")
	}
	return nil
}

// SetupHeaders 实现Provider接口
func (p *DeepSeekProvider) SetupHeaders(headers map[string]string) {
	headers["Authorization"] = "Bearer " + p.config.GetAPIKey()
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	// 添加额外的头部
	for k, v := range p.config.GetExtraHeaders() {
		headers[k] = v
	}
}

// CreateChatCompletion 实现Provider接口
func (p *DeepSeekProvider) CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	// 转换为DeepSeek内部请求格式
	deepseekReq := p.convertToDeepSeekRequest(req)
	deepseekReq.Stream = false

	// 发送请求
	respBody, err := p.doRequest(ctx, deepseekReq)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	// 读取响应
	body, err := io.ReadAll(respBody)
	if err != nil {
		return nil, err
	}

	// 解析DeepSeek响应
	var deepseekResp response.ChatCompletionsResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return nil, err
	}

	// 转换为通用响应格式
	return p.convertToOpenAIResponse(&deepseekResp), nil
}

// CreateChatCompletionStream 实现Provider接口
func (p *DeepSeekProvider) CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (io.ReadCloser, error) {
	// 转换为DeepSeek内部请求格式
	deepseekReq := p.convertToDeepSeekRequest(req)
	deepseekReq.Stream = true

	// 发送请求
	return p.doRequest(ctx, deepseekReq)
}

// doRequest 发送HTTP请求
func (p *DeepSeekProvider) doRequest(ctx context.Context, req *request.ChatCompletionsRequest) (io.ReadCloser, error) {
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
		return nil, fmt.Errorf("deepseek: HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// convertToDeepSeekRequest 转换为DeepSeek请求格式
func (p *DeepSeekProvider) convertToDeepSeekRequest(req *types.ChatCompletionRequest) *request.ChatCompletionsRequest {
	deepseekReq := &request.ChatCompletionsRequest{
		Model:            req.Model,
		Stream:           req.Stream,
		MaxTokens:        0,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Stop:             req.Stop,
	}

	if req.MaxTokens != nil {
		deepseekReq.MaxTokens = *req.MaxTokens
	}
	if req.FrequencyPenalty != nil {
		deepseekReq.FrequencyPenalty = *req.FrequencyPenalty
	}
	if req.PresencePenalty != nil {
		deepseekReq.PresencePenalty = int(*req.PresencePenalty)
	}

	// 转换消息
	for _, msg := range req.Messages {
		// 获取消息内容的字符串表示
		contentStr := getContentAsString(msg.Content)

		deepseekMsg := &request.Message{
			Role:    msg.Role,
			Content: contentStr,
			Name:    msg.Name,
		}
		if msg.ToolCallID != "" {
			deepseekMsg.ToolCallId = msg.ToolCallID
		}
		deepseekReq.Messages = append(deepseekReq.Messages, deepseekMsg)
	}

	// 转换响应格式
	if req.ResponseFormat != nil {
		deepseekReq.ResponseFormat = &request.ResponseFormat{
			Type: req.ResponseFormat.Type,
		}
	}

	// 转换流选项
	if req.StreamOptions != nil {
		deepseekReq.StreamOptions = &request.StreamOptions{
			IncludeUsage: req.StreamOptions.IncludeUsage,
		}
	}

	// 转换工具
	if len(req.Tools) > 0 {
		tools := make([]request.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = request.Tool{
				Type: tool.Type,
				Function: &request.ToolFunction{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		deepseekReq.Tools = &tools
	}

	// 转换工具选择
	if req.ToolChoice != nil {
		deepseekReq.ToolChoice = req.ToolChoice
	}

	// 转换logprobs
	deepseekReq.Logprobs = req.Logprobs
	if req.TopLogprobs != nil {
		deepseekReq.TopLogprobs = req.TopLogprobs
	}

	return deepseekReq
}

// convertToOpenAIResponse 转换为OpenAI响应格式
func (p *DeepSeekProvider) convertToOpenAIResponse(resp *response.ChatCompletionsResponse) *types.ChatCompletionResponse {
	openaiResp := &types.ChatCompletionResponse{
		ID:                resp.Id,
		Object:            resp.Object,
		Created:           int64(resp.Created),
		Model:             resp.Model,
		SystemFingerprint: resp.SystemFingerprint,
	}

	// 转换选择
	for _, choice := range resp.Choices {
		openaiChoice := types.ChatCompletionChoice{
			Index:        choice.Index,
			FinishReason: choice.FinishReason,
		}

		// 转换消息
		if choice.Message != nil {
			openaiChoice.Message = &types.ChatCompletionMessage{
				Role:             choice.Message.Role,
				Content:          choice.Message.Content,
				ReasoningContent: choice.Message.ReasoningContent,
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				for _, tc := range choice.Message.ToolCalls {
					openaiChoice.Message.ToolCalls = append(openaiChoice.Message.ToolCalls, types.ToolCall{
						ID:   tc.Id,
						Type: tc.Type,
						Function: types.ToolFunction{
							Name:       tc.Function.Name,
							Parameters: tc.Function.Arguments,
						},
					})
				}
			}
		}

		// 转换delta（流式响应）
		if choice.Delta != nil {
			openaiChoice.Delta = &types.ChatCompletionMessage{
				Content:          choice.Delta.Content,
				ReasoningContent: choice.Delta.ReasoningContent,
			}

			// 转换Delta中的工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					openaiChoice.Delta.ToolCalls = append(openaiChoice.Delta.ToolCalls, types.ToolCall{
						ID:   tc.Id,
						Type: tc.Type,
						Function: types.ToolFunction{
							Name:       tc.Function.Name,
							Parameters: tc.Function.Arguments,
						},
					})
				}
			}
		}

		openaiResp.Choices = append(openaiResp.Choices, openaiChoice)
	}

	// 转换使用统计
	if resp.Usage != nil {
		openaiResp.Usage = &types.Usage{
			PromptTokens:          resp.Usage.PromptTokens,
			CompletionTokens:      resp.Usage.CompletionTokens,
			TotalTokens:           resp.Usage.TotalTokens,
			PromptCacheHitTokens:  resp.Usage.PromptCacheHitTokens,
			PromptCacheMissTokens: resp.Usage.PromptCacheMissTokens,
			PromptTokensDetails: &types.PromptTokensDetails{
				CachedTokens: resp.Usage.PromptTokensDetails.CachedTokens,
			},
			CompletionTokensDetails: &types.CompletionTokensDetails{
				ReasoningTokens: resp.Usage.CompletionTokensDetails.ReasoningTokens,
			},
		}
	}

	return openaiResp
}

// getContentAsString 获取内容的字符串表示
func getContentAsString(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []types.MessageContent:
		// 只返回文本内容，忽略图像
		for _, part := range c {
			if part.Type == types.MessageContentTypeText {
				return part.Text
			}
		}
		return ""
	default:
		return ""
	}
}
