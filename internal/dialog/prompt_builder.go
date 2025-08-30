package dialog

import (
	"fmt"
	"strings"
	"time"
)

// PromptBuilder constructs prompts for LLM inference from dialog context
// Designed to create effective prompts for small models with limited context windows
type PromptBuilder struct {
	systemPrompt string
	personality  string
	history      []ConversationExchange
	context      DialogContext
	template     string
	maxTokens    int
}

// NewPromptBuilder creates a new prompt builder with default settings
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		maxTokens: 1500, // Conservative limit for small models
	}
}

// AddSystemPrompt sets the system-level instructions for the character
func (pb *PromptBuilder) AddSystemPrompt(prompt string) {
	pb.systemPrompt = prompt
}

// AddPersonality adds personality description to the prompt
func (pb *PromptBuilder) AddPersonality(personality string) {
	pb.personality = personality
}

// AddHistory includes recent conversation exchanges for context
func (pb *PromptBuilder) AddHistory(history []ConversationExchange) {
	pb.history = history
}

// AddContext includes the current dialog context
func (pb *PromptBuilder) AddContext(context DialogContext) {
	pb.context = context
}

// SetTemplate allows using a custom prompt template
func (pb *PromptBuilder) SetTemplate(template string) {
	pb.template = template
}

// SetMaxTokens limits the total prompt length
func (pb *PromptBuilder) SetMaxTokens(maxTokens int) {
	pb.maxTokens = maxTokens
}

// Build constructs the final prompt using default structure
func (pb *PromptBuilder) Build() string {
	var prompt strings.Builder

	// Add system prompt if available
	if pb.systemPrompt != "" {
		prompt.WriteString(pb.systemPrompt)
		prompt.WriteString("\n\n")
	}

	// Add character personality
	if pb.personality != "" {
		prompt.WriteString(fmt.Sprintf("You are a desktop pet character with the following personality: %s\n", pb.personality))
	} else {
		prompt.WriteString("You are a friendly desktop pet character.\n")
	}

	// Add character state context
	prompt.WriteString(pb.buildCharacterState())

	// Add conversation history if available
	if len(pb.history) > 0 {
		prompt.WriteString(pb.buildConversationHistory())
	}

	// Add current situation
	prompt.WriteString(pb.buildCurrentSituation())

	// Add response instructions
	prompt.WriteString(pb.buildResponseInstructions())

	result := prompt.String()

	// Truncate if too long (rough token estimation: 1 token ≈ 4 characters)
	if len(result) > pb.maxTokens*4 {
		result = result[:pb.maxTokens*4]
		// Try to end at a reasonable point
		if lastNewline := strings.LastIndex(result, "\n"); lastNewline > len(result)-100 {
			result = result[:lastNewline]
		}
	}

	return result
}

// BuildFromTemplate constructs the prompt using a custom template
func (pb *PromptBuilder) BuildFromTemplate(template string) string {
	result := template

	// Replace template variables
	replacements := map[string]string{
		"{personality}":          pb.personality,
		"{systemPrompt}":         pb.systemPrompt,
		"{characterState}":       pb.buildCharacterState(),
		"{conversationHistory}":  pb.buildConversationHistory(),
		"{currentSituation}":     pb.buildCurrentSituation(),
		"{responseInstructions}": pb.buildResponseInstructions(),
		"{trigger}":              pb.context.Trigger,
		"{mood}":                 fmt.Sprintf("%.1f", pb.context.CurrentMood),
		"{timeOfDay}":            pb.context.TimeOfDay,
		"{relationshipLevel}":    pb.context.RelationshipLevel,
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// buildCharacterState creates a description of the character's current state
func (pb *PromptBuilder) buildCharacterState() string {
	var state strings.Builder

	state.WriteString("Current character state:\n")

	pb.addMoodInfo(&state)
	pb.addTimeInfo(&state)
	pb.addRelationshipInfo(&state)
	pb.addPersonalityTraits(&state)
	pb.addAnimationInfo(&state)

	state.WriteString("\n")
	return state.String()
}

// addMoodInfo adds mood information to the character state
func (pb *PromptBuilder) addMoodInfo(state *strings.Builder) {
	if pb.context.CurrentMood > 0 {
		moodDesc := pb.describeMood(pb.context.CurrentMood)
		state.WriteString(fmt.Sprintf("- Mood: %s (%.1f/100)\n", moodDesc, pb.context.CurrentMood))
	}
}

// addTimeInfo adds time context to the character state
func (pb *PromptBuilder) addTimeInfo(state *strings.Builder) {
	if pb.context.TimeOfDay != "" {
		state.WriteString(fmt.Sprintf("- Time of day: %s\n", pb.context.TimeOfDay))
	}
}

// addRelationshipInfo adds relationship context to the character state
func (pb *PromptBuilder) addRelationshipInfo(state *strings.Builder) {
	if pb.context.RelationshipLevel != "" {
		state.WriteString(fmt.Sprintf("- Relationship level: %s\n", pb.context.RelationshipLevel))
	}
}

// addPersonalityTraits adds key personality traits to the character state
func (pb *PromptBuilder) addPersonalityTraits(state *strings.Builder) {
	if len(pb.context.PersonalityTraits) > 0 {
		state.WriteString("- Key traits: ")
		traits := pb.extractTopTraits()
		state.WriteString(strings.Join(traits, ", "))
		state.WriteString("\n")
	}
}

// extractTopTraits extracts the top 3 personality traits above threshold
func (pb *PromptBuilder) extractTopTraits() []string {
	traits := make([]string, 0, 3)
	for trait, value := range pb.context.PersonalityTraits {
		if value > 0.6 { // Only include strong traits
			traits = append(traits, fmt.Sprintf("%s (%.1f)", trait, value))
			if len(traits) >= 3 {
				break
			}
		}
	}
	return traits
}

// addAnimationInfo adds current animation state to the character state
func (pb *PromptBuilder) addAnimationInfo(state *strings.Builder) {
	if pb.context.CurrentAnimation != "" {
		state.WriteString(fmt.Sprintf("- Current animation: %s\n", pb.context.CurrentAnimation))
	}
}

// buildConversationHistory creates a summary of recent conversation
func (pb *PromptBuilder) buildConversationHistory() string {
	if len(pb.history) == 0 {
		return ""
	}

	var history strings.Builder
	history.WriteString("Recent conversation:\n")

	// Include up to 5 most recent exchanges for context
	start := 0
	if len(pb.history) > 5 {
		start = len(pb.history) - 5
	}

	for i := start; i < len(pb.history); i++ {
		exchange := pb.history[i]
		timeAgo := pb.formatTimeAgo(exchange.Timestamp)
		history.WriteString(fmt.Sprintf("- %s (%s): User %s → You said: \"%s\"\n",
			timeAgo, exchange.Trigger, exchange.Trigger, exchange.Response))
	}

	history.WriteString("\n")
	return history.String()
}

// buildCurrentSituation describes what just happened to trigger this response
func (pb *PromptBuilder) buildCurrentSituation() string {
	var situation strings.Builder

	situation.WriteString("Current situation:\n")
	situation.WriteString(fmt.Sprintf("- The user just performed: %s\n", pb.describeTrigger(pb.context.Trigger)))

	// Add turn information if this is part of an ongoing conversation
	if pb.context.ConversationTurn > 1 {
		situation.WriteString(fmt.Sprintf("- This is turn %d of the current conversation\n", pb.context.ConversationTurn))
	}

	// Add last response context if available
	if pb.context.LastResponse != "" {
		situation.WriteString(fmt.Sprintf("- Your last response was: \"%s\"\n", pb.context.LastResponse))
	}

	situation.WriteString("\n")
	return situation.String()
}

// buildResponseInstructions provides guidance for generating appropriate responses
func (pb *PromptBuilder) buildResponseInstructions() string {
	instructions := `Response guidelines:
- Keep responses short and natural (1-2 sentences maximum)
- Match your personality and current mood
- Respond appropriately to the user's action
- Use simple, conversational language
- Include an emoji if it fits naturally
- Stay in character as a desktop pet

Your response:`

	return instructions
}

// describeMood converts numeric mood to descriptive text
func (pb *PromptBuilder) describeMood(mood float64) string {
	switch {
	case mood >= 80:
		return "very happy"
	case mood >= 60:
		return "happy"
	case mood >= 40:
		return "neutral"
	case mood >= 20:
		return "sad"
	default:
		return "very sad"
	}
}

// describeTrigger converts trigger codes to natural language
func (pb *PromptBuilder) describeTrigger(trigger string) string {
	triggers := map[string]string{
		"click":      "clicked on you",
		"rightclick": "right-clicked on you",
		"hover":      "hovered over you",
		"feed":       "fed you",
		"pet":        "petted you",
		"play":       "wants to play",
		"talk":       "wants to talk",
		"gift":       "gave you a gift",
		"compliment": "complimented you",
		"ignore":     "ignored you",
		"idle":       "you've been idle",
		"timer":      "time passed",
	}

	if description, exists := triggers[trigger]; exists {
		return description
	}
	return trigger // Fallback to original trigger name
}

// formatTimeAgo converts timestamp to relative time description
func (pb *PromptBuilder) formatTimeAgo(timestamp time.Time) string {
	duration := time.Since(timestamp)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// EstimateTokenCount provides a rough estimate of token count for the prompt
// Uses a simple heuristic: 1 token ≈ 4 characters
func (pb *PromptBuilder) EstimateTokenCount(text string) int {
	return len(text) / 4
}
