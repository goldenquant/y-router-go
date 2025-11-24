package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DataLogger 数据流记录器
type DataLogger struct {
	enabled   bool
	directory string
	config    LoggingConfig
	sessions  map[string]*SessionLog
	mu        sync.RWMutex
}

// SessionLog 会话日志，包含一次请求的所有数据
type SessionLog struct {
	RequestID         string      `json:"request_id"`
	Timestamp         string      `json:"timestamp"`
	AnthropicRequest  interface{} `json:"anthropic_request,omitempty"`
	OpenAIRequest     interface{} `json:"openai_request,omitempty"`
	OpenAIResponse    interface{} `json:"openai_response,omitempty"`
	AnthropicResponse interface{} `json:"anthropic_response,omitempty"`
	StreamData        string      `json:"stream_data,omitempty"`
	IsStreaming       bool        `json:"is_streaming"`
}

// NewDataLogger 创建新的数据记录器
func NewDataLogger(config LoggingConfig) *DataLogger {
	return &DataLogger{
		enabled:   config.Enabled,
		directory: config.Directory,
		config:    config,
		sessions:  make(map[string]*SessionLog),
	}
}

// StartSession 开始一个新会话
func (l *DataLogger) StartSession(requestID string) *SessionLog {
	if !l.enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	session := &SessionLog{
		RequestID: requestID,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	l.sessions[requestID] = session
	return session
}

// LogAnthropicRequest 记录Anthropic请求
func (l *DataLogger) LogAnthropicRequest(requestID string, data interface{}) {
	if !l.enabled || !l.config.LogAnthropicRequest {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if session, exists := l.sessions[requestID]; exists {
		session.AnthropicRequest = data
	}
}

// LogOpenAIRequest 记录OpenAI请求
func (l *DataLogger) LogOpenAIRequest(requestID string, data interface{}) {
	if !l.enabled || !l.config.LogOpenAIRequest {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if session, exists := l.sessions[requestID]; exists {
		session.OpenAIRequest = data
	}
}

// LogOpenAIResponse 记录OpenAI响应
func (l *DataLogger) LogOpenAIResponse(requestID string, data interface{}) {
	if !l.enabled || !l.config.LogOpenAIResponse {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if session, exists := l.sessions[requestID]; exists {
		session.OpenAIResponse = data
	}
}

// LogAnthropicResponse 记录Anthropic响应
func (l *DataLogger) LogAnthropicResponse(requestID string, data interface{}) {
	if !l.enabled || !l.config.LogAnthropicResponse {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if session, exists := l.sessions[requestID]; exists {
		session.AnthropicResponse = data
	}
}

// LogStreamData 记录流式响应的完整数据
func (l *DataLogger) LogStreamData(requestID string, data string) {
	if !l.enabled || !l.config.LogAnthropicResponse {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if session, exists := l.sessions[requestID]; exists {
		session.StreamData = data
		session.IsStreaming = true
	}
}

// EndSession 结束会话并保存日志
func (l *DataLogger) EndSession(requestID string) error {
	if !l.enabled {
		return nil
	}

	l.mu.Lock()
	session, exists := l.sessions[requestID]
	if !exists {
		l.mu.Unlock()
		return nil
	}
	delete(l.sessions, requestID)
	l.mu.Unlock()

	// 创建日志目录
	if err := os.MkdirAll(l.directory, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 生成文件名：时间戳_请求ID.json
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", timestamp, requestID)
	fullPath := filepath.Join(l.directory, filename)

	// 格式化JSON
	jsonData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
