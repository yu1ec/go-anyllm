package openai

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
	// 注册OpenAI服务商创建函数
	providers.RegisterOpenAIProvider(func(config providers.ProviderConfig) (providers.Provider, error) {
		return NewOpenAIProvider(config)
	})
}

// OpenAIProvider OpenAI服务商实现
type OpenAIProvider struct {
	config     *OpenAIConfig
	httpClient *http.Client
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey       string
	BaseURL      string
	OrgID        string
	Timeout      int
	ExtraHeaders map[string]string
}

// GetAPIKey 实现ProviderConfig接口
func (c *OpenAIConfig) GetAPIKey() string {
	return c.APIKey
}

// GetBaseURL 实现ProviderConfig接口
func (c *OpenAIConfig) GetBaseURL() string {
	if c.BaseURL == "" {
		return "https://api.openai.com/v1"
	}
	return c.BaseURL
}

// GetTimeout 实现ProviderConfig接口
func (c *OpenAIConfig) GetTimeout() int {
	if c.Timeout == 0 {
		return 120
	}
	return c.Timeout
}

// GetExtraHeaders 实现ProviderConfig接口
func (c *OpenAIConfig) GetExtraHeaders() map[string]string {
	return c.ExtraHeaders
}

// NewOpenAIProvider 创建OpenAI服务商
func NewOpenAIProvider(config providers.ProviderConfig) (*OpenAIProvider, error) {
	openaiConfig, ok := config.(*OpenAIConfig)
	if !ok {
		// 尝试从通用配置创建
		openaiConfig = &OpenAIConfig{
			APIKey:       config.GetAPIKey(),
			BaseURL:      config.GetBaseURL(),
			Timeout:      config.GetTimeout(),
			ExtraHeaders: config.GetExtraHeaders(),
		}
	}

	provider := &OpenAIProvider{
		config: openaiConfig,
		httpClient: &http.Client{
			Timeout: time.Duration(openaiConfig.GetTimeout()) * time.Second,
		},
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// GetName 实现Provider接口
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetBaseURL 实现Provider接口
func (p *OpenAIProvider) GetBaseURL() string {
	return p.config.GetBaseURL()
}

// ValidateConfig 实现Provider接口
func (p *OpenAIProvider) ValidateConfig() error {
	if p.config.GetAPIKey() == "" {
		return fmt.Errorf("openai: API key is required")
	}
	return nil
}

// SetupHeaders 实现Provider接口
func (p *OpenAIProvider) SetupHeaders(headers map[string]string) {
	headers["Authorization"] = "Bearer " + p.config.GetAPIKey()
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	// 添加组织ID（如果设置）
	if p.config.OrgID != "" {
		headers["OpenAI-Organization"] = p.config.OrgID
	}

	// 添加额外的头部
	for k, v := range p.config.GetExtraHeaders() {
		headers[k] = v
	}
}

// CreateChatCompletion 实现Provider接口
func (p *OpenAIProvider) CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	// OpenAI使用标准格式，无需转换
	req.Stream = false

	// 发送请求
	respBody, err := p.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	// 读取响应
	body, err := io.ReadAll(respBody)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var resp types.ChatCompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateChatCompletionStream 实现Provider接口
func (p *OpenAIProvider) CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (io.ReadCloser, error) {
	// OpenAI使用标准格式，无需转换
	req.Stream = true

	// 发送请求
	return p.doRequest(ctx, req)
}

// doRequest 发送HTTP请求
func (p *OpenAIProvider) doRequest(ctx context.Context, req *types.ChatCompletionRequest) (io.ReadCloser, error) {
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
		return nil, fmt.Errorf("openai: HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}
