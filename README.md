# MiniLM Dialog System

A high-performance, CPU-optimized dialog system for character-based applications with support for small language models.

## Features

- **Complete dialog backend** with mock LLM implementation (production llama.cpp integration planned)
- **Multi-backend support** with automatic fallback mechanisms
- **Context-aware conversation management** with personality-driven responses
- **Character Asset Integration** with automated tooling for adding LLM configuration to existing character files
- **CPU-optimized architecture** ready for production LLM integration (currently uses intelligent mock responses)

## Quick Start

### Basic Usage

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "minilm/dialog"
)

func main() {
    // Create dialog manager
    manager := dialog.NewDialogManager(false)
    
    // Create and configure LLM backend
    backend := dialog.NewLLMBackend()
    
    config := dialog.LLMConfig{
        ModelPath:   "/path/to/model.gguf",
        MaxTokens:   50,
        Temperature: 0.8,
        TopP:        0.9,
        ContextSize: 2048,
        Threads:     4,
    }
    
    configJSON, err := json.Marshal(config)
    if err != nil {
        log.Fatal(err)
    }
    
    err = backend.Initialize(configJSON)
    if err != nil {
        log.Fatal(err)
    }
    
    // Register backend
    manager.RegisterBackend("llm", backend)
    manager.SetDefaultBackend("llm")
    
    // Create dialog context
    context := dialog.DialogContext{
        Trigger:       "click",
        InteractionID: "session-1",
        CurrentMood:   80,
        PersonalityTraits: map[string]float64{
            "friendly": 0.8,
            "playful":  0.6,
        },
        FallbackResponses: []string{"Hello!"},
        FallbackAnimation: "talking",
    }
    
    // Generate dialog
    response, err := manager.GenerateDialog(context)
    if err != nil {
        log.Printf("Dialog generation failed: %v", err)
        return
    }
    
    fmt.Printf("Character says: %s\n", response.Text)
    fmt.Printf("Animation: %s\n", response.Animation)
}
```

## Architecture

The system follows a clean layered architecture:

- **Public API** (`dialog` package) - Clean interface for application integration
- **Internal Implementation** (`internal/dialog` package) - Core logic and backends
- **Example Applications** (`cmd/` directory) - Working examples and tools

## Model Support

**Current Status: Mock Implementation**

The system currently uses intelligent mock responses that simulate LLM behavior for development and testing. Production llama.cpp integration is planned for future releases.

**Planned Support** for small, permissively licensed models that will run efficiently on CPU:

- **Llama 2** (7B and smaller variants)
- **Mistral 7B** 
- **Phi-3** (3.8B)
- **TinyLlama** (1.1B)
- **RWKV** models

All models will support GGUF quantization (Q4, Q8) for optimal CPU performance when implemented.

## Performance

**Current Status: Mock Implementation**

The system currently provides instant mock responses for development and testing.

**Planned Performance** on consumer hardware with real LLM integration:

- **4-core CPU (2018+)**: 2-5 tokens/second with Q4 quantization
- **8-core CPU (2020+)**: 5-10 tokens/second with Q4 quantization
- **Memory usage**: 4-8GB RAM for 7B models (Q4), 2-4GB for smaller models

## Development

Run the example:

```bash
go run cmd/example/main.go
```

Run tests:

```bash
go test ./...
```

## License

MIT License - see LICENSE file for details.
