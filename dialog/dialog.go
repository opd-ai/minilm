// Package dialog provides the public API for MiniLM's LLM-powered dialog system.
// This package exports the core dialog functionality for integration with DDS and other applications.
//
// The dialog system supports multiple backends (LLM, Markov chain, rule-based) with automatic
// fallback mechanisms, context-aware conversation management, and personality-driven responses.
//
// Example Usage:
//
//	manager := dialog.NewDialogManager(false)
//	backend := dialog.NewLLMBackend()
//
//	config := dialog.LLMConfig{
//		ModelPath: "/path/to/model.gguf",
//		MaxTokens: 50,
//		Temperature: 0.8,
//	}
//	configJSON, _ := json.Marshal(config)
//	backend.Initialize(configJSON)
//
//	manager.RegisterBackend("llm", backend)
//	manager.SetDefaultBackend("llm")
//
//	context := dialog.DialogContext{
//		Trigger: "click",
//		InteractionID: "session-1",
//		CurrentMood: 80,
//	}
//
//	response, err := manager.GenerateDialog(context)
//	if err != nil {
//		log.Printf("Dialog generation failed: %v", err)
//		return
//	}
//
//	fmt.Printf("Character says: %s\n", response.Text)
//	fmt.Printf("Animation: %s\n", response.Animation)
package dialog

import (
	"github.com/opd-ai/minilm/internal/dialog"
)

// Core interfaces and types - re-export from internal package for public API

// DialogBackend defines the interface for pluggable dialog generation systems.
// Implementations include LLM-based, Markov chain, and rule-based backends.
type DialogBackend = dialog.DialogBackend

// DialogContext provides complete context for dialog generation including
// character state, interaction history, and conversation context.
type DialogContext = dialog.DialogContext

// DialogResponse contains the generated response and associated metadata
// including confidence scores, emotional tone, and animation triggers.
type DialogResponse = dialog.DialogResponse

// UserFeedback captures user response to dialog for backend learning
// and adaptation mechanisms.
type UserFeedback = dialog.UserFeedback

// InteractionRecord captures a single interaction for context building
// and conversation history management.
type InteractionRecord = dialog.InteractionRecord

// BackendInfo provides metadata about a dialog backend implementation
// including capabilities, version, and licensing information.
type BackendInfo = dialog.BackendInfo

// DialogManager orchestrates multiple backends and handles fallbacks.
// It provides the main entry point for dialog generation with automatic
// backend selection and graceful degradation.
type DialogManager = dialog.DialogManager

// Configuration types for backend setup

// LLMConfig defines configuration options for the LLM backend including
// model parameters, context management, and performance settings.
type LLMConfig = dialog.LLMConfig

// MarkovChainConfig represents Markov chain configuration for personality
// extraction and training data management.
type MarkovChainConfig = dialog.MarkovChainConfig

// DialogBackendConfig represents JSON configuration for dialog backends
// including fallback chains and global settings.
type DialogBackendConfig = dialog.DialogBackendConfig

// LLMBackend implements DialogBackend using LLM inference.
// Supports both mock (for development) and production (llama.cpp) models.
type LLMBackend = dialog.LLMBackend

// Factory functions - re-export from internal package

// NewDialogManager creates a new dialog manager with no backends registered.
// The debug parameter enables detailed logging of backend selection and fallback behavior.
//
// Example:
//
//	manager := NewDialogManager(true) // Enable debug logging
//	manager.RegisterBackend("llm", NewLLMBackend())
//	manager.SetDefaultBackend("llm")
func NewDialogManager(debug bool) *DialogManager {
	return dialog.NewDialogManager(debug)
}

// NewLLMBackend creates a new LLM-powered dialog backend with conservative
// defaults optimized for consumer CPU hardware.
//
// Default settings:
//   - MaxTokens: 50 (suitable for desktop pet responses)
//   - Temperature: 0.7 (balanced creativity and consistency)
//   - ContextSize: 2048 (fits most consumer hardware)
//   - Threads: 4 (optimal for 4-8 core CPUs)
//   - Timeout: 2 seconds (responsive UX)
//
// Example:
//
//	backend := NewLLMBackend()
//	config := LLMConfig{
//		ModelPath: "/models/tinyllama-1.1b-q4.gguf",
//		MarkovConfig: MarkovChainConfig{
//			TrainingData: []string{
//				"Hello! I'm so happy to see you! 😊",
//				"Thanks for spending time with me!",
//			},
//		},
//	}
//	configJSON, _ := json.Marshal(config)
//	err := backend.Initialize(configJSON)
func NewLLMBackend() *LLMBackend {
	return dialog.NewLLMBackend()
}

// Utility functions for configuration management

// ValidateBackendConfig ensures the backend configuration is valid.
// Returns an error if required fields are missing or values are out of range.
//
// Validation rules:
//   - DefaultBackend is required when dialog system is enabled
//   - ConfidenceThreshold must be between 0 and 1
//   - ResponseTimeout must be non-negative
//
// Example:
//
//	config := DialogBackendConfig{
//		Enabled: true,
//		DefaultBackend: "llm",
//		ConfidenceThreshold: 0.5,
//	}
//	if err := ValidateBackendConfig(config); err != nil {
//		log.Fatalf("Invalid config: %v", err)
//	}
func ValidateBackendConfig(config DialogBackendConfig) error {
	return dialog.ValidateBackendConfig(config)
}

// LoadDialogBackendConfig loads backend configuration from JSON data.
// Sets sensible defaults for missing optional fields.
//
// Default values:
//   - ConfidenceThreshold: 0.5
//   - ResponseTimeout: 1000ms
//   - MemoryEnabled: true
//   - LearningEnabled: false
//
// Example:
//
//	jsonData := []byte(`{
//		"enabled": true,
//		"defaultBackend": "llm",
//		"backends": {
//			"llm": {"modelPath": "/models/model.gguf"}
//		}
//	}`)
//	config, err := LoadDialogBackendConfig(jsonData)
func LoadDialogBackendConfig(data []byte) (DialogBackendConfig, error) {
	return dialog.LoadDialogBackendConfig(data)
}

// UpdateBackendMemory records interaction outcomes for backend learning.
// This enables backends to adapt based on user interactions and feedback.
//
// The method finds the backend that can handle the given context and
// updates its memory with the interaction outcome.
//
// Example:
//
//	feedback := &UserFeedback{
//		Positive: true,
//		Engagement: 0.9,
//	}
//	UpdateBackendMemory(manager, context, response, feedback)
func UpdateBackendMemory(dm *DialogManager, context DialogContext, response DialogResponse, feedback *UserFeedback) {
	dm.UpdateBackendMemory(context, response, feedback)
}

// Version and metadata

const (
	// Version represents the current version of the dialog API
	Version = "1.0.0"

	// APICompatibility indicates the API compatibility level
	APICompatibility = "DDS-1.0"
)

// GetVersion returns version information about the dialog system.
// Useful for compatibility checking and debugging.
func GetVersion() string {
	return Version
}

// GetAPIInfo returns information about the dialog API including
// version, compatibility level, and supported features.
func GetAPIInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":       Version,
		"compatibility": APICompatibility,
		"features": []string{
			"llm_backend",
			"context_management",
			"personality_extraction",
			"fallback_chains",
			"memory_tracking",
		},
		"backends": []string{
			"llm",
			"mock",
		},
	}
}
