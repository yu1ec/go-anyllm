package alicloud

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	// 思考模式超时配置 (秒)
	ThinkingTimeout int // 思考阶段总超时时间，默认300秒(5分钟)
	OutputTimeout   int // 输出阶段无数据超时时间，默认60秒(1分钟)
	ReadTimeout     int // 单次读取超时时间，默认30秒
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
	// 检查是否开启了思考模式
	// 根据阿里云文档，思考模式只支持流式输出，所以需要特殊处理
	if req.EnableThinking != nil && *req.EnableThinking {
		// 如果开启了思考模式，使用流式调用然后聚合结果
		return p.handleThinkingModeNonStream(ctx, req)
	}

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

// handleThinkingModeNonStream 处理思考模式的非流式调用
// 由于思考模式只支持流式输出，所以我们需要内部使用流式调用然后聚合结果
func (p *AliCloudProvider) handleThinkingModeNonStream(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	// 创建流式调用
	stream, err := p.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	// 使用优化的流式读取器
	return p.readStreamWithTimeout(ctx, stream)
}

// readStreamWithTimeout 带超时控制的流式读取器
func (p *AliCloudProvider) readStreamWithTimeout(ctx context.Context, stream io.ReadCloser) (*types.ChatCompletionResponse, error) {
	// 读取状态跟踪
	var response types.ChatCompletionResponse
	var contentBuilder strings.Builder
	var reasoningContentBuilder strings.Builder

	isThinkingPhase := true // 是否在思考阶段
	lastDataTime := time.Now()
	thinkingStartTime := time.Now()

	// 分阶段超时配置（使用配置值或默认值）
	thinkingTimeout := time.Duration(p.getThinkingTimeout()) * time.Second
	outputTimeout := time.Duration(p.getOutputTimeout()) * time.Second
	readTimeout := time.Duration(p.getReadTimeout()) * time.Second

	// 创建带缓冲的读取器
	reader := bufio.NewReaderSize(stream, 8192) // 8KB缓冲区

	// 创建定时器
	readTimer := time.NewTimer(readTimeout)
	defer readTimer.Stop()

	for {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return p.buildPartialResponse(response, contentBuilder, reasoningContentBuilder, "context_cancelled")
		default:
		}

		// 检查分阶段超时
		now := time.Now()
		if isThinkingPhase {
			// 思考阶段：检查总思考时间
			if now.Sub(thinkingStartTime) > thinkingTimeout {
				return nil, fmt.Errorf("alicloud: thinking timeout after %v", thinkingTimeout)
			}
		} else {
			// 输出阶段：检查最后数据时间
			if now.Sub(lastDataTime) > outputTimeout {
				return p.buildPartialResponse(response, contentBuilder, reasoningContentBuilder, "output_timeout")
			}
		}

		// 非阻塞读取一行
		line, err := p.readLineWithTimeout(reader, readTimer, readTimeout)
		if err != nil {
			if err == io.EOF {
				break // 正常结束
			}

			// 检查是否是超时错误
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
				// 尝试返回部分结果而不是完全失败
				if contentBuilder.Len() > 0 || reasoningContentBuilder.Len() > 0 {
					return p.buildPartialResponse(response, contentBuilder, reasoningContentBuilder, "timeout_partial")
				}
				return nil, fmt.Errorf("alicloud: stream read timeout: %v", err)
			}

			return nil, fmt.Errorf("alicloud: error reading stream: %v", err)
		}

		// 更新最后数据接收时间
		lastDataTime = now

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 处理 SSE 格式
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				break
			}

			// 解析 JSON
			var chunk types.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue // 跳过无法解析的行
			}

			// 聚合响应信息
			if response.ID == "" {
				response.ID = chunk.ID
				response.Object = "chat.completion"
				response.Created = chunk.Created
				response.Model = chunk.Model
			}

			// 处理选择项
			if len(chunk.Choices) > 0 {
				choice := chunk.Choices[0]

				// 检测阶段切换：从思考到输出
				deltaContent := getContentAsString(choice.Delta.Content)
				if isThinkingPhase && deltaContent != "" {
					isThinkingPhase = false
					lastDataTime = now // 重置输出阶段计时
				}

				// 聚合内容
				if deltaContent != "" {
					contentBuilder.WriteString(deltaContent)
				}

				// 聚合思考内容
				if choice.Delta.ReasoningContent != "" {
					reasoningContentBuilder.WriteString(choice.Delta.ReasoningContent)
				}

				// 设置完成原因
				if choice.FinishReason != "" {
					if len(response.Choices) == 0 {
						response.Choices = append(response.Choices, types.ChatCompletionChoice{
							Index: 0,
							Message: &types.ChatCompletionMessage{
								Role: types.RoleAssistant,
							},
						})
					}
					response.Choices[0].FinishReason = choice.FinishReason
				}
			}

			// 处理使用信息
			if chunk.Usage != nil {
				response.Usage = chunk.Usage
			}
		}
	}

	return p.buildFinalResponse(response, contentBuilder, reasoningContentBuilder), nil
}

// readLineWithTimeout 带超时的行读取
func (p *AliCloudProvider) readLineWithTimeout(reader *bufio.Reader, timer *time.Timer, timeout time.Duration) (string, error) {
	// 重置定时器
	timer.Reset(timeout)

	// 使用channel来实现带超时的读取
	type readResult struct {
		line string
		err  error
	}

	ch := make(chan readResult, 1)
	go func() {
		line, err := reader.ReadString('\n')
		if err != nil {
			ch <- readResult{"", err}
			return
		}
		// 移除行尾的换行符
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		ch <- readResult{line, nil}
	}()

	select {
	case result := <-ch:
		return result.line, result.err
	case <-timer.C:
		return "", fmt.Errorf("read timeout after %v", timeout)
	}
}

// buildPartialResponse 构建部分响应（用于超时恢复）
func (p *AliCloudProvider) buildPartialResponse(response types.ChatCompletionResponse, contentBuilder, reasoningContentBuilder strings.Builder, reason string) (*types.ChatCompletionResponse, error) {
	if len(response.Choices) == 0 {
		response.Choices = append(response.Choices, types.ChatCompletionChoice{
			Index: 0,
			Message: &types.ChatCompletionMessage{
				Role: types.RoleAssistant,
			},
		})
	}

	// 设置部分内容
	content := contentBuilder.String()
	if content == "" && reasoningContentBuilder.Len() > 0 {
		// 如果只有思考内容，提供一个默认回复
		content = "[思考中断] 由于超时，思考过程未完成。"
	}

	response.Choices[0].Message.Content = content
	response.Choices[0].FinishReason = reason

	// 生成默认ID和时间戳（如果没有的话）
	if response.ID == "" {
		response.ID = fmt.Sprintf("chatcmpl-partial-%d", time.Now().Unix())
		response.Object = "chat.completion"
		response.Created = time.Now().Unix()
	}

	return &response, nil
}

// buildFinalResponse 构建最终响应
func (p *AliCloudProvider) buildFinalResponse(response types.ChatCompletionResponse, contentBuilder, reasoningContentBuilder strings.Builder) *types.ChatCompletionResponse {
	if len(response.Choices) == 0 {
		response.Choices = append(response.Choices, types.ChatCompletionChoice{
			Index: 0,
			Message: &types.ChatCompletionMessage{
				Role: types.RoleAssistant,
			},
		})
	}

	// 设置聚合的内容
	response.Choices[0].Message.Content = contentBuilder.String()

	// 如果有思考内容，可以添加到响应中（这里可以根据需要决定如何处理）
	if reasoningContentBuilder.Len() > 0 {
		// 可以将思考内容添加到 content 的开头，或者存储在其他字段中
		// 这里我们选择不在最终响应中包含思考过程，只返回最终答案
		// 如果需要包含思考过程，可以修改这里的逻辑
	}

	return &response
}

// 超时配置获取方法
func (p *AliCloudProvider) getThinkingTimeout() int {
	if p.config.ThinkingTimeout > 0 {
		return p.config.ThinkingTimeout
	}
	return 300 // 默认5分钟
}

func (p *AliCloudProvider) getOutputTimeout() int {
	if p.config.OutputTimeout > 0 {
		return p.config.OutputTimeout
	}
	return 60 // 默认1分钟
}

func (p *AliCloudProvider) getReadTimeout() int {
	if p.config.ReadTimeout > 0 {
		return p.config.ReadTimeout
	}
	return 30 // 默认30秒
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

		// 尝试读取错误响应体中的详细信息
		errorBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("alicloud: HTTP %d (failed to read error response: %v)", resp.StatusCode, readErr)
		}

		// 尝试解析错误响应为JSON格式
		var errorResp map[string]interface{}
		if parseErr := json.Unmarshal(errorBody, &errorResp); parseErr == nil {
			// 如果成功解析为JSON，尝试提取错误信息
			if errorObj, exists := errorResp["error"]; exists {
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					// 构建详细的错误信息
					var errorMsg string
					if message, exists := errorMap["message"]; exists {
						errorMsg = fmt.Sprintf("%v", message)
					}
					if code, exists := errorMap["code"]; exists {
						if errorMsg != "" {
							errorMsg = fmt.Sprintf("%s (code: %v)", errorMsg, code)
						} else {
							errorMsg = fmt.Sprintf("code: %v", code)
						}
					}
					if errorType, exists := errorMap["type"]; exists {
						if errorMsg != "" {
							errorMsg = fmt.Sprintf("%s, type: %v", errorMsg, errorType)
						} else {
							errorMsg = fmt.Sprintf("type: %v", errorType)
						}
					}

					// 如果有request_id，也包含进去
					if requestId, exists := errorResp["request_id"]; exists {
						errorMsg = fmt.Sprintf("%s, request_id: %v", errorMsg, requestId)
					}

					if errorMsg != "" {
						return nil, fmt.Errorf("alicloud: HTTP %d - %s", resp.StatusCode, errorMsg)
					}
				}
			}
		}

		// 如果无法解析JSON或没有标准错误结构，返回原始响应体
		return nil, fmt.Errorf("alicloud: HTTP %d - %s", resp.StatusCode, string(errorBody))
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
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // 支持string或多模态内容数组
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
