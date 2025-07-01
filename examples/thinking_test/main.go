package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yu1ec/go-anyllm/providers/alicloud"
	"github.com/yu1ec/go-anyllm/types"
)

func main() {
	// 获取API Key
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 DASHSCOPE_API_KEY 环境变量")
	}

	// 创建阿里云配置
	config := &alicloud.AliCloudConfig{
		APIKey: apiKey,

		// 可选：自定义超时配置
		ThinkingTimeout: 600, // 思考阶段超时：10分钟
		OutputTimeout:   120, // 输出阶段超时：2分钟
		ReadTimeout:     45,  // 单次读取超时：45秒
	}

	// 创建阿里云提供商
	provider, err := alicloud.NewAliCloudProvider(config)
	if err != nil {
		log.Fatalf("创建提供商失败: %v", err)
	}

	// 创建请求
	req := &types.ChatCompletionRequest{
		Model: "qwen-plus",
		Messages: []types.ChatCompletionMessage{
			{Role: "user", Content: "简单介绍一下你自己"},
		},
		EnableThinking: types.ToPtr(true), // 开启思考模式
		Stream:         false,             // 非流式调用
	}

	fmt.Println("测试思考模式非流式调用...")
	fmt.Printf("模型: %s\n", req.Model)
	fmt.Printf("思考模式: %v\n", req.EnableThinking != nil && *req.EnableThinking)
	fmt.Printf("流式调用: %v\n", req.Stream)
	fmt.Println("---")

	// 发送请求
	resp, err := provider.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	// 打印响应
	fmt.Println("请求成功！")
	fmt.Printf("响应ID: %s\n", resp.ID)
	fmt.Printf("模型: %s\n", resp.Model)

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		fmt.Printf("完成原因: %s\n", choice.FinishReason)
		fmt.Printf("响应内容: %s\n", choice.Message.Content)

		if choice.Message.ReasoningContent != "" {
			fmt.Printf("思考内容: %s\n", choice.Message.ReasoningContent)
		}
	}

	if resp.Usage != nil {
		fmt.Printf("Token使用: 输入=%d, 输出=%d, 总计=%d\n",
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			resp.Usage.TotalTokens)
	}

	fmt.Println("\n✅ 测试通过！思考模式在非流式调用中工作正常")
}
