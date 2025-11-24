# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based API router that acts as a protocol converter between Anthropic Claude API and OpenAI-compatible APIs. The service translates incoming Anthropic-format requests to OpenAI format, forwards them to upstream providers (like OpenRouter), and converts responses back to Anthropic format.

## Common Development Commands

### Build and Run
```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go

# Build executable
go build -o y-router.exe
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific test
go test -v

# Run with coverage
go test -cover
```

## Architecture

The codebase follows Go's design philosophy with simple, focused components:

### Core Components

- **main.go**: Application entry point, configuration loading, and HTTP server setup
- **handlers.go**: Main HTTP request handler (`/v1/messages`) with session management and request orchestration
- **format_request.go**: Anthropic → OpenAI request format conversion with model mapping and tool validation
- **format_response.go**: OpenAI → Anthropic response format conversion
- **stream_response.go**: Server-sent events (SSE) streaming response conversion
- **logger.go**: Session-based data logging system that records complete request/response cycles
- **html_handlers.go**: Static page handlers (currently minimal)

### Request Flow

1. **Receive**: Anthropic-formatted request at `/v1/messages`
2. **Session**: Generate unique request ID, start logging session
3. **Convert**: Transform to OpenAI format (tools, messages, model mapping)
4. **Proxy**: Forward to upstream provider (OpenRouter by default)
5. **Process**: Handle streaming or non-streaming responses
6. **Convert**: Transform response back to Anthropic format
7. **Log**: Save complete session data to single JSON file
8. **Respond**: Return Anthropic-formatted response to client

### Configuration System

- **Primary**: `config.json` file
- **Override**: Environment variables (`OPENROUTER_BASE_URL`, `PORT`)
- **Model Mapping**: Keyword-based model name translation
- **Data Logging**: Configurable session-based request/response recording

### Data Structures

Key types are defined across format files:
- `MessageCreateParamsBase`: Incoming Anthropic request structure
- `OpenAIRequest`: Outgoing OpenAI request structure
- `OpenAICompletionResponse`: OpenAI response structure
- `AnthropicResponse`: Outgoing Anthropic response structure
- `SessionLog`: Complete session data for logging

## Model Mapping

The `mapModel()` function in `format_request.go:92` handles model name translation:
- Direct pass-through for models containing `/` (OpenRouter IDs)
- Keyword-based mapping using `ModelMappings` config
- Supports partial name matching (e.g., "haiku" → "anthropic/claude-3.5-haiku")

## Tool Validation

The `validateOpenAIToolCalls()` function in `format_request.go:108` ensures proper tool call/tool result pairing:
- Removes orphaned tool calls without immediate tool results
- Removes tool results without preceding tool calls
- Maintains message sequence integrity

## Streaming Implementation

Streaming uses Go's `io.Pipe` for real-time SSE conversion:
- Converts OpenAI SSE chunks to Anthropic events
- Handles tool call streaming and text content separately
- Manages content block lifecycle (start/delta/stop events)
- Collects complete stream data for logging

## Data Logging System

Session-based logging in `logger.go`:
- **One file per session**: `{timestamp}_{requestID}.json`
- **Complete coverage**: Records all request/response formats
- **Stream handling**: Collects full SSE content before logging
- **Non-blocking**: Logging failures don't interrupt requests
- **Configurable**: Selective logging via `data_logging` config

## Authentication

Supports two API key methods:
- `X-Api-Key` header (primary)
- `Authorization: Bearer <key>` header (fallback)

## Environment Variables

- `OPENROUTER_BASE_URL`: Override upstream API URL
- `PORT`: Server port (default: 8080)

## Development Notes

- The codebase uses minimal external dependencies (only Gin framework)
- Error handling follows Go conventions with explicit returns
- All JSON marshaling/unmarshaling includes error checking
- Stream processing uses buffered scanning for robustness
- Session management ensures request isolation and proper cleanup

## Configuration Examples

See `config.example.json` for full configuration structure. Key settings:
- `openrouter_base_url`: Upstream API endpoint
- `model_mappings`: Model name translation rules
- `data_logging`: Session recording preferences

## Testing the Service

Use curl to test the endpoint:
```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"model": "claude-3-sonnet-20240229", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}]}'
```