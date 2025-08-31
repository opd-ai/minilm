package dialog

import (
	"strings"
	"testing"
)

// Debug test to find the exact matching issue
func TestDebugPromptMatching(t *testing.T) {
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

	prompt := strings.ToLower(fullPrompt)

	t.Logf("Full prompt (lowercase): %s", prompt)

	// Check each condition (updated to match the fixed code)
	conditions := []struct {
		name  string
		check bool
	}{
		{"hello|hi", strings.Contains(prompt, "hello") || strings.Contains(prompt, "hi")},
		{"fed you|feed|food", strings.Contains(prompt, "fed you") || strings.Contains(prompt, "feed") || strings.Contains(prompt, "food")},
		{"petted you|pat", strings.Contains(prompt, "petted you") || strings.Contains(prompt, "pat")},
		{"wants to talk|talk|chat", strings.Contains(prompt, "wants to talk") || strings.Contains(prompt, "talk") || strings.Contains(prompt, "chat")},
		{"clicked on you|click", strings.Contains(prompt, "clicked on you") || strings.Contains(prompt, "click")},
	}

	for _, cond := range conditions {
		if cond.check {
			t.Logf("MATCH: %s", cond.name)
		} else {
			t.Logf("NO MATCH: %s", cond.name)
		}
	}
}
