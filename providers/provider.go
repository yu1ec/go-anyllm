package providers

import (
	"context"
	"io"

	"github.com/yu1ec/go-anyllm/types"
)

// Provider 定义了AI服务商的通用接口
type Provider interface {
	// GetName 返回服务商名称
	GetName() string

	// GetBaseURL 返回服务商的基础URL
	GetBaseURL() string

	// CreateChatCompletion 创建聊天完成请求（非流式）
	CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)

	// CreateChatCompletionStream 创建聊天完成请求（流式）
	CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (io.ReadCloser, error)

	// ValidateConfig 验证配置是否有效
	ValidateConfig() error

	// SetupHeaders 设置请求头
	SetupHeaders(headers map[string]string)
}

// ProviderConfig 服务商配置接口
type ProviderConfig interface {
	GetAPIKey() string
	GetBaseURL() string
	GetTimeout() int
	GetExtraHeaders() map[string]string
}

// ProviderType 服务商类型
type ProviderType string

const (
	ProviderDeepSeek ProviderType = "deepseek"
	ProviderOpenAI   ProviderType = "openai"
	ProviderAliCloud ProviderType = "alicloud"
	ProviderBaidu    ProviderType = "baidu"
	ProviderTencent  ProviderType = "tencent"
)

// ProviderFactory 服务商工厂接口
type ProviderFactory interface {
	CreateProvider(providerType ProviderType, config ProviderConfig) (Provider, error)
	SupportedProviders() []ProviderType
}
