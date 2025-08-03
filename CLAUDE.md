# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go application demonstrating usage of the Eino AI framework with OpenAI integration. It's a simple example that shows how to:
- Create chat message templates with placeholders
- Initialize OpenAI chat models using environment variables
- Generate responses both synchronously and via streaming
- Handle conversation history and context

## Architecture

The codebase is organized into focused single-purpose files:

- `main.go` - Entry point that orchestrates template creation, model initialization, and response generation
- `openai.go` - OpenAI chat model configuration and initialization
- `template.go` - Message template creation using Eino's prompt system with FString formatting
- `generate.go` - Core LLM interaction functions for both sync and streaming responses
- `stream.go` - Stream response handling and reporting utilities

The application uses the Eino framework (`github.com/cloudwego/eino`) as the primary abstraction layer for AI model interactions, with OpenAI as the backend provider via `github.com/cloudwego/eino-ext/components/model/openai`.

## Development Commands

### Running the Application
```bash
go run .
```

### Building
```bash
go build -o eino-test .
```

### Running Tests
```bash
go test ./...
```

### Code Formatting
```bash
go fmt ./...
```

### Code Linting/Vetting
```bash
go vet ./...
```

### Dependency Management
```bash
go mod download    # Download dependencies
go mod tidy        # Clean up dependencies
go mod vendor      # Create vendor directory
```

## Environment Variables

The application requires these environment variables to function:

- `OPENAI_API_KEY` - Your OpenAI API key
- `OPENAI_MODEL_NAME` - Model name (e.g., "gpt-3.5-turbo", "gpt-4")
- `OPENAI_BASE_URL` - OpenAI API base URL (optional, defaults to OpenAI's official endpoint)

## Key Dependencies

- **Eino Framework**: Core AI abstraction framework
- **Eino OpenAI Extension**: OpenAI integration for Eino
- **Standard Go Libraries**: Context, logging, I/O operations

The application demonstrates a clean separation between template management, model configuration, and response generation, making it easy to extend with additional models or response processing logic.