package deepseek

import (
	"testing"

	"github.com/yu1ec/go-anyllm/providers"
	"github.com/yu1ec/go-anyllm/types"

	// 导入服务商包以触发注册
	_ "github.com/yu1ec/go-anyllm/providers/alicloud"
	_ "github.com/yu1ec/go-anyllm/providers/deepseek"
	_ "github.com/yu1ec/go-anyllm/providers/openai"
)

func TestUnifiedClientCreation(t *testing.T) {
	tests := []struct {
		name         string
		providerType providers.ProviderType
		apiKey       string
		expectError  bool
	}{
		{
			name:         "DeepSeek with API key",
			providerType: providers.ProviderDeepSeek,
			apiKey:       "test-api-key",
			expectError:  false,
		},
		{
			name:         "OpenAI with API key",
			providerType: providers.ProviderOpenAI,
			apiKey:       "test-api-key",
			expectError:  false,
		},
		{
			name:         "AliCloud with API key",
			providerType: providers.ProviderAliCloud,
			apiKey:       "test-api-key",
			expectError:  false,
		},
		{
			name:         "DeepSeek without API key",
			providerType: providers.ProviderDeepSeek,
			apiKey:       "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithProvider(tt.providerType, tt.apiKey)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && client != nil {
				providerName := client.GetProviderName()
				if providerName == "" {
					t.Errorf("Provider name should not be empty")
				}
				t.Logf("Created client with provider: %s", providerName)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("NewDeepSeekClient", func(t *testing.T) {
		client, err := NewDeepSeekClient("test-api-key")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if client.GetProviderName() != "deepseek" {
			t.Errorf("Expected provider name 'deepseek', got '%s'", client.GetProviderName())
		}
	})

	t.Run("NewOpenAIClient", func(t *testing.T) {
		client, err := NewOpenAIClient("test-api-key")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if client.GetProviderName() != "openai" {
			t.Errorf("Expected provider name 'openai', got '%s'", client.GetProviderName())
		}
	})

	t.Run("NewAliCloudClient", func(t *testing.T) {
		client, err := NewAliCloudClient("test-api-key")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if client.GetProviderName() != "alicloud" {
			t.Errorf("Expected provider name 'alicloud', got '%s'", client.GetProviderName())
		}
	})
}

func TestUnifiedClientConfig(t *testing.T) {
	config := &ClientConfig{
		Provider: providers.ProviderDeepSeek,
		APIKey:   "test-api-key",
		BaseURL:  "https://custom-endpoint.com",
		Timeout:  60,
		ExtraHeaders: map[string]string{
			"Custom-Header": "test-value",
		},
	}

	client, err := NewUnifiedClient(config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if client.GetProviderName() != "deepseek" {
		t.Errorf("Expected provider name 'deepseek', got '%s'", client.GetProviderName())
	}

	provider := client.GetProvider()
	if provider.GetBaseURL() != "https://custom-endpoint.com" {
		t.Errorf("Expected custom base URL, got '%s'", provider.GetBaseURL())
	}
}

func TestChatCompletionRequest(t *testing.T) {
	// 创建一个模拟的请求，不实际发送网络请求
	req := &types.ChatCompletionRequest{
		Model: "test-model",
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleUser,
				Content: "Hello, world!",
			},
		},
		MaxTokens:   types.ToPtr(100),
		Temperature: types.ToPtr(float32(0.7)),
		TopP:        types.ToPtr(float32(0.9)),
		Stream:      false,
	}

	// 验证请求结构
	if req.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", req.Model)
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != types.RoleUser {
		t.Errorf("Expected role 'user', got '%s'", req.Messages[0].Role)
	}

	if req.Messages[0].Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", req.Messages[0].Content)
	}

	if *req.MaxTokens != 100 {
		t.Errorf("Expected max tokens 100, got %d", *req.MaxTokens)
	}

	if *req.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *req.Temperature)
	}
}

func TestTypesConstants(t *testing.T) {
	// 测试角色常量
	if types.RoleSystem != "system" {
		t.Errorf("Expected RoleSystem to be 'system', got '%s'", types.RoleSystem)
	}
	if types.RoleUser != "user" {
		t.Errorf("Expected RoleUser to be 'user', got '%s'", types.RoleUser)
	}
	if types.RoleAssistant != "assistant" {
		t.Errorf("Expected RoleAssistant to be 'assistant', got '%s'", types.RoleAssistant)
	}

	// 测试响应格式常量
	if types.ResponseFormatText != "text" {
		t.Errorf("Expected ResponseFormatText to be 'text', got '%s'", types.ResponseFormatText)
	}
	if types.ResponseFormatJSONObject != "json_object" {
		t.Errorf("Expected ResponseFormatJSONObject to be 'json_object', got '%s'", types.ResponseFormatJSONObject)
	}

	// 测试完成原因常量
	if types.FinishReasonStop != "stop" {
		t.Errorf("Expected FinishReasonStop to be 'stop', got '%s'", types.FinishReasonStop)
	}
	if types.FinishReasonLength != "length" {
		t.Errorf("Expected FinishReasonLength to be 'length', got '%s'", types.FinishReasonLength)
	}
}

func TestToPtr(t *testing.T) {
	// 测试ToPtr辅助函数
	intVal := 42
	intPtr := types.ToPtr(intVal)
	if *intPtr != 42 {
		t.Errorf("Expected *intPtr to be 42, got %d", *intPtr)
	}

	strVal := "test"
	strPtr := types.ToPtr(strVal)
	if *strPtr != "test" {
		t.Errorf("Expected *strPtr to be 'test', got '%s'", *strPtr)
	}

	floatVal := float32(3.14)
	floatPtr := types.ToPtr(floatVal)
	if *floatPtr != 3.14 {
		t.Errorf("Expected *floatPtr to be 3.14, got %f", *floatPtr)
	}
}

// 基准测试
func BenchmarkClientCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client, err := NewDeepSeekClient("test-api-key")
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
		_ = client
	}
}

func BenchmarkRequestCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req := &types.ChatCompletionRequest{
			Model: "test-model",
			Messages: []types.ChatCompletionMessage{
				{
					Role:    types.RoleUser,
					Content: "Hello, world!",
				},
			},
			MaxTokens:   types.ToPtr(100),
			Temperature: types.ToPtr(float32(0.7)),
		}
		_ = req
	}
}
