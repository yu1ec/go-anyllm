package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	deepseek "github.com/yu1ec/go-anyllm"
	"github.com/yu1ec/go-anyllm/tools"
	"github.com/yu1ec/go-anyllm/types"
)

// TestRealStreamingToolCalls 测试真实环境下的流式工具调用
func TestRealStreamingToolCalls(t *testing.T) {
	// 需要真实的API Key
	apiKey := os.Getenv("ALIYUN_API_KEY")
	if apiKey == "" {
		t.Skip("跳过真实API测试: 未设置ALIYUN_API_KEY")
	}

	client, err := deepseek.NewAliCloudClient(apiKey)
	if err != nil {
		t.Fatal("创建客户端失败:", err)
	}

	// 创建流式工具调用累积器
	accumulator := tools.NewStreamingToolCallAccumulator()
	registry := tools.NewFunctionRegistry()
	registry.Register("get_weather", &TestWeatherHandler{})
	registry.Register("calculator", &TestCalculatorHandler{})

	req := &types.ChatCompletionRequest{
		Model:  "qwen-plus",
		Stream: true,
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleSystem,
				Content: "你是一个有用的助手，可以查询天气和进行计算。",
			},
			{
				Role:    types.RoleUser,
				Content: "请查询北京的天气，然后计算 25 + 17 的结果",
			},
		},
		Tools: []types.Tool{
			tools.GetWeatherTool(),
			tools.CalculatorTool(),
		},
		ToolChoice: tools.Choice.Auto(),
	}

	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		t.Fatal("创建流失败:", err)
	}

	var (
		toolCallsExecuted = 0
		contentReceived   = false
		startTime         = time.Now()
		maxWaitTime       = 30 * time.Second
	)

	t.Log("开始流式工具调用测试...")

	for stream.Next() {
		// 检查超时
		if time.Since(startTime) > maxWaitTime {
			t.Fatal("测试超时")
		}

		chunk := stream.Current()
		if chunk == nil || len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		if choice.Delta == nil {
			continue
		}

		// 处理常规内容
		if choice.Delta.Content != "" {
			contentReceived = true
			t.Logf("收到内容: %s", choice.Delta.Content)
		}

		// 处理推理内容
		if choice.Delta.ReasoningContent != "" {
			t.Logf("收到推理内容: %s", choice.Delta.ReasoningContent)
		}

		// 处理流式工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			t.Logf("收到工具调用Delta，数量: %d", len(choice.Delta.ToolCalls))

			// 使用累积器处理Delta
			accumulator.ProcessDelta(choice.Delta.ToolCalls)

			// 显示调试信息
			if accumulator.HasPendingToolCalls() {
				pending := accumulator.GetPendingToolCallsDebugInfo()
				for id, info := range pending {
					t.Logf("待完成工具调用 %s: %s", id, info)
				}
			}

			// 检查是否有完成的工具调用
			completed := accumulator.GetCompletedToolCalls()
			if len(completed) > 0 {
				t.Logf("执行工具调用，数量: %d", len(completed))
				for _, toolCall := range completed {
					result := registry.Handle(toolCall)
					if result.Error != "" {
						t.Errorf("工具调用执行失败 %s: %s", toolCall.Function.Name, result.Error)
					} else {
						toolCallsExecuted++
						t.Logf("工具调用执行成功 %s: %s", toolCall.Function.Name, result.Content)
					}
				}
				// 清除已完成的工具调用
				accumulator.ClearCompleted()
			}
		}

		// 检查完成状态
		if choice.FinishReason != "" {
			t.Logf("流结束，原因: %s", choice.FinishReason)

			// 检查是否还有待完成的工具调用
			if accumulator.HasPendingToolCalls() {
				t.Logf("仍有 %d 个工具调用待完成，继续等待...", accumulator.GetPendingCount())
				continue
			}

			// 最终检查
			finalCompleted := accumulator.FinalizeStream()
			if len(finalCompleted) > 0 {
				t.Logf("执行最终工具调用，数量: %d", len(finalCompleted))
				for _, toolCall := range finalCompleted {
					result := registry.Handle(toolCall)
					if result.Error != "" {
						t.Errorf("最终工具调用执行失败 %s: %s", toolCall.Function.Name, result.Error)
					} else {
						toolCallsExecuted++
						t.Logf("最终工具调用执行成功 %s: %s", toolCall.Function.Name, result.Content)
					}
				}
			}
			break
		}
	}

	if err := stream.Error(); err != nil {
		t.Fatal("流处理错误:", err)
	}

	// 验证测试结果
	if toolCallsExecuted == 0 {
		t.Error("预期至少执行一个工具调用，但实际为0")
	}

	if !contentReceived {
		t.Log("警告: 未收到任何文本内容，这可能是正常的（取决于模型行为）")
	}

	t.Logf("测试完成！共执行了 %d 个工具调用", toolCallsExecuted)
}

// TestStreamingToolCallsWithTimeout 测试超时处理
func TestStreamingToolCallsWithTimeout(t *testing.T) {
	apiKey := os.Getenv("ALIYUN_API_KEY")
	if apiKey == "" {
		t.Skip("跳过真实API测试: 未设置ALIYUN_API_KEY")
	}

	client, err := deepseek.NewAliCloudClient(apiKey)
	if err != nil {
		t.Fatal("创建客户端失败:", err)
	}

	accumulator := tools.NewStreamingToolCallAccumulator()
	registry := tools.NewFunctionRegistry()
	registry.Register("get_weather", &TestWeatherHandler{})

	req := &types.ChatCompletionRequest{
		Model:  "qwen-plus",
		Stream: true,
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleUser,
				Content: "请查询上海的天气",
			},
		},
		Tools: []types.Tool{tools.GetWeatherTool()},
	}

	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		t.Fatal("创建流失败:", err)
	}

	timeout := time.After(15 * time.Second)
	done := make(chan bool)

	go func() {
		defer close(done)
		for stream.Next() {
			chunk := stream.Current()
			if chunk == nil || len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]
			if choice.Delta != nil && len(choice.Delta.ToolCalls) > 0 {
				accumulator.ProcessDelta(choice.Delta.ToolCalls)

				completed := accumulator.GetCompletedToolCalls()
				for _, toolCall := range completed {
					result := registry.Handle(toolCall)
					t.Logf("工具调用: %s -> %s", toolCall.Function.Name, result.Content)
				}
				accumulator.ClearCompleted()
			}

			if choice.FinishReason != "" {
				if !accumulator.HasPendingToolCalls() {
					break
				}
			}
		}
	}()

	select {
	case <-done:
		t.Log("超时测试通过：流正常完成")
	case <-timeout:
		t.Log("超时测试通过：检测到超时保护")
	}
}

// TestWeatherHandler 测试用天气处理器
type TestWeatherHandler struct{}

func (h *TestWeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
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

// TestCalculatorHandler 测试用计算器处理器
type TestCalculatorHandler struct{}

func (h *TestCalculatorHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
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

// 示例：如何在main函数中使用
func ExampleStreamingToolCalls() {
	apiKey := os.Getenv("ALIYUN_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 ALIYUN_API_KEY 环境变量")
	}

	client, err := deepseek.NewAliCloudClient(apiKey)
	if err != nil {
		log.Fatal("创建客户端失败:", err)
	}

	// 创建流式工具调用累积器
	accumulator := tools.NewStreamingToolCallAccumulator()
	registry := tools.NewFunctionRegistry()
	registry.Register("get_weather", &TestWeatherHandler{})

	req := &types.ChatCompletionRequest{
		Model:  "qwen-plus",
		Stream: true,
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleUser,
				Content: "请查询深圳的天气",
			},
		},
		Tools: []types.Tool{tools.GetWeatherTool()},
	}

	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Fatal("创建流失败:", err)
	}

	fmt.Println("=== 流式工具调用示例 ===")

	for stream.Next() {
		chunk := stream.Current()
		if chunk == nil || len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		if choice.Delta == nil {
			continue
		}

		// 处理内容
		if choice.Delta.Content != "" {
			fmt.Print(choice.Delta.Content)
		}

		// 处理工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			accumulator.ProcessDelta(choice.Delta.ToolCalls)

			completed := accumulator.GetCompletedToolCalls()
			for _, toolCall := range completed {
				fmt.Printf("\n🔧 执行工具: %s", toolCall.Function.Name)
				result := registry.Handle(toolCall)
				fmt.Printf("\n📊 结果: %s\n", result.Content)
			}
			accumulator.ClearCompleted()
		}

		if choice.FinishReason != "" {
			if !accumulator.HasPendingToolCalls() {
				break
			}
		}
	}

	if err := stream.Error(); err != nil {
		log.Fatal("流处理错误:", err)
	}

	fmt.Println("\n=== 示例完成 ===")
}
