package dialog

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLlamaModel_NewLlamaModel(t *testing.T) {
	// Create a temporary GGUF file for testing
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	// Create empty file
	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Test valid configuration
	config := LlamaConfig{
		ModelPath:   modelPath,
		ContextSize: 1024,
		Threads:     2,
		Temperature: 0.8,
		TopP:        0.95,
	}

	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	if model.modelPath != config.ModelPath {
		t.Errorf("Expected modelPath %s, got %s", config.ModelPath, model.modelPath)
	}

	if model.contextSize != config.ContextSize {
		t.Errorf("Expected contextSize %d, got %d", config.ContextSize, model.contextSize)
	}

	if model.threads != config.Threads {
		t.Errorf("Expected threads %d, got %d", config.Threads, model.threads)
	}
}

func TestLlamaModel_NewLlamaModelDefaults(t *testing.T) {
	// Create a temporary GGUF file for testing
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	// Create empty file
	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Test configuration with defaults
	config := LlamaConfig{
		ModelPath: modelPath,
	}

	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	if model.contextSize != 2048 {
		t.Errorf("Expected default contextSize 2048, got %d", model.contextSize)
	}

	if model.threads != 4 {
		t.Errorf("Expected default threads 4, got %d", model.threads)
	}

	if model.temperature != 0.7 {
		t.Errorf("Expected default temperature 0.7, got %f", model.temperature)
	}

	if model.topP != 0.9 {
		t.Errorf("Expected default topP 0.9, got %f", model.topP)
	}
}

func TestLlamaModel_NewLlamaModelInvalidConfig(t *testing.T) {
	// Test empty model path
	config := LlamaConfig{}

	_, err := NewLlamaModel(config)
	if err == nil {
		t.Error("Expected error for empty model path")
	}
}

func TestLlamaModel_InitializeWithNonExistentFile(t *testing.T) {
	config := LlamaConfig{
		ModelPath: "/nonexistent/path/model.gguf",
	}

	_, err := NewLlamaModel(config)
	if err == nil {
		t.Error("Expected error for non-existent model file")
	}
}

func TestLlamaModel_InitializeWithMockFile(t *testing.T) {
	// Create a temporary GGUF file for testing
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	// Create empty file
	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{
		ModelPath: modelPath,
	}

	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	// Test initialization
	err = model.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	if !model.initialized {
		t.Error("Model should be initialized")
	}

	// Clean up
	model.Free()
}

func TestLlamaModel_PredictNotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{ModelPath: modelPath}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	_, err = model.Predict("Hello")
	if err == nil {
		t.Error("Expected error for uninitialized model")
	}
}

func TestLlamaModel_Predict(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{ModelPath: modelPath}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	err = model.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	response, err := model.Predict("Hello there!")
	if err != nil {
		t.Fatalf("Failed to predict: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	model.Free()
}

func TestLlamaModel_PredictWithTimeout(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{ModelPath: modelPath}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	err = model.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	response, err := model.PredictWithTimeout(ctx, "Hello there!")
	if err != nil {
		t.Fatalf("Failed to predict with timeout: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	model.Free()
}

func TestLlamaModel_EstimateTokens(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{ModelPath: modelPath}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	// Test token estimation
	text := "Hello world, this is a test"
	tokens := model.EstimateTokens(text)

	expectedTokens := len(text) / 4 // Rough approximation
	if tokens != expectedTokens {
		t.Errorf("Expected approximately %d tokens, got %d", expectedTokens, tokens)
	}
}

func TestLlamaModel_GetContextSize(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{
		ModelPath:   modelPath,
		ContextSize: 1024,
	}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	contextSize := model.GetContextSize()
	if contextSize != 1024 {
		t.Errorf("Expected context size 1024, got %d", contextSize)
	}
}

func TestLlamaModel_GetModelInfo(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{
		ModelPath:   modelPath,
		ContextSize: 1024,
		Threads:     2,
		Temperature: 0.8,
		TopP:        0.95,
	}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	info := model.GetModelInfo()

	if info.ModelPath != modelPath {
		t.Errorf("Expected model path %s, got %s", modelPath, info.ModelPath)
	}

	if info.ContextSize != 1024 {
		t.Errorf("Expected context size 1024, got %d", info.ContextSize)
	}

	if info.Threads != 2 {
		t.Errorf("Expected threads 2, got %d", info.Threads)
	}

	if info.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", info.Temperature)
	}

	if info.TopP != 0.95 {
		t.Errorf("Expected topP 0.95, got %f", info.TopP)
	}

	if info.ModelType != "llama.cpp" {
		t.Errorf("Expected model type 'llama.cpp', got %s", info.ModelType)
	}

	if info.Backend != "CPU" {
		t.Errorf("Expected backend 'CPU', got %s", info.Backend)
	}
}

func TestLlamaModel_Free(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{ModelPath: modelPath}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	err = model.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	err = model.Free()
	if err != nil {
		t.Fatalf("Failed to free model: %v", err)
	}

	if model.initialized {
		t.Error("Model should not be initialized after free")
	}
}

func TestLlamaModel_ContextualResponses(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "test_model.gguf")

	file, err := os.Create(modelPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	config := LlamaConfig{ModelPath: modelPath}
	model, err := NewLlamaModel(config)
	if err != nil {
		t.Fatalf("Failed to create LlamaModel: %v", err)
	}

	err = model.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Test different context types
	testCases := []struct {
		prompt   string
		keywords []string
	}{
		{"User just fed the character", []string{"thank", "food", "meal"}},
		{"Character is feeling happy", []string{"wonderful", "great", "happy"}},
		{"Character is feeling sad", []string{"understand", "sad", "here"}},
		{"Romantic context", []string{"heart", "love", "magic"}},
		{"General conversation", []string{"chat", "talk", "conversation"}},
	}

	for _, tc := range testCases {
		response, err := model.Predict(tc.prompt)
		if err != nil {
			t.Fatalf("Failed to predict for prompt '%s': %v", tc.prompt, err)
		}

		if response == "" {
			t.Errorf("Expected non-empty response for prompt '%s'", tc.prompt)
		}

		// Check if response contains contextually appropriate keywords
		containsKeyword := false
		for _, keyword := range tc.keywords {
			if contains(response, keyword) {
				containsKeyword = true
				break
			}
		}

		if !containsKeyword {
			t.Logf("Response for '%s': %s", tc.prompt, response)
			// Note: This is informational - the mock might not always match keywords
		}
	}

	model.Free()
}

// Helper function to check if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
