package types

import (
	"testing"
)

func TestChatCompletionRequest_ThinkingParameters(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "qwen-max",
		Messages: []ChatCompletionMessage{
			{
				Role:    RoleUser,
				Content: "测试问题",
			},
		},
	}

	// 测试设置思考参数
	req.WithEnableThinking(true).WithThinkingBudget(1000)

	// 验证参数设置
	if !req.IsThinkingEnabled() {
		t.Error("期望思考模式为开启状态")
	}

	if req.GetThinkingBudget() != 1000 {
		t.Errorf("期望思考预算为1000，实际为%d", req.GetThinkingBudget())
	}

	// 测试关闭思考
	req.WithEnableThinking(false)
	if req.IsThinkingEnabled() {
		t.Error("期望思考模式为关闭状态")
	}
}

func TestChatCompletionRequest_ThinkingParametersNil(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "qwen-max",
		Messages: []ChatCompletionMessage{
			{
				Role:    RoleUser,
				Content: "测试问题",
			},
		},
	}

	// 测试未设置参数的情况
	if req.IsThinkingEnabled() {
		t.Error("期望思考模式为未设置状态（false）")
	}

	if req.GetThinkingBudget() != 0 {
		t.Errorf("期望思考预算为0，实际为%d", req.GetThinkingBudget())
	}
}

func TestChatCompletionRequest_DirectParameterSetting(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "qwen-max",
		Messages: []ChatCompletionMessage{
			{
				Role:    RoleUser,
				Content: "测试问题",
			},
		},
		EnableThinking: ToPtr(true),
		ThinkingBudget: ToPtr(2000),
	}

	// 验证直接设置的参数
	if !req.IsThinkingEnabled() {
		t.Error("期望思考模式为开启状态")
	}

	if req.GetThinkingBudget() != 2000 {
		t.Errorf("期望思考预算为2000，实际为%d", req.GetThinkingBudget())
	}
}
