package providers

import (
	"fmt"
)

// DefaultProviderFactory 默认服务商工厂实现
type DefaultProviderFactory struct{}

// NewDefaultProviderFactory 创建默认服务商工厂
func NewDefaultProviderFactory() ProviderFactory {
	return &DefaultProviderFactory{}
}

// CreateProvider 创建服务商实例
func (f *DefaultProviderFactory) CreateProvider(providerType ProviderType, config ProviderConfig) (Provider, error) {
	switch providerType {
	case ProviderDeepSeek:
		// 动态创建DeepSeek服务商
		return createDeepSeekProvider(config)
	case ProviderOpenAI:
		// 动态创建OpenAI服务商
		return createOpenAIProvider(config)
	case ProviderAliCloud:
		// 动态创建阿里云服务商
		return createAliCloudProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// SupportedProviders 返回支持的服务商列表
func (f *DefaultProviderFactory) SupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderDeepSeek,
		ProviderOpenAI,
		ProviderAliCloud,
		ProviderBaidu,
		ProviderTencent,
	}
}

// GenericConfig 通用配置实现
type GenericConfig struct {
	APIKey       string
	BaseURL      string
	Timeout      int
	ExtraHeaders map[string]string
}

// GetAPIKey 实现ProviderConfig接口
func (c *GenericConfig) GetAPIKey() string {
	return c.APIKey
}

// GetBaseURL 实现ProviderConfig接口
func (c *GenericConfig) GetBaseURL() string {
	return c.BaseURL
}

// GetTimeout 实现ProviderConfig接口
func (c *GenericConfig) GetTimeout() int {
	if c.Timeout == 0 {
		return 120
	}
	return c.Timeout
}

// GetExtraHeaders 实现ProviderConfig接口
func (c *GenericConfig) GetExtraHeaders() map[string]string {
	return c.ExtraHeaders
}

// NewGenericConfig 创建通用配置
func NewGenericConfig(apiKey, baseURL string, timeout int) *GenericConfig {
	return &GenericConfig{
		APIKey:       apiKey,
		BaseURL:      baseURL,
		Timeout:      timeout,
		ExtraHeaders: make(map[string]string),
	}
}

// SetExtraHeader 设置额外的请求头
func (c *GenericConfig) SetExtraHeader(key, value string) {
	if c.ExtraHeaders == nil {
		c.ExtraHeaders = make(map[string]string)
	}
	c.ExtraHeaders[key] = value
}

// 这些函数将在各自的服务商包中实现，这里只是声明
var (
	createDeepSeekProvider func(config ProviderConfig) (Provider, error)
	createOpenAIProvider   func(config ProviderConfig) (Provider, error)
	createAliCloudProvider func(config ProviderConfig) (Provider, error)
)

// RegisterDeepSeekProvider 注册DeepSeek服务商创建函数
func RegisterDeepSeekProvider(creator func(config ProviderConfig) (Provider, error)) {
	createDeepSeekProvider = creator
}

// RegisterOpenAIProvider 注册OpenAI服务商创建函数
func RegisterOpenAIProvider(creator func(config ProviderConfig) (Provider, error)) {
	createOpenAIProvider = creator
}

// RegisterAliCloudProvider 注册阿里云服务商创建函数
func RegisterAliCloudProvider(creator func(config ProviderConfig) (Provider, error)) {
	createAliCloudProvider = creator
}
