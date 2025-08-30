package dialog

import (
	"testing"
	"time"
)

func TestContextManager_AddExchange(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	interactionID := "test-interaction-1"
	trigger := "click"
	response := "Hello there!"

	cm.AddExchange(interactionID, trigger, response)

	history := cm.GetHistory(interactionID, 10)
	if len(history) != 1 {
		t.Errorf("Expected 1 exchange, got %d", len(history))
	}

	if history[0].Trigger != trigger {
		t.Errorf("Expected trigger '%s', got '%s'", trigger, history[0].Trigger)
	}

	if history[0].Response != response {
		t.Errorf("Expected response '%s', got '%s'", response, history[0].Response)
	}
}

func TestContextManager_MaxHistoryLength(t *testing.T) {
	maxHistory := 3
	cm := NewContextManager(maxHistory)
	defer cm.Close()

	interactionID := "test-interaction-1"

	// Add more exchanges than the limit
	for i := 0; i < 5; i++ {
		cm.AddExchange(interactionID, "click", "response")
	}

	history := cm.GetHistory(interactionID, 10)
	if len(history) != maxHistory {
		t.Errorf("Expected %d exchanges (max history), got %d", maxHistory, len(history))
	}
}

func TestContextManager_GetHistoryLimited(t *testing.T) {
	cm := NewContextManager(10)
	defer cm.Close()

	interactionID := "test-interaction-1"

	// Add 5 exchanges
	for i := 0; i < 5; i++ {
		cm.AddExchange(interactionID, "click", "response")
	}

	// Request only 3 most recent
	history := cm.GetHistory(interactionID, 3)
	if len(history) != 3 {
		t.Errorf("Expected 3 exchanges, got %d", len(history))
	}
}

func TestContextManager_GetHistoryNonExistent(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	history := cm.GetHistory("non-existent", 10)
	if len(history) != 0 {
		t.Errorf("Expected 0 exchanges for non-existent interaction, got %d", len(history))
	}
}

func TestContextManager_UpdateFeedback(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	interactionID := "test-interaction-1"
	cm.AddExchange(interactionID, "click", "Hello!")

	// Update feedback for the exchange
	cm.UpdateFeedback(interactionID, true, 0.9)

	history := cm.GetHistory(interactionID, 10)
	if len(history) != 1 {
		t.Fatalf("Expected 1 exchange, got %d", len(history))
	}

	exchange := history[0]
	if !exchange.UserFeedback {
		t.Error("Expected positive feedback to be true")
	}

	if exchange.EngagementScore != 0.9 {
		t.Errorf("Expected engagement score 0.9, got %f", exchange.EngagementScore)
	}
}

func TestContextManager_UpdateFeedbackNonExistent(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	// Should not panic or error when updating non-existent interaction
	cm.UpdateFeedback("non-existent", true, 0.9)
}

func TestContextManager_GetConversationSummary(t *testing.T) {
	cm := NewContextManager(10)
	defer cm.Close()

	interactionID := "test-interaction-1"

	// Add exchanges with different triggers
	cm.AddExchange(interactionID, "click", "Hello!")
	cm.AddExchange(interactionID, "click", "Hi there!")
	cm.AddExchange(interactionID, "feed", "Thanks for food!")

	// Update feedback for some exchanges
	cm.UpdateFeedback(interactionID, true, 0.8)

	summary := cm.GetConversationSummary(interactionID)

	if summary.ExchangeCount != 3 {
		t.Errorf("Expected 3 exchanges, got %d", summary.ExchangeCount)
	}

	if summary.PositiveFeedback != 1 {
		t.Errorf("Expected 1 positive feedback, got %d", summary.PositiveFeedback)
	}

	// Should have "click" as dominant trigger (appears twice)
	found := false
	for _, trigger := range summary.DominantTriggers {
		if trigger == "click" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'click' to be a dominant trigger, got %v", summary.DominantTriggers)
	}
}

func TestContextManager_GetConversationSummaryEmpty(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	summary := cm.GetConversationSummary("non-existent")

	if summary.ExchangeCount != 0 {
		t.Errorf("Expected 0 exchanges for empty conversation, got %d", summary.ExchangeCount)
	}

	if summary.AvgEngagement != 0 {
		t.Errorf("Expected 0 average engagement for empty conversation, got %f", summary.AvgEngagement)
	}
}

func TestContextManager_ClearHistory(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	interactionID := "test-interaction-1"
	cm.AddExchange(interactionID, "click", "Hello!")

	// Verify history exists
	history := cm.GetHistory(interactionID, 10)
	if len(history) != 1 {
		t.Fatalf("Expected 1 exchange before clearing, got %d", len(history))
	}

	// Clear history
	cm.ClearHistory(interactionID)

	// Verify history is gone
	history = cm.GetHistory(interactionID, 10)
	if len(history) != 0 {
		t.Errorf("Expected 0 exchanges after clearing, got %d", len(history))
	}
}

func TestContextManager_GetActiveConversations(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	if cm.GetActiveConversations() != 0 {
		t.Errorf("Expected 0 active conversations initially, got %d", cm.GetActiveConversations())
	}

	// Add exchanges for different interactions
	cm.AddExchange("interaction-1", "click", "Hello!")
	cm.AddExchange("interaction-2", "feed", "Thanks!")

	if cm.GetActiveConversations() != 2 {
		t.Errorf("Expected 2 active conversations, got %d", cm.GetActiveConversations())
	}

	// Clear one conversation
	cm.ClearHistory("interaction-1")

	if cm.GetActiveConversations() != 1 {
		t.Errorf("Expected 1 active conversation after clearing one, got %d", cm.GetActiveConversations())
	}
}

func TestContextManager_DefaultMaxHistory(t *testing.T) {
	// Test with invalid max history (should default to 10)
	cm := NewContextManager(0)
	defer cm.Close()

	interactionID := "test-interaction-1"

	// Add 15 exchanges (more than default 10)
	for i := 0; i < 15; i++ {
		cm.AddExchange(interactionID, "click", "response")
	}

	history := cm.GetHistory(interactionID, 20)
	if len(history) != 10 {
		t.Errorf("Expected 10 exchanges (default max), got %d", len(history))
	}
}

func TestConversationExchange_Timestamps(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	interactionID := "test-interaction-1"
	start := time.Now()

	cm.AddExchange(interactionID, "click", "Hello!")

	history := cm.GetHistory(interactionID, 10)
	if len(history) != 1 {
		t.Fatalf("Expected 1 exchange, got %d", len(history))
	}

	exchange := history[0]
	if exchange.Timestamp.Before(start) {
		t.Error("Exchange timestamp should be after start time")
	}

	if exchange.Timestamp.After(time.Now().Add(time.Second)) {
		t.Error("Exchange timestamp should not be in the future")
	}
}

func TestContextManager_Close(t *testing.T) {
	cm := NewContextManager(5)

	// Add some data
	cm.AddExchange("test", "click", "Hello!")

	// Close should not panic
	cm.Close()

	// After close, the cleanup ticker should be stopped
	// (We can't easily test this without accessing internal state,
	// but we can verify Close doesn't panic)
}

func TestContextManager_ConcurrentAccess(t *testing.T) {
	cm := NewContextManager(10)
	defer cm.Close()

	interactionID := "test-interaction-1"

	// Test concurrent reads and writes
	done := make(chan bool, 3)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			cm.AddExchange(interactionID, "click", "response")
		}
		done <- true
	}()

	// Reader goroutine 1
	go func() {
		for i := 0; i < 50; i++ {
			cm.GetHistory(interactionID, 5)
		}
		done <- true
	}()

	// Reader goroutine 2
	go func() {
		for i := 0; i < 50; i++ {
			cm.GetConversationSummary(interactionID)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should have some exchanges (exact number may vary due to concurrency)
	history := cm.GetHistory(interactionID, 20)
	if len(history) == 0 {
		t.Error("Expected some exchanges after concurrent operations")
	}
}
