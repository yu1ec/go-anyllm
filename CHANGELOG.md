# 历史版本说明
## v0.1.2 (2025-07-04)
### 🔧 重要修复 & 优化

#### 🎯 类型系统优化
- 🔧 **工具函数类型分离**: 将 `ToolFunction` 分离为 `RequestToolFunction` 和 `ResponseToolFunction`
- 📝 **语义明确化**: 请求中使用 `Parameters` 字段，响应中使用 `Arguments` 字段
- 🛡️ **类型安全**: 增强了工具调用的类型安全性和可读性
- ✨ **向后兼容**: 保持所有现有API的完全兼容性

#### 🖼️ 多模态（视觉）功能支持
- 🌟 **阿里云视觉模型**: 支持 `qwen-vl-max-latest` 和 `qwen-vl-plus-latest` 模型
- 📷 **图像格式支持**: 支持 PNG、JPEG、WEBP 格式的图像处理
- 🔄 **流式多模态**: 多模态消息同样支持流式处理
- 📱 **灵活API设计**: 提供多种创建多模态消息的方式
- 🎨 **图像详细度控制**: 支持 low、high、auto 三种图像处理精度

### ✨ 新增功能

#### 🖼️ 多模态消息支持
- `NewTextMessage()` - 创建纯文本消息（向后兼容）
- `NewMultiModalMessage()` - 创建包含文本和图像的混合消息
- `NewTextContent()` - 创建文本内容项
- `NewImageContent()` - 创建图像内容项，支持详细度设置
- `GetContentAsString()` - 获取消息的文本表示
- `IsMultiModal()` - 检查消息是否包含多模态内容
- `GetImageContents()` - 提取消息中的所有图像内容

#### 🎯 图像处理增强
- **Base64编码支持**: `data:image/png;base64,{data}` 格式
- **详细度控制**: `ImageDetailLow`、`ImageDetailHigh`、`ImageDetailAuto` 常量
- **类型安全**: 完整的多模态类型定义和验证
- **示例代码**: 完整的多模态使用示例和文档

#### 🔧 工具系统优化
- **类型分离**: `RequestToolFunction` 用于请求，`ResponseToolFunction` 用于响应
- **字段语义**: 明确区分 `Parameters`（请求）和 `Arguments`（响应）
- **构建器增强**: `BuildForTypes()` 方法支持新的类型系统

### 🛠️ 技术改进

#### 📊 示例代码修复
- 修复了工具调用示例中的类型错误
- 更新了流式工具调用示例的字段引用
- 统一了所有示例代码的类型使用
- 增加了多模态功能的完整示例

#### 🔒 类型安全增强
- 更清晰的工具函数类型定义
- 减少了类型转换时的歧义
- 增强了编译时的类型检查
- 提供了更好的IDE支持和自动补全

#### 📈 向后兼容性
- 所有现有API保持完全兼容
- 渐进式类型系统升级
- 平滑的迁移路径
- 无破坏性变更

### 🚀 使用方式

#### 多模态消息创建
```go
// 方式一：直接构造
contents := []types.MessageContent{
    types.NewImageContent("data:image/png;base64," + base64Image, types.ImageDetailAuto),
    types.NewTextContent("请分析这张图片的内容。"),
}

// 方式二：使用辅助函数
req := &types.ChatCompletionRequest{
    Model: "qwen-vl-max-latest",
    Messages: []types.ChatCompletionMessage{
        types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
        types.NewMultiModalMessage(types.RoleUser, contents),
    },
}
```

#### 图像文件处理
```go
// 从文件加载并转换为base64
base64Image, err := imageFileToBase64("path/to/image.png")
content := types.NewImageContent(
    fmt.Sprintf("data:image/png;base64,%s", base64Image),
    types.ImageDetailAuto,
)
```

#### 类型安全的工具构建
```go
// 使用新的类型系统
tool := tools.NewTool("analyze_image", "分析图像内容").
    AddStringParam("description", "图像描述", true).
    BuildForTypes()  // 返回 types.Tool 类型
```

### 🐛 Bug修复

- 修复了工具调用示例中 `Parameters` 和 `Arguments` 字段的混用问题
- 修复了流式工具调用测试中的类型不匹配错误
- 统一了工具函数类型的使用规范
- 修复了示例代码中的编译错误

### 📄 新增文档

- [`docs/多模态（视觉）功能使用指南.md`](docs/多模态（视觉）功能使用指南.md) - 多模态功能完整使用指南
- [`docs/流式工具调用使用指南.md`](docs/流式工具调用使用指南.md) - 流式工具调用使用指南
- [`examples/multimodal_demo/main.go`](examples/multimodal_demo/main.go) - 多模态功能示例代码
- 完善的API文档和类型定义说明

### 🔗 兼容性

- v0.1.1 和 v0.1.0 对于方法参数调用上不兼容
- 现有的纯文本API继续正常工作
- 工具调用API保持完全兼容
- 渐进式类型系统升级，无破坏性变更
- 所有现有代码无需修改即可使用


## v0.1.1 (2025-07-03)

### 🔧 重要修复 & 优化

#### 🚨 阿里云思考模式修复
- 🔄 **思考模式兼容性**: 修复了非流式调用下思考模式的参数冲突问题
- 📊 **智能处理**: 当开启思考模式时，自动使用流式调用然后聚合结果
- 🛡️ **错误处理**: 改进了HTTP错误响应的处理，现在显示详细的错误信息
- ⚙️ **向后兼容**: 保持所有现有功能完全兼容

#### 🔧 流式工具调用核心修复
- 📝 **Delta结构修复**: 修复了`response.Delta`结构体缺少`ToolCalls`字段的编译错误
- 🔄 **流式JSON解析**: 解决了流式响应中JSON参数分块传输导致的解析失败问题
- 🛡️ **安全解析**: 新增`ParseToolCallArgumentsSafe`函数，支持不完整JSON的安全处理
- 🚫 **防过早退出**: 添加`HasPendingToolCalls()`方法防止流在工具调用完成前提前结束

#### ⚡ 流式工具调用累积器
- 📊 **状态管理**: 新增`StreamingToolCallAccumulator`用于管理流式工具调用状态
- 🔄 **自动累积**: 自动处理分多个chunk传输的JSON参数
- 📈 **进度监控**: 支持实时监控待完成工具调用的数量和状态
- 🧹 **内存优化**: 及时清理已完成的工具调用，优化内存使用

#### 🕐 阿里云思考模式超时优化
- 🎯 **分阶段超时**: 实现思考阶段、输出阶段和读取阶段的独立超时控制
- 🔄 **智能恢复**: 超时时尝试返回部分结果而不是完全失败
- 📊 **性能提升**: 使用8KB缓冲区和非阻塞读取提高处理效率
- ⚙️ **灵活配置**: 支持自定义`ThinkingTimeout`、`OutputTimeout`和`ReadTimeout`

### ✨ 新增功能

#### 🔧 增强的工具调用API
- `ParseToolCallArgumentsSafe[T]()` - 安全的参数解析，支持流式JSON
- `IsValidJSON()` - JSON有效性检测函数
- `HasPendingToolCalls()` - 检查是否有待完成的工具调用
- `GetPendingCount()` - 获取待完成工具调用数量
- `GetCompletedCount()` - 获取已完成工具调用数量
- `FinalizeStream()` - 流结束时的强制完成检查

#### 🎯 调试与监控
- `GetPendingToolCallsDebugInfo()` - 详细的调试信息输出
- `ForceCompleteToolCall()` - 强制完成指定工具调用
- 详细的状态追踪和时间戳记录
- 完善的错误信息和建议

### 🛠️ 技术改进

#### 🔒 并发安全
- 所有累积器操作都使用读写锁保护
- 线程安全的状态管理和数据处理
- 防止竞态条件的设计

#### 📊 性能优化
- 智能的JSON检测避免不必要的解析尝试
- 8KB缓冲区提高网络读取效率
- 及时清理完成的工具调用释放内存

#### 🧪 测试覆盖
- 新增流式工具调用集成测试
- 多工具调用并发处理测试
- 思考模式超时处理测试
- JSON解析安全性测试

### 🚀 使用方式

#### 安全的工具调用参数解析
```go
// 替换原有的ParseToolCallArguments
params, isComplete, err := tools.ParseToolCallArgumentsSafe[MyParams](toolCall)
if !isComplete {
    return "", fmt.Errorf("参数不完整，继续等待")
}
```

#### 流式工具调用处理
```go
accumulator := tools.NewStreamingToolCallAccumulator()
// 处理Delta
accumulator.ProcessDelta(choice.Delta.ToolCalls)
// 防止过早退出
if choice.FinishReason != "" && accumulator.HasPendingToolCalls() {
    continue // 继续等待工具调用完成
}
```

#### 自定义阿里云思考模式超时
```go
config := &alicloud.AliCloudConfig{
    APIKey:          "your-api-key",
    ThinkingTimeout: 300, // 思考5分钟
    OutputTimeout:   60,  // 输出等待1分钟
    ReadTimeout:     30,  // 读取30秒
}
```

### 🐛 Bug修复

- 修复了阿里云思考模式参数冲突导致的调用失败
- 修复了流式工具调用中JSON参数解析失败的问题
- 修复了Delta结构体缺少ToolCalls字段的编译错误
- 修复了流式响应过早退出导致工具调用丢失的问题
- 修复了思考模式下的超时处理不当问题

### 📄 新增文档

- [`docs/thinking_mode_fix.md`](docs/thinking_mode_fix.md) - 阿里云思考模式修复说明
- [`docs/流式JSON解析修复总结.md`](docs/流式JSON解析修复总结.md) - 流式JSON解析修复总结
- [`docs/流式工具调用修复说明.md`](docs/流式工具调用修复说明.md) - 流式工具调用修复说明
- [`docs/timeout_optimization.md`](docs/timeout_optimization.md) - 阿里云思考模式超时优化指南
- [`docs/流式工具调用最佳实践.md`](docs/流式工具调用最佳实践.md) - 流式工具调用最佳实践
- [`docs/防止过早退出示例.md`](docs/防止过早退出示例.md) - 防止过早退出示例

### 🔗 兼容性

- ✅ 完全向后兼容 v0.1.0 的所有功能
- ✅ 保留原有的`ParseToolCallArguments`函数
- ✅ 所有现有代码无需修改即可使用
- ✅ 新功能为可选升级，不影响现有流程

### 💡 升级建议

1. **立即升级**: 流式工具调用用户建议立即升级以获得稳定性修复
2. **安全解析**: 在流式工具调用中使用`ParseToolCallArgumentsSafe`替代原函数
3. **防过早退出**: 在检查`FinishReason`时添加`HasPendingToolCalls()`检查
4. **超时配置**: 根据实际场景调整思考模式的超时参数

---

*此版本主要关注稳定性和可靠性改进，特别是流式工具调用的完整性和错误处理能力的提升。*



## v0.1.0 (2025-07-01)

### 🎉 首次发布

这是go-anyllm统一AI客户端的首个正式版本，提供了完整的多服务商AI调用解决方案。

### ✨ 核心功能

#### 多服务商支持
- 🌐 **DeepSeek**: 支持deepseek-chat、deepseek-reasoner模型
- 🤖 **OpenAI**: 支持gpt-3.5-turbo、gpt-4、gpt-4-turbo等模型  
- ☁️ **阿里云通义千问**: 支持qwen-turbo、qwen-plus、qwen-max模型
- 🔄 **统一接口**: 相同代码可无缝切换不同AI服务商

#### 对话能力
- 💬 **标准对话**: 完整的ChatCompletion API支持
- 🔄 **OpenAI兼容**: 使用标准OpenAI API格式，易于迁移
- ⚙️ **灵活配置**: 支持自定义API端点、请求头、超时时间等

#### 流式响应
- 🚀 **实时流式聊天**: 支持Server-Sent Events流式响应
- 🔄 **迭代器模式**: 提供Next()/Current()/Error()的现代化流处理接口
- 📦 **响应收集**: 内置流式响应收集和聚合功能
- 🛡️ **错误处理**: 完善的流式错误处理和恢复机制

#### 工具调用 (Function Calling)
- 🔧 **流式工具调用**: 完整支持流模式下的工具调用
- 📊 **JSON累积器**: 自动处理分块传输的JSON参数
- 🛡️ **安全解析**: 提供ParseToolCallArgumentsSafe安全解析函数
- 🚫 **防过早退出**: 智能检测待完成工具调用，防止流提前结束
- 🔍 **调试支持**: 详细的工具调用状态追踪和调试信息

#### 思维模式支持
- 🧠 **Thinking模式**: 支持AI模型的思维过程展示
- 📝 **思维内容提取**: 自动解析和提取thinking标签内容
- 🔄 **流式思维**: 支持流式模式下的思维内容处理

### 🛠️ 技术特性

#### 架构设计
- 📦 **插件式架构**: 基于Provider模式的可扩展设计  
- 🔧 **工厂模式**: 统一的客户端创建和管理
- 🧩 **模块化**: 清晰的模块分离，易于维护和扩展

#### 可靠性
- ✅ **完整测试**: 单元测试、集成测试、压力测试覆盖
- 🔒 **并发安全**: 线程安全的状态管理和数据处理
- 📊 **性能优化**: 内存使用优化和高效的并发处理
- 🛡️ **错误恢复**: 完善的错误处理和异常恢复机制

#### 开发体验
- 📚 **丰富文档**: 详细的API文档和使用指南
- 🎯 **类型安全**: 完整的Go类型定义和检查
- 🔍 **调试友好**: 丰富的调试信息和错误提示
- 📝 **示例代码**: 多种使用场景的完整示例

### 🚀 使用场景

- **统一AI接入**: 需要支持多个AI服务商的应用
- **实时对话**: 聊天机器人、客服系统等
- **AI Agent**: 基于工具调用的智能代理
- **内容生成**: 文本生成、代码生成等应用
- **API迁移**: 从单一服务商迁移到多服务商架构

### 📋 兼容性

- **Go版本**: 要求Go 1.22+
- **API兼容**: 完全兼容OpenAI ChatCompletion API v1
- **向后兼容**: 保持与原DeepSeek客户端的完全兼容

### 🔗 快速开始

```bash
go get github.com/yu1ec/go-anyllm
```

### 📄 相关文档

- [README.md](README.md) - 基础使用指南
- [README_STREAMING_TOOLS.md](README_STREAMING_TOOLS.md) - 流式工具调用指南  
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - 架构设计文档
- [docs/TOOLS_GUIDE.md](docs/TOOLS_GUIDE.md) - 工具调用详细指南

---

*感谢 [go-deepseek/deepseek](https://github.com/go-deepseek/deepseek) 项目的设计启发*