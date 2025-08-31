package dialog

import (
	"strings"
	"testing"
	"time"
)

// test_mock_context_bug_1 reproduces the bug where different triggers return identical responses
func TestMockContextBug1_DifferentTriggersReturnSameResponse(t *testing.T) {
	// Initialize mock model
	mockModel := NewMockLLMModel()
	err := mockModel.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize mock model: %v", err)
	}

	// Test different triggers that should produce different responses
	testCases := []struct {
		trigger          string
		expectedKeyword  string // keyword that should be in the response for this trigger
		shouldNotContain string // default response text that shouldn't appear
	}{
		{"feed", "meal", "Hi there! How are you doing today?"},
		{"pet", "wonderful", "Hi there! How are you doing today?"},
		{"talk", "chat", "Hi there! How are you doing today?"},
		{"click", "", ""}, // click can have the default greeting
	}

	pb := NewPromptBuilder()

	for _, tc := range testCases {
		t.Run("trigger_"+tc.trigger, func(t *testing.T) {
			// Set up context for this trigger
			context := DialogContext{
				Trigger:       tc.trigger,
				InteractionID: "test-session",
				Timestamp:     time.Now(),
				CurrentStats: map[string]float64{
					"happiness": 75,
					"energy":    75,
					"trust":     60,
				},
				PersonalityTraits: map[string]float64{
					"cheerful":   0.9,
					"supportive": 0.8,
				},
				CurrentMood: 75,
			}

			// Build prompt
			pb.AddContext(context)
			prompt := pb.Build()

			// Get response from mock model
			response, err := mockModel.Predict(prompt)
			if err != nil {
				t.Fatalf("Mock model prediction failed: %v", err)
			}

			t.Logf("Trigger: %s", tc.trigger)
			t.Logf("Prompt: %s", prompt)
			t.Logf("Response: %s", response)

			// Check if the response is appropriate for the trigger
			if tc.expectedKeyword != "" {
				if !strings.Contains(strings.ToLower(response), tc.expectedKeyword) {
					t.Errorf("Expected response for trigger '%s' to contain '%s', but got: %s",
						tc.trigger, tc.expectedKeyword, response)
				}
			}

			// Check that non-contextual triggers don't get the default greeting
			if tc.shouldNotContain != "" && tc.trigger != "click" {
				if strings.Contains(response, tc.shouldNotContain) {
					t.Errorf("Response for trigger '%s' should not contain default greeting '%s', but got: %s",
						tc.trigger, tc.shouldNotContain, response)
				}
			}
		})
	}
}
