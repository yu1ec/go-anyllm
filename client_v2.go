package deepseek

import (
	"context"

	"github.com/yu1ec/go-anyllm/providers"
	"github.com/yu1ec/go-anyllm/response"
	"github.com/yu1ec/go-anyllm/types"

	// 导入服务商包以触发注册
	_ "github.com/yu1ec/go-anyllm/providers/alicloud"
	_ "github.com/yu1ec/go-anyllm/providers/deepseek"
	_ "github.com/yu1ec/go-anyllm/providers/openai"
)

// UnifiedClient 统一的AI客户端接口，兼容OpenAI风格
type UnifiedClient interface {
	// CreateChatCompletion 创建聊天完成（非流式）
	CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)

	// CreateChatCompletionStream 创建聊天完成（流式）
	CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (response.StreamReader, error)

	// GetProvider 获取当前使用的服务商
	GetProvider() providers.Provider

	// GetProviderName 获取服务商名称
	GetProviderName() string
}

// ClientConfig 统一客户端配置
type ClientConfig struct {
	Provider     providers.ProviderType
	APIKey       string
	BaseURL      string
	Timeout      int
	ExtraHeaders map[string]string

	// 特定服务商配置
	OpenAIOrgID string // OpenAI组织ID
}

// unifiedClient 统一客户端实现
type unifiedClient struct {
	provider providers.Provider
	factory  providers.ProviderFactory
}

// NewUnifiedClient 创建统一客户端
func NewUnifiedClient(config *ClientConfig) (UnifiedClient, error) {
	factory := providers.NewDefaultProviderFactory()

	// 创建服务商配置
	var providerConfig providers.ProviderConfig
	switch config.Provider {
	case providers.ProviderOpenAI:
		providerConfig = &providers.GenericConfig{
			APIKey:       config.APIKey,
			BaseURL:      config.BaseURL,
			Timeout:      config.Timeout,
			ExtraHeaders: config.ExtraHeaders,
		}
		// 添加OpenAI特有的组织ID
		if config.OpenAIOrgID != "" {
			if providerConfig.GetExtraHeaders() == nil {
				providerConfig.(*providers.GenericConfig).ExtraHeaders = make(map[string]string)
			}
			providerConfig.(*providers.GenericConfig).ExtraHeaders["OpenAI-Organization"] = config.OpenAIOrgID
		}
	default:
		providerConfig = &providers.GenericConfig{
			APIKey:       config.APIKey,
			BaseURL:      config.BaseURL,
			Timeout:      config.Timeout,
			ExtraHeaders: config.ExtraHeaders,
		}
	}

	// 创建服务商实例
	provider, err := factory.CreateProvider(config.Provider, providerConfig)
	if err != nil {
		return nil, err
	}

	return &unifiedClient{
		provider: provider,
		factory:  factory,
	}, nil
}

// CreateChatCompletion 实现UnifiedClient接口
func (c *unifiedClient) CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	return c.provider.CreateChatCompletion(ctx, req)
}

// CreateChatCompletionStream 实现UnifiedClient接口
func (c *unifiedClient) CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (response.StreamReader, error) {
	respBody, err := c.provider.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	return response.NewStreamReader(respBody), nil
}

// GetProvider 实现UnifiedClient接口
func (c *unifiedClient) GetProvider() providers.Provider {
	return c.provider
}

// GetProviderName 实现UnifiedClient接口
func (c *unifiedClient) GetProviderName() string {
	return c.provider.GetName()
}

// 便捷函数，保持向后兼容

// NewDeepSeekClient 创建DeepSeek客户端（向后兼容）
func NewDeepSeekClient(apiKey string) (UnifiedClient, error) {
	config := &ClientConfig{
		Provider: providers.ProviderDeepSeek,
		APIKey:   apiKey,
		Timeout:  120,
	}
	return NewUnifiedClient(config)
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(apiKey string, orgID ...string) (UnifiedClient, error) {
	config := &ClientConfig{
		Provider: providers.ProviderOpenAI,
		APIKey:   apiKey,
		Timeout:  120,
	}
	if len(orgID) > 0 {
		config.OpenAIOrgID = orgID[0]
	}
	return NewUnifiedClient(config)
}

// NewAliCloudClient 创建阿里云客户端
func NewAliCloudClient(apiKey string) (UnifiedClient, error) {
	config := &ClientConfig{
		Provider: providers.ProviderAliCloud,
		APIKey:   apiKey,
		Timeout:  120,
	}
	return NewUnifiedClient(config)
}

// NewClientWithProvider 使用指定服务商创建客户端
func NewClientWithProvider(providerType providers.ProviderType, apiKey string, baseURL ...string) (UnifiedClient, error) {
	config := &ClientConfig{
		Provider: providerType,
		APIKey:   apiKey,
		Timeout:  120,
	}
	if len(baseURL) > 0 {
		config.BaseURL = baseURL[0]
	}
	return NewUnifiedClient(config)
}
