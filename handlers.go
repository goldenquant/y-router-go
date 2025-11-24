package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// streamCollector 收集流式数据的完整内容
type streamCollector struct {
	inner     io.ReadCloser
	buffer    *bytes.Buffer
	requestID string
	logger    *DataLogger
}

func (sc *streamCollector) Read(p []byte) (n int, err error) {
	n, err = sc.inner.Read(p)
	if n > 0 {
		// 收集数据到缓冲区
		sc.buffer.Write(p[:n])
	}
	return n, err
}

func (sc *streamCollector) Close() error {
	// 在关闭时，将完整的流数据记录到日志
	if sc.logger.enabled && sc.logger.config.LogAnthropicResponse {
		sc.logger.LogStreamData(sc.requestID, sc.buffer.String())
	}
	return sc.inner.Close()
}

// wrapStreamWithCollector 包装流以收集完整数据
func wrapStreamWithCollector(stream io.ReadCloser, requestID string, logger *DataLogger) io.ReadCloser {
	return &streamCollector{
		inner:     stream,
		buffer:    &bytes.Buffer{},
		requestID: requestID,
		logger:    logger,
	}
}

// handleMessages 处理API消息请求
func handleMessages(c *gin.Context) {
	// 生成请求ID用于追踪
	requestID := generateRequestID()

	// 启动日志会话
	dataLogger.StartSession(requestID)

	// 确保在函数结束时保存日志
	defer func() {
		if err := dataLogger.EndSession(requestID); err != nil {
			// 记录错误但不影响响应
		}
	}()

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 解析Anthropic请求
	var anthropicRequest MessageCreateParamsBase
	if err := json.Unmarshal(body, &anthropicRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// 记录Anthropic请求
	dataLogger.LogAnthropicRequest(requestID, anthropicRequest)

	// 转换为OpenAI格式
	openaiRequest, err := formatAnthropicToOpenAI(anthropicRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert request format"})
		return
	}

	// 记录OpenAI请求
	dataLogger.LogOpenAIRequest(requestID, openaiRequest)

	// 获取API密钥
	bearerToken := c.GetHeader("X-Api-Key")
	if bearerToken == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			bearerToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
		return
	}

	// 准备OpenAI请求
	requestBody, err := json.Marshal(openaiRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal OpenAI request"})
		return
	}

	// 发送请求到OpenRouter
	req, err := http.NewRequest("POST", env.OpenRouterBaseUrl+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to upstream"})
		return
	}
	defer resp.Body.Close()

	// 处理错误响应
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "text/plain", errorBody)
		return
	}

	// 处理流式响应
	if openaiRequest.Stream {
		anthropicStream := streamOpenAIToAnthropic(resp.Body, openaiRequest.Model)

		// 如果启用了日志记录，包装流以收集完整数据
		if dataLogger.enabled && dataLogger.config.LogAnthropicResponse {
			anthropicStream = wrapStreamWithCollector(anthropicStream, requestID, dataLogger)
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		io.Copy(c.Writer, anthropicStream)
		anthropicStream.Close()
	} else {
		// 处理非流式响应
		var openaiResponse OpenAICompletionResponse
		if err := json.NewDecoder(resp.Body).Decode(&openaiResponse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode OpenAI response"})
			return
		}

		// 记录OpenAI响应
		dataLogger.LogOpenAIResponse(requestID, openaiResponse)

		// 转换为Anthropic格式
		anthropicResponse := formatOpenAIToAnthropic(openaiResponse, openaiRequest.Model)

		// 记录Anthropic响应
		dataLogger.LogAnthropicResponse(requestID, anthropicResponse)

		c.JSON(http.StatusOK, anthropicResponse)
	}
}