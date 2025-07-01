# 历史版本说明

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