/*
流式工具调用演示程序

本程序展示了如何使用 go-anyllm 库的流式工具调用功能，包括：

1. 同步工具调用处理器：
   - DemoWeatherHandler: 天气查询
   - DemoCalculatorHandler: 计算器
   - DemoUserProfileHandler: 用户档案查询

2. 流式工具调用处理器：
   - StreamingWriterHandler: 流式内容写作（逐词输出）
   - StreamingCodeGeneratorHandler: 流式代码生成（逐字符输出）

3. 统一工具调用处理器：
   - UnifiedDataAnalyzerHandler: 数据分析（同时支持同步和流式）

演示特性：
- 流式工具调用的实时输出
- 不同延迟模式的演示
- 自动检测工具是否支持流式处理
- 优雅的错误处理和状态显示

运行前请设置环境变量：
export ALIYUN_API_KEY="your_api_key"
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	deepseek "github.com/yu1ec/go-anyllm"
	"github.com/yu1ec/go-anyllm/tools"
	"github.com/yu1ec/go-anyllm/types"
)

// DemoWeatherHandler 演示用天气查询处理器
type DemoWeatherHandler struct{}

func (h *DemoWeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	// 使用安全解析函数
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}](toolCall)

	if err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}
	if !isComplete {
		return "", fmt.Errorf("参数不完整，需要继续等待")
	}

	// 模拟天气查询
	unit := "°C"
	if params.Unit == "fahrenheit" {
		unit = "°F"
	}

	return fmt.Sprintf("%s: 晴天, 气温25%s, 湿度60%%", params.Location, unit), nil
}

// DemoCalculatorHandler 演示用计算器处理器
type DemoCalculatorHandler struct{}

func (h *DemoCalculatorHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
		Expression string `json:"expression"`
	}](toolCall)

	if err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}
	if !isComplete {
		return "", fmt.Errorf("参数不完整，需要继续等待")
	}

	// 简单的数学计算
	switch params.Expression {
	case "25+17", "25 + 17":
		return "42", nil
	case "2+3*4", "2 + 3 * 4":
		return "14", nil
	default:
		return fmt.Sprintf("计算结果: %s (模拟)", params.Expression), nil
	}
}

// DemoUserProfileHandler 演示用用户档案查询处理器 - 测试复杂参数解析
type DemoUserProfileHandler struct{}

func (h *DemoUserProfileHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	// 定义复杂的参数结构
	type Filters struct {
		Status     string `json:"status"`
		Department string `json:"department"`
		Priority   int    `json:"priority"`
	}

	type UserProfileParams struct {
		UserID         string  `json:"user_id"`
		QueryType      string  `json:"query_type"`
		Fields         string  `json:"fields"`
		IncludeHistory bool    `json:"include_history"`
		MaxRecords     int     `json:"max_records"`
		DateRange      string  `json:"date_range"`
		Filters        Filters `json:"filters"`
	}

	// 使用安全解析函数
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[UserProfileParams](toolCall)
	if err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}
	if !isComplete {
		return "", fmt.Errorf("参数不完整，需要继续等待")
	}

	// 构建模拟响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("📋 用户档案查询结果 (ID: %s)\n", params.UserID))
	result.WriteString(fmt.Sprintf("📊 查询类型: %s\n", params.QueryType))

	if params.Fields != "" {
		result.WriteString(fmt.Sprintf("🔍 指定字段: %s\n", params.Fields))
	}

	if params.IncludeHistory {
		result.WriteString("📚 包含历史记录: 是\n")
		if params.MaxRecords > 0 {
			result.WriteString(fmt.Sprintf("📝 最大记录数: %d\n", params.MaxRecords))
		}
	}

	if params.DateRange != "" {
		result.WriteString(fmt.Sprintf("📅 日期范围: %s\n", params.DateRange))
	}

	// 处理过滤条件
	if params.Filters.Status != "" || params.Filters.Department != "" || params.Filters.Priority > 0 {
		result.WriteString("🔧 应用的过滤条件:\n")
		if params.Filters.Status != "" {
			result.WriteString(fmt.Sprintf("   - 状态: %s\n", params.Filters.Status))
		}
		if params.Filters.Department != "" {
			result.WriteString(fmt.Sprintf("   - 部门: %s\n", params.Filters.Department))
		}
		if params.Filters.Priority > 0 {
			result.WriteString(fmt.Sprintf("   - 优先级: %d\n", params.Filters.Priority))
		}
	}

	// 模拟用户数据
	result.WriteString("\n✅ 查询结果:\n")
	result.WriteString("   姓名: 张三\n")
	result.WriteString("   部门: 技术部\n")
	result.WriteString("   状态: active\n")
	result.WriteString("   注册时间: 2023-01-15\n")

	if params.IncludeHistory {
		result.WriteString("   最近活动: 2024-01-10 登录系统\n")
	}

	return result.String(), nil
}

// StreamingWriterHandler 流式写作处理器
type StreamingWriterHandler struct{}

func (h *StreamingWriterHandler) HandleToolCallStream(toolCall types.ToolCall) (<-chan tools.StreamChunk, error) {
	// 解析参数
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
		Topic  string `json:"topic"`
		Length string `json:"length"`
		Style  string `json:"style"`
	}](toolCall)

	if err != nil {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: err, Done: true}
		close(errChan)
		return errChan, nil
	}
	if !isComplete {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: fmt.Errorf("参数不完整，需要继续等待"), Done: true}
		close(errChan)
		return errChan, nil
	}

	resultChan := make(chan tools.StreamChunk, 50)

	go func() {
		defer close(resultChan)

		// 根据主题生成不同的内容
		var content []string
		switch params.Topic {
		case "技术":
			content = []string{
				"📚 关于", params.Topic, "的", params.Style, "文章：\n\n",
				"在当今快速发展的", "技术", "领域中，", "我们", "需要", "不断", "学习", "和", "适应", "新的", "变化", "。\n",
				"人工智能", "、", "云计算", "、", "大数据", "等", "技术", "正在", "改变", "我们的", "工作", "方式", "。\n",
				"掌握", "这些", "技术", "不仅", "能够", "提高", "工作", "效率", "，", "还能", "为", "未来", "的", "发展", "奠定", "基础", "。\n\n",
				"✨ 总结：", "持续", "学习", "，", "拥抱", "变化", "！",
			}
		case "生活":
			content = []string{
				"🌟 关于", params.Topic, "的", params.Style, "分享：\n\n",
				"生活", "是", "一场", "美妙", "的", "旅程", "，", "充满", "了", "惊喜", "和", "挑战", "。\n",
				"每一天", "都是", "新的", "开始", "，", "每一刻", "都", "值得", "珍惜", "。\n",
				"无论", "遇到", "什么", "困难", "，", "保持", "积极", "的", "心态", "都是", "最", "重要", "的", "。\n",
				"与", "家人", "朋友", "分享", "快乐", "，", "一起", "度过", "难关", "，", "这就是", "生活", "的", "意义", "。\n\n",
				"💫 感悟：", "珍惜", "当下", "，", "享受", "生活", "！",
			}
		default:
			content = []string{
				"📝 关于", "\"" + params.Topic + "\"", "的", params.Style, "内容：\n\n",
				"这是", "一个", "有趣", "的", "话题", "，", "值得", "我们", "深入", "探讨", "。\n",
				"通过", "不同", "的", "角度", "和", "视角", "，", "我们", "可以", "获得", "新的", "见解", "。\n",
				"希望", "这些", "内容", "能够", "为", "您", "提供", "有用", "的", "信息", "和", "启发", "。\n\n",
				"🎯 结论：", "持续", "探索", "，", "不断", "成长", "！",
			}
		}

		// 流式输出内容
		for _, word := range content {
			resultChan <- tools.StreamChunk{
				Content: word,
				Done:    false,
			}

			// 根据长度设置不同的延迟
			var delay int
			switch params.Length {
			case "短":
				delay = 50
			case "长":
				delay = 150
			default:
				delay = 100
			}

			// 模拟写作延迟
			time.Sleep(time.Duration(delay) * time.Millisecond)

			// 在句号后添加稍长的停顿
			if word == "。" || word == "！" || word == "？" {
				time.Sleep(300 * time.Millisecond)
			}
		}

		// 发送完成信号
		resultChan <- tools.StreamChunk{
			Content: "",
			Done:    true,
		}
	}()

	return resultChan, nil
}

// StreamingCodeGeneratorHandler 流式代码生成处理器
type StreamingCodeGeneratorHandler struct{}

func (h *StreamingCodeGeneratorHandler) HandleToolCallStream(toolCall types.ToolCall) (<-chan tools.StreamChunk, error) {
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
		Language    string `json:"language"`
		Description string `json:"description"`
		Complexity  string `json:"complexity"`
	}](toolCall)

	if err != nil {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: err, Done: true}
		close(errChan)
		return errChan, nil
	}
	if !isComplete {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: fmt.Errorf("参数不完整，需要继续等待"), Done: true}
		close(errChan)
		return errChan, nil
	}

	resultChan := make(chan tools.StreamChunk, 50)

	go func() {
		defer close(resultChan)

		// 生成代码模板
		var codeLines []string
		switch params.Language {
		case "go":
			codeLines = []string{
				"```go\n",
				"package main\n\n",
				"import (\n",
				"\t\"fmt\"\n",
				")\n\n",
				"// ", params.Description, "\n",
				"func main() {\n",
				"\tfmt.Println(\"Hello, ", params.Description, "!\")\n",
				"}\n",
				"```\n",
			}
		case "python":
			codeLines = []string{
				"```python\n",
				"# ", params.Description, "\n\n",
				"def main():\n",
				"    print(\"Hello, ", params.Description, "!\")\n",
				"    return True\n\n",
				"if __name__ == \"__main__\":\n",
				"    main()\n",
				"```\n",
			}
		default:
			codeLines = []string{
				"```\n",
				"// ", params.Description, "\n",
				"// 语言: ", params.Language, "\n",
				"// 复杂度: ", params.Complexity, "\n\n",
				"这里是示例代码...\n",
				"```\n",
			}
		}

		// 流式输出代码
		for _, line := range codeLines {
			// 逐字符输出，模拟打字效果
			for _, char := range line {
				resultChan <- tools.StreamChunk{
					Content: string(char),
					Done:    false,
				}
				time.Sleep(30 * time.Millisecond) // 快速打字效果
			}
		}

		resultChan <- tools.StreamChunk{
			Content: "",
			Done:    true,
		}
	}()

	return resultChan, nil
}

// UnifiedDataAnalyzerHandler 统一数据分析处理器（同时支持同步和流式）
type UnifiedDataAnalyzerHandler struct{}

func (h *UnifiedDataAnalyzerHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
		DataType   string `json:"data_type"`
		Operation  string `json:"operation"`
		Parameters string `json:"parameters"`
	}](toolCall)

	if err != nil {
		return "", err
	}
	if !isComplete {
		return "", fmt.Errorf("参数不完整，需要继续等待")
	}

	return fmt.Sprintf("📊 数据分析完成\n类型: %s\n操作: %s\n参数: %s\n结果: 分析成功",
		params.DataType, params.Operation, params.Parameters), nil
}

func (h *UnifiedDataAnalyzerHandler) HandleToolCallStream(toolCall types.ToolCall) (<-chan tools.StreamChunk, error) {
	params, isComplete, err := tools.ParseToolCallArgumentsSafe[struct {
		DataType   string `json:"data_type"`
		Operation  string `json:"operation"`
		Parameters string `json:"parameters"`
	}](toolCall)

	if err != nil {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: err, Done: true}
		close(errChan)
		return errChan, nil
	}
	if !isComplete {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: fmt.Errorf("参数不完整，需要继续等待"), Done: true}
		close(errChan)
		return errChan, nil
	}

	resultChan := make(chan tools.StreamChunk, 20)

	go func() {
		defer close(resultChan)

		steps := []string{
			"📊 开始数据分析...\n",
			"🔍 正在加载数据集 (" + params.DataType + ")...\n",
			"⚙️  执行操作: " + params.Operation + "\n",
			"📈 应用参数: " + params.Parameters + "\n",
			"🧮 计算统计指标...\n",
			"📋 生成分析报告...\n",
			"✅ 分析完成！\n\n",
			"📊 结果摘要:\n",
			"- 数据类型: " + params.DataType + "\n",
			"- 处理操作: " + params.Operation + "\n",
			"- 成功率: 98.5%\n",
			"- 处理时间: 2.3秒\n",
		}

		for i, step := range steps {
			resultChan <- tools.StreamChunk{
				Content: step,
				Done:    i == len(steps)-1,
			}
			time.Sleep(400 * time.Millisecond) // 模拟分析步骤延迟
		}
	}()

	return resultChan, nil
}

// StreamingWriterTool 流式写作工具定义
func StreamingWriterTool() types.Tool {
	return tools.NewTool("stream_writer", "流式内容写作工具，可以生成不同主题和风格的文章").
		AddStringParam("topic", "写作主题（如：技术、生活、工作等）", true).
		AddStringParam("style", "写作风格（如：正式、轻松、学术等）", false).
		AddStringParam("length", "内容长度（短、中、长）", false).
		BuildForTypes()
}

// StreamingCodeGeneratorTool 流式代码生成工具定义
func StreamingCodeGeneratorTool() types.Tool {
	return tools.NewTool("stream_code_generator", "流式代码生成工具，支持多种编程语言").
		AddStringParam("language", "编程语言（如：go、python、javascript等）", true).
		AddStringParam("description", "代码功能描述", true).
		AddStringParam("complexity", "代码复杂度（简单、中等、复杂）", false).
		BuildForTypes()
}

// UnifiedDataAnalyzerTool 统一数据分析工具定义
func UnifiedDataAnalyzerTool() types.Tool {
	return tools.NewTool("unified_data_analyzer", "数据分析工具，支持多种数据类型和分析操作").
		AddStringParam("data_type", "数据类型（如：csv、json、xlsx等）", true).
		AddStringParam("operation", "分析操作（如：统计、可视化、预测等）", true).
		AddStringParam("parameters", "分析参数（JSON格式）", false).
		BuildForTypes()
}

func main() {
	// 检查API Key
	apiKey := os.Getenv("ALIYUN_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 ALIYUN_API_KEY 环境变量")
	}

	// 创建客户端 阿里云客户端
	client, err := deepseek.NewAliCloudClient(apiKey)
	if err != nil {
		log.Fatal("创建客户端失败:", err)
	}

	// 创建流式工具调用累积器
	accumulator := tools.NewStreamingToolCallAccumulator()
	registry := tools.NewFunctionRegistry()
	registry.Register("get_weather", &DemoWeatherHandler{})
	registry.Register("calculator", &DemoCalculatorHandler{})
	registry.Register("query_user_profile", &DemoUserProfileHandler{}) // 注册新的自定义工具
	registry.RegisterStreaming("stream_writer", &StreamingWriterHandler{})
	registry.RegisterStreaming("stream_code_generator", &StreamingCodeGeneratorHandler{})
	registry.RegisterUnified("unified_data_analyzer", &UnifiedDataAnalyzerHandler{})

	// 创建请求
	req := &types.ChatCompletionRequest{
		Model:  "qwen-plus",
		Stream: true,
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleSystem,
				Content: "你是一个有用的助手，可以查询天气、进行计算和查询用户档案信息。",
			},
			{
				Role:    types.RoleUser,
				Content: "请为我写一篇关于'技术'主题的正式风格文章，长度为中等，然后用Go语言生成一个'Hello World'程序",
			},
		},
		Tools: []types.Tool{
			tools.GetWeatherTool(),
			tools.CalculatorTool(),
			tools.CustomUserProfileTool(), // 添加自定义工具
			StreamingWriterTool(),         // 流式写作工具
			StreamingCodeGeneratorTool(),  // 流式代码生成工具
			UnifiedDataAnalyzerTool(),     // 统一数据分析工具
		},
		ToolChoice: tools.Choice.Auto(),
	}

	// 发送流式请求
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Fatal("创建流失败:", err)
	}

	fmt.Println("=== 流式工具调用演示 ===")
	fmt.Println("正在处理您的请求...")

	var toolCallsExecuted int

	for stream.Next() {
		chunk := stream.Current()
		if chunk == nil || len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		if choice.Delta == nil {
			continue
		}

		// 处理文本内容
		if choice.Delta.Content != "" {
			fmt.Print(choice.Delta.Content)
		}

		// 处理推理内容（DeepSeek特有）
		if choice.Delta.ReasoningContent != "" {
			fmt.Printf("[思考] %s", choice.Delta.ReasoningContent)
		}

		// 处理流式工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			fmt.Printf("\n[DEBUG] 收到工具调用Delta，数量: %d\n", len(choice.Delta.ToolCalls))

			// 使用累积器处理Delta
			accumulator.ProcessDelta(choice.Delta.ToolCalls)

			// 检查是否有完成的工具调用
			completed := accumulator.GetCompletedToolCalls()
			if len(completed) > 0 {
				fmt.Printf("\n🔧 执行工具调用 (%d个):\n", len(completed))
				for _, toolCall := range completed {
					fmt.Printf("   调用: %s", toolCall.Function.Name)

					// 检查是否支持流式处理
					if registry.CanHandleStreaming(toolCall.Function.Name) {
						fmt.Printf(" [流式] ")

						// 使用流式处理
						streamChan, err := registry.HandleStreaming(toolCall)
						if err != nil {
							fmt.Printf(" ❌ 流式错误: %s\n", err)
							continue
						}

						fmt.Print("🔄 ")
						for chunk := range streamChan {
							if chunk.Error != nil {
								fmt.Printf(" ❌ 错误: %v\n", chunk.Error)
								break
							}
							fmt.Print(chunk.Content)
							if chunk.Done {
								fmt.Printf(" ✅ [流式完成]\n")
							}
						}
						toolCallsExecuted++
					} else {
						// 使用同步处理
						result := registry.Handle(toolCall)
						if result.Error != "" {
							fmt.Printf(" ❌ 错误: %s\n", result.Error)
						} else {
							fmt.Printf(" ✅ 结果: %s\n", result.Content)
							toolCallsExecuted++
						}
					}
				}
				// 清除已完成的工具调用
				accumulator.ClearCompleted()
			}

			// 显示待完成的工具调用状态（可选）
			if accumulator.HasPendingToolCalls() {
				fmt.Printf("[DEBUG] 还有 %d 个工具调用正在生成中...\n", accumulator.GetPendingCount())
			}
		}

		// 检查流结束
		if choice.FinishReason != "" {
			fmt.Printf("\n\n[流结束] 原因: %s\n", choice.FinishReason)

			// 检查是否还有待完成的工具调用
			if accumulator.HasPendingToolCalls() {
				fmt.Printf("[等待] 仍有 %d 个工具调用正在生成中，继续等待...\n", accumulator.GetPendingCount())
				continue // 不要退出，继续等待
			}

			// 最终检查：处理流结束时可能遗留的工具调用
			finalCompleted := accumulator.FinalizeStream()
			if len(finalCompleted) > 0 {
				fmt.Printf("\n🔧 执行最终工具调用 (%d个):\n", len(finalCompleted))
				for _, toolCall := range finalCompleted {
					// 检查是否支持流式处理
					if registry.CanHandleStreaming(toolCall.Function.Name) {
						fmt.Printf("   %s [流式] ", toolCall.Function.Name)

						// 使用流式处理
						streamChan, err := registry.HandleStreaming(toolCall)
						if err != nil {
							fmt.Printf("❌ 流式错误: %s\n", err)
							continue
						}

						fmt.Print("🔄 ")
						for chunk := range streamChan {
							if chunk.Error != nil {
								fmt.Printf("❌ 错误: %v\n", chunk.Error)
								break
							}
							fmt.Print(chunk.Content)
							if chunk.Done {
								fmt.Printf(" ✅ [流式完成]\n")
							}
						}
						toolCallsExecuted++
					} else {
						// 使用同步处理
						result := registry.Handle(toolCall)
						if result.Error != "" {
							fmt.Printf("   %s ❌ 错误: %s\n", toolCall.Function.Name, result.Error)
						} else {
							fmt.Printf("   %s ✅ 结果: %s\n", toolCall.Function.Name, result.Content)
							toolCallsExecuted++
						}
					}
				}
			}
			break
		}
	}

	// 检查流处理错误
	if err := stream.Error(); err != nil {
		log.Fatal("流处理错误:", err)
	}

	fmt.Printf("\n=== 流式工具调用演示完成 ===\n")
	fmt.Printf("📊 执行统计:\n")
	fmt.Printf("   - 总共执行了 %d 个工具调用\n", toolCallsExecuted)
	fmt.Printf("   - 累积器统计: 总计 %d, 已完成 %d, 待完成 %d\n",
		accumulator.GetTotalCount(),
		accumulator.GetCompletedCount(),
		accumulator.GetPendingCount())

	// 显示工具能力统计
	fmt.Printf("\n🔧 工具能力统计:\n")
	allTools := []string{"get_weather", "calculator", "query_user_profile", "stream_writer", "stream_code_generator", "unified_data_analyzer"}
	for _, toolName := range allTools {
		canStream := registry.CanHandleStreaming(toolName)
		streamSymbol := "❌"
		if canStream {
			streamSymbol = "✅"
		}
		fmt.Printf("   - %s: 流式支持 %s\n", toolName, streamSymbol)
	}
}
