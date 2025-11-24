# y-router-go

一个基于 Go 语言和 Gin 框架的 API 路由器，用于在 Anthropic Claude API 和 OpenAI 兼容 API 之间进行协议转换。

## 功能特性

- 🔄 **协议转换**: 将 Anthropic Claude API 格式转换为 OpenAI 兼容格式
- 🌊 **流式支持**: 支持流式响应处理
- 🚀 **高性能**: 基于 Gin 框架构建，提供高性能的 HTTP 服务
- 🔐 **安全认证**: 支持 API 密钥认证
- 📄 **静态页面**: 内置服务条款、隐私政策等页面
- 📝 **数据流记录**: 可配置的输入输出数据流记录功能，便于调试和审计

## 环境要求

- Go 1.20 或更高版本

## 安装与运行

### 1. 克隆项目

```bash
git clone <repository-url>
cd y-router-go
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置文件

创建 `config.json` 文件（可参考 `config.example.json`）：

```json
{
  "openrouter_base_url": "https://openrouter.ai/api/v1",
  "model_mappings": {
    "haiku": "anthropic/claude-3.5-haiku",
    "sonnet": "anthropic/claude-sonnet-4",
    "opus": "anthropic/claude-opus-4"
  },
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

### 4. 配置环境变量（可选）

```bash
# OpenRouter API 基础 URL（可选）
export OPENROUTER_BASE_URL="https://openrouter.ai/api/v1"

# 服务端口（可选，默认 8080）
export PORT="8080"
```

### 4. 运行服务

```bash
go run main.go
```

或者使用编译后的可执行文件：

```bash
./y-router.exe
```

## API 端点

### 主要 API

- `POST /v1/messages` - 消息处理端点，支持 Anthropic Claude API 格式

### 静态页面

- `GET /` - 首页
- `GET /terms` - 服务条款
- `GET /privacy` - 隐私政策
- `GET /install.sh` - 安装脚本

## 使用示例

### 发送消息请求

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1024,
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ]
  }'
```

## Claude Code 设置

```
export ANTHROPIC_BASE_URL="https://127.0.0.1:8080"
export ANTHROPIC_API_KEY="your-openrouter-api-key"
export ANTHROPIC_CUSTOM_HEADERS="x-api-key: $ANTHROPIC_API_KEY"
```

## 项目结构

```
y-router-go/
├── main.go              # 主程序入口
├── handlers.go          # HTTP 请求处理器
├── html_handlers.go     # 静态页面处理器
├── format_request.go    # 请求格式转换
├── format_response.go   # 响应格式转换
├── stream_response.go   # 流式响应处理
├── logger.go            # 数据流记录模块
├── config.json          # 配置文件
├── config.example.json  # 配置文件示例
├── LOGGING.md           # 数据流记录功能文档
├── go.mod               # Go 模块文件
├── go.sum               # 依赖校验文件
└── router.exe           # 编译后的可执行文件
```

## 技术栈

- **框架**: Gin (v1.9.1)
- **语言**: Go 1.20
- **HTTP 客户端**: 标准库 net/http

## 认证方式

支持两种 API 密钥认证方式：

1. 通过 `X-Api-Key` 请求头：
   ```
   X-Api-Key: your-api-key
   ```

2. 通过 `Authorization` 请求头：
   ```
   Authorization: Bearer your-api-key
   ```

## 数据流记录

本项目支持可配置的数据流记录功能，可以记录所有输入输出数据到文件。详细使用说明请参考 [LOGGING.md](LOGGING.md)。

### 快速启用

在 `config.json` 中设置：

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

日志文件会按照时间戳和请求ID命名，每个会话的所有数据保存在一个文件中：
- `20250124_143025_req_1706090225123.json` - 包含该会话的完整请求和响应数据
- `20250124_143520_req_1706090350456.json` - 包含另一个会话的完整数据

## 许可证

本项目采用开源许可证，具体请查看项目根目录的 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进本项目。