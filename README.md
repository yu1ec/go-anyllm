# 统一AI客户端 - 多服务商支持
这是一个兼容OpenAI风格的统一AI客户端，支持多个AI服务商，包括DeepSeek、OpenAI、阿里云等。

## 致谢

本项目参考了 [go-deepseek/deepseek](https://github.com/go-deepseek/deepseek) 项目的设计和实现。

## 特性

- 🌐 **多服务商支持**: DeepSeek、OpenAI、阿里云通义千问等
- 🔄 **OpenAI兼容**: 使用标准的OpenAI API格式
- 🚀 **统一接口**: 相同的代码可以切换不同的AI服务商
- �� **流式响应**: 支持实时流式聊天，支持迭代器模式
- 🔧 **灵活配置**: 支持自定义API端点和请求头
- 🔒 **向后兼容**: 保持与原有DeepSeek客户端的兼容性

## 快速开始

### 安装

```bash
go get github.com/yu1ec/go-anyllm
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yu1ec/go-anyllm"
    "github.com/yu1ec/go-anyllm/providers"
    "github.com/yu1ec/go-anyllm/types"
    
    // 导入服务商包
    _ "github.com/yu1ec/go-anyllm/providers/deepseek"
    _ "github.com/yu1ec/go-anyllm/providers/openai"
    _ "github.com/yu1ec/go-anyllm/providers/alicloud"
)

func main() {
    // 方式1: 使用便捷函数
    client, err := deepseek.NewDeepSeekClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    // 方式2: 使用统一配置
    client, err = deepseek.NewClientWithProvider(
        providers.ProviderOpenAI, 
        "your-openai-api-key",
    )
    if err != nil {
        log.Fatal(err)
    }

    // 创建聊天请求
    req := &types.ChatCompletionRequest{
        Model: "gpt-3.5-turbo", // 或 "deepseek-chat", "qwen-turbo"
        Messages: []types.ChatCompletionMessage{
            {
                Role:    types.RoleUser,
                Content: "你好！",
            },
        },
        MaxTokens:   types.ToPtr(100),
        Temperature: types.ToPtr(float32(0.7)),
    }

    // 发送请求
    resp, err := client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## 支持的服务商

### 1. DeepSeek

```go
client, err := deepseek.NewDeepSeekClient("your-deepseek-api-key")

// 支持的模型
req := &types.ChatCompletionRequest{
    Model: "deepseek-chat",     // 或 "deepseek-reasoner"
    // ...
}
```

### 2. OpenAI

```go
client, err := deepseek.NewOpenAIClient("your-openai-api-key", "org-id")

// 支持的模型
req := &types.ChatCompletionRequest{
    Model: "gpt-3.5-turbo",     // 或 "gpt-4", "gpt-4-turbo"
    // ...
}
```

### 3. 阿里云通义千问

```go
client, err := deepseek.NewAliCloudClient("your-dashscope-api-key")

// 支持的模型
req := &types.ChatCompletionRequest{
    Model: "qwen-turbo",        // 或 "qwen-plus", "qwen-max"
    // ...
}
```

## 高级配置

### 自定义配置

```go
config := &deepseek.ClientConfig{
    Provider:     providers.ProviderOpenAI,
    APIKey:       "your-api-key",
    BaseURL:      "https://custom-api-endpoint.com/v1",
    Timeout:      60,
    ExtraHeaders: map[string]string{
        "Custom-Header": "value",
    },
    OpenAIOrgID:  "your-org-id", // OpenAI专用
}

client, err := deepseek.NewUnifiedClient(config)
```

### 流式响应

新版本的 `StreamReader` 支持两种使用模式：

#### 1. 迭代器模式（推荐）

新增的迭代器模式提供了更直观和灵活的流处理方式：

- `Next() bool`: 移动到下一个响应，如果有下一个响应返回 `true`，否则返回 `false`
- `Current() *ChatCompletionsResponse`: 返回当前的响应，需要先调用 `Next()`
- `Error() error`: 返回最后一次操作的错误

```go
req := &types.ChatCompletionRequest{
    Model:  "deepseek-chat",
    Stream: true,
    Messages: []types.ChatCompletionMessage{
        {
            Role:    types.RoleUser,
            Content: "请写一首关于春天的诗",
        },
    },
}

stream, err := client.CreateChatCompletionStream(context.Background(), req)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

// 使用迭代器模式处理流式响应
fmt.Print("AI回复：")
for stream.Next() {
    chunk := stream.Current()
    if chunk != nil && len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
        content := chunk.Choices[0].Delta.Content
        fmt.Print(content)
    }
}

// 检查是否有错误
if err := stream.Error(); err != nil {
    log.Printf("流处理错误: %v", err)
}

fmt.Println("\n\n流式响应完成")
```

##### 收集完整响应

```go
func collectStreamResponse(stream response.StreamReader) (string, error) {
    var fullContent strings.Builder
    
    for stream.Next() {
        chunk := stream.Current()
        if chunk != nil && len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
            content := chunk.Choices[0].Delta.Content
            fullContent.WriteString(content)
        }
    }
    
    if err := stream.Error(); err != nil {
        return "", err
    }
    
    return fullContent.String(), nil
}
```

##### 处理工具调用

```go
func handleToolCalls(stream response.StreamReader) error {
    for stream.Next() {
        chunk := stream.Current()
        if chunk != nil && len(chunk.Choices) > 0 {
            choice := chunk.Choices[0]
            
            // 处理delta内容
            if choice.Delta != nil {
                if choice.Delta.Content != "" {
                    fmt.Print(choice.Delta.Content)
                }
                
                // 处理工具调用
                if len(choice.Delta.ToolCalls) > 0 {
                    for _, toolCall := range choice.Delta.ToolCalls {
                        fmt.Printf("工具调用: %s\n", toolCall.Function.Name)
                    }
                }
            }
            
            // 检查完成原因
            if choice.FinishReason != "" {
                fmt.Printf("\n完成原因: %s\n", choice.FinishReason)
            }
        }
    }
    
    return stream.Error()
}
```

#### 2. 传统模式（兼容性支持）

原有的 `Read()` 方法仍然可用，保持向后兼容：

```go
req := &types.ChatCompletionRequest{
    Model:  "deepseek-chat",
    Stream: true,
    Messages: []types.ChatCompletionMessage{
        {
            Role:    types.RoleUser,
            Content: "请写一首诗",
        },
    },
}

stream, err := client.CreateChatCompletionStream(context.Background(), req)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

// 传统方式仍然有效
for {
    response, err := stream.Read()
    if err != nil {
        if err == io.EOF {
            break // 正常结束
        }
        log.Printf("错误: %v", err)
        break
    }
    
    // 处理响应
    if len(response.Choices) > 0 && response.Choices[0].Delta != nil {
        fmt.Print(response.Choices[0].Delta.Content)
    }
}
```

#### 注意事项

1. **不要混合使用**：不要在同一个流上同时使用 `Next()/Current()` 和 `Read()` 方法
2. **先调用Next()**：必须先调用 `Next()` 才能调用 `Current()`
3. **错误检查**：在迭代结束后记得检查 `Error()` 方法
4. **一次性使用**：每个流只能被消费一次，不能重复迭代

#### 迭代器模式的优势

- **更直观**：类似于其他语言中的迭代器模式
- **更安全**：自动处理EOF和错误状态
- **更灵活**：可以在循环中更容易地添加条件判断和错误处理
- **更清晰**：代码结构更清晰，易于理解和维护

## API参考

### 主要接口

```go
type UnifiedClient interface {
    CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)
    CreateChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (*StreamReader, error)
    GetProvider() providers.Provider
    GetProviderName() string
}
```

### 请求类型

```go
type ChatCompletionRequest struct {
    Model            string                 `json:"model"`
    Messages         []ChatCompletionMessage `json:"messages"`
    MaxTokens        *int                   `json:"max_tokens,omitempty"`
    Temperature      *float32               `json:"temperature,omitempty"`
    TopP             *float32               `json:"top_p,omitempty"`
    Stream           bool                   `json:"stream,omitempty"`
    Stop             []string               `json:"stop,omitempty"`
    // ... 更多字段
}
```

### 响应类型

```go
type ChatCompletionResponse struct {
    ID      string                 `json:"id"`
    Object  string                 `json:"object"`
    Created int64                  `json:"created"`
    Model   string                 `json:"model"`
    Choices []ChatCompletionChoice `json:"choices"`
    Usage   *Usage                 `json:"usage,omitempty"`
}
```

## 环境变量

设置相应的API密钥环境变量：

```bash
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export OPENAI_API_KEY="your-openai-api-key"
export ALICLOUD_API_KEY="your-dashscope-api-key"
```

## 向后兼容

原有的DeepSeek客户端代码无需修改即可继续使用：

```go
// 原有代码仍然有效
client, err := deepseek.NewClient("your-api-key")
resp, err := client.CallChatCompletionsChat(ctx, req)
```

## 扩展新服务商

要添加新的AI服务商支持：

1. 实现 `providers.Provider` 接口
2. 注册服务商创建函数
3. 添加到工厂方法中

```go
// 实现Provider接口
type CustomProvider struct {
    // ...
}

func (p *CustomProvider) CreateChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
    // 实现逻辑
}

// 注册服务商
func init() {
    providers.RegisterCustomProvider(func(config providers.ProviderConfig) (providers.Provider, error) {
        return NewCustomProvider(config)
    })
}
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！ 