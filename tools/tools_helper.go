package tools

import (
	"encoding/json"
	"fmt"

	"github.com/yu1ec/go-anyllm/request"
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

// 常用工具定义的预设模板

// GetWeatherTool 获取天气信息的工具
func GetWeatherTool() types.Tool {
	return NewTool("get_weather", "获取指定地点的天气信息").
		AddStringParam("location", "城市和州，例如：北京, 中国", true).
		AddStringParam("unit", "温度单位", false, "celsius", "fahrenheit").
		BuildForTypes()
}

// CalculatorTool 计算器工具
func CalculatorTool() types.Tool {
	return NewTool("calculator", "执行数学计算").
		AddStringParam("expression", "要计算的数学表达式，例如：2+3*4", true).
		BuildForTypes()
}

// SearchTool 搜索工具
func SearchTool() types.Tool {
	return NewTool("search", "在互联网上搜索信息").
		AddStringParam("query", "搜索查询词", true).
		AddIntegerParam("max_results", "最大结果数量", false).
		BuildForTypes()
}

// SendEmailTool 发送邮件工具
func SendEmailTool() types.Tool {
	return NewTool("send_email", "发送电子邮件").
		AddStringParam("to", "收件人邮箱地址", true).
		AddStringParam("subject", "邮件主题", true).
		AddStringParam("body", "邮件正文", true).
		AddStringParam("cc", "抄送邮箱地址", false).
		BuildForTypes()
}

// FileOperationTool 文件操作工具
func FileOperationTool() types.Tool {
	return NewTool("file_operation", "执行文件操作").
		AddStringParam("operation", "操作类型", true, "read", "write", "delete", "list").
		AddStringParam("path", "文件路径", true).
		AddStringParam("content", "文件内容（仅用于写入操作）", false).
		BuildForTypes()
}
