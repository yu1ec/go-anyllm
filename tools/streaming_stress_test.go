package tools

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yu1ec/go-anyllm/response"
)

// TestStreamingToolCallsStress 压力测试：大量并发工具调用
func TestStreamingToolCallsStress(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	numCalls := 100
	var wg sync.WaitGroup

	t.Logf("开始压力测试：%d 个并发工具调用", numCalls)

	// 模拟大量并发的工具调用
	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			callID := fmt.Sprintf("call_%d", id)
			// 模拟分块传输
			chunks := []string{
				fmt.Sprintf(`{"param_%d":`, id),
				` "value_`,
				fmt.Sprintf(`%d"}`, id),
			}

			for j, chunk := range chunks {
				delta := []*response.ToolCall{{
					Id:   callID,
					Type: "function",
					Function: response.ToolFunction{
						Name:      "test_function",
						Arguments: chunk,
					},
				}}

				// 如果是第一个chunk，设置完整信息
				if j == 0 {
					delta[0].Type = "function"
					delta[0].Function.Name = "test_function"
				} else {
					// 后续chunk只有参数片段，但保持ID一致
					delta[0].Type = ""
					delta[0].Function.Name = ""
				}

				accumulator.ProcessDelta(delta)

				// 模拟网络延迟
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// 等待所有goroutine完成发送
	wg.Wait()

	// 等待所有工具调用完成
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			pending := accumulator.GetPendingCount()
			completed := accumulator.GetCompletedCount()
			t.Fatalf("超时：完成 %d，待完成 %d，总数 %d", completed, pending, numCalls)
		case <-ticker.C:
			completed := accumulator.GetCompletedCount()
			pending := accumulator.GetPendingCount()
			total := accumulator.GetTotalCount()

			t.Logf("进度：完成 %d，待完成 %d，总数 %d", completed, pending, total)

			if completed == numCalls {
				t.Logf("压力测试通过：%d 个工具调用全部完成", numCalls)
				return
			}
		}
	}
}

// TestStreamingToolCallsRaceCondition 测试竞态条件
func TestStreamingToolCallsRaceCondition(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 同时读写累积器
	numWriters := 10
	numReaders := 5
	duration := 2 * time.Second

	var wg sync.WaitGroup
	start := make(chan struct{})

	// 启动写入goroutines
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			<-start

			callID := fmt.Sprintf("writer_%d_call", writerID)
			endTime := time.Now().Add(duration)

			for time.Now().Before(endTime) {
				delta := []*response.ToolCall{{
					Id:   callID,
					Type: "function",
					Function: response.ToolFunction{
						Name:      "test_function",
						Arguments: fmt.Sprintf(`{"step":"%d"}`, time.Now().UnixNano()),
					},
				}}
				accumulator.ProcessDelta(delta)
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// 启动读取goroutines
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			<-start

			endTime := time.Now().Add(duration)
			readCount := 0

			for time.Now().Before(endTime) {
				// 并发读取各种状态
				_ = accumulator.GetCompletedToolCalls()
				_ = accumulator.GetPendingToolCalls()
				_ = accumulator.HasPendingToolCalls()
				_ = accumulator.GetTotalCount()
				_ = accumulator.GetPendingCount()
				_ = accumulator.GetCompletedCount()

				readCount++
				time.Sleep(5 * time.Millisecond)
			}

			t.Logf("读取器 %d 完成 %d 次读取", readerID, readCount)
		}(i)
	}

	// 开始测试
	close(start)
	wg.Wait()

	t.Logf("竞态条件测试完成：总计 %d 个工具调用", accumulator.GetTotalCount())
}

// TestStreamingToolCallsMemoryUsage 测试内存使用
func TestStreamingToolCallsMemoryUsage(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 创建大量工具调用然后清理
	cycles := 10
	callsPerCycle := 100

	for cycle := 0; cycle < cycles; cycle++ {
		t.Logf("内存测试周期 %d/%d", cycle+1, cycles)

		// 创建工具调用
		for i := 0; i < callsPerCycle; i++ {
			callID := fmt.Sprintf("cycle_%d_call_%d", cycle, i)
			delta := []*response.ToolCall{{
				Id:   callID,
				Type: "function",
				Function: response.ToolFunction{
					Name:      "memory_test",
					Arguments: fmt.Sprintf(`{"data":"test_data_%d_%d"}`, cycle, i),
				},
			}}
			accumulator.ProcessDelta(delta)
		}

		// 验证创建的工具调用数量
		expectedCount := callsPerCycle
		actualCount := accumulator.GetCompletedCount()
		if actualCount != expectedCount {
			t.Errorf("周期 %d：期望 %d 个完成的工具调用，实际 %d 个",
				cycle, expectedCount, actualCount)
		}

		// 清理完成的工具调用
		accumulator.ClearCompleted()

		// 验证清理效果
		if accumulator.GetTotalCount() != 0 {
			t.Errorf("周期 %d：清理后仍有 %d 个工具调用残留",
				cycle, accumulator.GetTotalCount())
		}
	}

	t.Logf("内存使用测试完成：%d 个周期，每周期 %d 个调用", cycles, callsPerCycle)
}

// TestStreamingToolCallsEdgeCases 测试边界情况
func TestStreamingToolCallsEdgeCases(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 测试1：空Delta数组
	accumulator.ProcessDelta([]*response.ToolCall{})
	if accumulator.GetTotalCount() != 0 {
		t.Error("处理空Delta后应该没有工具调用")
	}

	// 测试2：ID为空的Delta（应该被跳过）
	accumulator.ProcessDelta([]*response.ToolCall{{
		Id:   "",
		Type: "function",
		Function: response.ToolFunction{
			Name:      "test",
			Arguments: `{"test": "value"}`,
		},
	}})
	if accumulator.GetTotalCount() != 0 {
		t.Error("ID为空的Delta应该被跳过")
	}

	// 测试3：非常长的参数
	longValue := make([]byte, 1000)
	for i := range longValue {
		longValue[i] = 'A' // 填充有效字符而不是零字节
	}
	longArgs := `{"long_param": "` + string(longValue) + `"}`
	accumulator.ProcessDelta([]*response.ToolCall{{
		Id:   "long_call",
		Type: "function",
		Function: response.ToolFunction{
			Name:      "long_test",
			Arguments: longArgs,
		},
	}})
	if accumulator.GetCompletedCount() != 1 {
		t.Error("应该能处理长参数")
	}

	// 测试4：重复的工具调用ID（会累积参数）
	// 创建新的accumulator避免之前测试的干扰
	newAccumulator := NewStreamingToolCallAccumulator()
	newAccumulator.ProcessDelta([]*response.ToolCall{{
		Id:   "duplicate_call",
		Type: "function",
		Function: response.ToolFunction{
			Name:      "test_func",
			Arguments: `{"param1":`,
		},
	}})

	newAccumulator.ProcessDelta([]*response.ToolCall{{
		Id:   "duplicate_call",
		Type: "",
		Function: response.ToolFunction{
			Name:      "",
			Arguments: ` "value1"}`,
		},
	}})

	completed := newAccumulator.GetCompletedToolCalls()
	if len(completed) != 1 {
		t.Errorf("重复ID应该累积成一个完整的工具调用，实际有 %d 个", len(completed))
	}

	if len(completed) > 0 && completed[0].Function.Parameters != `{"param1": "value1"}` {
		t.Errorf("累积的参数不正确，期望 %s，实际 %s",
			`{"param1": "value1"}`, completed[0].Function.Parameters)
	}

	// 清理
	accumulator.ClearCompleted()
}

// TestStreamingToolCallsTimeout 测试超时场景
func TestStreamingToolCallsTimeout(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 创建一个不完整的工具调用
	accumulator.ProcessDelta([]*response.ToolCall{{
		Id:   "timeout_call",
		Type: "function",
		Function: response.ToolFunction{
			Name:      "timeout_test",
			Arguments: `{"incomplete":`,
		},
	}})

	// 验证工具调用未完成
	if accumulator.GetCompletedCount() != 0 {
		t.Error("不完整的工具调用不应该标记为完成")
	}

	if accumulator.GetPendingCount() != 1 {
		t.Error("应该有一个待完成的工具调用")
	}

	// 强制完成工具调用
	forcedCall := accumulator.ForceCompleteToolCall("timeout_call")
	if forcedCall == nil {
		t.Error("强制完成应该返回工具调用对象")
	}

	if forcedCall.Function.Parameters != `{"incomplete":` {
		t.Error("强制完成的工具调用参数不正确")
	}

	// 验证状态更新
	if accumulator.GetCompletedCount() != 1 {
		t.Error("强制完成后应该有一个完成的工具调用")
	}
}

// TestStreamingToolCallsDebugInfo 测试调试信息
func TestStreamingToolCallsDebugInfo(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 创建多个不同状态的工具调用
	testCases := []struct {
		id   string
		args string
	}{
		{"debug_call_1", `{"param1": "value1"`},
		{"debug_call_2", `{"param2": "value2"}`},
		{"debug_call_3", `{"param3":`},
	}

	for _, tc := range testCases {
		accumulator.ProcessDelta([]*response.ToolCall{{
			Id:   tc.id,
			Type: "function",
			Function: response.ToolFunction{
				Name:      "debug_test",
				Arguments: tc.args,
			},
		}})
	}

	// 获取调试信息
	debugInfo := accumulator.GetPendingToolCallsDebugInfo()

	// 验证调试信息
	expectedPending := 2 // debug_call_1 和 debug_call_3 应该是待完成的
	if len(debugInfo) != expectedPending {
		t.Errorf("期望 %d 个待完成的工具调用，实际 %d 个", expectedPending, len(debugInfo))
	}

	// 验证调试信息包含必要的字段
	for id, info := range debugInfo {
		if info == "" {
			t.Errorf("工具调用 %s 的调试信息为空", id)
		}

		// 调试信息应该包含函数名、类型、参数长度等
		expectedFields := []string{"Function:", "Type:", "Args Length:", "Is Valid JSON:"}
		for _, field := range expectedFields {
			if !contains(info, field) {
				t.Errorf("工具调用 %s 的调试信息缺少字段 %s", id, field)
			}
		}

		t.Logf("调试信息 %s: %s", id, info)
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkStreamingToolCallAccumulator 性能基准测试
func BenchmarkStreamingToolCallAccumulator(b *testing.B) {
	accumulator := NewStreamingToolCallAccumulator()

	// 准备测试数据
	deltas := make([][]*response.ToolCall, b.N)
	for i := 0; i < b.N; i++ {
		deltas[i] = []*response.ToolCall{{
			Id:   fmt.Sprintf("bench_call_%d", i),
			Type: "function",
			Function: response.ToolFunction{
				Name:      "benchmark_test",
				Arguments: fmt.Sprintf(`{"bench_param_%d": "value_%d"}`, i, i),
			},
		}}
	}

	b.ResetTimer()

	// 执行基准测试
	for i := 0; i < b.N; i++ {
		accumulator.ProcessDelta(deltas[i])
	}

	b.StopTimer()

	// 验证结果
	if accumulator.GetCompletedCount() != b.N {
		b.Errorf("期望 %d 个完成的工具调用，实际 %d 个", b.N, accumulator.GetCompletedCount())
	}
}

// BenchmarkStreamingToolCallAccumulatorConcurrent 并发性能基准测试
func BenchmarkStreamingToolCallAccumulatorConcurrent(b *testing.B) {
	accumulator := NewStreamingToolCallAccumulator()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			delta := []*response.ToolCall{{
				Id:   fmt.Sprintf("concurrent_call_%d", i),
				Type: "function",
				Function: response.ToolFunction{
					Name:      "concurrent_test",
					Arguments: fmt.Sprintf(`{"concurrent_param_%d": "value_%d"}`, i, i),
				},
			}}
			accumulator.ProcessDelta(delta)
			i++
		}
	})
}
