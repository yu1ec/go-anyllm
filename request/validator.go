package request

import (
	"errors"
	"fmt"
)

var (
	modelChat     = "deepseek-chat"
	modelReasoner = "deepseek-reasoner"
)

var roles map[string]struct{}
var rolesStr string
var modelStr string
var responseFormatStr string

func init() {
	rolesStr = fmt.Sprintf(`%s, %s, %s, %s`, RoleSystem, RoleUser, RoleAssistant, RoleTool)

	roles = make(map[string]struct{})
	roles[RoleSystem] = struct{}{}
	roles[RoleUser] = struct{}{}
	roles[RoleAssistant] = struct{}{}
	roles[RoleTool] = struct{}{}

	modelStr = fmt.Sprintf(`%s, %s`, modelChat, modelReasoner)
	responseFormatStr = fmt.Sprintf(`%s, %s`, ResponseFormatText, ResponseFormatJsonObject)
}

func ValidateChatCompletionsRequest(req *ChatCompletionsRequest) error {
	if req == nil {
		return errors.New("err: input request is nil")
	}

	if err := validateMessages(req.Messages); err != nil {
		return err
	}

	if err := validateModel(req); err != nil {
		return err
	}

	if err := validateMultipleFields(req); err != nil {
		return err
	}

	if err := validateResponseFormat(req); err != nil {
		return err
	}

	if err := validateStreamOptions(req); err != nil {
		return err
	}

	// 验证工具
	if err := validateTools(req); err != nil {
		return err
	}

	// 验证工具选择
	if err := validateToolChoice(req); err != nil {
		return err
	}

	if req.TopLogprobs != nil {
		if !req.Logprobs {
			return fmt.Errorf(`err: top_logprobs can not be set when "logprobs" is false`)
		}
		if !(*req.TopLogprobs >= 0 && *req.TopLogprobs <= 20) {
			return fmt.Errorf(`err: top_logprobs is invalid; it should be number between 0 and 20`)
		}
	}
	return nil
}

func validateMessages(messages []*Message) error {
	if messages == nil || len(messages) < 1 {
		return errors.New("err: at least one message required in request")
	}
	for ind, msg := range messages {
		if msg.Role == "" {
			return fmt.Errorf("err: role is blank for message at %d index; role must be one of [%s]", ind, rolesStr)
		}
		if _, found := roles[msg.Role]; !found {
			return fmt.Errorf("err: invalid role %q for message at %d index; role must be one of [%s]", msg.Role, ind, rolesStr)
		}
		if msg.Content == "" {
			return fmt.Errorf("err: content is blank for message at %d index", ind)
		}
		if msg.Role == RoleTool && msg.ToolCallId == "" {
			return fmt.Errorf("err: tool_call_id is blank for message at %d index; tool_call_id is mandatory with role %q", ind, msg.Role)
		}
	}
	return nil
}

func validateModel(req *ChatCompletionsRequest) error {
	if req.Model == "" {
		return errors.New("err: model required in request")
	}
	if !(req.Model == modelChat || req.Model == modelReasoner) {
		return fmt.Errorf("err: invalid model %q; model should be one of [%s]", req.Model, modelStr)
	}
	return nil
}

func validateMultipleFields(req *ChatCompletionsRequest) error {
	if !(req.FrequencyPenalty >= -2 && req.FrequencyPenalty <= 2) {
		return fmt.Errorf("err: frequency_penalty is invalid; it should be number between -2 and 2")
	}

	if !(req.MaxTokens == 0 || (req.MaxTokens >= 1 && req.MaxTokens <= 8192)) {
		return fmt.Errorf("err: max_tokens is invalid; it should be number between 1 and 8192 or 0")
	}

	if !(req.PresencePenalty >= -2 && req.PresencePenalty <= 2) {
		return fmt.Errorf("err: presence_penalty is invalid; it should be number between -2 and 2")
	}

	if req.Temperature != nil {
		if !(*req.Temperature >= 0 && *req.Temperature <= 2) {
			return fmt.Errorf("err: temperature is invalid; the valid range of temperature is [0, 2.0]")
		}
	}

	if req.TopP != nil {
		if !(*req.TopP > 0 && *req.TopP <= 1) {
			return fmt.Errorf("err: top_p is invalid; the valid range of top_p is (0, 1.0]")
		}
	}

	return nil
}

func validateResponseFormat(req *ChatCompletionsRequest) error {
	if req.ResponseFormat != nil {
		if !(req.ResponseFormat.Type == ResponseFormatText || req.ResponseFormat.Type == ResponseFormatJsonObject) {
			return fmt.Errorf(`err: invalid response_format type %q; it should be one of [%s]`, req.ResponseFormat.Type, responseFormatStr)
		}
	}
	return nil
}

func validateStreamOptions(req *ChatCompletionsRequest) error {
	if req.StreamOptions != nil {
		if !req.Stream {
			return errors.New(`err: stream_options should be set along with stream = true`)
		}
	}
	return nil
}

// validateTools 验证工具定义
func validateTools(req *ChatCompletionsRequest) error {
	if req.Tools == nil {
		return nil
	}

	for i, tool := range *req.Tools {
		if tool.Type == "" {
			return fmt.Errorf("err: tool type is blank for tool at %d index", i)
		}
		if tool.Type != "function" {
			return fmt.Errorf("err: invalid tool type %q for tool at %d index; type must be 'function'", tool.Type, i)
		}
		if tool.Function == nil {
			return fmt.Errorf("err: function is nil for tool at %d index", i)
		}
		if tool.Function.Name == "" {
			return fmt.Errorf("err: function name is blank for tool at %d index", i)
		}
		// 验证函数名称格式（只允许字母、数字、下划线和连字符，长度1-64）
		if len(tool.Function.Name) > 64 {
			return fmt.Errorf("err: function name too long for tool at %d index; max length is 64", i)
		}
		for _, char := range tool.Function.Name {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || char == '_' || char == '-') {
				return fmt.Errorf("err: invalid character in function name for tool at %d index; only letters, numbers, underscore and hyphen allowed", i)
			}
		}
	}
	return nil
}

// validateToolChoice 验证工具选择
func validateToolChoice(req *ChatCompletionsRequest) error {
	if req.ToolChoice == nil {
		return nil
	}

	// 如果设置了tool_choice，必须同时设置tools
	if req.Tools == nil || len(*req.Tools) == 0 {
		return errors.New("err: tool_choice can only be set when tools are provided")
	}

	// 检查tool_choice的类型
	switch choice := req.ToolChoice.(type) {
	case string:
		// 支持的字符串值："none", "auto", "required"
		if choice != "none" && choice != "auto" && choice != "required" {
			return fmt.Errorf("err: invalid tool_choice string %q; must be one of 'none', 'auto', 'required'", choice)
		}
	case map[string]interface{}:
		// 检查是否为命名工具选择格式
		toolType, hasType := choice["type"]
		if !hasType {
			return errors.New("err: tool_choice object must have 'type' field")
		}
		if toolType != "function" {
			return fmt.Errorf("err: invalid tool_choice type %q; must be 'function'", toolType)
		}

		functionField, hasFunction := choice["function"]
		if !hasFunction {
			return errors.New("err: tool_choice object must have 'function' field when type is 'function'")
		}

		function, ok := functionField.(map[string]interface{})
		if !ok {
			return errors.New("err: tool_choice 'function' field must be an object")
		}

		functionName, hasName := function["name"]
		if !hasName {
			return errors.New("err: tool_choice function must have 'name' field")
		}

		name, ok := functionName.(string)
		if !ok {
			return errors.New("err: tool_choice function name must be a string")
		}

		// 验证指定的函数名是否存在于tools中
		found := false
		for _, tool := range *req.Tools {
			if tool.Function != nil && tool.Function.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("err: tool_choice function %q not found in provided tools", name)
		}
	case ToolChoiceNamed:
		// 处理结构体类型的工具选择
		if choice.Type != "function" {
			return fmt.Errorf("err: invalid tool_choice type %q; must be 'function'", choice.Type)
		}
		if choice.Function.Name == "" {
			return errors.New("err: tool_choice function name cannot be empty")
		}

		// 验证指定的函数名是否存在于tools中
		found := false
		for _, tool := range *req.Tools {
			if tool.Function != nil && tool.Function.Name == choice.Function.Name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("err: tool_choice function %q not found in provided tools", choice.Function.Name)
		}
	default:
		return errors.New("err: tool_choice must be a string, object, or ToolChoiceNamed struct")
	}

	return nil
}
