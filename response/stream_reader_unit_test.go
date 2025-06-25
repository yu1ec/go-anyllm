package response

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessResponse(t *testing.T) {
	t.Run("response keep-alive return error", func(t *testing.T) {
		respBody := []byte(KEEP_ALIVE)
		_, err := processResponse(respBody)
		assert.Error(t, err)
	})

	t.Run("response done return error", func(t *testing.T) {
		respBody := []byte(`data: [DONE]`)
		_, err := processResponse(respBody)
		assert.Error(t, err)
		assert.Equal(t, err, io.EOF)
	})

	t.Run("response json return chat response", func(t *testing.T) {
		respBody := []byte(`data: {"id":"aceb72f7-ffab-422a-b498-62c9b4034f84","object":"chat.completion.chunk","created":1738119601,"model":"deepseek-chat","system_fingerprint":"fp_3a5770e1b4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}`)
		chatResp, err := processResponse(respBody)
		assert.NoError(t, err)
		assert.NotNil(t, chatResp)
		assert.Equal(t, "aceb72f7-ffab-422a-b498-62c9b4034f84", chatResp.Id)
	})
}

func TestTrimDataPrefix(t *testing.T) {
	t.Run("data prefix trimmed from json response", func(t *testing.T) {
		dataPrefix := `data: `
		jsonResp := `{"id":"aceb72f7-ffab-422a-b498-62c9b4034f84","object":"chat.completion.chunk","created":1738119601,"model":"deepseek-chat","system_fingerprint":"fp_3a5770e1b4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}`
		respBody := []byte(dataPrefix + jsonResp)
		gotBody := trimDataPrefix(respBody)
		assert.Equal(t, jsonResp, string(gotBody))
	})

	t.Run("data prefix not trimmed from done response", func(t *testing.T) {
		dataPrefix := `data: `
		doneResp := `[DONE]`
		respBody := []byte(dataPrefix + doneResp)
		gotBody := trimDataPrefix(respBody)
		assert.Equal(t, doneResp, string(gotBody))
	})
}

// 新增迭代器模式测试
func TestStreamReaderIterator(t *testing.T) {
	t.Run("iterator pattern with Next() and Current()", func(t *testing.T) {
		// 模拟流式响应数据
		streamData := `data: {"id":"test-1","object":"chat.completion.chunk","created":1738119601,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"test-2","object":"chat.completion.chunk","created":1738119602,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":null}]}

data: {"id":"test-3","object":"chat.completion.chunk","created":1738119603,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}

data: [DONE]

`
		reader := io.NopCloser(strings.NewReader(streamData))
		streamReader := NewStreamReader(reader)

		var responses []*ChatCompletionsResponse

		// 使用迭代器模式读取所有响应
		for streamReader.Next() {
			current := streamReader.Current()
			assert.NotNil(t, current)
			responses = append(responses, current)
		}

		// 检查错误
		assert.NoError(t, streamReader.Error())

		// 验证响应数量和内容
		assert.Len(t, responses, 3)
		assert.Equal(t, "test-1", responses[0].Id)
		assert.Equal(t, "test-2", responses[1].Id)
		assert.Equal(t, "test-3", responses[2].Id)

		// 验证内容
		assert.Equal(t, "Hello", responses[0].Choices[0].Delta.Content)
		assert.Equal(t, " World", responses[1].Choices[0].Delta.Content)
		assert.Equal(t, "!", responses[2].Choices[0].Delta.Content)
		assert.Equal(t, "stop", responses[2].Choices[0].FinishReason)
	})

	t.Run("iterator handles empty stream", func(t *testing.T) {
		streamData := `data: [DONE]

`
		reader := io.NopCloser(strings.NewReader(streamData))
		streamReader := NewStreamReader(reader)

		// 空流应该立即结束
		assert.False(t, streamReader.Next())
		assert.NoError(t, streamReader.Error())
		assert.Nil(t, streamReader.Current())
	})

	t.Run("iterator handles error in stream", func(t *testing.T) {
		streamData := `data: {"invalid json

`
		reader := io.NopCloser(strings.NewReader(streamData))
		streamReader := NewStreamReader(reader)

		// 应该返回false并设置错误
		assert.False(t, streamReader.Next())
		assert.Error(t, streamReader.Error())
	})

	t.Run("multiple calls to Next() after stream end", func(t *testing.T) {
		streamData := `data: {"id":"test-1","object":"chat.completion.chunk","created":1738119601,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: [DONE]

`
		reader := io.NopCloser(strings.NewReader(streamData))
		streamReader := NewStreamReader(reader)

		// 第一次调用应该成功
		assert.True(t, streamReader.Next())
		assert.NotNil(t, streamReader.Current())

		// 第二次调用应该返回false（遇到DONE）
		assert.False(t, streamReader.Next())
		assert.NoError(t, streamReader.Error())

		// 后续调用应该一直返回false
		assert.False(t, streamReader.Next())
		assert.False(t, streamReader.Next())
	})
}
