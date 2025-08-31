package dialog

import (
	"fmt"
	"strings"
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

// Test for bug #2: Race Condition in Context Manager Cleanup
func TestContextManager_test_bug2_cleanup_race_condition(t *testing.T) {
	cm := NewContextManager(100)
	defer cm.Close()

	// Test the cleanup function with many items to ensure the fix works correctly
	cutoff := time.Now().Add(-25 * time.Hour) // Older than 24 hour cleanup threshold

	// Manually populate conversations with old timestamps
	cm.mu.Lock()
	for i := 0; i < 1000; i++ {
		interactionID := "old_" + string(rune('a'+i%26)) + string(rune('a'+(i/26)%26)) + string(rune('a'+(i/676)%26))
		cm.conversations[interactionID] = &ConversationHistory{
			Exchanges:   []ConversationExchange{},
			LastUpdated: cutoff,
		}
	}
	// Add some new conversations that should NOT be cleaned up
	for i := 0; i < 100; i++ {
		interactionID := "new_" + string(rune('a'+i%26)) + string(rune('a'+(i/26)%26))
		cm.conversations[interactionID] = &ConversationHistory{
			Exchanges:   []ConversationExchange{},
			LastUpdated: time.Now(),
		}
	}
	cm.mu.Unlock()

	initialCount := cm.GetActiveConversations()
	if initialCount != 1100 {
		t.Errorf("Expected 1100 initial conversations, got %d", initialCount)
	}

	// Run cleanup - should remove old conversations safely
	// After fix: should use two-phase deletion (collect keys, then delete)
	cm.cleanupOldConversations()

	finalCount := cm.GetActiveConversations()

	// Should have removed 1000 old conversations, keeping 100 new ones
	if finalCount != 100 {
		t.Errorf("Expected 100 conversations after cleanup, got %d", finalCount)
	}

	// Verify that only new conversations remain
	cm.mu.RLock()
	for id := range cm.conversations {
		if !strings.HasPrefix(id, "new_") {
			t.Errorf("Found unexpected old conversation after cleanup: %s", id)
		}
	}
	cm.mu.RUnlock()
}

// TestDDS_test_bug8_context_manager_memory_leak_prevention tests for bug #8
func TestDDS_test_bug8_context_manager_memory_leak_prevention(t *testing.T) {
	cm := NewContextManager(5)
	defer cm.Close()

	// Simulate high-traffic scenario with many concurrent conversations
	// that would accumulate within the 24-hour cleanup window
	numConversations := 1000

	// Create many conversations with recent timestamps (all within 24 hours)
	for i := 0; i < numConversations; i++ {
		conversationID := fmt.Sprintf("user_%d", i)

		// Use the actual AddExchange method with trigger and response
		cm.AddExchange(conversationID, "user_message", fmt.Sprintf("Hello user %d", i))
	}

	// Bug #8: All conversations remain in memory because they're within 24-hour window
	activeCount := cm.GetActiveConversations()
	t.Logf("Active conversations after creating %d recent conversations: %d", numConversations, activeCount)

	if activeCount != numConversations {
		t.Errorf("Expected %d active conversations, got %d", numConversations, activeCount)
	}

	// The issue: No mechanism to limit memory usage when all conversations are "recent"
	// In a real high-traffic scenario, this could lead to unlimited memory growth
	// until the 24-hour cleanup threshold is reached

	// Simulate what would happen with real memory pressure
	// Even with thousands of active conversations, there's no LRU or count-based eviction
	expectedMemoryIssue := activeCount >= 500 // Arbitrary threshold for "too many"
	if expectedMemoryIssue {
		t.Logf("Bug #8 demonstrated: %d active conversations could cause memory pressure in production", activeCount)
		t.Logf("Current implementation has no conversation count limits or LRU eviction")

		// The fix should provide:
		// 1. Configurable cleanup intervals (not fixed at 1 hour)
		// 2. Configurable retention periods (not fixed at 24 hours)
		// 3. Maximum conversation count limits with LRU eviction
		// 4. Memory-aware cleanup policies
	}
}

// TestDDS_test_bug8_context_manager_lru_eviction_fix tests the fix for bug #8
func TestDDS_test_bug8_context_manager_lru_eviction_fix(t *testing.T) {
	// Test the enhanced ContextManager with conversation limits and LRU eviction
	maxConversations := 100
	cm := NewContextManagerWithConfig(5, maxConversations, 1*time.Minute, 30*time.Minute)
	defer cm.Close()

	// Create more conversations than the limit to trigger LRU eviction
	numConversations := 150

	for i := 0; i < numConversations; i++ {
		conversationID := fmt.Sprintf("user_%d", i)
		cm.AddExchange(conversationID, "user_message", fmt.Sprintf("Hello user %d", i))
	}

	// After the fix: Should not exceed the conversation limit due to LRU eviction
	activeCount := cm.GetActiveConversations()
	t.Logf("Active conversations after creating %d conversations with limit %d: %d", numConversations, maxConversations, activeCount)

	if activeCount > maxConversations {
		t.Errorf("Expected at most %d active conversations due to LRU eviction, got %d", maxConversations, activeCount)
	}

	// Verify that the most recent conversations are kept (LRU behavior)
	// The last maxConversations should still be present
	for i := numConversations - maxConversations; i < numConversations; i++ {
		recentID := fmt.Sprintf("user_%d", i)
		history := cm.GetHistory(recentID, 10)
		if len(history) == 0 {
			t.Errorf("Expected recent conversation %s to be preserved by LRU", recentID)
		}
	}

	// Verify that old conversations were evicted
	for i := 0; i < numConversations-maxConversations; i++ {
		oldID := fmt.Sprintf("user_%d", i)
		history := cm.GetHistory(oldID, 10)
		if len(history) > 0 {
			t.Errorf("Expected old conversation %s to be evicted by LRU", oldID)
		}
	}

	t.Logf("Bug #8 FIXED: LRU eviction successfully limits memory usage to %d conversations", activeCount)
}
