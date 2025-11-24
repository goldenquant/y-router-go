package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// OpenAIStreamDelta OpenAI流式响应增量
type OpenAIStreamDelta struct {
	Role      string               `json:"role,omitempty"`
	Content   string               `json:"content,omitempty"`
	ToolCalls []OpenAIToolCallDelta `json:"tool_calls,omitempty"`
}

// OpenAIToolCallDelta OpenAI工具调用增量
type OpenAIToolCallDelta struct {
	Index    int                    `json:"index"`
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function OpenAIFunctionDelta    `json:"function,omitempty"`
}

// OpenAIFunctionDelta OpenAI函数增量
type OpenAIFunctionDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// OpenAIStreamResponse OpenAI流式响应结构
type OpenAIStreamResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
	Usage   *OpenAIUsage         `json:"usage,omitempty"`
}

// OpenAIStreamChoice OpenAI流式选择
type OpenAIStreamChoice struct {
	Index int                `json:"index"`
	Delta OpenAIStreamDelta  `json:"delta"`
	FinishReason *string     `json:"finish_reason,omitempty"`
}

// AnthropicStreamEvent Anthropic流式事件结构
type AnthropicStreamEvent struct {
	Type string `json:"type"`
}

// MessageStartEvent 消息开始事件
type MessageStartEvent struct {
	Type    string               `json:"type"`
	Message AnthropicResponse    `json:"message"`
}

// ContentBlockStartEvent 内容块开始事件
type ContentBlockStartEvent struct {
	Type         string            `json:"type"`
	Index        int               `json:"index"`
	ContentBlock AnthropicContent  `json:"content_block"`
}

// ContentBlockDeltaEvent 内容块增量事件
type ContentBlockDeltaEvent struct {
	Type  string          `json:"type"`
	Index int             `json:"index"`
	Delta interface{}     `json:"delta"`
}

// ContentBlockStopEvent 内容块停止事件
type ContentBlockStopEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// MessageDeltaEvent 消息增量事件
type MessageDeltaEvent struct {
	Type    string      `json:"type"`
	Delta   interface{} `json:"delta"`
	Usage   interface{} `json:"usage"`
}

// MessageStopEvent 消息停止事件
type MessageStopEvent struct {
	Type string `json:"type"`
}

// streamOpenAIToAnthropic 将OpenAI流式响应转换为Anthropic流式响应
func streamOpenAIToAnthropic(openaiStream io.ReadCloser, model string) io.ReadCloser {
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixMilli())
	
	pr, pw := io.Pipe()
	
	go func() {
		defer pw.Close()
		defer openaiStream.Close()
		
		// 发送message_start事件
		messageStart := MessageStartEvent{
			Type: "message_start",
			Message: AnthropicResponse{
				ID:           messageID,
				Type:         "message",
				Role:         "assistant",
				Content:      []AnthropicContent{},
				Model:        model,
				StopReason:   "",
				StopSequence: nil,
				Usage:        AnthropicUsage{InputTokens: 0, OutputTokens: 0},
			},
		}
		sendSSEEvent(pw, "message_start", messageStart)

		contentBlockIndex := 0
		hasStartedTextBlock := false
		isToolUse := false
		currentToolCallID := ""
		toolCallJsonMap := make(map[string]string)

		// 用于收集usage信息
		var inputTokens, outputTokens int
		
		scanner := bufio.NewScanner(openaiStream)
		var buffer string
		
		for scanner.Scan() {
			line := scanner.Text()
			buffer += line + "\n"
			
			// 处理完整的行
			lines := strings.Split(buffer, "\n")
			// 保留最后一个可能不完整的行在缓冲区中
			if len(lines) > 1 {
				buffer = lines[len(lines)-1]
				lines = lines[:len(lines)-1]
			} else {
				continue
			}
			
			for _, processLine := range lines {
				processLine = strings.TrimSpace(processLine)
				if !strings.HasPrefix(processLine, "data: ") {
					continue
				}
				
				data := strings.TrimPrefix(processLine, "data: ")
				if data == "[DONE]" {
					continue
				}
				
				var parsed OpenAIStreamResponse
				if err := json.Unmarshal([]byte(data), &parsed); err != nil {
					continue
				}
				
				if len(parsed.Choices) == 0 {
					// 检查是否有usage信息（某些提供商在最后一个chunk中返回）
					if parsed.Usage != nil {
						inputTokens = parsed.Usage.PromptTokens
						outputTokens = parsed.Usage.CompletionTokens
					}
					continue
				}

				// 更新usage信息
				if parsed.Usage != nil {
					inputTokens = parsed.Usage.PromptTokens
					outputTokens = parsed.Usage.CompletionTokens
				}
				
				delta := parsed.Choices[0].Delta
				processStreamDelta(pw, delta, &contentBlockIndex, &hasStartedTextBlock, &isToolUse, &currentToolCallID, toolCallJsonMap)
			}
		}
		
		// 处理缓冲区中剩余的数据
		if strings.TrimSpace(buffer) != "" {
			lines := strings.Split(buffer, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "data: ") {
					data := strings.TrimPrefix(line, "data: ")
					if data == "[DONE]" {
						continue
					}
					
					var parsed OpenAIStreamResponse
					if err := json.Unmarshal([]byte(data), &parsed); err != nil {
						continue
					}
					
					if len(parsed.Choices) > 0 {
						delta := parsed.Choices[0].Delta
						processStreamDelta(pw, delta, &contentBlockIndex, &hasStartedTextBlock, &isToolUse, &currentToolCallID, toolCallJsonMap)
					}
					// 更新usage信息
					if parsed.Usage != nil {
						inputTokens = parsed.Usage.PromptTokens
						outputTokens = parsed.Usage.CompletionTokens
					}
				}
			}
		}
		
		// 关闭最后一个内容块
		if isToolUse || hasStartedTextBlock {
			contentBlockStop := ContentBlockStopEvent{
				Type:  "content_block_stop",
				Index: contentBlockIndex,
			}
			sendSSEEvent(pw, "content_block_stop", contentBlockStop)
		}
		
		// 发送message_delta和message_stop事件
		stopReason := "end_turn"
		if isToolUse {
			stopReason = "tool_use"
		}
		
		messageDelta := MessageDeltaEvent{
			Type:  "message_delta",
			Delta: map[string]interface{}{
				"stop_reason":   stopReason,
				"stop_sequence": nil,
			},
			Usage: map[string]interface{}{
				"input_tokens":  inputTokens,
				"output_tokens": outputTokens,
			},
		}
		sendSSEEvent(pw, "message_delta", messageDelta)
		
		messageStop := MessageStopEvent{
			Type: "message_stop",
		}
		sendSSEEvent(pw, "message_stop", messageStop)
	}()
	
	return pr
}

// processStreamDelta 处理流式增量数据
func processStreamDelta(pw *io.PipeWriter, delta OpenAIStreamDelta, contentBlockIndex *int, hasStartedTextBlock *bool, isToolUse *bool, currentToolCallID *string, toolCallJsonMap map[string]string) {
	// 处理工具调用
	if len(delta.ToolCalls) > 0 {
		for _, toolCall := range delta.ToolCalls {
			toolCallID := toolCall.ID
			
			if toolCallID != "" && toolCallID != *currentToolCallID {
				if *isToolUse || *hasStartedTextBlock {
					contentBlockStop := ContentBlockStopEvent{
						Type:  "content_block_stop",
						Index: *contentBlockIndex,
					}
					sendSSEEvent(pw, "content_block_stop", contentBlockStop)
				}
				
				*isToolUse = true
				*hasStartedTextBlock = false
				*currentToolCallID = toolCallID
				*contentBlockIndex++
				toolCallJsonMap[toolCallID] = ""
				
				toolBlock := AnthropicContent{
					Type:  "tool_use",
					ID:    toolCallID,
					Name:  toolCall.Function.Name,
					Input: map[string]interface{}{},
				}
				
				contentBlockStart := ContentBlockStartEvent{
					Type:         "content_block_start",
					Index:        *contentBlockIndex,
					ContentBlock: toolBlock,
				}
				sendSSEEvent(pw, "content_block_start", contentBlockStart)
			}
			
			if toolCall.Function.Arguments != "" && *currentToolCallID != "" {
				currentJson := toolCallJsonMap[*currentToolCallID]
				toolCallJsonMap[*currentToolCallID] = currentJson + toolCall.Function.Arguments
				
				contentBlockDelta := ContentBlockDeltaEvent{
					Type:  "content_block_delta",
					Index: *contentBlockIndex,
					Delta: map[string]interface{}{
						"type":         "input_json_delta",
						"partial_json": toolCall.Function.Arguments,
					},
				}
				sendSSEEvent(pw, "content_block_delta", contentBlockDelta)
			}
		}
	} else if delta.Content != "" {
		if *isToolUse {
			contentBlockStop := ContentBlockStopEvent{
				Type:  "content_block_stop",
				Index: *contentBlockIndex,
			}
			sendSSEEvent(pw, "content_block_stop", contentBlockStop)
			*isToolUse = false
			*currentToolCallID = ""
			*contentBlockIndex++
		}
		
		if !*hasStartedTextBlock {
			contentBlockStart := ContentBlockStartEvent{
				Type:  "content_block_start",
				Index: *contentBlockIndex,
				ContentBlock: AnthropicContent{
					Type: "text",
					Text: "",
				},
			}
			sendSSEEvent(pw, "content_block_start", contentBlockStart)
			*hasStartedTextBlock = true
		}
		
		contentBlockDelta := ContentBlockDeltaEvent{
			Type:  "content_block_delta",
			Index: *contentBlockIndex,
			Delta: map[string]interface{}{
				"type": "text_delta",
				"text": delta.Content,
			},
		}
		sendSSEEvent(pw, "content_block_delta", contentBlockDelta)
	}
}

// sendSSEEvent 发送SSE事件
func sendSSEEvent(pw *io.PipeWriter, eventType string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	sseMessage := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, jsonData)
	pw.Write([]byte(sseMessage))
}