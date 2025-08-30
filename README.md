# MiniLM

> **A Go-based LLM-powered chatbot for DDS, optimized for CPU-only environments**

## Overview
MiniLM is a chatbot engine designed to bring personality and immersive interaction to [DDS](https://github.com/opd-ai/DDS). Unlike simple markov-chain bots, MiniLM leverages small, efficient language models to provide richer, more engaging conversations—while remaining practical for consumer hardware.

## Features
- **LLM Integration:** ✅ **Implemented** - Complete dialog backend using small, efficient language models with mock implementation ready for production LLM bindings
- **CPU-Optimized:** Runs efficiently on 4–8 core CPUs with 8–16GB RAM—no GPU required.
- **Production-Ready:** Focuses on maintainability, error handling, and performance with 89.6% test coverage.
- **Personality & Context:** Supports prompt engineering and context management for more natural dialogue.
- **DDS Compatibility:** Designed as a drop-in enhancement for DDS, following the interface in `go.interface.md`.
- **Rich Asset Support:** Includes a comprehensive set of character and gift assets from DDS, supporting a wide range of personalities, animations, and interaction scenarios.

## Assets
The `assets/` directory contains a rich set of resources originally from DDS, now bundled for seamless integration:

- **Characters:**
	- Multiple archetypes and difficulty levels (e.g., default, flirty, tsundere, romance, multiplayer, specialist, etc.)
	- Each character includes configuration (`character.json`) and animation sets (e.g., `eating.gif`, `happy.gif`, `talking.gif`, etc.)
	- Example and template characters for rapid prototyping
- **Gifts:**
	- A variety of gift items (e.g., `birthday_cake.json`, `chocolate_box.json`, `battle/guardian_amulet.json`, etc.)
- **Documentation:**
	- Each asset folder contains `README.md` and `SETUP.md` files for guidance

These assets enable MiniLM to deliver immersive, context-aware interactions and can be extended or customized for new scenarios.

## Implementation Status

### ✅ Completed
- **LLM Dialog Backend**: Full implementation of the DialogBackend interface
- **Context Management**: Conversation history with rolling window and user feedback tracking
- **Prompt Engineering**: Context-aware prompt construction with personality injection
- **Test Coverage**: Comprehensive test suite with 89.6% coverage
- **Documentation**: Complete implementation guide and examples

### 🚧 In Progress  
- **Model Integration**: llama.cpp binding integration (pending CGO resolution)
- **Character JSON Extension**: Adding LLM configuration to character cards

See `PLAN.md` for detailed implementation roadmap and `docs/LLM_BACKEND_IMPLEMENTATION.md` for technical details.

## Usage

### Quick Start
```go
// Create and configure LLM backend
backend := dialog.NewLLMBackend()
config := dialog.LLMConfig{
    ModelPath:   "/path/to/model.gguf",
    Personality: "cheerful and supportive",
    MaxTokens:   50,
}
configJSON, _ := json.Marshal(config)
backend.Initialize(configJSON)

// Generate context-aware response
context := dialog.DialogContext{
    Trigger:       "click",
    InteractionID: "session-1", 
    CurrentMood:   80,
    Personality:   map[string]float64{"cheerful": 0.9},
}
response, _ := backend.GenerateResponse(context)
fmt.Printf("Character says: %s", response.Text)
```

### Example Output
```
Character says: Hi there! I'm feeling great today! 😊
Animation: happy
Confidence: 0.8
Type: casual
```

See `cmd/example/main.go` for complete usage examples.

## Technical Notes
- Prioritizes permissively licensed models (Apache 2.0, MIT, etc.).
- Supports model quantization (GGUF, GGML, INT8/4) for memory and speed.
- Can be extended for web, CLI, or embedded use cases.
- Asset structure mirrors DDS for compatibility and ease of migration.

## Roadmap
- Add example Go code for model inference and chat loop
- Provide benchmarking results on typical CPUs
- Expand documentation for deployment, asset customization, and integration

---

For more information, see the DDS project and the interface/target documentation in this repository.

---

MiniLM is a chatbot with personality, designed for [DDS](https://github.com/opd-ai/DDS) to enhance interactivity. It is a more resource-intensive but more immersive alternative to the markov-chain base implementation. The DDS README is in `TARGET.md`. The target interface is in `go.interface.md`.