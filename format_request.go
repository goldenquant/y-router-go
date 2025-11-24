package main

import (
	"encoding/json"
	"strings"
)

// MessageCreateParamsBase 定义请求参数结构
type MessageCreateParamsBase struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	System      interface{} `json:"system,omitempty"`
	Temperature *float64    `json:"temperature,omitempty"`
	Tools       []Tool      `json:"tools,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
}

// Message 定义消息结构
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// Tool 定义工具结构
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ContentPart 定义内容部分结构
type ContentPart struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
}

// SystemMessage 定义系统消息结构
type SystemMessage struct {
	Role    string       `json:"role"`
	Content []ContentPart `json:"content"`
}

// OpenAIMessage OpenAI消息格式
type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    interface{}      `json:"content,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

// OpenAIToolCall OpenAI工具调用格式
type OpenAIToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Function OpenAIFunctionCall   `json:"function"`
}

// OpenAIFunctionCall OpenAI函数调用格式
type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAIRequest OpenAI请求格式
type OpenAIRequest struct {
	Model       string         `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature *float64       `json:"temperature,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	Tools       []OpenAITool   `json:"tools,omitempty"`
}

// OpenAITool OpenAI工具格式
type OpenAITool struct {
	Type     string               `json:"type"`
	Function OpenAIFunctionTool   `json:"function"`
}

// OpenAIFunctionTool OpenAI函数工具格式
type OpenAIFunctionTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// mapModel 映射模型名称
func mapModel(anthropicModel string) string {
	// 如果模型已经包含'/'，则是OpenRouter模型ID - 直接返回
	if strings.Contains(anthropicModel, "/") {
		return anthropicModel
	}
	
	// 遍历配置中的映射规则
	for keyword, mappedModel := range env.ModelMappings {
		if strings.Contains(anthropicModel, keyword) {
			return mappedModel
		}
	}
	
	return anthropicModel
}

// validateOpenAIToolCalls 验证OpenAI格式的消息以确保完整的tool_calls/tool消息配对
func validateOpenAIToolCalls(messages []OpenAIMessage) []OpenAIMessage {
	var validatedMessages []OpenAIMessage
	
	for i := 0; i < len(messages); i++ {
		currentMessage := messages[i]
		
		// 处理带有tool_calls的助手消息
		if currentMessage.Role == "assistant" && len(currentMessage.ToolCalls) > 0 {
			var validToolCalls []OpenAIToolCall
			var removedToolCallIDs []string
			
			// 收集所有紧随其后的工具消息
			var immediateToolMessages []OpenAIMessage
			j := i + 1
			for j < len(messages) && messages[j].Role == "tool" {
				immediateToolMessages = append(immediateToolMessages, messages[j])
				j++
			}
			
			// 对于每个tool_call，检查是否有紧随其后的工具消息
			for _, toolCall := range currentMessage.ToolCalls {
				hasImmediateToolMessage := false
				for _, toolMsg := range immediateToolMessages {
					if toolMsg.ToolCallID == toolCall.ID {
						hasImmediateToolMessage = true
						break
					}
				}
				
				if hasImmediateToolMessage {
					validToolCalls = append(validToolCalls, toolCall)
				} else {
					removedToolCallIDs = append(removedToolCallIDs, toolCall.ID)
				}
			}
			
			// 更新助手消息
			if len(validToolCalls) > 0 {
				currentMessage.ToolCalls = validToolCalls
			} else {
				currentMessage.ToolCalls = nil
			}
			
			// 只有当消息有内容或有效的tool_calls时才包含
			if currentMessage.Content != nil || len(currentMessage.ToolCalls) > 0 {
				validatedMessages = append(validatedMessages, currentMessage)
			}
		} else if currentMessage.Role == "tool" {
			hasImmediateToolCall := false
			
			// 检查紧邻的前一个助手消息是否有匹配的tool_call
			if i > 0 {
				prevMessage := messages[i-1]
				if prevMessage.Role == "assistant" && len(prevMessage.ToolCalls) > 0 {
					for _, toolCall := range prevMessage.ToolCalls {
						if toolCall.ID == currentMessage.ToolCallID {
							hasImmediateToolCall = true
							break
						}
					}
				} else if prevMessage.Role == "tool" {
					// 在工具消息序列之前检查助手消息
					for k := i - 1; k >= 0; k-- {
						if messages[k].Role == "tool" {
							continue
						}
						if messages[k].Role == "assistant" && len(messages[k].ToolCalls) > 0 {
							for _, toolCall := range messages[k].ToolCalls {
								if toolCall.ID == currentMessage.ToolCallID {
									hasImmediateToolCall = true
									break
								}
							}
						}
						break
					}
				}
			}
			
			if hasImmediateToolCall {
				validatedMessages = append(validatedMessages, currentMessage)
			}
		} else {
			validatedMessages = append(validatedMessages, currentMessage)
		}
	}
	
	return validatedMessages
}

// formatAnthropicToOpenAI 将Anthropic格式转换为OpenAI格式
func formatAnthropicToOpenAI(body MessageCreateParamsBase) (OpenAIRequest, error) {
	var openAIMessages []OpenAIMessage
	
	// 转换消息
	for _, anthropicMessage := range body.Messages {
		if anthropicMessage.Role == "assistant" {
			assistantMessage := OpenAIMessage{
				Role:    "assistant",
				Content: nil,
			}
			
			var textContent string
			var toolCalls []OpenAIToolCall
			
			if contentStr, ok := anthropicMessage.Content.(string); ok {
				// 简单字符串内容
				textContent = contentStr
			} else if contentArray, ok := anthropicMessage.Content.([]interface{}); ok {
				// 复杂内容数组
				for _, contentItem := range contentArray {
					contentBytes, _ := json.Marshal(contentItem)
					var contentPart ContentPart
					if err := json.Unmarshal(contentBytes, &contentPart); err == nil {
						if contentPart.Type == "text" {
							textContent += contentPart.Text + "\n"
						} else if contentPart.Type == "tool_use" {
							argsBytes, _ := json.Marshal(contentPart.Input)
							toolCalls = append(toolCalls, OpenAIToolCall{
								ID:   contentPart.ID,
								Type: "function",
								Function: OpenAIFunctionCall{
									Name:      contentPart.Name,
									Arguments: string(argsBytes),
								},
							})
						}
					}
				}
			}
			
			trimmedTextContent := strings.TrimSpace(textContent)
			if trimmedTextContent != "" {
				assistantMessage.Content = trimmedTextContent
			}
			if len(toolCalls) > 0 {
				assistantMessage.ToolCalls = toolCalls
			}
			
			if assistantMessage.Content != nil || len(assistantMessage.ToolCalls) > 0 {
				openAIMessages = append(openAIMessages, assistantMessage)
			}
		} else if anthropicMessage.Role == "user" {
			var userTextMessageContent string
			var subsequentToolMessages []OpenAIMessage
			
			if contentStr, ok := anthropicMessage.Content.(string); ok {
				// 简单字符串内容
				userTextMessageContent = contentStr
			} else if contentArray, ok := anthropicMessage.Content.([]interface{}); ok {
				// 复杂内容数组
				for _, contentItem := range contentArray {
					contentBytes, _ := json.Marshal(contentItem)
					var contentPart ContentPart
					if err := json.Unmarshal(contentBytes, &contentPart); err == nil {
						if contentPart.Type == "text" {
							userTextMessageContent += contentPart.Text + "\n"
						} else if contentPart.Type == "tool_result" {
							var toolContent interface{}
							if contentStr, ok := contentPart.Content.(string); ok {
								toolContent = contentStr
							} else {
								toolContent = contentPart.Content
							}
							
							subsequentToolMessages = append(subsequentToolMessages, OpenAIMessage{
								Role:       "tool",
								ToolCallID: contentPart.ToolUseID,
								Content:    toolContent,
							})
						}
					}
				}
			}
			
			trimmedUserText := strings.TrimSpace(userTextMessageContent)
			if trimmedUserText != "" {
				openAIMessages = append(openAIMessages, OpenAIMessage{
					Role:    "user",
					Content: trimmedUserText,
				})
			}
			openAIMessages = append(openAIMessages, subsequentToolMessages...)
		}
	}
	
	// 处理系统消息
	var systemMessages []OpenAIMessage
	if body.System != nil {
		if systemArray, ok := body.System.([]interface{}); ok {
			for _, item := range systemArray {
				itemBytes, _ := json.Marshal(item)
				var sysItem struct {
					Text string `json:"text"`
				}
				if err := json.Unmarshal(itemBytes, &sysItem); err == nil {
					content := ContentPart{
						Type: "text",
						Text: sysItem.Text,
					}
					if strings.Contains(body.Model, "claude") {
						content.Input = map[string]interface{}{"cache_control": map[string]string{"type": "ephemeral"}}
					}
					systemMessages = append(systemMessages, OpenAIMessage{
						Role:    "system",
						Content: []ContentPart{content},
					})
				}
			}
		} else if systemStr, ok := body.System.(string); ok {
			content := ContentPart{
				Type: "text",
				Text: systemStr,
			}
			if strings.Contains(body.Model, "claude") {
				content.Input = map[string]interface{}{"cache_control": map[string]string{"type": "ephemeral"}}
			}
			systemMessages = append(systemMessages, OpenAIMessage{
				Role:    "system",
				Content: []ContentPart{content},
			})
		}
	}
	
	// 构建最终请求
	data := OpenAIRequest{
		Model:       mapModel(body.Model),
		Messages:    append(systemMessages, validateOpenAIToolCalls(openAIMessages)...),
		Temperature: body.Temperature,
		Stream:      body.Stream,
	}
	
	// 处理工具
	if len(body.Tools) > 0 {
		var tools []OpenAITool
		for _, item := range body.Tools {
			tools = append(tools, OpenAITool{
				Type: "function",
				Function: OpenAIFunctionTool{
					Name:        item.Name,
					Description: item.Description,
					Parameters:  item.InputSchema,
				},
			})
		}
		data.Tools = tools
	}
	
	return data, nil
}