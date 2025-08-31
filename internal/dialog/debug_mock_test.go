package dialog

import (
	"testing"
)

// Debug test to understand the mock model matching behavior
func TestDebugMockModelMatching(t *testing.T) {
	mockModel := NewMockLLMModel()
	err := mockModel.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize mock model: %v", err)
	}

	// Test direct keyword matching
	testPrompts := []string{
		"wants to talk",
		"clicked on you",
		"petted you",
		"fed you",
		"Your response:",
	}

	for _, prompt := range testPrompts {
		response, err := mockModel.Predict(prompt)
		if err != nil {
			t.Errorf("Error for prompt '%s': %v", prompt, err)
		} else {
			t.Logf("Prompt: '%s' -> Response: '%s'", prompt, response)
		}
	}

	// Test with full prompt from actual scenario
	fullPrompt := `You are a friendly desktop pet character.
Current character state:
- Mood: happy (75.0/100)
- Key traits: cheerful (0.9), supportive (0.8)

Current situation:
- The user just performed: wants to talk

Response guidelines:
- Keep responses short and natural (1-2 sentences maximum)
- Match your personality and current mood
- Respond appropriately to the user's action
- Use simple, conversational language
- Include an emoji if it fits naturally
- Stay in character as a desktop pet

Your response:`

	response, err := mockModel.Predict(fullPrompt)
	if err != nil {
		t.Errorf("Error for full prompt: %v", err)
	} else {
		t.Logf("Full prompt response: '%s'", response)
	}
}
