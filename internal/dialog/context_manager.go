package dialog

import (
	"sync"
	"time"
)

// ConversationExchange represents a single turn in a conversation
type ConversationExchange struct {
	Timestamp       time.Time `json:"timestamp"`
	Trigger         string    `json:"trigger"`         // User action that triggered response
	Response        string    `json:"response"`        // Character's response
	UserFeedback    bool      `json:"userFeedback"`    // Whether user gave positive feedback
	EngagementScore float64   `json:"engagementScore"` // Engagement level (0-1)
}

// ConversationHistory tracks the recent conversation exchanges for a character
type ConversationHistory struct {
	InteractionID string                 `json:"interactionId"`
	Exchanges     []ConversationExchange `json:"exchanges"`
	LastUpdated   time.Time              `json:"lastUpdated"`
	MaxLength     int                    `json:"maxLength"`
}

// ContextManager handles conversation history and context for dialog generation
// Maintains a rolling window of recent exchanges to provide context for LLM prompts
type ContextManager struct {
	conversations map[string]*ConversationHistory
	maxHistory    int
	cleanupTicker *time.Ticker
	mu            sync.RWMutex
}

// NewContextManager creates a new context manager with specified history length
func NewContextManager(maxHistory int) *ContextManager {
	if maxHistory <= 0 {
		maxHistory = 10 // Default to 10 exchanges
	}

	cm := &ContextManager{
		conversations: make(map[string]*ConversationHistory),
		maxHistory:    maxHistory,
	}

	// Start cleanup routine to remove old conversations (runs every hour)
	cm.cleanupTicker = time.NewTicker(1 * time.Hour)
	go cm.cleanupRoutine()

	return cm
}

// AddExchange records a new conversation exchange
func (cm *ContextManager) AddExchange(interactionID, trigger, response string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get or create conversation history
	history, exists := cm.conversations[interactionID]
	if !exists {
		history = &ConversationHistory{
			InteractionID: interactionID,
			Exchanges:     make([]ConversationExchange, 0, cm.maxHistory),
			MaxLength:     cm.maxHistory,
		}
		cm.conversations[interactionID] = history
	}

	// Add new exchange
	exchange := ConversationExchange{
		Timestamp: time.Now(),
		Trigger:   trigger,
		Response:  response,
	}

	history.Exchanges = append(history.Exchanges, exchange)
	history.LastUpdated = time.Now()

	// Maintain rolling window by removing oldest exchanges if needed
	if len(history.Exchanges) > history.MaxLength {
		history.Exchanges = history.Exchanges[1:]
	}
}

// GetHistory retrieves recent conversation history for context building
func (cm *ContextManager) GetHistory(interactionID string, maxExchanges int) []ConversationExchange {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	history, exists := cm.conversations[interactionID]
	if !exists {
		return []ConversationExchange{}
	}

	exchanges := history.Exchanges

	// Limit to requested number of exchanges (most recent first)
	if maxExchanges > 0 && len(exchanges) > maxExchanges {
		start := len(exchanges) - maxExchanges
		exchanges = exchanges[start:]
	}

	// Return copy to avoid race conditions
	result := make([]ConversationExchange, len(exchanges))
	copy(result, exchanges)
	return result
}

// UpdateFeedback records user feedback for the most recent exchange
func (cm *ContextManager) UpdateFeedback(interactionID string, positive bool, engagement float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	history, exists := cm.conversations[interactionID]
	if !exists || len(history.Exchanges) == 0 {
		return
	}

	// Update the most recent exchange
	lastIdx := len(history.Exchanges) - 1
	history.Exchanges[lastIdx].UserFeedback = positive
	history.Exchanges[lastIdx].EngagementScore = engagement
	history.LastUpdated = time.Now()
}

// GetConversationSummary provides a summary of the conversation for prompt building
func (cm *ContextManager) GetConversationSummary(interactionID string) ConversationSummary {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	history, exists := cm.conversations[interactionID]
	if !exists {
		return ConversationSummary{
			ExchangeCount:    0,
			AvgEngagement:    0,
			PositiveFeedback: 0,
			LastInteraction:  time.Time{},
			DominantTriggers: []string{},
			RecentTopics:     []string{},
		}
	}

	return cm.calculateSummary(history)
}

// ConversationSummary provides aggregated statistics about a conversation
type ConversationSummary struct {
	ExchangeCount    int       `json:"exchangeCount"`
	AvgEngagement    float64   `json:"avgEngagement"`
	PositiveFeedback int       `json:"positiveFeedback"`
	LastInteraction  time.Time `json:"lastInteraction"`
	DominantTriggers []string  `json:"dominantTriggers"`
	RecentTopics     []string  `json:"recentTopics"`
}

// calculateSummary computes conversation statistics
func (cm *ContextManager) calculateSummary(history *ConversationHistory) ConversationSummary {
	exchanges := history.Exchanges
	if len(exchanges) == 0 {
		return ConversationSummary{}
	}

	// Calculate engagement and feedback statistics
	totalEngagement := 0.0
	positiveFeedback := 0
	triggerCounts := make(map[string]int)

	for _, exchange := range exchanges {
		totalEngagement += exchange.EngagementScore
		if exchange.UserFeedback {
			positiveFeedback++
		}
		triggerCounts[exchange.Trigger]++
	}

	avgEngagement := totalEngagement / float64(len(exchanges))

	// Find dominant triggers
	dominantTriggers := make([]string, 0, 3)
	for trigger, count := range triggerCounts {
		if count >= 2 { // Only include triggers used multiple times
			dominantTriggers = append(dominantTriggers, trigger)
		}
	}

	return ConversationSummary{
		ExchangeCount:    len(exchanges),
		AvgEngagement:    avgEngagement,
		PositiveFeedback: positiveFeedback,
		LastInteraction:  history.LastUpdated,
		DominantTriggers: dominantTriggers,
		RecentTopics:     []string{}, // TODO: Implement topic extraction
	}
}

// ClearHistory removes all conversation history for a specific interaction
func (cm *ContextManager) ClearHistory(interactionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.conversations, interactionID)
}

// GetActiveConversations returns the number of active conversations being tracked
func (cm *ContextManager) GetActiveConversations() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.conversations)
}

// cleanupRoutine periodically removes old conversations to prevent memory leaks
func (cm *ContextManager) cleanupRoutine() {
	for range cm.cleanupTicker.C {
		cm.cleanupOldConversations()
	}
}

// cleanupOldConversations removes conversations that haven't been active recently
func (cm *ContextManager) cleanupOldConversations() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Remove conversations older than 24 hours

	for id, history := range cm.conversations {
		if history.LastUpdated.Before(cutoff) {
			delete(cm.conversations, id)
		}
	}
}

// Close stops the cleanup routine and releases resources
func (cm *ContextManager) Close() {
	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conversations = nil
}
