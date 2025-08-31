package dialog

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNewDialogManager verifies that the public API correctly creates dialog managers
func TestNewDialogManager(t *testing.T) {
	manager := NewDialogManager(true)

	if manager == nil {
		t.Fatal("NewDialogManager should return a non-nil manager")
	}

	// Verify the manager is functional
	backends := manager.GetRegisteredBackends()
	if len(backends) != 0 {
		t.Errorf("New manager should have no backends, got %d", len(backends))
	}
}

// TestNewLLMBackend verifies that the public API correctly creates LLM backends
func TestNewLLMBackend(t *testing.T) {
	backend := NewLLMBackend()

	if backend == nil {
		t.Fatal("NewLLMBackend should return a non-nil backend")
	}

	// Verify backend info
	info := backend.GetBackendInfo()
	if info.Name != "llm_backend" {
		t.Errorf("Expected backend name 'llm_backend', got '%s'", info.Name)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}
}

// TestPublicAPIIntegration tests the complete public API workflow
func TestPublicAPIIntegration(t *testing.T) {
	// Create and configure dialog system using public API
	manager := NewDialogManager(false)
	backend := NewLLMBackend()

	// Configure backend
	config := LLMConfig{
		ModelPath:   "/path/to/model.gguf",
		MaxTokens:   50,
		Temperature: 0.8,
		TopP:        0.9,
		MarkovConfig: MarkovChainConfig{
			TrainingData: []string{
				"Hello there! I'm happy to see you! 😊",
				"How are you doing today?",
				"Thanks for being such a great friend!",
			},
		},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = backend.Initialize(configJSON)
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Register backend with manager
	manager.RegisterBackend("llm", backend)
	err = manager.SetDefaultBackend("llm")
	if err != nil {
		t.Fatalf("Failed to set default backend: %v", err)
	}

	// Create dialog context
	context := DialogContext{
		Trigger:       "click",
		InteractionID: "test-session",
		Timestamp:     time.Now(),
		CurrentMood:   75.0,
		PersonalityTraits: map[string]float64{
			"friendly": 0.8,
			"helpful":  0.7,
		},
		FallbackResponses: []string{"Hello!"},
		FallbackAnimation: "talking",
	}

	// Generate response
	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatalf("Failed to generate dialog: %v", err)
	}

	// Verify response
	if response.Text == "" {
		t.Error("Response text should not be empty")
	}

	if response.Animation == "" {
		t.Error("Response animation should not be empty")
	}

	if response.Confidence <= 0 {
		t.Error("Response confidence should be positive")
	}

	if response.ResponseType == "" {
		t.Error("Response type should not be empty")
	}
}

// TestValidateBackendConfig tests the public configuration validation function
func TestValidateBackendConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  DialogBackendConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DialogBackendConfig{
				Enabled:             true,
				DefaultBackend:      "llm",
				ConfidenceThreshold: 0.5,
				ResponseTimeout:     1000,
			},
			wantErr: false,
		},
		{
			name: "disabled config should pass",
			config: DialogBackendConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "missing default backend",
			config: DialogBackendConfig{
				Enabled:             true,
				ConfidenceThreshold: 0.5,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence threshold",
			config: DialogBackendConfig{
				Enabled:             true,
				DefaultBackend:      "llm",
				ConfidenceThreshold: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBackendConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBackendConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLoadDialogBackendConfig tests the public configuration loading function
func TestLoadDialogBackendConfig(t *testing.T) {
	jsonData := []byte(`{
		"enabled": true,
		"defaultBackend": "llm",
		"confidenceThreshold": 0.7,
		"memoryEnabled": true,
		"backends": {
			"llm": {
				"modelPath": "/models/test.gguf",
				"maxTokens": 50
			}
		}
	}`)

	config, err := LoadDialogBackendConfig(jsonData)
	if err != nil {
		t.Fatalf("LoadDialogBackendConfig() failed: %v", err)
	}

	if !config.Enabled {
		t.Error("Config should be enabled")
	}

	if config.DefaultBackend != "llm" {
		t.Errorf("Expected default backend 'llm', got '%s'", config.DefaultBackend)
	}

	if config.ConfidenceThreshold != 0.7 {
		t.Errorf("Expected confidence threshold 0.7, got %f", config.ConfidenceThreshold)
	}

	if !config.MemoryEnabled {
		t.Error("Memory should be enabled")
	}
}

// TestVersionInfo tests the version and API information functions
func TestVersionInfo(t *testing.T) {
	version := GetVersion()
	if version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", version)
	}

	apiInfo := GetAPIInfo()
	if apiInfo == nil {
		t.Fatal("GetAPIInfo should return non-nil map")
	}

	if apiInfo["version"] != "1.0.0" {
		t.Errorf("Expected API version '1.0.0', got '%v'", apiInfo["version"])
	}

	if apiInfo["compatibility"] != "DDS-1.0" {
		t.Errorf("Expected compatibility 'DDS-1.0', got '%v'", apiInfo["compatibility"])
	}

	features, ok := apiInfo["features"].([]string)
	if !ok {
		t.Error("Features should be a string slice")
	} else if len(features) == 0 {
		t.Error("Features should not be empty")
	}
}

// TestTypeCompatibility ensures all public types match internal types
func TestTypeCompatibility(t *testing.T) {
	// Test that public types are properly aliased to internal types

	// DialogContext compatibility
	context := DialogContext{
		Trigger:       "test",
		InteractionID: "test-id",
		Timestamp:     time.Now(),
	}

	if context.Trigger != "test" {
		t.Error("DialogContext should preserve fields correctly")
	}

	// DialogResponse compatibility
	response := DialogResponse{
		Text:       "Test response",
		Animation:  "talking",
		Confidence: 0.8,
	}

	if response.Text != "Test response" {
		t.Error("DialogResponse should preserve fields correctly")
	}

	// BackendInfo compatibility
	info := BackendInfo{
		Name:    "test_backend",
		Version: "1.0.0",
	}

	if info.Name != "test_backend" {
		t.Error("BackendInfo should preserve fields correctly")
	}
}

// TestBackendChaining tests the public API's backend fallback functionality
func TestBackendChaining(t *testing.T) {
	manager := NewDialogManager(false)

	// Create primary backend
	primaryBackend := NewLLMBackend()
	config := LLMConfig{
		ModelPath: "/fake/primary.gguf",
		MaxTokens: 50,
	}
	configJSON, _ := json.Marshal(config)
	primaryBackend.Initialize(configJSON)

	// Create fallback backend
	fallbackBackend := NewLLMBackend()
	fallbackConfig := LLMConfig{
		ModelPath: "/fake/fallback.gguf",
		MaxTokens: 30,
	}
	fallbackConfigJSON, _ := json.Marshal(fallbackConfig)
	fallbackBackend.Initialize(fallbackConfigJSON)

	// Register backends
	manager.RegisterBackend("primary", primaryBackend)
	manager.RegisterBackend("fallback", fallbackBackend)
	manager.SetDefaultBackend("primary")
	manager.SetFallbackChain([]string{"fallback"})

	// Test dialog generation
	context := DialogContext{
		Trigger:           "click",
		InteractionID:     "chain-test",
		FallbackResponses: []string{"Default response"},
		FallbackAnimation: "default",
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatalf("Dialog generation should not fail with fallback: %v", err)
	}

	if response.Text == "" {
		t.Error("Response should not be empty with fallback chain")
	}
}

// TestConcurrentAccess tests thread safety of the public API
func TestConcurrentAccess(t *testing.T) {
	manager := NewDialogManager(false)
	backend := NewLLMBackend()

	config := LLMConfig{
		ModelPath: "/fake/concurrent.gguf",
		MaxTokens: 50,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	manager.RegisterBackend("concurrent", backend)
	manager.SetDefaultBackend("concurrent")

	// Run concurrent dialog generation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			context := DialogContext{
				Trigger:           "click",
				InteractionID:     "concurrent-test",
				FallbackResponses: []string{"Concurrent response"},
			}

			_, err := manager.GenerateDialog(context)
			if err != nil {
				t.Errorf("Concurrent dialog %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Test for bug #3: README example should work with public API
func TestREADME_test_bug3_example_code_works(t *testing.T) {
	// This test validates that the README example code actually works
	// It reproduces the exact code shown in README.md

	// Create dialog manager
	manager := NewDialogManager(false)

	// Create and configure LLM backend
	backend := NewLLMBackend()

	config := LLMConfig{
		ModelPath:   "/path/to/model.gguf",
		MaxTokens:   50,
		Temperature: 0.8,
		TopP:        0.9,
		ContextSize: 2048,
		Threads:     4,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Config marshal failed: %v", err)
	}

	err = backend.Initialize(configJSON)
	if err != nil {
		t.Fatalf("Backend initialization failed: %v", err)
	}

	// Register backend
	manager.RegisterBackend("llm", backend)
	manager.SetDefaultBackend("llm")

	// Create dialog context
	context := DialogContext{
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
		t.Fatalf("Dialog generation failed: %v", err)
	}

	// Verify response
	if response.Text == "" {
		t.Error("Expected non-empty response text")
	}

	if response.Animation == "" {
		t.Error("Expected non-empty animation")
	}

	if response.Confidence <= 0 {
		t.Error("Expected positive confidence score")
	}

	t.Logf("Generated response: %s", response.Text)
	t.Logf("Animation: %s", response.Animation)
	t.Logf("Confidence: %.2f", response.Confidence)
}
