# 多模态（视觉）功能使用指南

## 概述

go-anyllm 现已支持多模态功能，可以处理包含文本和图像的消息。目前支持阿里云（DashScope）的视觉模型，如 `qwen-vl-max-latest`。

## 功能特性

✅ **向后兼容**: 现有的纯文本API继续正常工作  
✅ **多模态支持**: 支持文本 + 图像的混合消息  
✅ **流式响应**: 多模态消息也支持流式处理  
✅ **灵活的API**: 提供多种创建多模态消息的方式  
✅ **类型安全**: 完整的类型定义和验证  

## 支持的模型

| 提供商 | 模型 | 支持的格式 |
|--------|------|------------|
| 阿里云 | `qwen-vl-max-latest` | PNG, JPEG, WEBP |
| 阿里云 | `qwen-vl-plus-latest` | PNG, JPEG, WEBP |

## 基本使用

### 1. 环境配置

```bash
export DASHSCOPE_API_KEY="your-api-key"
```

### 2. 创建客户端

```go
import (
    deepseek "github.com/yu1ec/go-anyllm"
    "github.com/yu1ec/go-anyllm/providers"
    "github.com/yu1ec/go-anyllm/types"
)

client, err := deepseek.NewUnifiedClient(&deepseek.ClientConfig{
    Provider: providers.ProviderAliCloud,
    APIKey:   os.Getenv("DASHSCOPE_API_KEY"),
})
```

### 3. 纯文本消息（向后兼容）

```go
req := &types.ChatCompletionRequest{
    Model: "qwen-vl-max-latest",
    Messages: []types.ChatCompletionMessage{
        types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
        types.NewTextMessage(types.RoleUser, "请介绍一下Go语言的特点。"),
    },
}

resp, err := client.CreateChatCompletion(context.Background(), req)
```

### 4. 多模态消息

#### 方式一：直接构造

```go
req := &types.ChatCompletionRequest{
    Model: "qwen-vl-max-latest",
    Messages: []types.ChatCompletionMessage{
        types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
        {
            Role: types.RoleUser,
            Content: []types.MessageContent{
                {
                    Type: types.MessageContentTypeImageURL,
                    ImageURL: &types.ImageURL{
                        URL: "data:image/png;base64," + base64Image,
                    },
                },
                {
                    Type: types.MessageContentTypeText,
                    Text: "图中描绘的是什么景象?",
                },
            },
        },
    },
}
```

#### 方式二：使用辅助函数

```go
contents := []types.MessageContent{
    types.NewImageContent("data:image/png;base64," + base64Image, types.ImageDetailAuto),
    types.NewTextContent("请分析这张图片的内容。"),
}

req := &types.ChatCompletionRequest{
    Model: "qwen-vl-max-latest",
    Messages: []types.ChatCompletionMessage{
        types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
        types.NewMultiModalMessage(types.RoleUser, contents),
    },
}
```

## 图像处理

### 支持的图像格式

- **PNG**: `data:image/png;base64,{base64_data}`
- **JPEG**: `data:image/jpeg;base64,{base64_data}`
- **WEBP**: `data:image/webp;base64,{base64_data}`

### 图像详细度设置

```go
// 低精度 - 更快，消耗更少token
types.NewImageContent(imageURL, types.ImageDetailLow)

// 高精度 - 更详细，消耗更多token
types.NewImageContent(imageURL, types.ImageDetailHigh)

// 自动选择 - 根据图像大小自动选择
types.NewImageContent(imageURL, types.ImageDetailAuto)
```

### 从文件加载图像

```go
func imageFileToBase64(imagePath string) (string, error) {
    imageFile, err := os.Open(imagePath)
    if err != nil {
        return "", err
    }
    defer imageFile.Close()

    imageData, err := io.ReadAll(imageFile)
    if err != nil {
        return "", err
    }

    return base64.StdEncoding.EncodeToString(imageData), nil
}

// 使用
base64Image, err := imageFileToBase64("path/to/your/image.png")
if err != nil {
    log.Fatal(err)
}

content := types.NewImageContent(
    fmt.Sprintf("data:image/png;base64,%s", base64Image),
    types.ImageDetailAuto,
)
```

## 流式处理

多模态消息同样支持流式处理：

```go
stream, err := client.CreateChatCompletionStream(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

for stream.Next() {
    chunk := stream.Current()
    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
        content := chunk.Choices[0].Delta.Content
        if content != "" {
            fmt.Print(content)
        }
    }
}

if err := stream.Error(); err != nil {
    log.Printf("流式响应错误: %v", err)
}
```

## API 参考

### 类型定义

```go
// 消息内容项
type MessageContent struct {
    Type     string    `json:"type"`                // "text" 或 "image_url"
    Text     string    `json:"text,omitempty"`      // 文本内容
    ImageURL *ImageURL `json:"image_url,omitempty"` // 图像URL
}

// 图像URL结构
type ImageURL struct {
    URL    string `json:"url"`              // 图像URL或base64数据
    Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}

// 聊天消息
type ChatCompletionMessage struct {
    Role    string      `json:"role"`
    Content interface{} `json:"content,omitempty"` // 支持string或[]MessageContent
    // ... 其他字段
}
```

### 辅助函数

```go
// 创建文本消息
func NewTextMessage(role, content string) ChatCompletionMessage

// 创建多模态消息
func NewMultiModalMessage(role string, contents []MessageContent) ChatCompletionMessage

// 创建文本内容
func NewTextContent(text string) MessageContent

// 创建图像内容
func NewImageContent(imageURL string, detail ...string) MessageContent

// 获取消息内容的字符串表示
func (m *ChatCompletionMessage) GetContentAsString() string

// 检查消息是否包含多模态内容
func (m *ChatCompletionMessage) IsMultiModal() bool

// 获取消息中的所有图像内容
func (m *ChatCompletionMessage) GetImageContents() []MessageContent
```

### 常量

```go
// 内容类型
const (
    MessageContentTypeText     = "text"
    MessageContentTypeImageURL = "image_url"
)

// 图像详细度
const (
    ImageDetailLow  = "low"
    ImageDetailHigh = "high"
    ImageDetailAuto = "auto"
)
```

## 完整示例

查看 `examples/multimodal_demo/main.go` 获取完整的示例代码。

运行示例：

```bash
# 设置API密钥
export DASHSCOPE_API_KEY="your-api-key"

# 运行示例
go run examples/multimodal_demo/main.go
```

## 注意事项

1. **API密钥**: 确保设置了正确的 `DASHSCOPE_API_KEY` 环境变量
2. **模型支持**: 只有视觉模型（如 `qwen-vl-max-latest`）支持多模态
3. **文件大小**: 建议图像文件大小不超过20MB
4. **格式支持**: 确保图像格式与Content-Type匹配
5. **向后兼容**: 现有的纯文本API继续工作，无需修改

## 故障排查

### 常见错误

1. **模型不支持**: 确保使用支持视觉的模型
2. **API密钥错误**: 检查 `DASHSCOPE_API_KEY` 是否正确设置
3. **图像格式错误**: 确保base64编码和Content-Type匹配
4. **文件太大**: 压缩图像或使用较低的详细度设置

### 调试技巧

```go
// 检查消息是否包含图像
if msg.IsMultiModal() {
    images := msg.GetImageContents()
    fmt.Printf("发现 %d 张图片\n", len(images))
}

// 获取文本内容（忽略图像）
textContent := msg.GetContentAsString()
fmt.Printf("文本内容: %s\n", textContent)
```

这样就完成了阿里云多模态功能的实现！🎉 