package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// OpenAIChoice OpenAI选择结构
type OpenAIChoice struct {
	Index        int             `json:"index"`
	Message      OpenAIMessage   `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// OpenAIUsage OpenAI用量统计
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAICompletionResponse OpenAI完成响应结构
type OpenAICompletionResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   *OpenAIUsage   `json:"usage,omitempty"`
}

// AnthropicContent Anthropic内容结构
type AnthropicContent struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	ID   string      `json:"id,omitempty"`
	Name string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
}

// AnthropicUsage Anthropic用量统计
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicResponse Anthropic响应结构
type AnthropicResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []AnthropicContent `json:"content"`
	StopReason   string             `json:"stop_reason"`
	StopSequence interface{}        `json:"stop_sequence"`
	Model        string             `json:"model"`
	Usage        AnthropicUsage     `json:"usage"`
}

// formatOpenAIToAnthropic 将OpenAI格式转换为Anthropic格式
func formatOpenAIToAnthropic(completion OpenAICompletionResponse, model string) AnthropicResponse {
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixMilli())
	
	var content []AnthropicContent
	
	choice := completion.Choices[0]
	
	if choice.Message.Content != nil {
		// 处理文本内容
		if contentStr, ok := choice.Message.Content.(string); ok {
			content = []AnthropicContent{
				{
					Type: "text",
					Text: contentStr,
				},
			}
		}
	} else if len(choice.Message.ToolCalls) > 0 {
		// 处理工具调用
		for _, toolCall := range choice.Message.ToolCalls {
			var input interface{}
			if toolCall.Function.Arguments != "" {
				var args interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err == nil {
					input = args
				}
			}
			
			content = append(content, AnthropicContent{
				Type:  "tool_use",
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: input,
			})
		}
	}
	
	stopReason := "end_turn"
	if choice.FinishReason == "tool_calls" {
		stopReason = "tool_use"
	}

	// 转换usage信息
	var usage AnthropicUsage
	if completion.Usage != nil {
		usage = AnthropicUsage{
			InputTokens:  completion.Usage.PromptTokens,
			OutputTokens: completion.Usage.CompletionTokens,
		}
	}

	return AnthropicResponse{
		ID:           messageID,
		Type:         "message",
		Role:         "assistant",
		Content:      content,
		StopReason:   stopReason,
		StopSequence: nil,
		Model:        model,
		Usage:        usage,
	}
}