package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yu1ec/go-anyllm/tools"
	"github.com/yu1ec/go-anyllm/types"
)

// 示例1：纯同步处理器
type SyncCalculator struct{}

func (sc *SyncCalculator) HandleToolCall(toolCall types.ToolCall) (string, error) {
	// 解析参数
	type Args struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	args, err := tools.ParseToolCallArguments[Args](toolCall)
	if err != nil {
		return "", err
	}

	result := args.A + args.B
	return fmt.Sprintf("计算结果: %d + %d = %d", args.A, args.B, result), nil
}

// 示例2：纯流式处理器
type StreamingWriter struct{}

func (sw *StreamingWriter) HandleToolCallStream(toolCall types.ToolCall) (<-chan tools.StreamChunk, error) {
	// 解析参数
	type Args struct {
		Text string `json:"text"`
	}

	args, err := tools.ParseToolCallArguments[Args](toolCall)
	if err != nil {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: err, Done: true}
		close(errChan)
		return errChan, nil
	}

	resultChan := make(chan tools.StreamChunk, 10)

	go func() {
		defer close(resultChan)

		// 模拟逐字输出
		for _, char := range args.Text {
			resultChan <- tools.StreamChunk{
				Content: string(char),
				Done:    false,
			}
			time.Sleep(100 * time.Millisecond) // 模拟延迟
		}

		resultChan <- tools.StreamChunk{
			Content: "",
			Done:    true,
		}
	}()

	return resultChan, nil
}

// 示例3：统一处理器（同时支持同步和流式）
type UnifiedProcessor struct{}

func (up *UnifiedProcessor) HandleToolCall(toolCall types.ToolCall) (string, error) {
	// 同步版本
	type Args struct {
		Message string `json:"message"`
	}

	args, err := tools.ParseToolCallArguments[Args](toolCall)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("处理完成: %s", args.Message), nil
}

func (up *UnifiedProcessor) HandleToolCallStream(toolCall types.ToolCall) (<-chan tools.StreamChunk, error) {
	// 流式版本
	type Args struct {
		Message string `json:"message"`
	}

	args, err := tools.ParseToolCallArguments[Args](toolCall)
	if err != nil {
		errChan := make(chan tools.StreamChunk, 1)
		errChan <- tools.StreamChunk{Error: err, Done: true}
		close(errChan)
		return errChan, nil
	}

	resultChan := make(chan tools.StreamChunk, 5)

	go func() {
		defer close(resultChan)

		chunks := []string{"正在处理", "...", "处理中", "...", fmt.Sprintf("完成: %s", args.Message)}
		for i, chunk := range chunks {
			resultChan <- tools.StreamChunk{
				Content: chunk,
				Done:    i == len(chunks)-1,
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	return resultChan, nil
}

// 示例4：可选流式处理器
type OptionalHandler struct {
	enableStreaming bool
}

func (oh *OptionalHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	return "这是同步结果", nil
}

func (oh *OptionalHandler) CanStream() bool {
	return oh.enableStreaming
}

func (oh *OptionalHandler) HandleToolCallStream(toolCall types.ToolCall) (<-chan tools.StreamChunk, error) {
	resultChan := make(chan tools.StreamChunk, 2)

	go func() {
		defer close(resultChan)
		resultChan <- tools.StreamChunk{Content: "这是流式结果", Done: true}
	}()

	return resultChan, nil
}

func main() {
	// 创建函数注册表
	registry := tools.NewFunctionRegistry()

	// 注册不同类型的处理器
	registry.Register("sync_calc", &SyncCalculator{})
	registry.RegisterStreaming("stream_writer", &StreamingWriter{})
	registry.RegisterUnified("unified_processor", &UnifiedProcessor{})
	registry.Register("optional_handler", &OptionalHandler{enableStreaming: true})

	// 测试同步处理
	fmt.Println("=== 测试同步处理 ===")
	syncToolCall := types.ToolCall{
		ID:   "test-1",
		Type: "function",
		Function: types.ToolFunction{
			Name:       "sync_calc",
			Parameters: `{"a": 10, "b": 20}`,
		},
	}

	result := registry.Handle(syncToolCall)
	fmt.Printf("同步结果: %s\n", result.Content)
	if result.Error != "" {
		fmt.Printf("错误: %s\n", result.Error)
	}

	// 测试流式处理
	fmt.Println("\n=== 测试流式处理 ===")
	streamToolCall := types.ToolCall{
		ID:   "test-2",
		Type: "function",
		Function: types.ToolFunction{
			Name:       "stream_writer",
			Parameters: `{"text": "Hello流式输出!"}`,
		},
	}

	streamChan, err := registry.HandleStreaming(streamToolCall)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("流式输出: ")
	for chunk := range streamChan {
		if chunk.Error != nil {
			fmt.Printf("错误: %v\n", chunk.Error)
			break
		}
		fmt.Print(chunk.Content)
		if chunk.Done {
			fmt.Println(" [完成]")
		}
	}

	// 测试统一处理器的流式模式
	fmt.Println("\n=== 测试统一处理器（流式） ===")
	unifiedToolCall := types.ToolCall{
		ID:   "test-3",
		Type: "function",
		Function: types.ToolFunction{
			Name:       "unified_processor",
			Parameters: `{"message": "测试消息"}`,
		},
	}

	streamChan2, err := registry.HandleStreaming(unifiedToolCall)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("统一处理器流式: ")
	for chunk := range streamChan2 {
		if chunk.Error != nil {
			fmt.Printf("错误: %v\n", chunk.Error)
			break
		}
		fmt.Print(chunk.Content)
		if chunk.Done {
			fmt.Println(" [完成]")
		}
	}

	// 测试能力检查
	fmt.Println("\n=== 测试能力检查 ===")
	functions := []string{"sync_calc", "stream_writer", "unified_processor", "optional_handler"}
	for _, fn := range functions {
		canStream := registry.CanHandleStreaming(fn)
		fmt.Printf("%s 支持流式处理: %t\n", fn, canStream)
	}
}
