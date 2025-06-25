# 流式响应迭代器模式使用指南

## 概述

新版本的 `StreamReader` 支持迭代器模式，您可以使用 `stream.Next()` 和 `stream.Current()` 方法来处理流式响应，这提供了更直观和灵活的流处理方式。

## 新增方法

- `Next() bool`: 移动到下一个响应，如果有下一个响应返回 `true`，否则返回 `false`
- `Current() *ChatCompletionsResponse`: 返回当前的响应，需要先调用 `Next()`
- `Error() error`: 返回最后一次操作的错误

## 使用示例

### 基本迭代器模式

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    deepseek "github.com/yu1ec/go-anyllm"
    "github.com/yu1ec/go-anyllm/types"
)

func main() {
    // 创建客户端
    client, err := deepseek.NewDeepSeekClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    // 创建流式请求
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

    // 发送流式请求
    stream, err := client.CreateChatCompletionStream(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

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
}
```

### 收集完整响应

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

### 处理工具调用

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

### 错误处理

```go
func processStreamWithErrorHandling(stream response.StreamReader) {
    for stream.Next() {
        chunk := stream.Current()
        
        // 处理响应
        if chunk != nil {
            // 您的处理逻辑
            processChunk(chunk)
        }
    }
    
    // 检查流处理过程中的错误
    if err := stream.Error(); err != nil {
        log.Printf("流处理错误: %v", err)
        return
    }
    
    fmt.Println("流处理成功完成")
}
```

## 向后兼容性

原有的 `Read()` 方法仍然可用，您可以继续使用之前的方式：

```go
// 旧的方式仍然有效
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
    processResponse(response)
}
```

## 注意事项

1. **不要混合使用**：不要在同一个流上同时使用 `Next()/Current()` 和 `Read()` 方法
2. **先调用Next()**：必须先调用 `Next()` 才能调用 `Current()`
3. **错误检查**：在迭代结束后记得检查 `Error()` 方法
4. **一次性使用**：每个流只能被消费一次，不能重复迭代

## 优势

- **更直观**：类似于其他语言中的迭代器模式
- **更安全**：自动处理EOF和错误状态
- **更灵活**：可以在循环中更容易地添加条件判断和错误处理
- **更清晰**：代码结构更清晰，易于理解和维护 