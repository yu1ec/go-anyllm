package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yu1ec/go-anyllm/request"
	"github.com/yu1ec/go-anyllm/response"
	"github.com/yu1ec/go-anyllm/types"
)

// 常用的JSON Schema类型常量
const (
	TypeString  = "string"
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeBoolean = "boolean"
	TypeArray   = "array"
	TypeObject  = "object"
)

// PropertyDefinition JSON Schema属性定义
type PropertyDefinition struct {
	Type        string                         `json:"type"`
	Description string                         `json:"description,omitempty"`
	Enum        []interface{}                  `json:"enum,omitempty"`
	Items       *PropertyDefinition            `json:"items,omitempty"`
	Properties  map[string]*PropertyDefinition `json:"properties,omitempty"`
	Required    []string                       `json:"required,omitempty"`
	Default     interface{}                    `json:"default,omitempty"`
}

// FunctionSchema 函数参数的JSON Schema定义
type FunctionSchema struct {
	Type       string                         `json:"type"`
	Properties map[string]*PropertyDefinition `json:"properties"`
	Required   []string                       `json:"required"`
}

// ToolBuilder 工具构建器
type ToolBuilder struct {
	tool *request.Tool
}

// NewTool 创建新的工具构建器
func NewTool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: &request.Tool{
			Type: "function",
			Function: &request.ToolFunction{
				Name:        name,
				Description: description,
				Parameters: &FunctionSchema{
					Type:       "object",
					Properties: make(map[string]*PropertyDefinition),
					Required:   []string{},
				},
			},
		},
	}
}

// AddStringParam 添加字符串参数
func (tb *ToolBuilder) AddStringParam(name, description string, required bool, enum ...string) *ToolBuilder {
	prop := &PropertyDefinition{
		Type:        TypeString,
		Description: description,
	}
	if len(enum) > 0 {
		prop.Enum = make([]interface{}, len(enum))
		for i, v := range enum {
			prop.Enum[i] = v
		}
	}

	schema := tb.tool.Function.Parameters.(*FunctionSchema)
	schema.Properties[name] = prop

	if required {
		schema.Required = append(schema.Required, name)
	}
	return tb
}

// AddNumberParam 添加数字参数
func (tb *ToolBuilder) AddNumberParam(name, description string, required bool) *ToolBuilder {
	prop := &PropertyDefinition{
		Type:        TypeNumber,
		Description: description,
	}

	schema := tb.tool.Function.Parameters.(*FunctionSchema)
	schema.Properties[name] = prop

	if required {
		schema.Required = append(schema.Required, name)
	}
	return tb
}

// AddIntegerParam 添加整数参数
func (tb *ToolBuilder) AddIntegerParam(name, description string, required bool) *ToolBuilder {
	prop := &PropertyDefinition{
		Type:        TypeInteger,
		Description: description,
	}

	schema := tb.tool.Function.Parameters.(*FunctionSchema)
	schema.Properties[name] = prop

	if required {
		schema.Required = append(schema.Required, name)
	}
	return tb
}

// AddBooleanParam 添加布尔参数
func (tb *ToolBuilder) AddBooleanParam(name, description string, required bool) *ToolBuilder {
	prop := &PropertyDefinition{
		Type:        TypeBoolean,
		Description: description,
	}

	schema := tb.tool.Function.Parameters.(*FunctionSchema)
	schema.Properties[name] = prop

	if required {
		schema.Required = append(schema.Required, name)
	}
	return tb
}

// AddArrayParam 添加数组参数
func (tb *ToolBuilder) AddArrayParam(name, description string, itemType string, required bool) *ToolBuilder {
	prop := &PropertyDefinition{
		Type:        TypeArray,
		Description: description,
		Items: &PropertyDefinition{
			Type: itemType,
		},
	}

	schema := tb.tool.Function.Parameters.(*FunctionSchema)
	schema.Properties[name] = prop

	if required {
		schema.Required = append(schema.Required, name)
	}
	return tb
}

// AddObjectParam 添加对象参数
func (tb *ToolBuilder) AddObjectParam(name, description string, properties map[string]*PropertyDefinition, requiredFields []string, required bool) *ToolBuilder {
	prop := &PropertyDefinition{
		Type:        TypeObject,
		Description: description,
		Properties:  properties,
		Required:    requiredFields,
	}

	schema := tb.tool.Function.Parameters.(*FunctionSchema)
	schema.Properties[name] = prop

	if required {
		schema.Required = append(schema.Required, name)
	}
	return tb
}

// Build 构建工具
func (tb *ToolBuilder) Build() *request.Tool {
	return tb.tool
}

// BuildForTypes 构建types包的Tool
func (tb *ToolBuilder) BuildForTypes() types.Tool {
	return types.Tool{
		Type: tb.tool.Type,
		Function: types.ToolFunction{
			Name:        tb.tool.Function.Name,
			Description: tb.tool.Function.Description,
			Parameters:  tb.tool.Function.Parameters,
		},
	}
}

// ToolChoice 工具选择辅助结构
type ToolChoice struct{}

// Auto 自动选择工具
func (tc ToolChoice) Auto() interface{} {
	return "auto"
}

// None 不使用任何工具
func (tc ToolChoice) None() interface{} {
	return "none"
}

// Required 必须使用工具
func (tc ToolChoice) Required() interface{} {
	return "required"
}

// Function 指定特定函数
func (tc ToolChoice) Function(functionName string) interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name": functionName,
		},
	}
}

// FunctionStruct 指定特定函数（结构体方式）
func (tc ToolChoice) FunctionStruct(functionName string) request.ToolChoiceNamed {
	return request.ToolChoiceNamed{
		Type: "function",
		Function: request.ToolChoiceFunction{
			Name: functionName,
		},
	}
}

// 全局ToolChoice实例
var Choice = ToolChoice{}

// ToolCallHandler 工具调用处理器接口
type ToolCallHandler interface {
	HandleToolCall(toolCall types.ToolCall) (string, error)
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	Error      string `json:"error,omitempty"`
}

// FunctionRegistry 函数注册表
type FunctionRegistry struct {
	handlers map[string]ToolCallHandler
}

// NewFunctionRegistry 创建新的函数注册表
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		handlers: make(map[string]ToolCallHandler),
	}
}

// Register 注册工具处理器
func (fr *FunctionRegistry) Register(functionName string, handler ToolCallHandler) {
	fr.handlers[functionName] = handler
}

// Handle 处理工具调用
func (fr *FunctionRegistry) Handle(toolCall types.ToolCall) *ToolCallResult {
	handler, exists := fr.handlers[toolCall.Function.Name]
	if !exists {
		return &ToolCallResult{
			ToolCallID: toolCall.ID,
			Error:      fmt.Sprintf("function %s not found", toolCall.Function.Name),
		}
	}

	content, err := handler.HandleToolCall(toolCall)
	result := &ToolCallResult{
		ToolCallID: toolCall.ID,
		Content:    content,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// ToToolMessage 将工具调用结果转换为消息
func (result *ToolCallResult) ToToolMessage() types.ChatCompletionMessage {
	content := result.Content
	if result.Error != "" {
		content = fmt.Sprintf("Error: %s", result.Error)
	}

	return types.ChatCompletionMessage{
		Role:       types.RoleTool,
		Content:    content,
		ToolCallID: result.ToolCallID,
	}
}

// ParseToolCallArguments 解析工具调用参数
func ParseToolCallArguments[T any](toolCall types.ToolCall) (T, error) {
	var result T

	// 尝试解析Parameters字段
	switch params := toolCall.Function.Parameters.(type) {
	case string:
		// 如果是字符串，尝试JSON解析
		if err := json.Unmarshal([]byte(params), &result); err != nil {
			return result, fmt.Errorf("failed to parse tool call arguments from string: %w", err)
		}
	case []byte:
		// 如果是字节数组，尝试JSON解析
		if err := json.Unmarshal(params, &result); err != nil {
			return result, fmt.Errorf("failed to parse tool call arguments from bytes: %w", err)
		}
	default:
		// 如果是其他类型，先序列化再反序列化
		data, err := json.Marshal(params)
		if err != nil {
			return result, fmt.Errorf("failed to marshal tool call parameters: %w", err)
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return result, fmt.Errorf("failed to parse tool call arguments: %w", err)
		}
	}

	return result, nil
}

// ParseToolCallArgumentsSafe 安全解析工具调用参数（支持流式JSON）
func ParseToolCallArgumentsSafe[T any](toolCall types.ToolCall) (T, bool, error) {
	var result T

	// 尝试解析Parameters字段
	switch params := toolCall.Function.Parameters.(type) {
	case string:
		// 检查JSON是否完整
		if !IsValidJSON(params) {
			return result, false, nil // 返回false表示JSON不完整，需要继续累积
		}
		// 如果是字符串，尝试JSON解析
		if err := json.Unmarshal([]byte(params), &result); err != nil {
			return result, false, fmt.Errorf("failed to parse tool call arguments from string: %w", err)
		}
		return result, true, nil
	case []byte:
		// 检查JSON是否完整
		if !IsValidJSON(string(params)) {
			return result, false, nil
		}
		// 如果是字节数组，尝试JSON解析
		if err := json.Unmarshal(params, &result); err != nil {
			return result, false, fmt.Errorf("failed to parse tool call arguments from bytes: %w", err)
		}
		return result, true, nil
	default:
		// 如果是其他类型，先序列化再反序列化
		data, err := json.Marshal(params)
		if err != nil {
			return result, false, fmt.Errorf("failed to marshal tool call parameters: %w", err)
		}
		if !IsValidJSON(string(data)) {
			return result, false, nil
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return result, false, fmt.Errorf("failed to parse tool call arguments: %w", err)
		}
		return result, true, nil
	}
}

// IsValidJSON 检查字符串是否为有效的JSON
func IsValidJSON(s string) bool {
	if s == "" {
		return false
	}

	var js interface{}
	err := json.Unmarshal([]byte(s), &js)
	return err == nil
}

// StreamingToolCallAccumulator 流式工具调用累积器
type StreamingToolCallAccumulator struct {
	toolCalls      map[string]*StreamingToolCall
	lastToolCallID string // 最近处理的工具调用ID，用于处理ID为空的Delta
	mutex          sync.RWMutex
}

// StreamingToolCall 流式工具调用状态
type StreamingToolCall struct {
	ID              string
	Type            string
	FunctionName    string
	ArgumentsBuffer strings.Builder
	IsComplete      bool
	LastUpdateTime  time.Time
}

// NewStreamingToolCallAccumulator 创建新的流式工具调用累积器
func NewStreamingToolCallAccumulator() *StreamingToolCallAccumulator {
	return &StreamingToolCallAccumulator{
		toolCalls: make(map[string]*StreamingToolCall),
	}
}

// ProcessDelta 处理Delta中的工具调用
func (acc *StreamingToolCallAccumulator) ProcessDelta(deltaToolCalls []*response.ToolCall) {
	acc.mutex.Lock()
	defer acc.mutex.Unlock()

	for _, delta := range deltaToolCalls {
		var targetID string

		// 处理ID为空的情况：关联到最近的工具调用
		if delta.Id == "" {
			if acc.lastToolCallID == "" {
				// 如果没有之前的工具调用ID，跳过这个Delta
				continue
			}
			targetID = acc.lastToolCallID
		} else {
			targetID = delta.Id
			acc.lastToolCallID = targetID // 更新最近的工具调用ID
		}

		// 获取或创建工具调用
		if acc.toolCalls[targetID] == nil {
			acc.toolCalls[targetID] = &StreamingToolCall{
				ID:             targetID,
				Type:           delta.Type,
				FunctionName:   delta.Function.Name,
				LastUpdateTime: time.Now(),
			}
		}

		streamingCall := acc.toolCalls[targetID]
		streamingCall.LastUpdateTime = time.Now()

		// 更新类型（如果有）
		if delta.Type != "" {
			streamingCall.Type = delta.Type
		}

		// 更新函数名（如果有）
		if delta.Function.Name != "" {
			streamingCall.FunctionName = delta.Function.Name
		}

		// 累积参数
		if delta.Function.Arguments != "" {
			streamingCall.ArgumentsBuffer.WriteString(delta.Function.Arguments)
		}

		// 检查JSON是否完整
		currentArgs := streamingCall.ArgumentsBuffer.String()
		if IsValidJSON(currentArgs) {
			streamingCall.IsComplete = true
		}
	}
}

// GetCompletedToolCalls 获取已完成的工具调用
func (acc *StreamingToolCallAccumulator) GetCompletedToolCalls() []types.ToolCall {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	var completed []types.ToolCall
	for _, streamingCall := range acc.toolCalls {
		if streamingCall.IsComplete {
			completed = append(completed, types.ToolCall{
				ID:   streamingCall.ID,
				Type: streamingCall.Type,
				Function: types.ToolFunction{
					Name:       streamingCall.FunctionName,
					Parameters: streamingCall.ArgumentsBuffer.String(),
				},
			})
		}
	}

	return completed
}

// GetPendingToolCalls 获取待完成的工具调用（用于调试）
func (acc *StreamingToolCallAccumulator) GetPendingToolCalls() map[string]string {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	pending := make(map[string]string)
	for id, streamingCall := range acc.toolCalls {
		if !streamingCall.IsComplete {
			pending[id] = streamingCall.ArgumentsBuffer.String()
		}
	}

	return pending
}

// ClearCompleted 清除已完成的工具调用
func (acc *StreamingToolCallAccumulator) ClearCompleted() {
	acc.mutex.Lock()
	defer acc.mutex.Unlock()

	for id, streamingCall := range acc.toolCalls {
		if streamingCall.IsComplete {
			delete(acc.toolCalls, id)
		}
	}
}

// HasPendingToolCalls 检查是否有待完成的工具调用
func (acc *StreamingToolCallAccumulator) HasPendingToolCalls() bool {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	for _, streamingCall := range acc.toolCalls {
		if !streamingCall.IsComplete {
			return true
		}
	}
	return false
}

// GetPendingCount 获取待完成的工具调用数量
func (acc *StreamingToolCallAccumulator) GetPendingCount() int {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	count := 0
	for _, streamingCall := range acc.toolCalls {
		if !streamingCall.IsComplete {
			count++
		}
	}
	return count
}

// GetCompletedCount 获取已完成的工具调用数量
func (acc *StreamingToolCallAccumulator) GetCompletedCount() int {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	count := 0
	for _, streamingCall := range acc.toolCalls {
		if streamingCall.IsComplete {
			count++
		}
	}
	return count
}

// GetTotalCount 获取总工具调用数量
func (acc *StreamingToolCallAccumulator) GetTotalCount() int {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	return len(acc.toolCalls)
}

// FinalizeStream 在流结束时强制检查所有累积的工具调用，将有效JSON标记为完成
func (acc *StreamingToolCallAccumulator) FinalizeStream() []types.ToolCall {
	acc.mutex.Lock()
	defer acc.mutex.Unlock()

	// 遍历所有待完成的工具调用
	for _, streamingCall := range acc.toolCalls {
		if !streamingCall.IsComplete {
			currentArgs := streamingCall.ArgumentsBuffer.String()
			// 强制检查JSON是否有效
			if IsValidJSON(currentArgs) {
				streamingCall.IsComplete = true
			}
		}
	}

	// 返回所有已完成的工具调用
	var completed []types.ToolCall
	for _, streamingCall := range acc.toolCalls {
		if streamingCall.IsComplete {
			completed = append(completed, types.ToolCall{
				ID:   streamingCall.ID,
				Type: streamingCall.Type,
				Function: types.ToolFunction{
					Name:       streamingCall.FunctionName,
					Parameters: streamingCall.ArgumentsBuffer.String(),
				},
			})
		}
	}

	return completed
}

// ForceCompleteToolCall 强制完成指定ID的工具调用（用于调试或特殊情况）
func (acc *StreamingToolCallAccumulator) ForceCompleteToolCall(toolCallId string) *types.ToolCall {
	acc.mutex.Lock()
	defer acc.mutex.Unlock()

	// 查找指定ID的工具调用
	streamingCall, exists := acc.toolCalls[toolCallId]
	if !exists {
		return nil
	}

	// 强制标记为完成
	streamingCall.IsComplete = true
	streamingCall.LastUpdateTime = time.Now()

	// 构建并返回工具调用对象
	toolCall := &types.ToolCall{
		ID:   streamingCall.ID,
		Type: streamingCall.Type,
		Function: types.ToolFunction{
			Name:       streamingCall.FunctionName,
			Parameters: streamingCall.ArgumentsBuffer.String(),
		},
	}

	return toolCall
}

// GetPendingToolCallsDebugInfo 返回当前待处理工具调用的详细信息，用于调试
func (acc *StreamingToolCallAccumulator) GetPendingToolCallsDebugInfo() map[string]string {
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()

	debugInfo := make(map[string]string)

	for id, streamingCall := range acc.toolCalls {
		if !streamingCall.IsComplete {
			currentArgs := streamingCall.ArgumentsBuffer.String()
			info := fmt.Sprintf("Function: %s | Type: %s | Args Length: %d | Last Update: %s | Args Content: %s | Is Valid JSON: %t",
				streamingCall.FunctionName,
				streamingCall.Type,
				len(currentArgs),
				streamingCall.LastUpdateTime.Format("15:04:05.000"),
				currentArgs,
				IsValidJSON(currentArgs))
			debugInfo[id] = info
		}
	}

	return debugInfo
}
