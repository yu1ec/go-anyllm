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