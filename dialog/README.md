# MiniLM Dialog API

A production-ready, CPU-optimized LLM dialog system designed for desktop virtual pets and interactive applications.

## Overview

The MiniLM Dialog API provides a complete solution for LLM-powered conversations with the following key features:

- **Multiple Backend Support**: LLM, Markov chain, and rule-based dialog backends
- **Automatic Fallback**: Graceful degradation when primary backends fail
- **Context Management**: Conversation history and personality-driven responses  
- **Resource Optimized**: Designed for consumer hardware (4-8 cores, 8-16GB RAM)
- **DDS Compatible**: Drop-in replacement for existing dialog systems

## Quick Start

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "minilm/dialog"
)

func main() {
    // Create dialog manager
    manager := dialog.NewDialogManager(false)
    
    // Create and configure LLM backend
    backend := dialog.NewLLMBackend()
    config := dialog.LLMConfig{
        ModelPath:   "/models/tinyllama-1.1b-q4.gguf",
        MaxTokens:   50,
        Temperature: 0.8,
        MarkovConfig: dialog.MarkovChainConfig{
            TrainingData: []string{
                "Hello! I'm so happy to see you! 😊",
                "How are you doing today?",
                "Thanks for being such a wonderful friend!",
            },
        },
    }
    
    configJSON, _ := json.Marshal(config)
    if err := backend.Initialize(configJSON); err != nil {
        log.Fatal(err)
    }
    
    // Register backend
    manager.RegisterBackend("llm", backend)
    manager.SetDefaultBackend("llm")
    
    // Generate dialog
    context := dialog.DialogContext{
        Trigger:       "click",
        InteractionID: "demo-session",
        Timestamp:     time.Now(),
        CurrentMood:   80.0,
        PersonalityTraits: map[string]float64{
            "friendly": 0.9,
            "cheerful": 0.8,
        },
    }
    
    response, err := manager.GenerateDialog(context)
    if err != nil {
        log.Printf("Dialog failed: %v", err)
        return
    }
    
    fmt.Printf("Character: %s\n", response.Text)
    fmt.Printf("Animation: %s\n", response.Animation)
    fmt.Printf("Confidence: %.2f\n", response.Confidence)
}
```

## Architecture

### Core Components

#### DialogManager
The central orchestrator that manages multiple backends and handles fallback chains:

```go
manager := dialog.NewDialogManager(debug)
manager.RegisterBackend("llm", llmBackend)
manager.RegisterBackend("markov", markovBackend)
manager.SetDefaultBackend("llm")
manager.SetFallbackChain([]string{"markov"})
```

#### LLMBackend
Production-ready LLM backend with CPU optimization:

```go
backend := dialog.NewLLMBackend()
config := dialog.LLMConfig{
    ModelPath:        "/models/model.gguf",
    MaxTokens:        50,          // Short responses for pets
    Temperature:      0.7,         // Balanced creativity
    ContextSize:      2048,        // Fits consumer hardware
    Threads:          4,           // Optimal for 4-8 core CPUs
    MaxHistoryLength: 5,           // Rolling conversation window
    TimeoutMs:        2000,        // Responsive UX
}
```

### Dialog Flow

1. **Context Creation**: Build DialogContext with character state and interaction details
2. **Backend Selection**: Manager selects best available backend
3. **Prompt Construction**: Context-aware prompt building with personality injection
4. **Response Generation**: LLM inference with timeout protection
5. **Fallback Handling**: Automatic fallback if primary backend fails
6. **Memory Update**: Record interaction for future context

## Configuration

### Backend Configuration

Complete backend configuration with validation:

```go
config := dialog.DialogBackendConfig{
    Enabled:             true,
    DefaultBackend:      "llm",
    FallbackChain:       []string{"markov", "simple"},
    ConfidenceThreshold: 0.5,
    MemoryEnabled:       true,
    LearningEnabled:     false,
    ResponseTimeout:     2000,
    DebugMode:           false,
    Backends: map[string]json.RawMessage{
        "llm": llmConfigJSON,
    },
}

// Validate configuration
if err := dialog.ValidateBackendConfig(config); err != nil {
    log.Fatal("Invalid config:", err)
}
```

### Character Integration

Integrate with existing character systems using Markov training data:

```json
{
  "name": "Friendly Pet",
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "llm",
    "backends": {
      "llm": {
        "modelPath": "/models/tinyllama-1.1b-q4.gguf",
        "maxTokens": 50,
        "temperature": 0.8,
        "markov_chain": {
          "trainingData": [
            "Hello there! I'm so happy to see you again! 😊",
            "How are you doing today? I've been thinking about you!",
            "Thanks for being such a great friend! I appreciate you so much."
          ],
          "fallbackPhrases": [
            "Hi there! 👋",
            "What's up?",
            "Nice to see you!"
          ]
        }
      }
    }
  }
}
```

## Performance Optimization

### CPU Optimization
- **Model Selection**: Use quantized models (Q4, Q8) under 500MB
- **Context Management**: Rolling window of 5-10 recent exchanges
- **Token Limiting**: Max 50 tokens per response for desktop pets
- **Thread Control**: Configure threads based on CPU cores

### Memory Management
- **Resource Cleanup**: Proper model deallocation with `Free()` methods
- **Context Pruning**: Automatic conversation history trimming
- **Fallback Chains**: Graceful degradation to lighter backends

### Response Time
- **Timeout Protection**: 2-second default timeout with context cancellation
- **Async Generation**: Non-blocking response generation
- **Cache Efficiency**: Reuse loaded models across conversations

## Error Handling

The API provides comprehensive error handling with automatic recovery:

```go
response, err := manager.GenerateDialog(context)
if err != nil {
    // Primary backend failed, but fallback may have succeeded
    // Check response.ResponseType for "fallback" indicator
    log.Printf("Backend error (recovered): %v", err)
}

// Always check confidence for quality assessment
if response.Confidence < 0.3 {
    // Low confidence response, may want to retry or use different backend
    log.Printf("Low confidence response: %.2f", response.Confidence)
}
```

## Testing

Comprehensive test suite with 100% coverage of public API:

```bash
# Run tests
go test ./dialog -v

# Check coverage  
go test ./dialog -cover

# Performance testing
go test ./dialog -bench=.
```

## Production Deployment

### Model Setup
1. Download quantized GGUF models (recommended: TinyLlama 1.1B Q4)
2. Install llama.cpp dependencies (when CGO support available)
3. Configure model paths in character files

### Resource Requirements
- **CPU**: 4-8 cores (Intel i5/AMD Ryzen 5 or better)
- **RAM**: 8-16GB total (models use <500MB)
- **Storage**: 500MB-2GB for model files
- **Response Time**: <500ms on target hardware

### Integration Steps
1. Import the dialog package: `import "minilm/dialog"`
2. Create DialogManager with desired backends
3. Configure backends with appropriate model paths
4. Register backends and set fallback chains
5. Generate responses using DialogContext

## API Reference

### Core Types

- **DialogManager**: Main orchestrator for multi-backend dialog
- **LLMBackend**: LLM-powered backend with CPU optimization  
- **DialogContext**: Complete context for dialog generation
- **DialogResponse**: Generated response with metadata
- **LLMConfig**: Configuration for LLM backend parameters

### Factory Functions

- `NewDialogManager(debug bool) *DialogManager`
- `NewLLMBackend() *LLMBackend`

### Configuration Functions

- `ValidateBackendConfig(config DialogBackendConfig) error`
- `LoadDialogBackendConfig(data []byte) (DialogBackendConfig, error)`

### Version Information

- `GetVersion() string` - Returns current API version
- `GetAPIInfo() map[string]interface{}` - Returns detailed API metadata

## Examples

See `cmd/example/main.go` for complete usage examples including:
- Backend configuration and initialization
- Context creation and dialog generation
- Error handling and fallback scenarios
- Performance monitoring and optimization

## License

MIT License - see LICENSE file for details.

## Contributing

1. Ensure all tests pass: `go test ./...`
2. Maintain test coverage above 85%
3. Follow Go conventions and document public APIs
4. Add integration tests for new backends
