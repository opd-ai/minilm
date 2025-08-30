package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"minilm/internal/dialog"
)

// Example demonstrates how to use the LLM-powered dialog system
func main() {
	fmt.Println("MiniLM Dialog System Example")
	fmt.Println("============================")

	// Initialize and configure the dialog system
	manager, llmBackend := setupDialogSystem()

	// Run interaction demonstrations
	runInteractionExamples(manager)

	// Show system information
	displaySystemInfo(manager)

	// Clean up resources
	cleanupResources(llmBackend)

	fmt.Println("Example completed successfully!")
}

// setupDialogSystem initializes and configures the dialog manager and LLM backend
func setupDialogSystem() (*dialog.DialogManager, *dialog.LLMBackend) {
	manager := dialog.NewDialogManager(true) // Enable debug mode
	llmBackend := dialog.NewLLMBackend()

	config := createLLMConfig()
	initializeBackend(llmBackend, config)
	registerBackend(manager, llmBackend)

	fmt.Println("Backend initialized successfully!")
	fmt.Printf("Backend info: %+v\n\n", llmBackend.GetBackendInfo())

	return manager, llmBackend
}

// createLLMConfig creates the LLM configuration with training data and settings
func createLLMConfig() dialog.LLMConfig {
	return dialog.LLMConfig{
		ModelPath:   "/path/to/your/model.gguf", // Replace with actual model path
		MaxTokens:   50,                         // Short responses for desktop pets
		Temperature: 0.8,                        // Slightly creative responses
		TopP:        0.9,
		ContextSize: 2048,
		Threads:     4,
		MarkovConfig: dialog.MarkovChainConfig{
			TrainingData: []string{
				"Hello there! I'm so happy to see you again! 😊",
				"How are you doing today? I've been thinking about you!",
				"Your company means everything to me! I'm so grateful.",
				"Thanks for being such a great friend! I appreciate you so much.",
				"What would you like to do together? I'm here for you!",
			},
		},
		MaxHistoryLength: 5,
		TimeoutMs:        2000,
		FallbackEnabled:  true,
	}
}

// initializeBackend initializes the LLM backend with the provided configuration
func initializeBackend(llmBackend *dialog.LLMBackend, config dialog.LLMConfig) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	err = llmBackend.Initialize(configJSON)
	if err != nil {
		log.Fatalf("Failed to initialize LLM backend: %v", err)
	}
}

// registerBackend registers the LLM backend with the dialog manager
func registerBackend(manager *dialog.DialogManager, llmBackend *dialog.LLMBackend) {
	manager.RegisterBackend("llm", llmBackend)
	manager.SetDefaultBackend("llm")
}

// runInteractionExamples demonstrates various user interactions with the dialog system
func runInteractionExamples(manager *dialog.DialogManager) {
	interactions := defineInteractionScenarios()

	for i, interaction := range interactions {
		fmt.Printf("=== Interaction %d: %s ===\n", i+1, interaction.description)

		context := buildDialogContext(interaction, i)
		processInteraction(manager, context)

		fmt.Println()
		time.Sleep(100 * time.Millisecond) // Brief pause between interactions
	}
}

// defineInteractionScenarios creates the list of interaction scenarios for demonstration
func defineInteractionScenarios() []struct {
	trigger     string
	description string
	mood        float64
} {
	return []struct {
		trigger     string
		description string
		mood        float64
	}{
		{"click", "User clicks on the pet", 75},
		{"feed", "User feeds the pet", 85},
		{"pet", "User pets the character", 90},
		{"talk", "User wants to have a conversation", 80},
		{"idle", "Pet has been idle for a while", 70},
	}
}

// buildDialogContext creates a dialog context for the given interaction scenario
func buildDialogContext(interaction struct {
	trigger     string
	description string
	mood        float64
}, turnIndex int) dialog.DialogContext {
	return dialog.DialogContext{
		Trigger:       interaction.trigger,
		InteractionID: "demo-session",
		Timestamp:     time.Now(),
		CurrentStats: map[string]float64{
			"happiness": interaction.mood,
			"energy":    75,
			"trust":     60,
		},
		PersonalityTraits: map[string]float64{
			"cheerful":   0.9,
			"supportive": 0.8,
			"playful":    0.7,
			"energetic":  0.6,
		},
		CurrentMood:       interaction.mood,
		CurrentAnimation:  "idle",
		TimeOfDay:         "afternoon",
		RelationshipLevel: "friend",
		ConversationTurn:  turnIndex + 1,
		FallbackResponses: []string{
			"Hello there! 👋",
			"How are you doing?",
			"Nice to see you!",
		},
		FallbackAnimation: "talking",
	}
}

// processInteraction handles a single interaction with the dialog system
func processInteraction(manager *dialog.DialogManager, context dialog.DialogContext) {
	response, err := manager.GenerateDialog(context)
	if err != nil {
		log.Printf("Error generating dialog: %v", err)
		return
	}

	displayInteractionResults(context, response)
	simulateUserFeedback(manager, context, response)
}

// displayInteractionResults shows the results of a dialog interaction
func displayInteractionResults(context dialog.DialogContext, response dialog.DialogResponse) {
	fmt.Printf("Trigger: %s\n", context.Trigger)
	fmt.Printf("Mood: %.1f/100\n", context.CurrentMood)
	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Animation: %s\n", response.Animation)
	fmt.Printf("Confidence: %.2f\n", response.Confidence)
	fmt.Printf("Type: %s\n", response.ResponseType)
	fmt.Printf("Emotional Tone: %s\n", response.EmotionalTone)
	if len(response.Topics) > 0 {
		fmt.Printf("Topics: %v\n", response.Topics)
	}
}

// simulateUserFeedback creates and applies user feedback for the interaction
func simulateUserFeedback(manager *dialog.DialogManager, context dialog.DialogContext, response dialog.DialogResponse) {
	feedback := &dialog.UserFeedback{
		Positive:     true,
		ResponseTime: time.Duration(2+context.ConversationTurn) * time.Second,
		FollowUpType: "positive_reaction",
		Engagement:   0.8 + float64(context.ConversationTurn)*0.05, // Increasing engagement
	}

	manager.UpdateBackendMemory(context, response, feedback)
}

// displaySystemInfo shows information about the dialog system configuration
func displaySystemInfo(manager *dialog.DialogManager) {
	fmt.Println("=== System Information ===")
	fmt.Printf("Registered backends: %v\n", manager.GetRegisteredBackends())

	if info, err := manager.GetBackendInfo("llm"); err == nil {
		fmt.Printf("LLM Backend: %s v%s\n", info.Name, info.Version)
		fmt.Printf("Description: %s\n", info.Description)
		fmt.Printf("Capabilities: %v\n", info.Capabilities)
		fmt.Printf("Author: %s\n", info.Author)
		fmt.Printf("License: %s\n", info.License)
	}
}

// cleanupResources performs cleanup operations for the dialog system
func cleanupResources(llmBackend *dialog.LLMBackend) {
	fmt.Println("\nCleaning up...")
	if closer, ok := interface{}(llmBackend).(interface{ Close() error }); ok {
		closer.Close()
	}
}

// ExampleBasicUsage demonstrates the simplest way to use the dialog system
func ExampleBasicUsage() {
	// Create and initialize backend
	backend := dialog.NewLLMBackend()
	config := dialog.LLMConfig{
		ModelPath: "/path/to/model.gguf",
		MarkovConfig: dialog.MarkovChainConfig{
			TrainingData: []string{
				"Hello! I'm friendly and helpful!",
				"How can I assist you today?",
				"Nice to meet you! I'm here to help.",
			},
		},
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	// Create dialog context
	context := dialog.DialogContext{
		Trigger:           "click",
		InteractionID:     "basic-example",
		Timestamp:         time.Now(),
		CurrentMood:       80,
		FallbackResponses: []string{"Hello!"},
	}

	// Generate response
	response, err := backend.GenerateResponse(context)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Animation: %s\n", response.Animation)
}

// ExampleWithCharacterJSON demonstrates loading character configuration from JSON
func ExampleWithCharacterJSON() {
	// Example character configuration that could be loaded from a character.json file
	characterConfig := `{
		"name": "Cheerful Companion",
		"personality": "upbeat, encouraging, and slightly mischievous",
		"dialogBackend": {
			"enabled": true,
			"defaultBackend": "llm",
			"backends": {
				"llm": {
					"modelPath": "/models/tinyllama-1.1b-q4.gguf",
					"maxTokens": 40,
					"temperature": 0.9,
					"personality": "upbeat, encouraging, and slightly mischievous",
					"promptTemplate": "You are a {personality} desktop companion. Current mood: {mood}/100. User just {trigger}. Respond briefly and stay in character:",
					"fallbackEnabled": true
				}
			},
			"confidenceThreshold": 0.6,
			"memoryEnabled": true
		}
	}`

	var charConfig struct {
		Name          string                     `json:"name"`
		Personality   string                     `json:"personality"`
		DialogBackend dialog.DialogBackendConfig `json:"dialogBackend"`
	}

	if err := json.Unmarshal([]byte(characterConfig), &charConfig); err != nil {
		log.Printf("Failed to parse character config: %v", err)
		return
	}

	fmt.Printf("Loaded character: %s\n", charConfig.Name)
	fmt.Printf("Personality: %s\n", charConfig.Personality)
	fmt.Printf("Dialog enabled: %t\n", charConfig.DialogBackend.Enabled)
	fmt.Printf("Default backend: %s\n", charConfig.DialogBackend.DefaultBackend)
}
