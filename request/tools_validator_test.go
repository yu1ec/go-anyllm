package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTools(t *testing.T) {
	t.Run("nil tools should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{}
		err := validateTools(req)
		assert.NoError(t, err)
	})

	t.Run("empty tools should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{},
		}
		err := validateTools(req)
		assert.NoError(t, err)
	})

	t.Run("valid tools should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name:        "get_weather",
						Description: "Get weather information",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type":        "string",
									"description": "The city and state",
								},
							},
							"required": []string{"location"},
						},
					},
				},
			},
		}
		err := validateTools(req)
		assert.NoError(t, err)
	})

	t.Run("tool with blank type should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
		}
		err := validateTools(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool type is blank")
	})

	t.Run("tool with invalid type should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "invalid",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
		}
		err := validateTools(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tool type")
		assert.Contains(t, err.Error(), "type must be 'function'")
	})

	t.Run("tool with nil function should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type:     "function",
					Function: nil,
				},
			},
		}
		err := validateTools(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "function is nil")
	})

	t.Run("tool with blank function name should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "",
					},
				},
			},
		}
		err := validateTools(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "function name is blank")
	})

	t.Run("tool with too long function name should fail", func(t *testing.T) {
		longName := ""
		for i := 0; i < 65; i++ {
			longName += "a"
		}
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: longName,
					},
				},
			},
		}
		err := validateTools(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "function name too long")
	})

	t.Run("tool with invalid characters in function name should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "invalid@name",
					},
				},
			},
		}
		err := validateTools(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character in function name")
	})

	t.Run("tool with valid function name characters should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "valid_function-name_123",
					},
				},
			},
		}
		err := validateTools(req)
		assert.NoError(t, err)
	})
}

func TestValidateToolChoice(t *testing.T) {
	t.Run("nil tool_choice should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{}
		err := validateToolChoice(req)
		assert.NoError(t, err)
	})

	t.Run("tool_choice without tools should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			ToolChoice: "auto",
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_choice can only be set when tools are provided")
	})

	t.Run("tool_choice with empty tools should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools:      &[]Tool{},
			ToolChoice: "auto",
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_choice can only be set when tools are provided")
	})

	t.Run("valid string tool_choice should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: "auto",
		}
		err := validateToolChoice(req)
		assert.NoError(t, err)

		req.ToolChoice = "none"
		err = validateToolChoice(req)
		assert.NoError(t, err)

		req.ToolChoice = "required"
		err = validateToolChoice(req)
		assert.NoError(t, err)
	})

	t.Run("invalid string tool_choice should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: "invalid",
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tool_choice string")
	})

	t.Run("valid object tool_choice should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name": "test_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.NoError(t, err)
	})

	t.Run("object tool_choice without type should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: map[string]interface{}{
				"function": map[string]interface{}{
					"name": "test_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_choice object must have 'type' field")
	})

	t.Run("object tool_choice with invalid type should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: map[string]interface{}{
				"type": "invalid",
				"function": map[string]interface{}{
					"name": "test_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tool_choice type")
	})

	t.Run("object tool_choice with nonexistent function should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name": "nonexistent_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_choice function \"nonexistent_function\" not found in provided tools")
	})

	t.Run("valid struct tool_choice should pass", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: ToolChoiceNamed{
				Type: "function",
				Function: ToolChoiceFunction{
					Name: "test_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.NoError(t, err)
	})

	t.Run("struct tool_choice with invalid type should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: ToolChoiceNamed{
				Type: "invalid",
				Function: ToolChoiceFunction{
					Name: "test_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tool_choice type")
	})

	t.Run("struct tool_choice with nonexistent function should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: ToolChoiceNamed{
				Type: "function",
				Function: ToolChoiceFunction{
					Name: "nonexistent_function",
				},
			},
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_choice function \"nonexistent_function\" not found in provided tools")
	})

	t.Run("invalid tool_choice type should fail", func(t *testing.T) {
		req := &ChatCompletionsRequest{
			Tools: &[]Tool{
				{
					Type: "function",
					Function: &ToolFunction{
						Name: "test_function",
					},
				},
			},
			ToolChoice: 123, // invalid type
		}
		err := validateToolChoice(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool_choice must be a string, object, or ToolChoiceNamed struct")
	})
}
