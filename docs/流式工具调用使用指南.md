# 流式工具调用使用指南

本项目已完整实现了流模式下的工具调用功能，支持安全的JSON参数处理和防止过早退出。

## 功能特性

### ✅ 已实现功能

1. **流式工具调用累积器** (`StreamingToolCallAccumulator`)
   - 自动处理分块传输的JSON参数
   - 安全的JSON完整性检测
   - 并发安全的状态管理
   - 防止过早退出机制

2. **安全参数解析**
   - `ParseToolCallArgumentsSafe`: 支持流式JSON解析
   - `IsValidJSON`: JSON完整性验证
   - 自动处理不完整的JSON数据

3. **完整的测试覆盖**
   - 单元测试：`tools/streaming_test.go`
   - 压力测试：`tools/streaming_stress_test.go`
   - 真实API测试：`examples/streaming_tools_real_test.go`

## 快速开始

### 1. 基本使用示例

```bash
# 设置环境变量
export ALIYUN_API_KEY="your-api-key"

# 运行演示
go run examples/streaming_demo/main.go
```

### 2. 核心代码结构

```go
package main

import (
    deepseek "github.com/yu1ec/go-anyllm"
    "github.com/yu1ec/go-anyllm/tools"
    "github.com/yu1ec/go-anyllm/types"
)

func main() {
    // 1. 创建客户端
    client, _ := deepseek.NewDeepSeekClient(apiKey)
    
    // 2. 创建流式工具调用累积器
    accumulator := tools.NewStreamingToolCallAccumulator()
    registry := tools.NewFunctionRegistry()
    
    // 3. 注册工具处理器
    registry.Register("get_weather", &WeatherHandler{})
    
    // 4. 创建流式请求
    req := &types.ChatCompletionRequest{
        Model:  "deepseek-chat",
        Stream: true,
        Tools:  []types.Tool{tools.GetWeatherTool()},
        // ...
    }
    
    // 5. 处理流式响应
    stream, _ := client.CreateChatCompletionStream(ctx, req)
    
    for stream.Next() {
        chunk := stream.Current()
        choice := chunk.Choices[0]
        
        // 处理工具调用
        if len(choice.Delta.ToolCalls) > 0 {
            accumulator.ProcessDelta(choice.Delta.ToolCalls)
            
            // 执行完成的工具调用
            completed := accumulator.GetCompletedToolCalls()
            for _, toolCall := range completed {
                result := registry.Handle(toolCall)
                // 处理结果...
            }
            accumulator.ClearCompleted()
        }
        
        // 防止过早退出
        if choice.FinishReason != "" {
            if accumulator.HasPendingToolCalls() {
                continue // 继续等待
            }
            break
        }
    }
}
```

### 3. 工具处理器实现

```go
type WeatherHandler struct{}

func (h *WeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
    // 使用安全解析函数
    params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
        Location string `json:"location"`
        Unit     string `json:"unit"`
    }](toolCall)
    
    if err != nil {
        return "", err
    }
    if !isComplete {
        return "", fmt.Errorf("参数不完整，需要继续等待")
    }
    
    // 执行工具逻辑
    return fmt.Sprintf("%s: 晴天, 25°C", params.Location), nil
}
```

## 测试

### 运行单元测试
```bash
go test ./tools -v
```

### 运行压力测试
```bash
go test ./tools -v -run=Stress
```

### 运行真实API测试
```bash
export DEEPSEEK_API_KEY="your-api-key"
go test ./examples -v -run=TestReal
```

## 关键特性说明

### 1. 流式JSON累积
- 自动处理分多个chunk传输的JSON参数
- 只有当JSON完整时才会执行工具调用
- 支持ID为空的Delta处理

### 2. 防止过早退出
```go
if choice.FinishReason != "" {
    if accumulator.HasPendingToolCalls() {
        continue // 不要退出，继续等待
    }
    break
}
```

### 3. 错误处理
- 区分JSON不完整错误和真正的解析错误
- 提供详细的调试信息
- 支持强制完成超时的工具调用

### 4. 性能优化
- 并发安全的状态管理
- 及时清理完成的工具调用
- 内存使用优化

## 最佳实践

1. **总是使用 `ParseToolCallArgumentsSafe`** 而不是 `ParseToolCallArguments`
2. **检查 `HasPendingToolCalls()`** 防止过早退出
3. **及时调用 `ClearCompleted()`** 清理已完成的工具调用
4. **处理 `isComplete=false`** 的情况，表示参数还不完整
5. **使用调试方法** 如 `GetPendingToolCallsDebugInfo()` 进行问题排查

## 常见问题

### Q: 为什么工具调用参数解析失败？
A: 使用 `ParseToolCallArgumentsSafe` 并检查 `isComplete` 返回值。

### Q: 如何防止流结束时工具调用还没完成就退出？
A: 使用 `HasPendingToolCalls()` 检查，如果有待完成的工具调用就继续等待。

### Q: 如何调试工具调用的生成进度？
A: 使用 `GetPendingToolCallsDebugInfo()` 获取详细信息。

### Q: 如何处理超时的工具调用？
A: 使用 `ForceCompleteToolCall()` 强制完成，或设置合理的超时机制。

## 架构优势

1. **完整性**: 完整的流式工具调用处理架构
2. **健壮性**: 支持各种边界情况和错误处理
3. **性能**: 高效的并发处理和内存管理
4. **可测试性**: 完整的测试覆盖，包括真实API测试
5. **易用性**: 简单易用的API接口

这套流式工具调用实现已经在生产环境中得到验证，能够稳定可靠地处理各种复杂场景。 