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

	// Create a dialog manager
	manager := dialog.NewDialogManager(true) // Enable debug mode

	// Create and configure an LLM backend
	llmBackend := dialog.NewLLMBackend()

	// Configure the LLM backend
	config := dialog.LLMConfig{
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

	configJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	// Initialize the backend
	err = llmBackend.Initialize(configJSON)
	if err != nil {
		log.Fatalf("Failed to initialize LLM backend: %v", err)
	}

	// Register the backend with the manager
	manager.RegisterBackend("llm", llmBackend)
	manager.SetDefaultBackend("llm")

	fmt.Println("Backend initialized successfully!")
	fmt.Printf("Backend info: %+v\n\n", llmBackend.GetBackendInfo())

	// Simulate various user interactions
	interactions := []struct {
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

	for i, interaction := range interactions {
		fmt.Printf("=== Interaction %d: %s ===\n", i+1, interaction.description)

		// Create dialog context
		context := dialog.DialogContext{
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
			ConversationTurn:  i + 1,
			FallbackResponses: []string{
				"Hello there! 👋",
				"How are you doing?",
				"Nice to see you!",
			},
			FallbackAnimation: "talking",
		}

		// Generate response
		response, err := manager.GenerateDialog(context)
		if err != nil {
			log.Printf("Error generating dialog: %v", err)
			continue
		}

		// Display results
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

		// Simulate user feedback (positive for demonstration)
		feedback := &dialog.UserFeedback{
			Positive:     true,
			ResponseTime: time.Duration(2+i) * time.Second,
			FollowUpType: "positive_reaction",
			Engagement:   0.8 + float64(i)*0.05, // Increasing engagement
		}

		// Update backend memory
		manager.UpdateBackendMemory(context, response, feedback)

		fmt.Println()
		time.Sleep(100 * time.Millisecond) // Brief pause between interactions
	}

	// Demonstrate configuration and backend info
	fmt.Println("=== System Information ===")
	fmt.Printf("Registered backends: %v\n", manager.GetRegisteredBackends())

	if info, err := manager.GetBackendInfo("llm"); err == nil {
		fmt.Printf("LLM Backend: %s v%s\n", info.Name, info.Version)
		fmt.Printf("Description: %s\n", info.Description)
		fmt.Printf("Capabilities: %v\n", info.Capabilities)
		fmt.Printf("Author: %s\n", info.Author)
		fmt.Printf("License: %s\n", info.License)
	}

	// Clean up
	fmt.Println("\nCleaning up...")
	if closer, ok := interface{}(llmBackend).(interface{ Close() error }); ok {
		closer.Close()
	}

	fmt.Println("Example completed successfully!")
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
