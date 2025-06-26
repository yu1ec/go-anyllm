package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
				Content: "请查询用户ID为'user123'的详细档案信息，包含历史记录，最多显示10条记录，只需要姓名和部门字段，过滤条件是状态为active的用户",
			},
		},
		Tools: []types.Tool{
			tools.GetWeatherTool(),
			tools.CalculatorTool(),
			tools.CustomUserProfileTool(), // 添加自定义工具
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

					result := registry.Handle(toolCall)
					if result.Error != "" {
						fmt.Printf(" ❌ 错误: %s\n", result.Error)
					} else {
						fmt.Printf(" ✅ 结果: %s\n", result.Content)
						toolCallsExecuted++
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
					result := registry.Handle(toolCall)
					if result.Error != "" {
						fmt.Printf("   %s ❌ 错误: %s\n", toolCall.Function.Name, result.Error)
					} else {
						fmt.Printf("   %s ✅ 结果: %s\n", toolCall.Function.Name, result.Content)
						toolCallsExecuted++
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

	fmt.Printf("\n=== 处理完成 ===\n")
	fmt.Printf("总共执行了 %d 个工具调用\n", toolCallsExecuted)

	// 显示最终统计
	fmt.Printf("累积器统计: 总计 %d, 已完成 %d, 待完成 %d\n",
		accumulator.GetTotalCount(),
		accumulator.GetCompletedCount(),
		accumulator.GetPendingCount())
}
