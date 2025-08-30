package dialog

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDialogManager_NewDialogManager(t *testing.T) {
	dm := NewDialogManager(true)

	if dm == nil {
		t.Fatal("NewDialogManager should return a non-nil manager")
	}

	if len(dm.backends) != 0 {
		t.Error("New dialog manager should have no backends registered")
	}

	if dm.defaultBackend != "" {
		t.Error("New dialog manager should have no default backend")
	}

	if len(dm.fallbackChain) != 0 {
		t.Error("New dialog manager should have empty fallback chain")
	}

	if !dm.debug {
		t.Error("Debug mode should be enabled when specified")
	}
}

func TestDialogManager_RegisterBackend(t *testing.T) {
	dm := NewDialogManager(false)
	backend := NewLLMBackend()

	dm.RegisterBackend("test_backend", backend)

	if len(dm.backends) != 1 {
		t.Error("Should have one backend registered")
	}

	retrievedBackend, exists := dm.GetBackend("test_backend")
	if !exists {
		t.Error("Should be able to retrieve registered backend")
	}

	if retrievedBackend != backend {
		t.Error("Retrieved backend should be the same as registered")
	}
}

func TestDialogManager_SetDefaultBackend(t *testing.T) {
	dm := NewDialogManager(false)
	backend := NewLLMBackend()

	// Should fail with unregistered backend
	err := dm.SetDefaultBackend("nonexistent")
	if err == nil {
		t.Error("Should fail to set nonexistent backend as default")
	}

	// Should succeed with registered backend
	dm.RegisterBackend("test_backend", backend)
	err = dm.SetDefaultBackend("test_backend")
	if err != nil {
		t.Errorf("Should succeed setting registered backend as default: %v", err)
	}

	if dm.defaultBackend != "test_backend" {
		t.Error("Default backend should be set correctly")
	}
}

func TestDialogManager_SetFallbackChain(t *testing.T) {
	dm := NewDialogManager(false)
	backend1 := NewLLMBackend()
	backend2 := NewLLMBackend()

	dm.RegisterBackend("backend1", backend1)
	dm.RegisterBackend("backend2", backend2)

	// Should succeed with registered backends
	err := dm.SetFallbackChain([]string{"backend1", "backend2"})
	if err != nil {
		t.Errorf("Should succeed setting fallback chain with registered backends: %v", err)
	}

	// Should fail with unregistered backend
	err = dm.SetFallbackChain([]string{"backend1", "nonexistent"})
	if err == nil {
		t.Error("Should fail setting fallback chain with unregistered backend")
	}
}

func TestDialogManager_GenerateDialog(t *testing.T) {
	dm := NewDialogManager(false)

	// Create and initialize a backend
	backend := NewLLMBackend()
	config := LLMConfig{ModelPath: "/fake/path.gguf"}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	dm.RegisterBackend("test_backend", backend)
	dm.SetDefaultBackend("test_backend")

	context := DialogContext{
		Trigger:           "click",
		InteractionID:     "test-1",
		Timestamp:         time.Now(),
		FallbackResponses: []string{"Default response"},
		FallbackAnimation: "talking",
	}

	response, err := dm.GenerateDialog(context)
	if err != nil {
		t.Errorf("Should generate dialog successfully: %v", err)
	}

	if response.Text == "" {
		t.Error("Generated response should have text")
	}

	if response.Animation == "" {
		t.Error("Generated response should have animation")
	}
}

func TestDialogManager_GenerateDialogFallback(t *testing.T) {
	dm := NewDialogManager(false)

	// No backends registered - should use fallback
	context := DialogContext{
		Trigger:           "click",
		InteractionID:     "test-1",
		Timestamp:         time.Now(),
		FallbackResponses: []string{"Fallback response"},
		FallbackAnimation: "idle",
	}

	response, err := dm.GenerateDialog(context)
	if err != nil {
		t.Errorf("Should generate fallback dialog: %v", err)
	}

	if response.ResponseType != "fallback" {
		t.Error("Response should be marked as fallback")
	}

	if response.Animation != "idle" {
		t.Errorf("Expected fallback animation 'idle', got '%s'", response.Animation)
	}
}

func TestDialogManager_GetRegisteredBackends(t *testing.T) {
	dm := NewDialogManager(false)
	backend1 := NewLLMBackend()
	backend2 := NewLLMBackend()

	if len(dm.GetRegisteredBackends()) != 0 {
		t.Error("Should have no registered backends initially")
	}

	dm.RegisterBackend("backend1", backend1)
	dm.RegisterBackend("backend2", backend2)

	backends := dm.GetRegisteredBackends()
	if len(backends) != 2 {
		t.Errorf("Should have 2 registered backends, got %d", len(backends))
	}

	// Check that both backends are in the list
	found1, found2 := false, false
	for _, name := range backends {
		if name == "backend1" {
			found1 = true
		}
		if name == "backend2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Should find both registered backend names")
	}
}

func TestDialogManager_GetBackendInfo(t *testing.T) {
	dm := NewDialogManager(false)
	backend := NewLLMBackend()
	dm.RegisterBackend("test_backend", backend)

	// Should succeed for registered backend
	info, err := dm.GetBackendInfo("test_backend")
	if err != nil {
		t.Errorf("Should get backend info successfully: %v", err)
	}

	if info.Name != "llm_backend" {
		t.Errorf("Expected backend name 'llm_backend', got '%s'", info.Name)
	}

	// Should fail for unregistered backend
	_, err = dm.GetBackendInfo("nonexistent")
	if err == nil {
		t.Error("Should fail to get info for nonexistent backend")
	}
}

func TestDialogManager_UpdateBackendMemory(t *testing.T) {
	dm := NewDialogManager(false)
	backend := NewLLMBackend()

	// Initialize backend
	config := LLMConfig{ModelPath: "/fake/path.gguf"}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	dm.RegisterBackend("test_backend", backend)

	context := DialogContext{
		Trigger:       "click",
		InteractionID: "test-1",
	}

	response := DialogResponse{
		Text:       "Test response",
		Confidence: 0.8,
	}

	feedback := &UserFeedback{
		Positive:   true,
		Engagement: 0.9,
	}

	// Should not panic or error
	dm.UpdateBackendMemory(context, response, feedback)
}

func TestValidateBackendConfig(t *testing.T) {
	testCases := []struct {
		name      string
		config    DialogBackendConfig
		shouldErr bool
	}{
		{
			name: "valid config",
			config: DialogBackendConfig{
				DefaultBackend:      "test_backend",
				Enabled:             true,
				ConfidenceThreshold: 0.5,
				ResponseTimeout:     1000,
			},
			shouldErr: false,
		},
		{
			name: "disabled config",
			config: DialogBackendConfig{
				Enabled: false,
				// Other fields can be invalid when disabled
			},
			shouldErr: false,
		},
		{
			name: "missing default backend",
			config: DialogBackendConfig{
				Enabled: true,
				// DefaultBackend is missing
			},
			shouldErr: true,
		},
		{
			name: "invalid confidence threshold",
			config: DialogBackendConfig{
				DefaultBackend:      "test_backend",
				Enabled:             true,
				ConfidenceThreshold: 1.5, // Invalid: > 1
			},
			shouldErr: true,
		},
		{
			name: "negative response timeout",
			config: DialogBackendConfig{
				DefaultBackend:  "test_backend",
				Enabled:         true,
				ResponseTimeout: -100, // Invalid: negative
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBackendConfig(tc.config)
			if tc.shouldErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestLoadDialogBackendConfig(t *testing.T) {
	// Test valid config
	validJSON := `{
		"enabled": true,
		"defaultBackend": "llm",
		"fallbackChain": ["markov"],
		"confidenceThreshold": 0.7,
		"responseTimeout": 2000,
		"memoryEnabled": true,
		"learningEnabled": false,
		"debugMode": true,
		"backends": {
			"llm": {"modelPath": "/path/to/model.gguf"}
		}
	}`

	config, err := LoadDialogBackendConfig([]byte(validJSON))
	if err != nil {
		t.Fatalf("Should load valid config successfully: %v", err)
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

	if config.ResponseTimeout != 2000 {
		t.Errorf("Expected response timeout 2000, got %d", config.ResponseTimeout)
	}

	// Test config with defaults
	minimalJSON := `{"enabled": true, "defaultBackend": "test"}`
	config, err = LoadDialogBackendConfig([]byte(minimalJSON))
	if err != nil {
		t.Fatalf("Should load minimal config with defaults: %v", err)
	}

	if config.ConfidenceThreshold != 0.5 {
		t.Errorf("Expected default confidence threshold 0.5, got %f", config.ConfidenceThreshold)
	}

	if config.ResponseTimeout != 1000 {
		t.Errorf("Expected default response timeout 1000, got %d", config.ResponseTimeout)
	}

	if !config.MemoryEnabled {
		t.Error("Memory should be enabled by default")
	}

	// Test invalid JSON
	invalidJSON := `{"enabled": true, "defaultBackend":}`
	_, err = LoadDialogBackendConfig([]byte(invalidJSON))
	if err == nil {
		t.Error("Should fail to load invalid JSON")
	}

	// Test invalid config
	invalidConfigJSON := `{"enabled": true}` // Missing required defaultBackend
	_, err = LoadDialogBackendConfig([]byte(invalidConfigJSON))
	if err == nil {
		t.Error("Should fail validation for invalid config")
	}
}

func TestInteractionRecord(t *testing.T) {
	record := InteractionRecord{
		Type:      "click",
		Response:  "Hello!",
		Timestamp: time.Now(),
		Stats:     map[string]float64{"happiness": 80},
		Outcome:   "positive",
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Should marshal InteractionRecord: %v", err)
	}

	var unmarshaled InteractionRecord
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Should unmarshal InteractionRecord: %v", err)
	}

	if unmarshaled.Type != record.Type {
		t.Errorf("Expected type '%s', got '%s'", record.Type, unmarshaled.Type)
	}

	if unmarshaled.Response != record.Response {
		t.Errorf("Expected response '%s', got '%s'", record.Response, unmarshaled.Response)
	}

	if unmarshaled.Outcome != record.Outcome {
		t.Errorf("Expected outcome '%s', got '%s'", record.Outcome, unmarshaled.Outcome)
	}
}

func TestUserFeedback(t *testing.T) {
	feedback := UserFeedback{
		Positive:     true,
		ResponseTime: 5 * time.Second,
		FollowUpType: "click",
		Engagement:   0.8,
		CustomData:   map[string]interface{}{"test": "value"},
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(feedback)
	if err != nil {
		t.Fatalf("Should marshal UserFeedback: %v", err)
	}

	var unmarshaled UserFeedback
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Should unmarshal UserFeedback: %v", err)
	}

	if unmarshaled.Positive != feedback.Positive {
		t.Errorf("Expected positive %t, got %t", feedback.Positive, unmarshaled.Positive)
	}

	if unmarshaled.Engagement != feedback.Engagement {
		t.Errorf("Expected engagement %f, got %f", feedback.Engagement, unmarshaled.Engagement)
	}
}
