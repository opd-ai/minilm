package dialog

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPromptBuilder_Build(t *testing.T) {
	pb := NewPromptBuilder()

	// Set up test data
	pb.AddPersonality("cheerful and energetic")
	pb.AddHistory([]ConversationExchange{
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Trigger:   "click",
			Response:  "Hello there!",
		},
		{
			Timestamp: time.Now().Add(-2 * time.Minute),
			Trigger:   "feed",
			Response:  "Thanks for the food!",
		},
	})

	context := DialogContext{
		Trigger:     "click",
		CurrentMood: 80,
		TimeOfDay:   "morning",
		PersonalityTraits: map[string]float64{
			"cheerful":  0.9,
			"energetic": 0.8,
			"friendly":  0.7,
		},
		CurrentAnimation: "idle",
	}
	pb.AddContext(context)

	prompt := pb.Build()

	// Verify key components are included
	if !strings.Contains(prompt, "cheerful and energetic") {
		t.Error("Prompt should contain personality description")
	}

	if !strings.Contains(prompt, "morning") {
		t.Error("Prompt should contain time of day")
	}

	if !strings.Contains(prompt, "80") {
		t.Error("Prompt should contain mood information")
	}

	if !strings.Contains(prompt, "Hello there!") {
		t.Error("Prompt should contain conversation history")
	}

	if !strings.Contains(prompt, "clicked on you") {
		t.Error("Prompt should contain current trigger description")
	}

	if !strings.Contains(prompt, "Response guidelines") {
		t.Error("Prompt should contain response instructions")
	}
}

func TestPromptBuilder_BuildWithSystemPrompt(t *testing.T) {
	pb := NewPromptBuilder()
	pb.AddSystemPrompt("You are a helpful desktop assistant.")
	pb.AddPersonality("friendly")

	context := DialogContext{Trigger: "click"}
	pb.AddContext(context)

	prompt := pb.Build()

	if !strings.Contains(prompt, "You are a helpful desktop assistant.") {
		t.Error("Prompt should contain system prompt")
	}

	// System prompt should appear before personality
	systemIndex := strings.Index(prompt, "You are a helpful desktop assistant.")
	personalityIndex := strings.Index(prompt, "friendly")

	if systemIndex == -1 || personalityIndex == -1 {
		t.Fatal("Both system prompt and personality should be in the prompt")
	}

	if systemIndex > personalityIndex {
		t.Error("System prompt should appear before personality description")
	}
}

func TestPromptBuilder_BuildFromTemplate(t *testing.T) {
	pb := NewPromptBuilder()
	pb.AddPersonality("cheerful")
	pb.AddSystemPrompt("System instructions here")

	context := DialogContext{
		Trigger:     "click",
		CurrentMood: 75,
		TimeOfDay:   "afternoon",
	}
	pb.AddContext(context)

	template := "System: {systemPrompt}\nPersonality: {personality}\nMood: {mood}\nTrigger: {trigger}\nTime: {timeOfDay}"

	result := pb.BuildFromTemplate(template)

	expected := "System: System instructions here\nPersonality: cheerful\nMood: 75.0\nTrigger: click\nTime: afternoon"

	if result != expected {
		t.Errorf("Template replacement failed.\nExpected: %s\nGot: %s", expected, result)
	}
}

func TestPromptBuilder_BuildFromTemplateWithMissingVars(t *testing.T) {
	pb := NewPromptBuilder()
	// Don't set personality

	context := DialogContext{Trigger: "click"}
	pb.AddContext(context)

	template := "Personality: {personality}, Trigger: {trigger}"
	result := pb.BuildFromTemplate(template)

	// Should replace with empty string for missing personality
	expected := "Personality: , Trigger: click"
	if result != expected {
		t.Errorf("Template with missing vars failed.\nExpected: %s\nGot: %s", expected, result)
	}
}

func TestPromptBuilder_BuildCharacterState(t *testing.T) {
	pb := NewPromptBuilder()

	context := DialogContext{
		CurrentMood:       85,
		TimeOfDay:         "evening",
		RelationshipLevel: "friend",
		PersonalityTraits: map[string]float64{
			"cheerful":  0.9,
			"energetic": 0.8,
			"shy":       0.3, // Should be excluded (< 0.6)
		},
		CurrentAnimation: "happy",
	}
	pb.AddContext(context)

	state := pb.buildCharacterState()

	if !strings.Contains(state, "very happy") {
		t.Error("Character state should describe mood as 'very happy' for mood 85")
	}

	if !strings.Contains(state, "evening") {
		t.Error("Character state should contain time of day")
	}

	if !strings.Contains(state, "friend") {
		t.Error("Character state should contain relationship level")
	}

	if !strings.Contains(state, "cheerful") {
		t.Error("Character state should contain strong personality traits")
	}

	if strings.Contains(state, "shy") {
		t.Error("Character state should not contain weak personality traits")
	}

	if !strings.Contains(state, "happy") {
		t.Error("Character state should contain current animation")
	}
}

func TestPromptBuilder_BuildConversationHistory(t *testing.T) {
	pb := NewPromptBuilder()

	history := []ConversationExchange{
		{
			Timestamp: time.Now().Add(-10 * time.Minute),
			Trigger:   "click",
			Response:  "Hello!",
		},
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Trigger:   "feed",
			Response:  "Thanks for food!",
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Trigger:   "pet",
			Response:  "That feels nice!",
		},
	}
	pb.AddHistory(history)

	historyText := pb.buildConversationHistory()

	if !strings.Contains(historyText, "Recent conversation:") {
		t.Error("History should have header")
	}

	if !strings.Contains(historyText, "Hello!") {
		t.Error("History should contain first response")
	}

	if !strings.Contains(historyText, "Thanks for food!") {
		t.Error("History should contain second response")
	}

	if !strings.Contains(historyText, "That feels nice!") {
		t.Error("History should contain third response")
	}

	if !strings.Contains(historyText, "User click") {
		t.Error("History should contain trigger references")
	}
}

func TestPromptBuilder_BuildConversationHistoryEmpty(t *testing.T) {
	pb := NewPromptBuilder()
	// No history added

	historyText := pb.buildConversationHistory()

	if historyText != "" {
		t.Error("Empty history should return empty string")
	}
}

func TestPromptBuilder_BuildConversationHistoryLimited(t *testing.T) {
	pb := NewPromptBuilder()

	// Add more than 5 exchanges
	history := make([]ConversationExchange, 8)
	for i := 0; i < 8; i++ {
		history[i] = ConversationExchange{
			Timestamp: time.Now().Add(time.Duration(-i-1) * time.Minute), // Ensure proper ordering
			Trigger:   "click",
			Response:  fmt.Sprintf("Response %d", i), // Use numbers for clarity
		}
	}
	pb.AddHistory(history)

	historyText := pb.buildConversationHistory()

	// Should only include last 5 exchanges (3, 4, 5, 6, 7)
	if strings.Contains(historyText, "Response 0") || strings.Contains(historyText, "Response 1") || strings.Contains(historyText, "Response 2") {
		t.Error("History should not contain oldest exchanges beyond limit of 5")
	}

	if !strings.Contains(historyText, "Response 3") {
		t.Error("History should contain recent exchanges within limit")
	}
}

func TestPromptBuilder_BuildCurrentSituation(t *testing.T) {
	pb := NewPromptBuilder()

	context := DialogContext{
		Trigger:          "rightclick",
		ConversationTurn: 3,
		LastResponse:     "Previous response here",
	}
	pb.AddContext(context)

	situation := pb.buildCurrentSituation()

	if !strings.Contains(situation, "Current situation:") {
		t.Error("Situation should have header")
	}

	if !strings.Contains(situation, "right-clicked on you") {
		t.Error("Situation should describe the trigger")
	}

	if !strings.Contains(situation, "turn 3") {
		t.Error("Situation should mention conversation turn")
	}

	if !strings.Contains(situation, "Previous response here") {
		t.Error("Situation should include last response")
	}
}

func TestPromptBuilder_DescribeMood(t *testing.T) {
	pb := NewPromptBuilder()

	testCases := []struct {
		mood     float64
		expected string
	}{
		{90, "very happy"},
		{70, "happy"},
		{50, "neutral"},
		{30, "sad"},
		{10, "very sad"},
	}

	for _, tc := range testCases {
		result := pb.describeMood(tc.mood)
		if result != tc.expected {
			t.Errorf("Mood %.0f: expected '%s', got '%s'", tc.mood, tc.expected, result)
		}
	}
}

func TestPromptBuilder_DescribeTrigger(t *testing.T) {
	pb := NewPromptBuilder()

	testCases := []struct {
		trigger  string
		expected string
	}{
		{"click", "clicked on you"},
		{"rightclick", "right-clicked on you"},
		{"hover", "hovered over you"},
		{"feed", "fed you"},
		{"pet", "petted you"},
		{"play", "wants to play"},
		{"talk", "wants to talk"},
		{"gift", "gave you a gift"},
		{"unknown_trigger", "unknown_trigger"}, // Fallback
	}

	for _, tc := range testCases {
		result := pb.describeTrigger(tc.trigger)
		if result != tc.expected {
			t.Errorf("Trigger '%s': expected '%s', got '%s'", tc.trigger, tc.expected, result)
		}
	}
}

func TestPromptBuilder_FormatTimeAgo(t *testing.T) {
	pb := NewPromptBuilder()
	now := time.Now()

	testCases := []struct {
		timestamp time.Time
		expected  string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-1 * time.Minute), "1 minute ago"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-1 * time.Hour), "1 hour ago"},
		{now.Add(-3 * time.Hour), "3 hours ago"},
		{now.Add(-25 * time.Hour), "1 day ago"},
		{now.Add(-48 * time.Hour), "2 days ago"},
	}

	for _, tc := range testCases {
		result := pb.formatTimeAgo(tc.timestamp)
		if result != tc.expected {
			t.Errorf("Time ago for %v: expected '%s', got '%s'", tc.timestamp, tc.expected, result)
		}
	}
}

func TestPromptBuilder_SetMaxTokens(t *testing.T) {
	pb := NewPromptBuilder()
	pb.SetMaxTokens(100)

	pb.AddPersonality("cheerful")
	context := DialogContext{Trigger: "click"}
	pb.AddContext(context)

	// Create a very long personality description that would exceed token limit
	longPersonality := strings.Repeat("very very long personality description ", 50)
	pb.AddPersonality(longPersonality)

	prompt := pb.Build()

	// Rough token estimation: 1 token ≈ 4 characters
	maxChars := 100 * 4
	if len(prompt) > maxChars+100 { // Allow some margin
		t.Errorf("Prompt should be truncated to roughly %d characters, got %d", maxChars, len(prompt))
	}
}

func TestPromptBuilder_EstimateTokenCount(t *testing.T) {
	pb := NewPromptBuilder()

	testCases := []struct {
		text     string
		expected int
	}{
		{"hello", 1},       // 5 chars / 4 = 1
		{"hello world", 2}, // 11 chars / 4 = 2
		{"a much longer text that spans multiple words", 11}, // 46 chars / 4 = 11
	}

	for _, tc := range testCases {
		result := pb.EstimateTokenCount(tc.text)
		if result != tc.expected {
			t.Errorf("Token count for '%s': expected %d, got %d", tc.text, tc.expected, result)
		}
	}
}

func TestPromptBuilder_BuildResponseInstructions(t *testing.T) {
	pb := NewPromptBuilder()

	instructions := pb.buildResponseInstructions()

	expectedPhrases := []string{
		"Response guidelines:",
		"short and natural",
		"1-2 sentences",
		"personality",
		"desktop pet",
		"Your response:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(instructions, phrase) {
			t.Errorf("Instructions should contain '%s'", phrase)
		}
	}
}

func TestPromptBuilder_NoPersonalityDefault(t *testing.T) {
	pb := NewPromptBuilder()
	// Don't add personality

	context := DialogContext{Trigger: "click"}
	pb.AddContext(context)

	prompt := pb.Build()

	if !strings.Contains(prompt, "friendly desktop pet character") {
		t.Error("Prompt should contain default personality description when none is provided")
	}
}
