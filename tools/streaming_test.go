package tools

import (
	"testing"
	"time"

	"github.com/yu1ec/go-anyllm/response"
	"github.com/yu1ec/go-anyllm/types"
)

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"{", false},
		{`{"location"`, false},
		{`{"location": "北京"`, false},
		{`{"location": "北京"}`, true},
		{`{"location": "北京", "unit": "celsius"}`, true},
		{"invalid", false},
		{`[1, 2, 3]`, true},
		{`"string"`, true},
		{`123`, true},
		{`true`, true},
	}

	for _, test := range tests {
		result := IsValidJSON(test.input)
		if result != test.expected {
			t.Errorf("IsValidJSON(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestParseToolCallArgumentsSafe(t *testing.T) {
	type TestParams struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}

	tests := []struct {
		name           string
		toolCall       types.ToolCall
		expectedParams TestParams
		expectedOK     bool
		expectError    bool
	}{
		{
			name: "完整JSON",
			toolCall: types.ToolCall{
				ID:   "call_1",
				Type: "function",
				Function: types.ResponseToolFunction{
					Name:      "get_weather",
					Arguments: `{"locationq": "北京", "unit": "celsius"}`,
				},
			},
			expectedParams: TestParams{Location: "北京", Unit: "celsius"},
			expectedOK:     true,
			expectError:    false,
		},
		{
			name: "不完整JSON",
			toolCall: types.ToolCall{
				ID:   "call_2",
				Type: "function",
				Function: types.ResponseToolFunction{
					Name:      "get_weather",
					Arguments: `{"location": "北京"`,
				},
			},
			expectedParams: TestParams{},
			expectedOK:     false,
			expectError:    false,
		},
		{
			name: "空参数",
			toolCall: types.ToolCall{
				ID:   "call_3",
				Type: "function",
				Function: types.ResponseToolFunction{
					Name:      "get_weather",
					Arguments: "",
				},
			},
			expectedParams: TestParams{},
			expectedOK:     false,
			expectError:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params, ok, err := ParseToolCallArgumentsSafe[TestParams](test.toolCall)

			if test.expectError && err == nil {
				t.Errorf("期望错误但没有发生")
			}
			if !test.expectError && err != nil {
				t.Errorf("不期望错误但发生了: %v", err)
			}
			if ok != test.expectedOK {
				t.Errorf("期望 ok=%v, 得到 %v", test.expectedOK, ok)
			}
			if test.expectedOK && params != test.expectedParams {
				t.Errorf("期望参数 %+v, 得到 %+v", test.expectedParams, params)
			}
		})
	}
}

func TestStreamingToolCallAccumulator(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 模拟分块传输的工具调用
	deltas := [][]*response.ToolCall{
		// 第一块：开始工具调用
		{
			{
				Id:   "call_1",
				Type: "function",
				Function: response.ToolFunction{
					Name:      "get_weather",
					Arguments: `{"location": "北`,
				},
			},
		},
		// 第二块：继续参数
		{
			{
				Id:   "call_1",
				Type: "",
				Function: response.ToolFunction{
					Name:      "",
					Arguments: `京", "unit": "cel`,
				},
			},
		},
		// 第三块：完成参数
		{
			{
				Id:   "call_1",
				Type: "",
				Function: response.ToolFunction{
					Name:      "",
					Arguments: `sius"}`,
				},
			},
		},
	}

	// 处理第一块，应该不完整
	accumulator.ProcessDelta(deltas[0])
	completed := accumulator.GetCompletedToolCalls()
	if len(completed) != 0 {
		t.Errorf("第一块后不应该有完成的工具调用，但得到 %d 个", len(completed))
	}

	// 处理第二块，仍然不完整
	accumulator.ProcessDelta(deltas[1])
	completed = accumulator.GetCompletedToolCalls()
	if len(completed) != 0 {
		t.Errorf("第二块后不应该有完成的工具调用，但得到 %d 个", len(completed))
	}

	// 处理第三块，应该完整了
	accumulator.ProcessDelta(deltas[2])
	completed = accumulator.GetCompletedToolCalls()
	if len(completed) != 1 {
		t.Errorf("第三块后应该有 1 个完成的工具调用，但得到 %d 个", len(completed))
	}

	if len(completed) > 0 {
		call := completed[0]
		if call.ID != "call_1" {
			t.Errorf("期望工具调用ID为 'call_1', 得到 '%s'", call.ID)
		}
		if call.Function.Name != "get_weather" {
			t.Errorf("期望函数名为 'get_weather', 得到 '%s'", call.Function.Name)
		}
		expectedParams := `{"location": "北京", "unit": "celsius"}`
		if call.Function.Arguments != expectedParams {
			t.Errorf("期望参数为 %s, 得到 %s", expectedParams, call.Function.Arguments)
		}
	}
}

func TestStreamingToolCallAccumulatorMultiple(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 模拟两个并发的工具调用
	deltas := []*response.ToolCall{
		// 第一个工具调用开始
		{
			Id:   "call_1",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "get_weather",
				Arguments: `{"location":`,
			},
		},
		// 第二个工具调用开始
		{
			Id:   "call_2",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "calculator",
				Arguments: `{"expression":`,
			},
		},
		// 第一个工具调用继续
		{
			Id:   "call_1",
			Type: "",
			Function: response.ToolFunction{
				Name:      "",
				Arguments: ` "北京"}`,
			},
		},
		// 第二个工具调用完成
		{
			Id:   "call_2",
			Type: "",
			Function: response.ToolFunction{
				Name:      "",
				Arguments: ` "2+3"}`,
			},
		},
	}

	// 处理前两个delta
	accumulator.ProcessDelta([]*response.ToolCall{deltas[0], deltas[1]})
	completed := accumulator.GetCompletedToolCalls()
	if len(completed) != 0 {
		t.Errorf("前两个delta后不应该有完成的工具调用")
	}

	// 处理第三个delta
	accumulator.ProcessDelta([]*response.ToolCall{deltas[2]})
	completed = accumulator.GetCompletedToolCalls()
	if len(completed) != 1 {
		t.Errorf("第三个delta后应该有 1 个完成的工具调用，但得到 %d 个", len(completed))
	}

	// 处理第四个delta
	accumulator.ProcessDelta([]*response.ToolCall{deltas[3]})
	completed = accumulator.GetCompletedToolCalls()
	if len(completed) != 2 {
		t.Errorf("第四个delta后应该有 2 个完成的工具调用，但得到 %d 个", len(completed))
	}

	// 验证两个工具调用都正确完成
	callMap := make(map[string]types.ToolCall)
	for _, call := range completed {
		callMap[call.ID] = call
	}

	if call, exists := callMap["call_1"]; exists {
		expectedParams := `{"location": "北京"}`
		if call.Function.Arguments != expectedParams {
			t.Errorf("call_1 期望参数 %s, 得到 %s", expectedParams, call.Function.Arguments)
		}
	} else {
		t.Errorf("缺少 call_1")
	}

	if call, exists := callMap["call_2"]; exists {
		expectedParams := `{"expression": "2+3"}`
		if call.Function.Arguments != expectedParams {
			t.Errorf("call_2 期望参数 %s, 得到 %s", expectedParams, call.Function.Arguments)
		}
	} else {
		t.Errorf("缺少 call_2")
	}
}

func TestStreamingToolCallAccumulatorClearCompleted(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 添加一个完整的工具调用
	delta := []*response.ToolCall{
		{
			Id:   "call_1",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "get_weather",
				Arguments: `{"location": "北京"}`,
			},
		},
	}

	accumulator.ProcessDelta(delta)

	// 验证有完成的工具调用
	completed := accumulator.GetCompletedToolCalls()
	if len(completed) != 1 {
		t.Errorf("应该有 1 个完成的工具调用")
	}

	// 清除完成的工具调用
	accumulator.ClearCompleted()

	// 验证已清除
	completed = accumulator.GetCompletedToolCalls()
	if len(completed) != 0 {
		t.Errorf("清除后不应该有完成的工具调用")
	}

	// 验证待完成的也被清除
	pending := accumulator.GetPendingToolCalls()
	if len(pending) != 0 {
		t.Errorf("清除后不应该有待完成的工具调用")
	}
}

func TestStreamingToolCallAccumulatorTimeout(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 添加一个不完整的工具调用
	delta := []*response.ToolCall{
		{
			Id:   "call_1",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "get_weather",
				Arguments: `{"location":`,
			},
		},
	}

	accumulator.ProcessDelta(delta)

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 检查待完成的工具调用
	pending := accumulator.GetPendingToolCalls()
	if len(pending) != 1 {
		t.Errorf("应该有 1 个待完成的工具调用")
	}

	if args, exists := pending["call_1"]; exists {
		expected := `{"location":`
		if args != expected {
			t.Errorf("期望待完成参数 %s, 得到 %s", expected, args)
		}
	} else {
		t.Errorf("缺少待完成的工具调用 call_1")
	}
}

func TestHasPendingToolCalls(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 初始状态：没有待完成的工具调用
	if accumulator.HasPendingToolCalls() {
		t.Errorf("初始状态不应该有待完成的工具调用")
	}

	// 添加一个不完整的工具调用
	delta := []*response.ToolCall{
		{
			Id:   "call_1",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "get_weather",
				Arguments: `{"location":`,
			},
		},
	}

	accumulator.ProcessDelta(delta)

	// 现在应该有待完成的工具调用
	if !accumulator.HasPendingToolCalls() {
		t.Errorf("添加不完整工具调用后应该有待完成的工具调用")
	}

	// 完成工具调用
	completeDelta := []*response.ToolCall{
		{
			Id:   "call_1",
			Type: "",
			Function: response.ToolFunction{
				Name:      "",
				Arguments: ` "北京"}`,
			},
		},
	}

	accumulator.ProcessDelta(completeDelta)

	// 现在不应该有待完成的工具调用
	if accumulator.HasPendingToolCalls() {
		t.Errorf("完成工具调用后不应该有待完成的工具调用")
	}
}

func TestToolCallCounts(t *testing.T) {
	accumulator := NewStreamingToolCallAccumulator()

	// 初始状态：所有计数都应该为0
	if accumulator.GetTotalCount() != 0 {
		t.Errorf("初始总数应该为0，得到 %d", accumulator.GetTotalCount())
	}
	if accumulator.GetPendingCount() != 0 {
		t.Errorf("初始待完成数应该为0，得到 %d", accumulator.GetPendingCount())
	}
	if accumulator.GetCompletedCount() != 0 {
		t.Errorf("初始已完成数应该为0，得到 %d", accumulator.GetCompletedCount())
	}

	// 添加两个工具调用：一个完整，一个不完整
	deltas := []*response.ToolCall{
		{
			Id:   "call_1",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "get_weather",
				Arguments: `{"location": "北京"}`, // 完整
			},
		},
		{
			Id:   "call_2",
			Type: "function",
			Function: response.ToolFunction{
				Name:      "calculator",
				Arguments: `{"expression":`, // 不完整
			},
		},
	}

	accumulator.ProcessDelta(deltas)

	// 验证计数
	if accumulator.GetTotalCount() != 2 {
		t.Errorf("总数应该为2，得到 %d", accumulator.GetTotalCount())
	}
	if accumulator.GetCompletedCount() != 1 {
		t.Errorf("已完成数应该为1，得到 %d", accumulator.GetCompletedCount())
	}
	if accumulator.GetPendingCount() != 1 {
		t.Errorf("待完成数应该为1，得到 %d", accumulator.GetPendingCount())
	}
	if !accumulator.HasPendingToolCalls() {
		t.Errorf("应该有待完成的工具调用")
	}

	// 完成第二个工具调用
	completeDelta := []*response.ToolCall{
		{
			Id:   "call_2",
			Type: "",
			Function: response.ToolFunction{
				Name:      "",
				Arguments: ` "2+3"}`,
			},
		},
	}

	accumulator.ProcessDelta(completeDelta)

	// 验证最终计数
	if accumulator.GetTotalCount() != 2 {
		t.Errorf("总数应该为2，得到 %d", accumulator.GetTotalCount())
	}
	if accumulator.GetCompletedCount() != 2 {
		t.Errorf("已完成数应该为2，得到 %d", accumulator.GetCompletedCount())
	}
	if accumulator.GetPendingCount() != 0 {
		t.Errorf("待完成数应该为0，得到 %d", accumulator.GetPendingCount())
	}
	if accumulator.HasPendingToolCalls() {
		t.Errorf("不应该有待完成的工具调用")
	}
}
