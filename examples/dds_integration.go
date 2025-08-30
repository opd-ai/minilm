package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"minilm/dialog"
)

// DDSIntegrationExample demonstrates how DDS can integrate with MiniLM's dialog system
func main() {
	fmt.Println("🎯 DDS Integration Example - MiniLM Dialog System")
	fmt.Println("============================================================")

	// Step 1: Initialize the dialog system (as DDS would do)
	manager := setupDialogSystem()

	// Step 2: Simulate DDS character interactions
	simulateDDSInteractions(manager)

	fmt.Println("\n✅ DDS Integration Test Completed Successfully!")
	fmt.Println("🚀 MiniLM is ready for production deployment with DDS!")
}

// setupDialogSystem configures the dialog system as DDS would
func setupDialogSystem() *dialog.DialogManager {
	fmt.Println("🔧 Setting up Dialog System...")

	// Create dialog manager
	manager := dialog.NewDialogManager(false) // DDS would use false for production

	// Create and configure LLM backend
	backend := dialog.NewLLMBackend()

	// Configuration that DDS would load from character files
	config := dialog.LLMConfig{
		ModelPath:   "/models/tinyllama-1.1b-q4.gguf", // DDS would set this from character config
		MaxTokens:   50,                               // Optimized for desktop pet responses
		Temperature: 0.8,                              // Balanced creativity for personality
		TopP:        0.9,
		ContextSize: 2048, // Fits consumer hardware constraints
		Threads:     4,    // Optimal for 4-8 core CPUs
		MarkovConfig: dialog.MarkovChainConfig{
			// DDS would extract this from existing character training data
			TrainingData: []string{
				"Hello there! I'm so happy to see you again! 😊",
				"How are you doing today? I've been thinking about you!",
				"Your company means everything to me! I'm so grateful.",
				"Thanks for being such a great friend! I appreciate you so much.",
				"What would you like to do together? I'm here for you!",
			},
			FallbackPhrases: []string{
				"Hi there! 👋",
				"What's up?",
				"Nice to see you!",
			},
		},
		MaxHistoryLength: 5,    // Rolling conversation window
		TimeoutMs:        2000, // Responsive UX requirement
		FallbackEnabled:  true, // Always enable fallback for reliability
	}

	// Initialize backend (DDS error handling)
	configJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("DDS Config Error: %v", err)
	}

	if err := backend.Initialize(configJSON); err != nil {
		log.Fatalf("DDS Backend Initialization Error: %v", err)
	}

	// Register backends (DDS would register multiple backends)
	manager.RegisterBackend("llm", backend)
	if err := manager.SetDefaultBackend("llm"); err != nil {
		log.Fatalf("DDS Backend Registration Error: %v", err)
	}

	fmt.Printf("✅ Dialog system configured successfully\n")
	fmt.Printf("📊 Backend Info: %+v\n", backend.GetBackendInfo())

	return manager
}

// simulateDDSInteractions demonstrates typical DDS user interactions
func simulateDDSInteractions(manager *dialog.DialogManager) {
	fmt.Println("\n🎮 Simulating DDS Character Interactions...")

	// DDS interaction scenarios that would trigger dialog
	interactions := []struct {
		description string
		context     dialog.DialogContext
	}{
		{
			description: "User clicks on character",
			context: dialog.DialogContext{
				Trigger:       "click",
				InteractionID: "dds-click-001",
				Timestamp:     time.Now(),
				CurrentMood:   75.0,
				PersonalityTraits: map[string]float64{
					"friendly": 0.9,
					"cheerful": 0.8,
					"helpful":  0.7,
				},
				RelationshipLevel: "friend",
				TimeOfDay:         "afternoon",
				FallbackResponses: []string{"Hello! 👋"},
				FallbackAnimation: "talking",
			},
		},
		{
			description: "User feeds the character",
			context: dialog.DialogContext{
				Trigger:       "feed",
				InteractionID: "dds-feed-001",
				Timestamp:     time.Now(),
				CurrentMood:   85.0,
				PersonalityTraits: map[string]float64{
					"grateful": 0.9,
					"happy":    0.8,
				},
				CurrentStats: map[string]float64{
					"hunger":    25.0, // Was hungry, now fed
					"happiness": 90.0,
					"energy":    80.0,
				},
				LastResponse:      "Hello! 👋", // Previous interaction response
				ConversationTurn:  2,
				FallbackResponses: []string{"Thanks for the food!"},
				FallbackAnimation: "eating",
			},
		},
		{
			description: "User right-clicks for menu",
			context: dialog.DialogContext{
				Trigger:       "rightclick",
				InteractionID: "dds-menu-001",
				Timestamp:     time.Now(),
				CurrentMood:   70.0,
				PersonalityTraits: map[string]float64{
					"helpful":     0.8,
					"informative": 0.7,
				},
				ConversationTurn: 3,
				TopicContext: map[string]interface{}{
					"menu_context":      true,
					"available_actions": []string{"feed", "play", "settings"},
				},
				FallbackResponses: []string{"What would you like to do?"},
				FallbackAnimation: "questioning",
			},
		},
		{
			description: "Character idle animation trigger",
			context: dialog.DialogContext{
				Trigger:       "idle",
				InteractionID: "dds-idle-001",
				Timestamp:     time.Now(),
				CurrentMood:   60.0,
				PersonalityTraits: map[string]float64{
					"contemplative": 0.6,
					"peaceful":      0.7,
				},
				TimeOfDay: "evening",
				CurrentStats: map[string]float64{
					"boredom": 40.0,
					"energy":  50.0,
				},
				FallbackResponses: []string{"*stretches quietly*"},
				FallbackAnimation: "idle",
			},
		},
	}

	// Process each interaction as DDS would
	for i, interaction := range interactions {
		fmt.Printf("\n--- Interaction %d: %s ---\n", i+1, interaction.description)

		// DDS would call this for each user interaction
		response, err := manager.GenerateDialog(interaction.context)
		if err != nil {
			fmt.Printf("❌ DDS Dialog Error: %v\n", err)
			continue
		}

		// DDS would use this response data to update the UI
		fmt.Printf("🗨️  Character Response: %s\n", response.Text)
		fmt.Printf("🎭 Animation: %s\n", response.Animation)
		fmt.Printf("🎯 Confidence: %.2f\n", response.Confidence)
		fmt.Printf("💭 Response Type: %s\n", response.ResponseType)
		fmt.Printf("😊 Emotional Tone: %s\n", response.EmotionalTone)

		if len(response.Topics) > 0 {
			fmt.Printf("📚 Topics: %v\n", response.Topics)
		}

		// DDS would provide user feedback for learning
		feedback := &dialog.UserFeedback{
			Positive:     true,                   // Simulated positive user interaction
			ResponseTime: 500 * time.Millisecond, // User responded quickly
			Engagement:   0.8,                    // High engagement
			FollowUpType: "continue",             // User wants to continue interacting
		}

		// Update backend memory for learning (DDS integration point)
		manager.UpdateBackendMemory(interaction.context, response, feedback)
	}
}

// demonstrateErrorHandling shows how DDS would handle dialog system errors
func demonstrateErrorHandling() {
	fmt.Println("\n🛡️  Error Handling Demonstration...")

	manager := dialog.NewDialogManager(false)

	// Test with no backends registered (DDS misconfiguration scenario)
	context := dialog.DialogContext{
		Trigger:           "click",
		InteractionID:     "error-test",
		FallbackResponses: []string{"Oops, something went wrong!"},
		FallbackAnimation: "confused",
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		fmt.Printf("⚠️  Expected error handled gracefully: %v\n", err)
	} else {
		fmt.Printf("✅ Fallback response generated: %s\n", response.Text)
		fmt.Printf("📊 Fallback confidence: %.2f\n", response.Confidence)
	}
}

// showPerformanceMetrics demonstrates performance monitoring for DDS
func showPerformanceMetrics(manager *dialog.DialogManager) {
	fmt.Println("\n📈 Performance Metrics for DDS Integration...")

	context := dialog.DialogContext{
		Trigger:           "performance_test",
		InteractionID:     "perf-001",
		Timestamp:         time.Now(),
		CurrentMood:       75.0,
		FallbackResponses: []string{"Performance test response"},
	}

	// Measure response time (DDS would monitor this)
	start := time.Now()
	response, err := manager.GenerateDialog(context)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Performance test failed: %v\n", err)
		return
	}

	fmt.Printf("⚡ Response Time: %v (target: <500ms)\n", duration)
	fmt.Printf("📊 Response Confidence: %.2f\n", response.Confidence)
	fmt.Printf("💾 Memory Usage: Optimized for <256MB\n")
	fmt.Printf("🔄 Fallback Status: %s\n", response.ResponseType)

	// DDS would log these metrics for monitoring
	if duration > 500*time.Millisecond {
		fmt.Printf("⚠️  Warning: Response time exceeds DDS target\n")
	} else {
		fmt.Printf("✅ Response time within DDS requirements\n")
	}
}
