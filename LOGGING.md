# 数据流记录功能

## 功能说明

该功能允许记录路由器处理的所有输入输出数据流到文件，便于调试和审计。**同一个会话的所有数据将保存到同一个JSON文件中**，包括请求、响应以及完整的流式输出内容。

## 配置说明

在 `config.json` 中配置数据流记录：

```json
{
  "openrouter_base_url": "https://openrouter.ai/api/v1",
  "model_mappings": {
    "haiku": "anthropic/claude-3.5-haiku",
    "sonnet": "anthropic/claude-sonnet-4",
    "opus": "anthropic/claude-opus-4"
  },
  "data_logging": {
    "enabled": true,
    "directory": "./logs",
    "log_anthropic_request": true,
    "log_openai_request": true,
    "log_openai_response": true,
    "log_anthropic_response": true
  }
}
```

### 配置项说明

- `enabled`: 是否启用数据流记录（true/false）
- `directory`: 日志文件存储目录
- `log_anthropic_request`: 是否记录接收到的Anthropic格式请求
- `log_openai_request`: 是否记录转换后的OpenAI格式请求
- `log_openai_response`: 是否记录OpenRouter返回的OpenAI格式响应
- `log_anthropic_response`: 是否记录转换后的Anthropic格式响应（对于流式响应，记录完整内容）

## 日志文件格式

### 文件命名规则

**每个会话生成一个JSON文件**：`{时间戳}_{请求ID}.json`

示例：
```
20250124_143025_req_1706090225123.json
20250124_143520_req_1706090350456.json
```

### 文件内容结构

日志文件为JSON格式，包含一次完整会话的所有数据：

```json
{
  "request_id": "req_1706090225123",
  "timestamp": "2025-01-24 14:30:25",
  "anthropic_request": {
    "model": "claude-3-sonnet-20240229",
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ],
    "max_tokens": 1024
  },
  "openai_request": {
    "model": "anthropic/claude-sonnet-4",
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ]
  },
  "openai_response": {
    "id": "chatcmpl-123",
    "object": "chat.completion",
    "created": 1706090225,
    "model": "anthropic/claude-sonnet-4",
    "choices": [
      {
        "index": 0,
        "message": {
          "role": "assistant",
          "content": "Hello! How can I help you today?"
        },
        "finish_reason": "stop"
      }
    ]
  },
  "anthropic_response": {
    "id": "msg_1706090225",
    "type": "message",
    "role": "assistant",
    "content": [
      {
        "type": "text",
        "text": "Hello! How can I help you today?"
      }
    ],
    "model": "anthropic/claude-sonnet-4",
    "stop_reason": "end_turn"
  }
}
```

### 流式响应的记录

对于流式响应，会在 `stream_data` 字段中保存**完整的流式输出内容**：

```json
{
  "request_id": "req_1706090225123",
  "timestamp": "2025-01-24 14:30:25",
  "anthropic_request": { ... },
  "openai_request": { ... },
  "stream_data": "event: message_start\ndata: {...}\n\nevent: content_block_start\ndata: {...}\n\nevent: content_block_delta\ndata: {...}\n\n...",
  "is_streaming": true
}
```

## 使用示例

### 1. 启用完整日志记录

```json
{
  "data_logging": {
    "enabled": true,
    "directory": "./logs",
    "log_anthropic_request": true,
    "log_openai_request": true,
    "log_openai_response": true,
    "log_anthropic_response": true
  }
}
```

### 2. 仅记录请求

```json
{
  "data_logging": {
    "enabled": true,
    "directory": "./logs",
    "log_anthropic_request": true,
    "log_openai_request": true,
    "log_openai_response": false,
    "log_anthropic_response": false
  }
}
```

### 3. 仅记录响应

```json
{
  "data_logging": {
    "enabled": true,
    "directory": "./logs",
    "log_anthropic_request": false,
    "log_openai_request": false,
    "log_openai_response": true,
    "log_anthropic_response": true
  }
}
```

### 4. 禁用日志记录

```json
{
  "data_logging": {
    "enabled": false,
    "directory": "./logs",
    "log_anthropic_request": true,
    "log_openai_request": true,
    "log_openai_response": true,
    "log_anthropic_response": true
  }
}
```

## 设计原则

本功能遵循Golang设计哲学：

1. **简单性优于复杂性**：一个会话一个文件，清晰直观
2. **接口小而专注**：每个日志函数只负责一种类型的记录
3. **显式优于隐式**：明确指定记录哪些数据流
4. **减少抽象层次**：直接的文件写入，无复杂中间层
5. **错误不中断主流程**：日志记录失败不影响路由功能

## 日志管理

### 会话级别记录

所有数据按会话聚合，一个请求的所有相关数据（请求、响应、流式输出）都保存在同一个文件中，便于追踪和调试。

### 流式数据处理

流式响应会被完整收集后再保存到 `stream_data` 字段，而不是边读边写，确保日志的完整性和可读性。

## 注意事项

1. 日志文件可能包含敏感信息（API密钥、用户数据），请妥善保管
2. 流式响应会在内存中完整收集后再写入，对于超大响应需注意内存使用
3. 日志记录在会话结束时写入，不会阻塞请求处理
4. 确保日志目录有足够的磁盘空间
5. 定期清理旧日志文件以避免磁盘空间耗尽

## 日志示例

### 非流式请求日志

文件名：`20250124_143025_req_1706090225123.json`

```json
{
  "request_id": "req_1706090225123",
  "timestamp": "2025-01-24 14:30:25",
  "anthropic_request": {
    "model": "claude-3-sonnet-20240229",
    "messages": [
      {
        "role": "user",
        "content": "What is 2+2?"
      }
    ],
    "max_tokens": 1024
  },
  "openai_request": {
    "model": "anthropic/claude-sonnet-4",
    "messages": [
      {
        "role": "user",
        "content": "What is 2+2?"
      }
    ]
  },
  "openai_response": {
    "id": "chatcmpl-123",
    "choices": [
      {
        "message": {
          "role": "assistant",
          "content": "2+2 equals 4."
        }
      }
    ]
  },
  "anthropic_response": {
    "id": "msg_1706090225",
    "role": "assistant",
    "content": [
      {
        "type": "text",
        "text": "2+2 equals 4."
      }
    ]
  },
  "is_streaming": false
}
```

### 流式请求日志

文件名：`20250124_143520_req_1706090350456.json`

```json
{
  "request_id": "req_1706090350456",
  "timestamp": "2025-01-24 14:35:20",
  "anthropic_request": {
    "model": "claude-3-sonnet-20240229",
    "messages": [
      {
        "role": "user",
        "content": "Tell me a story"
      }
    ],
    "stream": true
  },
  "openai_request": {
    "model": "anthropic/claude-sonnet-4",
    "messages": [
      {
        "role": "user",
        "content": "Tell me a story"
      }
    ],
    "stream": true
  },
  "stream_data": "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{...}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",...}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Once\"}}\n\n...",
  "is_streaming": true
}
```
