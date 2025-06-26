# Tools 功能使用指南

本项目现已完全支持工具调用（Function Calling）功能，包括工具定义、验证、调用和响应处理。

## 功能概述

### ✅ 已实现的功能

1. **工具定义和验证** - 完整的tools和tool_choice验证
2. **工具调用处理** - 支持所有提供商的工具调用转换
3. **辅助工具包** - 提供便捷的工具构建器和处理器
4. **预设工具模板** - 常用工具的开箱即用模板
5. **类型安全** - 完整的TypeScript风格类型支持

### 🚀 新增内容

- `request/validator.go` - 新增 `validateTools()` 和 `validateToolChoice()` 函数
- `tools/tools_helper.go` - 全新的工具辅助包
- `request/tools_validator_test.go` - 完整的单元测试
- `examples/tools_example.go` - 详细的使用示例

## 快速开始

### 1. 基本工具定义

```go
import "github.com/yu1ec/go-anyllm/tools"

// 使用工具构建器
weatherTool := tools.NewTool("get_weather", "获取天气信息").
    AddStringParam("location", "城市名称", true).
    AddStringParam("unit", "温度单位", false, "celsius", "fahrenheit").
    BuildForTypes()

// 使用预设模板
calculatorTool := tools.CalculatorTool()
```

### 2. 工具选择策略

```go
import "github.com/yu1ec/go-anyllm/tools"

// 不同的工具选择方式
req.ToolChoice = tools.Choice.Auto()           // 自动选择
req.ToolChoice = tools.Choice.None()           // 不使用工具  
req.ToolChoice = tools.Choice.Required()       // 必须使用工具
req.ToolChoice = tools.Choice.Function("get_weather") // 指定函数
```

### 3. 工具调用处理

```go
// 实现工具处理器
type WeatherHandler struct{}

func (h *WeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
    // 解析参数
    type WeatherParams struct {
        Location string `json:"location"`
        Unit     string `json:"unit"`
    }
    
    params, err := tools.ParseToolCallArguments[WeatherParams](toolCall)
    if err != nil {
        return "", err
    }
    
    // 执行实际逻辑
    return fmt.Sprintf("%s: 晴天, 25°C", params.Location), nil
}

// 注册和使用
registry := tools.NewFunctionRegistry()
registry.Register("get_weather", &WeatherHandler{})

// 处理工具调用
result := registry.Handle(toolCall)
toolMessage := result.ToToolMessage()
```

## 详细API文档

### 工具构建器 (ToolBuilder)

#### 创建工具
```go
builder := tools.NewTool(name, description)
```

#### 添加参数
```go
// 字符串参数
builder.AddStringParam(name, description, required, enum...)

// 数字参数  
builder.AddNumberParam(name, description, required)

// 整数参数
builder.AddIntegerParam(name, description, required)

// 布尔参数
builder.AddBooleanParam(name, description, required)

// 数组参数
builder.AddArrayParam(name, description, itemType, required)

// 对象参数
builder.AddObjectParam(name, description, properties, requiredFields, required)
```

#### 构建工具
```go
// 构建为request包的Tool
tool := builder.Build()

// 构建为types包的Tool  
tool := builder.BuildForTypes()
```

### 工具选择 (ToolChoice)

```go
// 全局实例
tools.Choice.Auto()                    // "auto"
tools.Choice.None()                    // "none"  
tools.Choice.Required()                // "required"
tools.Choice.Function("function_name") // 指定函数(map格式)
tools.Choice.FunctionStruct("name")    // 指定函数(struct格式)
```

### 函数注册表 (FunctionRegistry)

```go
// 创建注册表
registry := tools.NewFunctionRegistry()

// 注册处理器
registry.Register(functionName, handler)

// 处理工具调用
result := registry.Handle(toolCall)

// 转换为消息
message := result.ToToolMessage()
```

### 工具调用处理器接口

```go
type ToolCallHandler interface {
    HandleToolCall(toolCall types.ToolCall) (string, error)
}
```

### 参数解析

```go
// 类型安全的参数解析
params, err := tools.ParseToolCallArguments[YourParamsType](toolCall)
```

## 预设工具模板

### 可用的预设工具

```go
// 天气查询
weatherTool := tools.GetWeatherTool()

// 计算器
calculatorTool := tools.CalculatorTool()

// 搜索
searchTool := tools.SearchTool()

// 发送邮件
emailTool := tools.SendEmailTool()

// 文件操作
fileTool := tools.FileOperationTool()
```

## 完整使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    deepseek "github.com/yu1ec/go-anyllm"
    "github.com/yu1ec/go-anyllm/providers"
    "github.com/yu1ec/go-anyllm/tools"
    "github.com/yu1ec/go-anyllm/types"
)

func main() {
    // 1. 创建客户端
    config := &deepseek.ClientConfig{
        Provider: providers.ProviderDeepSeek,
        APIKey:   "your-api-key",
        Timeout:  120,
    }
    
    client, err := deepseek.NewUnifiedClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建工具
    weatherTool := tools.GetWeatherTool()
    
    // 3. 发送请求
    req := &types.ChatCompletionRequest{
        Model: "deepseek-chat",
        Messages: []types.ChatCompletionMessage{
            {
                Role:    types.RoleUser,
                Content: "北京天气怎么样？",
            },
        },
        Tools: []types.Tool{weatherTool},
        ToolChoice: tools.Choice.Auto(),
    }
    
    resp, err := client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 处理工具调用
    if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
        message := resp.Choices[0].Message
        if len(message.ToolCalls) > 0 {
            // 创建工具注册表
            registry := tools.NewFunctionRegistry()
            registry.Register("get_weather", &WeatherHandler{})
            
            // 处理每个工具调用
            for _, toolCall := range message.ToolCalls {
                result := registry.Handle(toolCall)
                fmt.Printf("工具调用结果: %s\n", result.Content)
            }
        }
    }
}

// 工具处理器实现
type WeatherHandler struct{}

func (h *WeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
    type WeatherParams struct {
        Location string `json:"location"`
        Unit     string `json:"unit"`
    }
    
    params, err := tools.ParseToolCallArguments[WeatherParams](toolCall)
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%s: 晴天, 25°C", params.Location), nil
}
```

## 验证功能

### 自动验证

当启用请求验证时（`DisableRequestValidation: false`），工具定义会自动验证：

- 工具类型必须为 "function"
- 函数名称不能为空，长度不超过64字符
- 函数名称只能包含字母、数字、下划线和连字符
- tool_choice只能在提供tools时设置
- tool_choice指定的函数必须存在于tools中

### 支持的tool_choice格式

```go
// 字符串格式
"auto" | "none" | "required"

// 对象格式
map[string]interface{}{
    "type": "function",
    "function": map[string]interface{}{
        "name": "function_name",
    },
}

// 结构体格式
request.ToolChoiceNamed{
    Type: "function",
    Function: request.ToolChoiceFunction{
        Name: "function_name",
    },
}
```

## 最佳实践

### 1. 工具设计
- 保持工具功能单一和明确
- 提供清晰的参数描述
- 使用合适的参数类型和约束

### 2. 错误处理
- 始终验证工具调用参数
- 为工具调用失败提供有用的错误信息
- 使用类型安全的参数解析

### 3. 性能优化
- 复用工具注册表实例
- 缓存经常使用的工具定义
- 合理设置请求超时

### 4. 安全考虑
- 验证工具调用的参数
- 限制危险操作的访问
- 记录工具调用的审计日志

## 调试技巧

### 1. 启用详细日志
```go
// 查看工具调用的详细信息
fmt.Printf("工具调用: %+v\n", toolCall)
fmt.Printf("参数: %v\n", toolCall.Function.Parameters)
```

### 2. 验证工具定义
```go
// 手动验证工具
err := request.ValidateChatCompletionsRequest(req)
if err != nil {
    log.Printf("验证失败: %v", err)
}
```

### 3. 测试工具处理器
```go
// 单独测试工具处理器
handler := &YourHandler{}
result, err := handler.HandleToolCall(mockToolCall)
```

## 支持的提供商

当前所有提供商都支持工具调用：

- ✅ DeepSeek
- ✅ OpenAI  
- ✅ 阿里云通义千问

每个提供商的工具调用会自动转换为统一的OpenAI格式，确保一致的使用体验。

## Stream + Enable Think 模式下的 Tools 处理

在 Stream + Enable Think 模式下，模型会先进行思考（reasoning），然后进行工具调用和内容生成。这种模式提供了更强的推理能力，但需要特殊的处理方式来正确解析工具调用。

### 特点说明

1. **思考内容先输出**：模型会先输出思考内容（reasoning_content）
2. **工具调用在流中分块传输**：工具调用信息会在多个流块中逐步构建
3. **需要累积解析**：需要累积所有相关的Delta信息来完整重构工具调用
4. **服务商差异**：不同服务商在流式工具调用的实现上可能有细微差异

### 重要说明

⚠️ **注意事项**：
- 在统一接口(`types`包)中，Delta对应`ChatCompletionMessage`，包含ToolCalls字段
- 在响应接口(`response`包)中，Delta结构体本身不包含ToolCalls，工具调用信息通过Message传递
- 阿里云的`enable_thinking`参数仅在阿里云通义千问模型中支持
- DeepSeek-R1模型具有内置推理能力，不需要额外的enable_thinking参数

### 基本使用方式

#### 1. 阿里云通义千问 + Enable Think

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"

    deepseek "github.com/yu1ec/go-anyllm"
    "github.com/yu1ec/go-anyllm/providers"
    "github.com/yu1ec/go-anyllm/tools"
    "github.com/yu1ec/go-anyllm/types"
)

func main() {
    // 1. 创建阿里云客户端
    config := &deepseek.ClientConfig{
        Provider: providers.ProviderAliCloud,
        APIKey:   "your-alicloud-api-key",
        Timeout:  120,
    }
    
    client, err := deepseek.NewUnifiedClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建工具
    weatherTool := tools.GetWeatherTool()
    
    // 3. 创建带思考模式的流式请求
    req := &types.ChatCompletionRequest{
        Model: "qwen-max",
        Messages: []types.ChatCompletionMessage{
            {
                Role:    types.RoleUser,
                Content: "北京明天天气如何？请仔细分析后回答。",
            },
        },
        Tools:      []types.Tool{weatherTool},
        ToolChoice: tools.Choice.Auto(),
        Stream:     true,
    }
    
    // 4. 启用思考模式和设置预算
    req.WithEnableThinking(true).WithThinkingBudget(2000)
    
    // 5. 发送流式请求
    stream, err := client.CreateChatCompletionStream(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()
    
    // 6. 处理流式响应和工具调用
    err = handleStreamWithThinkingAndTools(stream)
    if err != nil {
        log.Fatal(err)
    }
}

// handleStreamWithThinkingAndTools 处理带思考和工具调用的流式响应
func handleStreamWithThinkingAndTools(stream response.StreamReader) error {
    var (
        reasoningContent strings.Builder
        assistantContent strings.Builder
        toolCallsBuffer  = make(map[string]*types.ToolCall) // 用于累积工具调用
    )
    
    fmt.Println("=== AI 思考过程 ===")
    
    for stream.Next() {
        chunk := stream.Current()
        if chunk == nil || len(chunk.Choices) == 0 {
            continue
        }
        
        choice := chunk.Choices[0]
        if choice.Delta == nil {
            continue
        }
        
        // 1. 处理思考内容（reasoning_content）
        if choice.Delta.ReasoningContent != "" {
            fmt.Print(choice.Delta.ReasoningContent)
            reasoningContent.WriteString(choice.Delta.ReasoningContent)
        }
        
        // 2. 处理常规回复内容
        if choice.Delta.Content != "" {
            if reasoningContent.Len() > 0 {
                fmt.Println("\n\n=== AI 回复 ===")
                reasoningContent.Reset() // 清空，表示思考阶段结束
            }
            fmt.Print(choice.Delta.Content)
            assistantContent.WriteString(choice.Delta.Content)
        }
        
        // 3. 处理工具调用（增量式解析）
        // 注意：在流式响应中，工具调用信息既可以通过Delta.ToolCalls增量传递，
        // 也可以通过Message.ToolCalls传递完整信息
        
        // 3a. 处理Delta中的增量工具调用（实时流式）
        if choice.Delta != nil && len(choice.Delta.ToolCalls) > 0 {
            if reasoningContent.Len() > 0 {
                fmt.Println("\n\n=== 增量工具调用 ===")
                reasoningContent.Reset()
            }
            
            for _, deltaToolCall := range choice.Delta.ToolCalls {
                // 累积构建工具调用
                if deltaToolCall.Id != "" {
                    if toolCallsBuffer[deltaToolCall.Id] == nil {
                        toolCallsBuffer[deltaToolCall.Id] = &types.ToolCall{
                            ID:   deltaToolCall.Id,
                            Type: deltaToolCall.Type,
                            Function: types.ToolFunction{
                                Name:       deltaToolCall.Function.Name,
                                Parameters: deltaToolCall.Function.Arguments,
                            },
                        }
                    } else {
                        // 累积参数（如果参数是逐步传输的）
                        if deltaToolCall.Function.Arguments != "" {
                            toolCallsBuffer[deltaToolCall.Id].Function.Parameters = deltaToolCall.Function.Arguments
                        }
                    }
                    
                    fmt.Printf("增量工具调用: %s\n", toolCallsBuffer[deltaToolCall.Id].Function.Name)
                }
            }
        }
        
        // 3b. 处理Message中的完整工具调用（兼容性支持）
        if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
            if reasoningContent.Len() > 0 {
                fmt.Println("\n\n=== 工具调用 ===")
                reasoningContent.Reset()
            }
            
            for _, deltaToolCall := range choice.Message.ToolCalls {
                // 累积构建工具调用
                if deltaToolCall.Id != "" {
                    if toolCallsBuffer[deltaToolCall.Id] == nil {
                        toolCallsBuffer[deltaToolCall.Id] = &types.ToolCall{
                            ID:   deltaToolCall.Id,
                            Type: deltaToolCall.Type,
                            Function: types.ToolFunction{
                                Name:       deltaToolCall.Function.Name,
                                Parameters: deltaToolCall.Function.Arguments, // 注意：response包使用Arguments字段
                            },
                        }
                    } else {
                        // 累积参数（如果参数是逐步传输的）
                        if deltaToolCall.Function.Arguments != "" {
                            // 合并参数（这里简化处理，实际可能需要JSON合并）
                            toolCallsBuffer[deltaToolCall.Id].Function.Parameters = deltaToolCall.Function.Arguments
                        }
                    }
                    
                    fmt.Printf("工具调用: %s\n", toolCallsBuffer[deltaToolCall.Id].Function.Name)
                }
            }
        }
        
        // 4. 检查完成状态
        if choice.FinishReason != "" {
            fmt.Printf("\n\n完成原因: %s\n", choice.FinishReason)
            
            // 如果有工具调用，执行它们
            if len(toolCallsBuffer) > 0 {
                fmt.Println("\n=== 执行工具调用 ===")
                registry := tools.NewFunctionRegistry()
                registry.Register("get_weather", &WeatherHandler{})
                
                for _, toolCall := range toolCallsBuffer {
                    result := registry.Handle(*toolCall)
                    fmt.Printf("工具 %s 执行结果: %s\n", toolCall.Function.Name, result.Content)
                    if result.Error != "" {
                        fmt.Printf("工具执行错误: %s\n", result.Error)
                    }
                }
            }
            break
        }
    }
    
    return stream.Error()
}

// WeatherHandler 天气工具处理器示例
type WeatherHandler struct{}

func (h *WeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
    type WeatherParams struct {
        Location string `json:"location"`
        Unit     string `json:"unit"`
    }
    
    params, err := tools.ParseToolCallArguments[WeatherParams](toolCall)
    if err != nil {
        return "", err
    }
    
    // 模拟天气查询
    return fmt.Sprintf("%s: 晴天, 气温25°C, 湿度60%%", params.Location), nil
}
```

#### 2. DeepSeek-R1 推理模式

DeepSeek-R1模型具有内置的推理能力，在流式响应中会输出推理内容：

```go
func handleDeepSeekReasonerWithTools() {
    // 1. 创建DeepSeek客户端
    config := &deepseek.ClientConfig{
        Provider: providers.ProviderDeepSeek,
        APIKey:   "your-deepseek-api-key",
        Timeout:  120,
    }
    
    client, err := deepseek.NewUnifiedClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建工具
    calculatorTool := tools.CalculatorTool()
    
    // 3. 使用DeepSeek-R1模型
    req := &types.ChatCompletionRequest{
        Model: "deepseek-reasoner",
        Messages: []types.ChatCompletionMessage{
            {
                Role:    types.RoleUser,
                Content: "请计算 (15 + 27) * 3 - 8，并解释计算过程。",
            },
        },
        Tools:      []types.Tool{calculatorTool},
        ToolChoice: tools.Choice.Auto(),
        Stream:     true,
    }
    
    // 4. 发送请求并处理响应
    stream, err := client.CreateChatCompletionStream(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()
    
    handleDeepSeekReasoningStream(stream)
}

func handleDeepSeekReasoningStream(stream response.StreamReader) {
    var (
        reasoningContent strings.Builder
        finalContent     strings.Builder
        inReasoningMode  = true
    )
    
    fmt.Println("=== DeepSeek 推理过程 ===")
    
    for stream.Next() {
        chunk := stream.Current()
        if chunk == nil || len(chunk.Choices) == 0 {
            continue
        }
        
        choice := chunk.Choices[0]
        if choice.Delta == nil {
            continue
        }
        
        // 处理推理内容
        if choice.Delta.ReasoningContent != "" {
            if !inReasoningMode {
                fmt.Println("\n\n=== 继续推理 ===")
                inReasoningMode = true
            }
            fmt.Print(choice.Delta.ReasoningContent)
            reasoningContent.WriteString(choice.Delta.ReasoningContent)
        }
        
        // 处理最终回复
        if choice.Delta.Content != "" {
            if inReasoningMode {
                fmt.Println("\n\n=== 最终回复 ===")
                inReasoningMode = false
            }
            fmt.Print(choice.Delta.Content)
            finalContent.WriteString(choice.Delta.Content)
        }
        
        // 处理工具调用
        if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
            fmt.Println("\n\n=== 工具调用检测 ===")
            for _, toolCall := range choice.Message.ToolCalls {
                fmt.Printf("调用工具: %s\n", toolCall.Function.Name)
                // 在实际场景中，这里需要累积完整的工具调用信息
            }
        }
    }
    
    if err := stream.Error(); err != nil {
        log.Printf("流处理错误: %v", err)
    }
}
```

### 关键处理要点

#### 1. 工具调用的分块重建

在流式模式下，工具调用信息可能分多个chunks传输：

```go
type ToolCallAccumulator struct {
    toolCalls map[string]*types.ToolCall
    mutex     sync.RWMutex
}

func NewToolCallAccumulator() *ToolCallAccumulator {
    return &ToolCallAccumulator{
        toolCalls: make(map[string]*types.ToolCall),
    }
}

func (acc *ToolCallAccumulator) ProcessDelta(deltaToolCalls []types.ToolCall) {
    acc.mutex.Lock()
    defer acc.mutex.Unlock()
    
    for _, delta := range deltaToolCalls {
        if delta.ID == "" {
            continue
        }
        
        if acc.toolCalls[delta.ID] == nil {
            // 创建新的工具调用
            acc.toolCalls[delta.ID] = &types.ToolCall{
                ID:   delta.ID,
                Type: delta.Type,
                Function: types.ToolFunction{
                    Name:        delta.Function.Name,
                    Parameters:  delta.Function.Parameters,
                },
            }
        } else {
            // 累积更新已存在的工具调用
            if delta.Function.Name != "" {
                acc.toolCalls[delta.ID].Function.Name = delta.Function.Name
            }
            if delta.Function.Parameters != nil {
                acc.toolCalls[delta.ID].Function.Parameters = delta.Function.Parameters
            }
        }
    }
}

func (acc *ToolCallAccumulator) GetCompletedToolCalls() []types.ToolCall {
    acc.mutex.RLock()
    defer acc.mutex.RUnlock()
    
    var result []types.ToolCall
    for _, toolCall := range acc.toolCalls {
        result = append(result, *toolCall)
    }
    return result
}
```

#### 2. 思考内容的处理

```go
func separateThinkingFromResponse(stream response.StreamReader) (reasoning, content string, toolCalls []types.ToolCall, err error) {
    var (
        reasoningBuilder strings.Builder
        contentBuilder   strings.Builder
        accumulator      = NewToolCallAccumulator()
        thinkingPhase    = true
    )
    
    for stream.Next() {
        chunk := stream.Current()
        if chunk == nil || len(chunk.Choices) == 0 {
            continue
        }
        
        choice := chunk.Choices[0]
        if choice.Delta == nil {
            continue
        }
        
        // 思考阶段
        if thinkingPhase && choice.Delta.ReasoningContent != "" {
            reasoningBuilder.WriteString(choice.Delta.ReasoningContent)
        }
        
        // 回复阶段
        if choice.Delta.Content != "" {
            thinkingPhase = false // 开始回复阶段
            contentBuilder.WriteString(choice.Delta.Content)
        }
        
        // 工具调用
        if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
            thinkingPhase = false
            // 将response.ToolCall转换为types.ToolCall
            var deltaToolCalls []types.ToolCall
            for _, respToolCall := range choice.Message.ToolCalls {
                deltaToolCalls = append(deltaToolCalls, types.ToolCall{
                    ID:   respToolCall.Id,
                    Type: respToolCall.Type,
                    Function: types.ToolFunction{
                        Name:       respToolCall.Function.Name,
                        Parameters: respToolCall.Function.Arguments,
                    },
                })
            }
            accumulator.ProcessDelta(deltaToolCalls)
        }
        
        // 完成检查
        if choice.FinishReason != "" {
            break
        }
    }
    
    return reasoningBuilder.String(), 
           contentBuilder.String(), 
           accumulator.GetCompletedToolCalls(), 
           stream.Error()
}
```

#### 3. 多轮对话处理

当工具调用完成后，需要将结果返回给模型继续对话：

```go
func handleMultiTurnWithThinking(client deepseek.UnifiedClient, initialReq *types.ChatCompletionRequest) error {
    messages := make([]types.ChatCompletionMessage, len(initialReq.Messages))
    copy(messages, initialReq.Messages)
    
    for {
        // 创建当前轮次的请求
        currentReq := &types.ChatCompletionRequest{
            Model:          initialReq.Model,
            Messages:       messages,
            Tools:          initialReq.Tools,
            ToolChoice:     initialReq.ToolChoice,
            Stream:         true,
            EnableThinking: initialReq.EnableThinking,
            ThinkingBudget: initialReq.ThinkingBudget,
        }
        
        // 发送流式请求
        stream, err := client.CreateChatCompletionStream(context.Background(), currentReq)
        if err != nil {
            return err
        }
        
        // 处理响应
        reasoning, content, toolCalls, err := separateThinkingFromResponse(stream)
        stream.Close()
        
        if err != nil {
            return err
        }
        
        // 添加助手的消息
        assistantMsg := types.ChatCompletionMessage{
            Role:             types.RoleAssistant,
            Content:          content,
            ReasoningContent: reasoning,
            ToolCalls:        toolCalls,
        }
        messages = append(messages, assistantMsg)
        
        // 如果没有工具调用，结束对话
        if len(toolCalls) == 0 {
            fmt.Printf("最终回复: %s\n", content)
            break
        }
        
        // 执行工具调用
        registry := tools.NewFunctionRegistry()
        registry.Register("get_weather", &WeatherHandler{})
        registry.Register("calculator", &CalculatorHandler{})
        
        for _, toolCall := range toolCalls {
            result := registry.Handle(toolCall)
            
            // 添加工具调用结果消息
            toolMessage := types.ChatCompletionMessage{
                Role:       types.RoleTool,
                Content:    result.Content,
                ToolCallID: toolCall.ID,
            }
            messages = append(messages, toolMessage)
            
            fmt.Printf("工具 %s 执行结果: %s\n", toolCall.Function.Name, result.Content)
        }
        
        fmt.Println("继续对话...")
    }
    
    return nil
}
```

### 最佳实践

#### 1. 流式处理优化
- 使用goroutines并行处理思考内容和工具调用
- 实现工具调用的增量解析和验证
- 合理控制输出缓冲区大小

#### 2. 错误恢复
- 实现工具调用失败的重试机制
- 对不完整的工具调用进行检测和处理
- 提供思考内容的截断和恢复

#### 3. 用户体验
- 实时显示思考过程，增强透明度
- 为工具调用提供进度指示
- 支持中断长时间的思考过程

#### 4. 性能考虑
- 合理设置thinking_budget避免过度消耗
- 实现工具调用的批量处理
- 优化Delta信息的内存使用

这种模式特别适合需要复杂推理和多步骤工具调用的场景，如数据分析、问题解决、多步骤计算等。

### 常见问题和解决方案

#### 1. 工具调用信息不完整

**问题**：流式传输中工具调用的参数可能分多个chunk传输，导致解析不完整。

**解决方案**：
```go
// 实现更智能的参数累积
func (acc *ToolCallAccumulator) smartMergeParameters(toolCallID string, newParams interface{}) {
    acc.mutex.Lock()
    defer acc.mutex.Unlock()
    
    if acc.toolCalls[toolCallID] == nil {
        return
    }
    
    // 如果是字符串参数，需要JSON解析和合并
    if strParams, ok := newParams.(string); ok {
        if acc.toolCalls[toolCallID].Function.Parameters == nil {
            acc.toolCalls[toolCallID].Function.Parameters = strParams
        } else {
            // 合并JSON字符串参数
            if existing, ok := acc.toolCalls[toolCallID].Function.Parameters.(string); ok {
                merged := mergeJSONStrings(existing, strParams)
                acc.toolCalls[toolCallID].Function.Parameters = merged
            }
        }
    }
}

func mergeJSONStrings(existing, new string) string {
    // 实现JSON字符串的智能合并逻辑
    // 这里简化处理，实际应用中需要更复杂的JSON合并
    if existing == "" {
        return new
    }
    if new == "" {
        return existing
    }
    // 可以使用第三方JSON merge库或自定义实现
    return new // 简化处理
}
```

#### 2. 思考过程过长导致超时

**问题**：某些复杂问题的思考过程可能很长，导致请求超时。

**解决方案**：
```go
// 设置合理的思考预算和超时时间
req := &types.ChatCompletionRequest{
    Model:          "qwen-max",
    Messages:       messages,
    Tools:          tools,
    Stream:         true,
    EnableThinking: types.ToPtr(true),
    ThinkingBudget: types.ToPtr(1500), // 限制思考token数量
}

// 设置合理的上下文超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

stream, err := client.CreateChatCompletionStream(ctx, req)
```

#### 3. 多轮对话中的状态管理

**问题**：在多轮对话中，需要正确管理思考内容和工具调用的状态。

**解决方案**：
```go
type ConversationState struct {
    Messages        []types.ChatCompletionMessage
    ThinkingHistory []string
    ToolCallHistory []types.ToolCall
    mutex           sync.RWMutex
}

func (cs *ConversationState) AddThinking(reasoning string) {
    cs.mutex.Lock()
    defer cs.mutex.Unlock()
    cs.ThinkingHistory = append(cs.ThinkingHistory, reasoning)
}

func (cs *ConversationState) AddToolCall(toolCall types.ToolCall) {
    cs.mutex.Lock()
    defer cs.mutex.Unlock()
    cs.ToolCallHistory = append(cs.ToolCallHistory, toolCall)
}
```

### 性能优化建议

1. **并发处理**：使用goroutines并行处理思考内容显示和工具调用准备
2. **内存管理**：及时清理累积的Delta数据，避免内存泄漏
3. **错误重试**：实现智能重试机制，处理网络波动等临时问题
4. **缓存策略**：对相同的工具调用结果进行合理缓存

### 调试技巧

```go
// 启用详细的流式调试日志
func debugStreamProcessing(stream response.StreamReader) {
    chunkCount := 0
    for stream.Next() {
        chunkCount++
        chunk := stream.Current()
        
        log.Printf("Chunk %d: %+v", chunkCount, chunk)
        
        if chunk != nil && len(chunk.Choices) > 0 {
            choice := chunk.Choices[0]
            if choice.Delta != nil {
                log.Printf("  Delta Content: %q", choice.Delta.Content)
                log.Printf("  Delta Reasoning: %q", choice.Delta.ReasoningContent)
            }
            if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
                log.Printf("  Tool Calls: %+v", choice.Message.ToolCalls)
            }
        }
    }
}
```

这种模式特别适合需要复杂推理和多步骤工具调用的场景，如数据分析、问题解决、多步骤计算等。 