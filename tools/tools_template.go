package tools

import "github.com/yu1ec/go-anyllm/types"

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

// CustomUserProfileTool 自定义用户档案查询工具 - 用于测试复杂参数
func CustomUserProfileTool() types.Tool {
	return NewTool("query_user_profile", "查询用户档案信息，支持多种查询条件和选项").
		AddStringParam("user_id", "用户ID（必需）", true).
		AddStringParam("query_type", "查询类型", true, "basic", "detailed", "full").
		AddStringParam("fields", "需要返回的字段，用逗号分隔", false).
		AddBooleanParam("include_history", "是否包含历史记录", false).
		AddIntegerParam("max_records", "最大记录数量", false).
		AddStringParam("date_range", "日期范围", false).
		AddObjectParam("filters", "高级过滤条件", map[string]*PropertyDefinition{
			"status": {
				Type:        TypeString,
				Description: "用户状态过滤",
				Enum:        []interface{}{"active", "inactive", "pending"},
			},
			"department": {
				Type:        TypeString,
				Description: "部门过滤",
			},
			"priority": {
				Type:        TypeInteger,
				Description: "优先级（1-5）",
			},
		}, []string{}, false).
		BuildForTypes()
}
