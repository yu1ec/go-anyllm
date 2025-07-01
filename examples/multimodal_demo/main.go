package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	deepseek "github.com/yu1ec/go-anyllm"
	"github.com/yu1ec/go-anyllm/providers"
	"github.com/yu1ec/go-anyllm/types"
)

func main() {
	// 获取API密钥
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 DASHSCOPE_API_KEY 环境变量")
	}

	// 创建统一客户端
	client, err := deepseek.NewUnifiedClient(&deepseek.ClientConfig{
		Provider: providers.ProviderAliCloud,
		APIKey:   apiKey,
	})
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 示例：纯文本消息（向后兼容）
	fmt.Println("=== 纯文本消息示例 ===")
	textExample(client)

	// 示例：多模态消息（文本 + 图像）
	fmt.Println("\n=== 多模态消息示例 ===")
	multiModalExample(client)

	// 示例：使用辅助函数创建多模态消息
	fmt.Println("\n=== 使用辅助函数的多模态示例 ===")
	helperFunctionExample(client)
}

// 纯文本消息示例
func textExample(client deepseek.UnifiedClient) {
	req := &types.ChatCompletionRequest{
		Model: "qwen-vl-max-latest",
		Messages: []types.ChatCompletionMessage{
			types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
			types.NewTextMessage(types.RoleUser, "请介绍一下Go语言的特点。"),
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Printf("纯文本请求失败: %v", err)
		return
	}

	fmt.Printf("回复: %s\n", resp.Choices[0].Message.GetContentAsString())
}

// 多模态消息示例
func multiModalExample(client deepseek.UnifiedClient) {
	// 读取示例图像并转换为base64（这里使用一个简单的示例）
	// 实际使用时，您需要提供真实的图像数据
	base64Image := getExampleBase64Image()

	req := &types.ChatCompletionRequest{
		Model: "qwen-vl-max-latest",
		Messages: []types.ChatCompletionMessage{
			types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
			{
				Role: types.RoleUser,
				Content: []types.MessageContent{
					{
						Type: types.MessageContentTypeImageURL,
						ImageURL: &types.ImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
						},
					},
					{
						Type: types.MessageContentTypeText,
						Text: "图中描绘的是什么景象?",
					},
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Printf("多模态请求失败: %v", err)
		return
	}

	fmt.Printf("回复: %s\n", resp.Choices[0].Message.GetContentAsString())
}

// 使用辅助函数的多模态示例
func helperFunctionExample(client deepseek.UnifiedClient) {
	base64Image := getExampleBase64Image()

	// 使用辅助函数创建多模态内容
	contents := []types.MessageContent{
		types.NewImageContent(fmt.Sprintf("data:image/png;base64,%s", base64Image), types.ImageDetailAuto),
		types.NewTextContent("请分析这张图片的内容，并详细描述你看到了什么。"),
	}

	req := &types.ChatCompletionRequest{
		Model: "qwen-vl-max-latest",
		Messages: []types.ChatCompletionMessage{
			types.NewTextMessage(types.RoleSystem, "You are a helpful assistant specializing in image analysis."),
			types.NewMultiModalMessage(types.RoleUser, contents),
		},
	}

	// 使用流式响应
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Printf("流式多模态请求失败: %v", err)
		return
	}

	fmt.Print("流式回复: ")
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fmt.Print(content)
			}
		}
	}

	if err := stream.Error(); err != nil {
		log.Printf("流式响应错误: %v", err)
		return
	}

	fmt.Println()
}

// 获取示例base64图像数据
// 这里使用一个1x1像素的PNG图片作为示例
func getExampleBase64Image() string {
	// 1x1像素的透明PNG图片的base64编码
	// 实际使用时，您应该提供真实的图像数据
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	return base64.StdEncoding.EncodeToString(pngData)
}

// 从文件读取图像并转换为base64的辅助函数
func imageToBase64(imagePath string) (string, error) {
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer imageFile.Close()

	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}
