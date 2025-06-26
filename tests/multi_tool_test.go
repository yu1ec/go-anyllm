package tests

import (
	"context"
	"os"
	"testing"
	"time"

	deepseek "github.com/yu1ec/go-anyllm"
	"github.com/yu1ec/go-anyllm/tools"
	"github.com/yu1ec/go-anyllm/types"
)

// TestMultipleToolCallsRequired 测试需要多个工具调用的场景
func TestMultipleToolCallsRequired(t *testing.T) {
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

	// 明确要求使用多个工具的提示
	req := &types.ChatCompletionRequest{
		Model:  "qwen-plus",
		Stream: true,
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleSystem,
				Content: "你必须使用工具来回答用户的问题。对于每个不同的任务，你都需要调用相应的工具。",
			},
			{
				Role:    types.RoleUser,
				Content: "请帮我做三件事：1. 查询北京的天气 2. 计算 25 + 17 的结果 3. 查询上海的天气。请分别使用对应的工具来完成这些任务。",
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
		uniqueToolTypes   = make(map[string]int) // 记录不同类型工具的调用次数
		startTime         = time.Now()
		maxWaitTime       = 60 * time.Second
	)

	t.Log("开始多工具调用测试...")

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
			t.Logf("收到内容: %s", choice.Delta.Content)
		}

		// 处理流式工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			t.Logf("收到工具调用Delta，数量: %d", len(choice.Delta.ToolCalls))

			// 使用累积器处理Delta
			accumulator.ProcessDelta(choice.Delta.ToolCalls)

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
						uniqueToolTypes[toolCall.Function.Name]++
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
						uniqueToolTypes[toolCall.Function.Name]++
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
	t.Logf("测试完成！共执行了 %d 个工具调用", toolCallsExecuted)
	t.Logf("工具类型分布: %+v", uniqueToolTypes)

	// 验证是否调用了多个工具
	if toolCallsExecuted < 2 {
		t.Errorf("期望至少执行2个工具调用（天气查询和计算器），但实际只执行了%d个", toolCallsExecuted)
	}

	// 验证是否使用了不同类型的工具
	if len(uniqueToolTypes) < 2 {
		t.Errorf("期望使用至少2种不同的工具，但实际只使用了%d种: %v", len(uniqueToolTypes), uniqueToolTypes)
	}

	// 检查具体的工具调用
	if uniqueToolTypes["get_weather"] == 0 {
		t.Error("期望调用天气查询工具，但没有调用")
	}
	if uniqueToolTypes["calculator"] == 0 {
		t.Error("期望调用计算器工具，但没有调用")
	}

	t.Logf("多工具调用测试验证通过！")
}

// TestSequentialVsParallelToolCalls 测试顺序vs并行工具调用
func TestSequentialVsParallelToolCalls(t *testing.T) {
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
	registry.Register("calculator", &TestCalculatorHandler{})

	// 测试并行工具调用的提示
	req := &types.ChatCompletionRequest{
		Model:  "qwen-plus",
		Stream: true,
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleSystem,
				Content: "你可以同时调用多个工具来提高效率。",
			},
			{
				Role:    types.RoleUser,
				Content: "请同时查询北京和上海的天气，并且计算 100 + 200 的结果。这些任务可以并行执行。",
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
		toolCallTimestamps = make(map[string]time.Time)
		startTime          = time.Now()
		maxWaitTime        = 60 * time.Second
	)

	t.Log("开始并行工具调用测试...")

	for stream.Next() {
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

		if len(choice.Delta.ToolCalls) > 0 {
			accumulator.ProcessDelta(choice.Delta.ToolCalls)

			completed := accumulator.GetCompletedToolCalls()
			if len(completed) > 0 {
				for _, toolCall := range completed {
					// 记录工具调用的时间戳
					toolCallTimestamps[toolCall.ID] = time.Now()

					result := registry.Handle(toolCall)
					if result.Error == "" {
						t.Logf("工具调用 %s (%s) 在 %v 执行成功: %s",
							toolCall.ID, toolCall.Function.Name,
							time.Since(startTime), result.Content)
					}
				}
				accumulator.ClearCompleted()
			}
		}

		if choice.FinishReason != "" {
			if !accumulator.HasPendingToolCalls() {
				break
			}
		}
	}

	// 分析工具调用的时间分布
	if len(toolCallTimestamps) >= 2 {
		var timestamps []time.Time
		for _, ts := range toolCallTimestamps {
			timestamps = append(timestamps, ts)
		}

		// 检查是否有工具调用在相近时间内发生（表示并行执行）
		t.Logf("工具调用时间分析:")
		for id, ts := range toolCallTimestamps {
			t.Logf("  %s: %v", id, ts.Sub(startTime))
		}
	}

	t.Logf("并行工具调用测试完成，共记录了 %d 个工具调用", len(toolCallTimestamps))
}
