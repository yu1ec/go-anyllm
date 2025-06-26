package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	deepseek "github.com/yu1ec/go-anyllm"
	"github.com/yu1ec/go-anyllm/providers"
	"github.com/yu1ec/go-anyllm/tools"
	"github.com/yu1ec/go-anyllm/types"
)

// WeatherHandler 天气查询处理器
type WeatherHandler struct{}

func (h *WeatherHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	// 解析参数
	type WeatherParams struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}

	params, err := tools.ParseToolCallArguments[WeatherParams](toolCall)
	if err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 模拟天气查询
	temperature := 25
	if params.Unit == "fahrenheit" {
		temperature = int(float64(temperature)*9/5 + 32)
	}

	unit := "°C"
	if params.Unit == "fahrenheit" {
		unit = "°F"
	}

	return fmt.Sprintf("%s 当前天气：晴天，温度 %d%s，湿度 60%%", params.Location, temperature, unit), nil
}

// CalculatorHandler 计算器处理器
type CalculatorHandler struct{}

func (h *CalculatorHandler) HandleToolCall(toolCall types.ToolCall) (string, error) {
	// 解析参数
	type CalculatorParams struct {
		Expression string `json:"expression"`
	}

	params, err := tools.ParseToolCallArguments[CalculatorParams](toolCall)
	if err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 简单的计算器实现（仅支持基本运算）
	result, err := evaluateExpression(params.Expression)
	if err != nil {
		return "", fmt.Errorf("计算失败: %w", err)
	}

	return fmt.Sprintf("计算结果：%s = %g", params.Expression, result), nil
}

// 简单的表达式计算器（仅支持 +, -, *, /）
func evaluateExpression(expr string) (float64, error) {
	// 移除空格
	expr = strings.ReplaceAll(expr, " ", "")

	// 简单的两数运算解析
	operators := []string{"+", "-", "*", "/"}
	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.Split(expr, op)
			if len(parts) == 2 {
				a, err := strconv.ParseFloat(parts[0], 64)
				if err != nil {
					return 0, err
				}
				b, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					return 0, err
				}

				switch op {
				case "+":
					return a + b, nil
				case "-":
					return a - b, nil
				case "*":
					return a * b, nil
				case "/":
					if b == 0 {
						return 0, fmt.Errorf("除零错误")
					}
					return a / b, nil
				}
			}
		}
	}

	// 如果没有运算符，尝试解析为数字
	return strconv.ParseFloat(expr, 64)
}

func main() {
	// 1. 创建客户端
	config := &deepseek.ClientConfig{
		Provider: providers.ProviderDeepSeek,
		APIKey:   "your-api-key-here", // 请设置真实的API密钥
		BaseURL:  "https://api.deepseek.com",
		Timeout:  120,
	}

	client, err := deepseek.NewUnifiedClient(config)
	if err != nil {
		log.Fatal("创建客户端失败:", err)
	}

	// 2. 创建工具注册表
	registry := tools.NewFunctionRegistry()

	// 3. 注册工具处理器
	registry.Register("get_weather", &WeatherHandler{})
	registry.Register("calculator", &CalculatorHandler{})

	// 4. 定义工具
	weatherTool := tools.GetWeatherTool()
	calculatorTool := tools.CalculatorTool()

	// 5. 创建带工具的聊天请求
	req := &types.ChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []types.ChatCompletionMessage{
			{
				Role:    types.RoleSystem,
				Content: "你是一个有用的助手，可以查询天气和进行计算。",
			},
			{
				Role:    types.RoleUser,
				Content: "请查询北京的天气，然后计算 15 + 27 的结果。",
			},
		},
		Tools:      []types.Tool{weatherTool, calculatorTool},
		ToolChoice: tools.Choice.Auto(), // 自动选择合适的工具
		Stream:     false,
	}

	fmt.Println("发送请求...")

	// 6. 发送请求
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatal("请求失败:", err)
	}

	// 7. 处理响应
	if len(resp.Choices) == 0 {
		log.Fatal("没有返回任何选择")
	}

	choice := resp.Choices[0]
	fmt.Printf("完成原因: %s\n", choice.FinishReason)

	if choice.Message != nil {
		fmt.Printf("助手回复: %s\n", choice.Message.Content)

		// 8. 处理工具调用
		if len(choice.Message.ToolCalls) > 0 {
			fmt.Printf("\n检测到 %d 个工具调用:\n", len(choice.Message.ToolCalls))

			// 处理每个工具调用
			var toolMessages []types.ChatCompletionMessage
			for i, toolCall := range choice.Message.ToolCalls {
				fmt.Printf("\n工具调用 %d:\n", i+1)
				fmt.Printf("  ID: %s\n", toolCall.ID)
				fmt.Printf("  函数: %s\n", toolCall.Function.Name)
				fmt.Printf("  参数: %v\n", toolCall.Function.Parameters)

				// 使用注册表处理工具调用
				result := registry.Handle(toolCall)
				fmt.Printf("  结果: %s\n", result.Content)
				if result.Error != "" {
					fmt.Printf("  错误: %s\n", result.Error)
				}

				// 添加工具调用结果到消息历史
				toolMessages = append(toolMessages, result.ToToolMessage())
			}

			// 9. 发送工具调用结果，获取最终回复
			if len(toolMessages) > 0 {
				fmt.Println("\n发送工具调用结果，获取最终回复...")

				// 将工具调用添加到消息历史
				finalReq := &types.ChatCompletionRequest{
					Model: "deepseek-chat",
					Messages: append(req.Messages, []types.ChatCompletionMessage{
						*choice.Message, // 助手的工具调用消息
					}...),
					Stream: false,
				}

				// 添加工具调用结果
				finalReq.Messages = append(finalReq.Messages, toolMessages...)

				// 发送最终请求
				finalResp, err := client.CreateChatCompletion(context.Background(), finalReq)
				if err != nil {
					log.Printf("获取最终回复失败: %v", err)
				} else if len(finalResp.Choices) > 0 && finalResp.Choices[0].Message != nil {
					fmt.Printf("\n最终回复: %s\n", finalResp.Choices[0].Message.Content)
				}
			}
		}
	}

	fmt.Println("\n=== 工具使用示例 ===")
	demonstrateToolBuilder()
	demonstrateToolChoice()
}

// demonstrateToolBuilder 演示工具构建器的使用
func demonstrateToolBuilder() {
	fmt.Println("\n1. 使用工具构建器创建自定义工具:")

	// 创建复杂的工具
	complexTool := tools.NewTool("complex_calculation", "执行复杂的数学计算").
		AddStringParam("operation", "运算类型", true, "sin", "cos", "sqrt", "log").
		AddNumberParam("value", "计算值", true).
		AddBooleanParam("degrees", "是否使用度数（仅适用于三角函数）", false).
		BuildForTypes()

	fmt.Printf("工具名称: %s\n", complexTool.Function.Name)
	fmt.Printf("工具描述: %s\n", complexTool.Function.Description)
	fmt.Printf("工具类型: %s\n", complexTool.Type)

	// 创建带数组参数的工具
	arrayTool := tools.NewTool("process_list", "处理数字列表").
		AddArrayParam("numbers", "数字列表", "number", true).
		AddStringParam("operation", "对列表执行的操作", true, "sum", "average", "max", "min").
		BuildForTypes()

	fmt.Printf("\n数组工具: %s\n", arrayTool.Function.Name)
}

// demonstrateToolChoice 演示工具选择的使用
func demonstrateToolChoice() {
	fmt.Println("\n2. 工具选择示例:")

	// 不同的工具选择方式
	choices := map[string]interface{}{
		"自动选择":   tools.Choice.Auto(),
		"不使用工具":  tools.Choice.None(),
		"必须使用工具": tools.Choice.Required(),
		"指定函数":   tools.Choice.Function("get_weather"),
	}

	for name, choice := range choices {
		fmt.Printf("%s: %v\n", name, choice)
	}

	// 结构体方式的工具选择
	structChoice := tools.Choice.FunctionStruct("calculator")
	fmt.Printf("结构体选择: Type=%s, Function.Name=%s\n",
		structChoice.Type, structChoice.Function.Name)
}

// 使用预设工具的简单示例
func simpleToolExample() {
	// 获取预设的工具
	weatherTool := tools.GetWeatherTool()
	calculatorTool := tools.CalculatorTool()
	searchTool := tools.SearchTool()
	emailTool := tools.SendEmailTool()
	fileTool := tools.FileOperationTool()

	allTools := []types.Tool{
		weatherTool,
		calculatorTool,
		searchTool,
		emailTool,
		fileTool,
	}

	fmt.Printf("可用工具数量: %d\n", len(allTools))
	for i, tool := range allTools {
		fmt.Printf("%d. %s: %s\n", i+1, tool.Function.Name, tool.Function.Description)
	}
}
