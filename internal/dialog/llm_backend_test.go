package dialog

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLLMBackend_Initialize(t *testing.T) {
	backend := NewLLMBackend()

	// Test basic configuration
	config := LLMConfig{
		ModelPath:   "/path/to/model.gguf",
		MaxTokens:   100,
		Temperature: 0.8,
		TopP:        0.9,
		ContextSize: 1024,
		Threads:     2,
		MarkovConfig: MarkovChainConfig{
			TrainingData: []string{
				"I'm cheerful and helpful!",
				"How can I assist you today?",
				"I'm always here to help with a smile!",
			},
		},
		MaxHistoryLength: 5,
		TimeoutMs:        3000,
		FallbackEnabled:  true,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = backend.Initialize(configJSON)
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	if !backend.initialized {
		t.Error("Backend should be initialized")
	}

	if backend.maxTokens != 100 {
		t.Errorf("Expected maxTokens to be 100, got %d", backend.maxTokens)
	}

	if len(backend.markovConfig.TrainingData) == 0 {
		t.Error("Expected MarkovConfig training data to be set")
	}

	if backend.markovConfig.TrainingData[0] != "I'm cheerful and helpful!" {
		t.Errorf("Expected first training data to be 'I'm cheerful and helpful!', got '%s'", backend.markovConfig.TrainingData[0])
	}
}

func TestLLMBackend_InitializeWithDefaults(t *testing.T) {
	backend := NewLLMBackend()

	// Test with minimal configuration
	config := LLMConfig{
		ModelPath: "/path/to/model.gguf",
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = backend.Initialize(configJSON)
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Check defaults are applied
	if backend.maxTokens != 50 { // Default value
		t.Errorf("Expected default maxTokens to be 50, got %d", backend.maxTokens)
	}

	if backend.temperature != 0.7 { // Default value
		t.Errorf("Expected default temperature to be 0.7, got %f", backend.temperature)
	}
}

func TestLLMBackend_InitializeInvalidConfig(t *testing.T) {
	backend := NewLLMBackend()

	// Test with missing required field
	config := LLMConfig{
		// ModelPath is missing
		MaxTokens: 100,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = backend.Initialize(configJSON)
	if err == nil {
		t.Error("Expected error for missing modelPath, but got nil")
	}
}

func TestLLMBackend_GenerateResponse(t *testing.T) {
	backend := NewLLMBackend()

	// Initialize with basic config
	config := LLMConfig{
		ModelPath: "/path/to/model.gguf",
		MarkovConfig: MarkovChainConfig{
			TrainingData: []string{
				"I'm friendly and supportive!",
				"How can I help you today?",
				"I'm here to support you in any way I can!",
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

	// Test response generation
	context := DialogContext{
		Trigger:       "click",
		InteractionID: "test-interaction-1",
		Timestamp:     time.Now(),
		CurrentStats:  map[string]float64{"happiness": 80},
		PersonalityTraits: map[string]float64{
			"cheerful":  0.8,
			"friendly":  0.9,
			"energetic": 0.6,
		},
		CurrentMood:       75,
		CurrentAnimation:  "idle",
		TimeOfDay:         "morning",
		FallbackResponses: []string{"Hello!", "Hi there!"},
		FallbackAnimation: "talking",
	}

	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	// Validate response structure
	if response.Text == "" {
		t.Error("Response text should not be empty")
	}

	if response.Animation == "" {
		t.Error("Response animation should not be empty")
	}

	if response.Confidence <= 0 || response.Confidence > 1 {
		t.Errorf("Response confidence should be between 0 and 1, got %f", response.Confidence)
	}

	if response.ResponseType == "" {
		t.Error("Response type should not be empty")
	}
}

func TestLLMBackend_GenerateResponseNotInitialized(t *testing.T) {
	backend := NewLLMBackend()

	context := DialogContext{
		Trigger:       "click",
		InteractionID: "test-interaction-1",
		Timestamp:     time.Now(),
	}

	_, err := backend.GenerateResponse(context)
	if err == nil {
		t.Error("Expected error for uninitialized backend, but got nil")
	}
}

func TestLLMBackend_CanHandle(t *testing.T) {
	backend := NewLLMBackend()

	// Should not handle when not initialized
	context := DialogContext{Trigger: "click"}
	if backend.CanHandle(context) {
		t.Error("Uninitialized backend should not handle requests")
	}

	// Initialize
	config := LLMConfig{ModelPath: "/path/to/model.gguf"}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	// Should handle when initialized
	if !backend.CanHandle(context) {
		t.Error("Initialized backend should handle requests")
	}
}

func TestLLMBackend_GetBackendInfo(t *testing.T) {
	backend := NewLLMBackend()
	info := backend.GetBackendInfo()

	if info.Name != "llm_backend" {
		t.Errorf("Expected backend name to be 'llm_backend', got '%s'", info.Name)
	}

	if info.Version == "" {
		t.Error("Backend version should not be empty")
	}

	if len(info.Capabilities) == 0 {
		t.Error("Backend should have capabilities listed")
	}

	// Check for expected capabilities
	expectedCaps := []string{"context_aware", "personality_driven"}
	for _, expected := range expectedCaps {
		found := false
		for _, cap := range info.Capabilities {
			if cap == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected capability '%s' not found in %v", expected, info.Capabilities)
		}
	}
}

func TestLLMBackend_UpdateMemory(t *testing.T) {
	backend := NewLLMBackend()

	// Initialize
	config := LLMConfig{ModelPath: "/path/to/model.gguf"}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	context := DialogContext{
		InteractionID: "test-interaction-1",
		Trigger:       "click",
	}

	response := DialogResponse{
		Text:       "Hello there!",
		Animation:  "talking",
		Confidence: 0.8,
	}

	feedback := &UserFeedback{
		Positive:   true,
		Engagement: 0.9,
	}

	// Should not error
	err := backend.UpdateMemory(context, response, feedback)
	if err != nil {
		t.Errorf("UpdateMemory should not error, got: %v", err)
	}
}

func TestLLMBackend_CleanResponse(t *testing.T) {
	backend := NewLLMBackend()

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    `"Hello there!"`,
			expected: "Hello there!",
		},
		{
			input:    `'Hi friend!'`,
			expected: "Hi friend!",
		},
		{
			input:    "  Hello with spaces  ",
			expected: "Hello with spaces",
		},
		{
			input:    "",
			expected: "Hello! 👋",
		},
		{
			input:    "This is a very long response that should be truncated. It contains multiple sentences for testing. This sentence should be removed during truncation. And this one too.",
			expected: "", // We'll check truncation separately
		},
	}

	for i, tc := range testCases {
		result := backend.cleanResponse(tc.input)
		if i == 4 { // The long response test case
			// For long responses, just check that it was truncated
			if len(result) >= len(tc.input) {
				t.Errorf("Test case %d: Expected response to be truncated, but it wasn't", i)
			}
		} else if result != tc.expected {
			t.Errorf("Test case %d: Expected '%s', got '%s'", i, tc.expected, result)
		}
	}
}

func TestLLMBackend_SelectAnimation(t *testing.T) {
	backend := NewLLMBackend()

	testCases := []struct {
		response string
		expected string
	}{
		{
			response: "I'm so happy! 😊",
			expected: "happy",
		},
		{
			response: "I'm feeling sad today 😢",
			expected: "sad",
		},
		{
			response: "Time to eat some food!",
			expected: "eating",
		},
		{
			response: "Just a normal response",
			expected: "talking",
		},
	}

	context := DialogContext{}

	for i, tc := range testCases {
		result := backend.selectAnimation(context, tc.response)
		if result != tc.expected {
			t.Errorf("Test case %d: Expected animation '%s', got '%s'", i, tc.expected, result)
		}
	}
}

func TestLLMBackend_ClassifyResponse(t *testing.T) {
	backend := NewLLMBackend()

	testCases := []struct {
		response string
		expected string
	}{
		{
			response: "I love you so much! ❤️",
			expected: "romantic",
		},
		{
			response: "Let me help you with that",
			expected: "helpful",
		},
		{
			response: "What would you like to do?",
			expected: "inquisitive",
		},
		{
			response: "Just a normal chat",
			expected: "casual",
		},
	}

	for i, tc := range testCases {
		result := backend.classifyResponse(tc.response)
		if result != tc.expected {
			t.Errorf("Test case %d: Expected type '%s', got '%s'", i, tc.expected, result)
		}
	}
}

func TestLLMBackend_DetectEmotionalTone(t *testing.T) {
	backend := NewLLMBackend()

	testCases := []struct {
		response string
		expected string
	}{
		{
			response: "Wow, that's exciting!",
			expected: "excited", // Should be excited due to "!"
		},
		{
			response: "I'm happy to see you",
			expected: "happy", // Remove emoji to test "happy" keyword
		},
		{
			response: "I'm a bit shy about this...",
			expected: "shy",
		},
		{
			response: "Just a regular response",
			expected: "neutral",
		},
	}

	for i, tc := range testCases {
		result := backend.detectEmotionalTone(tc.response)
		if result != tc.expected {
			t.Errorf("Test case %d: Expected tone '%s', got '%s'", i, tc.expected, result)
		}
	}
}

func TestLLMBackend_ExtractTopics(t *testing.T) {
	backend := NewLLMBackend()

	testCases := []struct {
		response string
		expected []string
	}{
		{
			response: "Let's get some food to eat!",
			expected: []string{"food"},
		},
		{
			response: "Want to play a game?",
			expected: []string{"gaming"},
		},
		{
			response: "I love you with all my heart",
			expected: []string{"romance"},
		},
		{
			response: "Time to get back to work and study",
			expected: []string{"work", "study"},
		},
		{
			response: "Just a normal chat",
			expected: []string{},
		},
	}

	for i, tc := range testCases {
		result := backend.extractTopics(tc.response)
		if len(result) != len(tc.expected) {
			t.Errorf("Test case %d: Expected %d topics, got %d", i, len(tc.expected), len(result))
			continue
		}

		for _, expected := range tc.expected {
			found := false
			for _, topic := range result {
				if topic == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Test case %d: Expected topic '%s' not found in %v", i, expected, result)
			}
		}
	}
}

func TestLLMBackend_CreateFallbackResponse(t *testing.T) {
	backend := NewLLMBackend()

	context := DialogContext{
		Trigger:           "click",
		FallbackResponses: []string{"Fallback 1", "Fallback 2"},
		FallbackAnimation: "idle",
	}

	response := backend.createFallbackResponse(context)

	if response.Text == "" {
		t.Error("Fallback response text should not be empty")
	}

	if response.Animation != "talking" {
		t.Errorf("Expected fallback animation 'talking', got '%s'", response.Animation)
	}

	if response.Confidence >= 0.5 {
		t.Errorf("Fallback response should have low confidence, got %f", response.Confidence)
	}

	if response.ResponseType != "fallback" {
		t.Errorf("Expected response type 'fallback', got '%s'", response.ResponseType)
	}
}

func TestLLMBackend_Close(t *testing.T) {
	backend := NewLLMBackend()

	// Initialize first
	config := LLMConfig{ModelPath: "/path/to/model.gguf"}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	if !backend.initialized {
		t.Error("Backend should be initialized before testing Close")
	}

	err := backend.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}

	if backend.initialized {
		t.Error("Backend should not be initialized after Close")
	}

	if backend.model != nil {
		t.Error("Model should be nil after Close")
	}
}
