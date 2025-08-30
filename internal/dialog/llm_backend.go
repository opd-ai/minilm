package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MockLLMModel provides a mock implementation for testing without actual LLM dependencies
// In production, this would be replaced with actual llama.cpp or other LLM bindings
type MockLLMModel struct {
	responses   []string
	delay       time.Duration
	initialized bool
	contextSize int
	mu          sync.RWMutex
}

// NewMockLLMModel creates a mock model with predefined responses
func NewMockLLMModel() *MockLLMModel {
	return &MockLLMModel{
		responses: []string{
			"Hi there! How are you doing today? 😊",
			"That's nice! I'm feeling pretty good myself.",
			"What would you like to do? I'm up for anything!",
			"Thanks for spending time with me! 💕",
			"Hmm, that's interesting. Tell me more!",
			"I appreciate your attention! *happy smile*",
			"This is fun! Let's keep chatting.",
			"You're so sweet! Thank you! 🌟",
			"I'm here whenever you need me!",
			"That makes me happy! *excited bounce*",
		},
		delay:       200 * time.Millisecond, // Simulate processing time
		initialized: false,
		contextSize: 2048,
	}
}

// Initialize initializes the mock model (no-op for compatibility)
func (m *MockLLMModel) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = true
	return nil
}

// Predict simulates LLM prediction with mock responses
func (m *MockLLMModel) Predict(prompt string) (string, error) {
	m.mu.RLock()
	if !m.initialized {
		m.mu.RUnlock()
		return "", fmt.Errorf("mock model not initialized")
	}
	m.mu.RUnlock()

	// Simulate processing delay
	time.Sleep(m.delay)

	// Simple keyword-based response selection for more realistic behavior
	prompt = strings.ToLower(prompt)

	switch {
	case strings.Contains(prompt, "hello") || strings.Contains(prompt, "hi"):
		return "Hi there! How are you doing today? 😊", nil
	case strings.Contains(prompt, "feed") || strings.Contains(prompt, "food"):
		return "Thanks for the meal! *nom nom* 😋", nil
	case strings.Contains(prompt, "pet") || strings.Contains(prompt, "pat"):
		return "That feels wonderful! *purrs happily* 😊", nil
	case strings.Contains(prompt, "talk") || strings.Contains(prompt, "chat"):
		return "I love chatting with you! What's on your mind? 💭", nil
	case strings.Contains(prompt, "sad") || strings.Contains(prompt, "down"):
		return "Aww, I'm here for you! *gentle hug* 🤗", nil
	case strings.Contains(prompt, "happy") || strings.Contains(prompt, "joy"):
		return "Your happiness makes me happy too! 😄✨", nil
	default:
		// Return a random response for unmatched prompts
		return m.responses[int(time.Now().UnixNano())%len(m.responses)], nil
	}
}

// PredictWithTimeout generates text with a timeout context
func (m *MockLLMModel) PredictWithTimeout(ctx context.Context, prompt string) (string, error) {
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := m.Predict(prompt)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("mock prediction timed out: %w", ctx.Err())
	}
}

// EstimateTokens provides a rough estimate of token count for a text
func (m *MockLLMModel) EstimateTokens(text string) int {
	// Rough approximation: 4 characters per token for English text
	return len(text) / 4
}

// GetContextSize returns the maximum context size for this model
func (m *MockLLMModel) GetContextSize() int {
	return m.contextSize
}

// GetModelInfo returns information about the mock model
func (m *MockLLMModel) GetModelInfo() ModelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return ModelInfo{
		ModelPath:   "mock://model",
		ContextSize: m.contextSize,
		Threads:     1,
		Temperature: 0.7,
		TopP:        0.9,
		Initialized: m.initialized,
		ModelType:   "mock",
		Backend:     "CPU",
	}
}

// Free releases mock model resources (no-op for compatibility)
func (m *MockLLMModel) Free() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = false
	return nil
}

// LLMBackend implements DialogBackend using LLM inference
// Supports both mock (for development) and production (llama.cpp) models
type LLMBackend struct {
	// Model and inference configuration
	model              ProductionLLMModel // Production model interface (LlamaModel or MockLLMModel)
	mockModel          *MockLLMModel      // Legacy mock for fallback
	useProductionModel bool               // Whether to use production or mock model
	modelPath          string
	maxTokens          int
	temperature        float32
	topP               float32
	contextSize        int
	threads            int

	// Markov-based personality configuration (reuses existing character data)
	markovConfig    MarkovChainConfig
	trainingData    []string // Personality examples from Markov training data
	fallbackPhrases []string // Fallback responses from Markov config

	// Context management
	contextManager   *ContextManager
	maxHistoryLength int

	// Performance and reliability
	timeout         time.Duration
	fallbackEnabled bool
	initialized     bool
	mu              sync.RWMutex

	// Backend metadata
	info BackendInfo
}

// LLMConfig defines configuration options for the LLM backend
// Uses existing Markov chain configuration for personality and training data
type LLMConfig struct {
	// Model configuration
	ModelPath   string  `json:"modelPath"`   // Path to GGUF model file
	MaxTokens   int     `json:"maxTokens"`   // Maximum tokens per response (default: 50)
	Temperature float32 `json:"temperature"` // Sampling temperature (default: 0.7)
	TopP        float32 `json:"topP"`        // Top-p sampling (default: 0.9)
	ContextSize int     `json:"contextSize"` // Model context window (default: 2048)
	Threads     int     `json:"threads"`     // CPU threads to use (default: 4)

	// Markov-based personality configuration (compatible with existing character format)
	MarkovConfig MarkovChainConfig `json:"markov_chain"` // Reuse existing Markov configuration

	// Context management
	MaxHistoryLength int `json:"maxHistoryLength"` // Max conversation history (default: 10)

	// Performance settings
	TimeoutMs       int  `json:"timeoutMs"`       // Response timeout in ms (default: 2000)
	FallbackEnabled bool `json:"fallbackEnabled"` // Enable fallback on failure (default: true)
}

// MarkovChainConfig represents the existing Markov chain configuration
// This allows LLM backend to reuse existing character personality data
type MarkovChainConfig struct {
	ChainOrder      int      `json:"chainOrder"`
	MinWords        int      `json:"minWords"`
	MaxWords        int      `json:"maxWords"`
	TemperatureMin  float64  `json:"temperatureMin"`
	TemperatureMax  float64  `json:"temperatureMax"`
	UsePersonality  bool     `json:"usePersonality"`
	TrainingData    []string `json:"trainingData"`    // Personality-rich training sentences
	FallbackPhrases []string `json:"fallbackPhrases"` // Fallback responses
}

// NewLLMBackend creates a new LLM-powered dialog backend
// Uses conservative defaults optimized for consumer CPU hardware
func NewLLMBackend() *LLMBackend {
	return &LLMBackend{
		maxTokens:        50,
		temperature:      0.7,
		topP:             0.9,
		contextSize:      2048,
		threads:          4,
		maxHistoryLength: 10,
		timeout:          2 * time.Second,
		fallbackEnabled:  true,
		contextManager:   NewContextManager(10),
		info: BackendInfo{
			Name:        "llm_backend",
			Version:     "1.0.0",
			Description: "LLM-powered dialog backend using llama.cpp for CPU inference",
			Capabilities: []string{
				"context_aware",
				"personality_driven",
				"learning_enabled",
				"streaming_capable",
			},
			Author:  "MiniLM Project",
			License: "MIT",
		},
	}
}

// Initialize sets up the LLM backend with the provided JSON configuration
// This method handles model loading and validation
func (llm *LLMBackend) Initialize(config json.RawMessage) error {
	llm.mu.Lock()
	defer llm.mu.Unlock()

	var cfg LLMConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("failed to parse LLM config: %w", err)
	}

	// Apply configuration with defaults
	if err := llm.applyConfig(cfg); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}

	// Load the model
	if err := llm.loadModel(); err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	llm.initialized = true
	return nil
}

// applyConfig applies the provided configuration with sensible defaults
func (llm *LLMBackend) applyConfig(cfg LLMConfig) error {
	if err := llm.validateAndSetModelPath(cfg.ModelPath); err != nil {
		return err
	}

	llm.applyOptionalParameters(cfg)
	llm.configureMarkovSettings(cfg)

	return nil
}

// validateAndSetModelPath validates the model path and sets it on the backend
func (llm *LLMBackend) validateAndSetModelPath(modelPath string) error {
	if modelPath == "" {
		return fmt.Errorf("modelPath is required")
	}
	llm.modelPath = modelPath
	return nil
}

// applyOptionalParameters applies optional configuration parameters with defaults
func (llm *LLMBackend) applyOptionalParameters(cfg LLMConfig) {
	llm.applyLLMParameters(cfg)
	llm.applyContextParameters(cfg)
	llm.applyTimeoutParameters(cfg)
}

// applyLLMParameters configures core LLM generation parameters
func (llm *LLMBackend) applyLLMParameters(cfg LLMConfig) {
	if cfg.MaxTokens > 0 {
		llm.maxTokens = cfg.MaxTokens
	}
	if cfg.Temperature > 0 {
		llm.temperature = cfg.Temperature
	}
	if cfg.TopP > 0 {
		llm.topP = cfg.TopP
	}
}

// applyContextParameters configures context and execution parameters
func (llm *LLMBackend) applyContextParameters(cfg LLMConfig) {
	if cfg.ContextSize > 0 {
		llm.contextSize = cfg.ContextSize
	}
	if cfg.Threads > 0 {
		llm.threads = cfg.Threads
	}
	if cfg.MaxHistoryLength > 0 {
		llm.maxHistoryLength = cfg.MaxHistoryLength
		llm.contextManager = NewContextManager(cfg.MaxHistoryLength)
	}
}

// applyTimeoutParameters configures timeout-related parameters
func (llm *LLMBackend) applyTimeoutParameters(cfg LLMConfig) {
	if cfg.TimeoutMs > 0 {
		llm.timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond
	}
}

// configureMarkovSettings configures Markov-based personality settings
func (llm *LLMBackend) configureMarkovSettings(cfg LLMConfig) {
	llm.markovConfig = cfg.MarkovConfig
	llm.trainingData = cfg.MarkovConfig.TrainingData
	llm.fallbackPhrases = cfg.MarkovConfig.FallbackPhrases
	llm.fallbackEnabled = cfg.FallbackEnabled
}

// loadModel initializes either production LLM model or mock model
// Attempts to load production model first, falls back to mock if needed
func (llm *LLMBackend) loadModel() error {
	// Try to load production model if path points to actual GGUF file
	if strings.HasSuffix(llm.modelPath, ".gguf") {
		// Check if file exists to determine if we should attempt production loading
		// In a production environment, this would also check for llama.cpp availability
		productionModel, err := llm.tryLoadProductionModel()
		if err == nil {
			llm.model = productionModel
			llm.useProductionModel = true
			return nil
		}

		// Log the production model failure but continue with mock
		// In production, you might want to return the error instead
		fmt.Printf("Production model loading failed (%v), falling back to mock model\n", err)
	}

	// Use mock model as fallback or if not using production model
	mockModel := NewMockLLMModel()
	if err := mockModel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize mock model: %w", err)
	}

	llm.model = mockModel
	llm.mockModel = mockModel
	llm.useProductionModel = false
	return nil
}

// tryLoadProductionModel attempts to load a production LLM model
func (llm *LLMBackend) tryLoadProductionModel() (ProductionLLMModel, error) {
	config := LlamaConfig{
		ModelPath:   llm.modelPath,
		ContextSize: llm.contextSize,
		Threads:     llm.threads,
		Temperature: llm.temperature,
		TopP:        llm.topP,
	}

	model, err := NewLlamaModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create production model: %w", err)
	}

	if err := model.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize production model: %w", err)
	}

	return model, nil
}

// GenerateResponse produces a dialog response using the LLM
// Implements context-aware conversation with character personality
func (llm *LLMBackend) GenerateResponse(ctx DialogContext) (DialogResponse, error) {
	llm.mu.RLock()
	if !llm.initialized {
		llm.mu.RUnlock()
		return DialogResponse{}, fmt.Errorf("LLM backend not initialized")
	}
	llm.mu.RUnlock()

	// Build the prompt from context and character data
	prompt := llm.buildPrompt(ctx)

	// Generate response with timeout
	responseCtx, cancel := context.WithTimeout(context.Background(), llm.timeout)
	defer cancel()

	response, err := llm.generateWithTimeout(responseCtx, prompt)
	if err != nil {
		if llm.fallbackEnabled {
			return llm.createFallbackResponse(ctx), nil
		}
		return DialogResponse{}, fmt.Errorf("failed to generate response: %w", err)
	}

	// Update conversation context
	llm.contextManager.AddExchange(ctx.InteractionID, ctx.Trigger, response)

	// Create structured response
	dialogResponse := DialogResponse{
		Text:             response,
		Animation:        llm.selectAnimation(ctx, response),
		Confidence:       0.8, // High confidence for successful LLM generation
		ResponseType:     llm.classifyResponse(response),
		EmotionalTone:    llm.detectEmotionalTone(response),
		Topics:           llm.extractTopics(response),
		MemoryImportance: 0.7, // Default importance for LLM responses
		LearningValue:    0.6,
	}

	return dialogResponse, nil
}

// generateWithTimeout generates a response with the given context and timeout
func (llm *LLMBackend) generateWithTimeout(ctx context.Context, prompt string) (string, error) {
	// Channel to receive the result
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Generate response in a goroutine
	go func() {
		// Use mock model - in production this would use actual LLM prediction
		result, err := llm.model.Predict(prompt)
		if err != nil {
			errorChan <- err
			return
		}

		// Clean and validate the response
		cleaned := llm.cleanResponse(result)
		resultChan <- cleaned
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("response generation timed out after %v", llm.timeout)
	}
}

// buildPrompt constructs a prompt from the dialog context and character configuration
func (llm *LLMBackend) buildPrompt(ctx DialogContext) string {
	builder := NewPromptBuilder()

	// Extract personality from Markov training data
	personality := llm.extractPersonality()
	if personality != "" {
		builder.AddPersonality(personality)
	}

	// Add conversation history
	history := llm.contextManager.GetHistory(ctx.InteractionID, 5)
	builder.AddHistory(history)

	// Add current context
	builder.AddContext(ctx)

	return builder.Build()
}

// extractPersonality creates a personality description from Markov training data
func (llm *LLMBackend) extractPersonality() string {
	if len(llm.markovConfig.TrainingData) == 0 {
		return "You are a helpful AI assistant."
	}

	// Take first few training examples as personality indicators
	var personalityExamples []string
	limit := 3
	if len(llm.markovConfig.TrainingData) < limit {
		limit = len(llm.markovConfig.TrainingData)
	}

	for i := 0; i < limit; i++ {
		personalityExamples = append(personalityExamples, llm.markovConfig.TrainingData[i])
	}

	// Create personality description from examples
	personality := "Based on these example responses, respond in a similar tone and style:\n"
	for _, example := range personalityExamples {
		personality += "- " + example + "\n"
	}

	return personality
}

// cleanResponse processes the raw LLM output to ensure it's suitable for display
func (llm *LLMBackend) cleanResponse(response string) string {
	// Remove common LLM artifacts
	cleaned := strings.TrimSpace(response)

	// Remove leading/trailing quotes if present
	if (strings.HasPrefix(cleaned, `"`) && strings.HasSuffix(cleaned, `"`)) ||
		(strings.HasPrefix(cleaned, `'`) && strings.HasSuffix(cleaned, `'`)) {
		cleaned = cleaned[1 : len(cleaned)-1]
	}

	// Limit length for UI display (roughly 2-3 sentences)
	if len(cleaned) > 150 {
		sentences := strings.Split(cleaned, ". ")
		if len(sentences) > 2 {
			cleaned = strings.Join(sentences[:2], ". ") + "."
		}
	}

	// Ensure we have some content
	if len(strings.TrimSpace(cleaned)) == 0 {
		cleaned = "Hello! 👋"
	}

	return cleaned
}

// selectAnimation chooses an appropriate animation based on response content
func (llm *LLMBackend) selectAnimation(ctx DialogContext, response string) string {
	response = strings.ToLower(response)

	// Simple keyword-based animation selection
	if strings.Contains(response, "happy") || strings.Contains(response, "joy") || strings.Contains(response, "😊") {
		return "happy"
	}
	if strings.Contains(response, "sad") || strings.Contains(response, "sorry") || strings.Contains(response, "😢") {
		return "sad"
	}
	if strings.Contains(response, "eat") || strings.Contains(response, "food") || strings.Contains(response, "hungry") {
		return "eating"
	}

	// Default to talking animation
	return "talking"
}

// classifyResponse determines the type of response generated
func (llm *LLMBackend) classifyResponse(response string) string {
	response = strings.ToLower(response)

	if strings.Contains(response, "love") || strings.Contains(response, "heart") {
		return "romantic"
	}
	if strings.Contains(response, "help") || strings.Contains(response, "support") {
		return "helpful"
	}
	if strings.Contains(response, "?") {
		return "inquisitive"
	}

	return "casual"
}

// detectEmotionalTone analyzes the emotional content of the response
func (llm *LLMBackend) detectEmotionalTone(response string) string {
	response = strings.ToLower(response)

	// Simple emotion detection based on keywords and punctuation
	if strings.Contains(response, "!") || strings.Contains(response, "exciting") {
		return "excited"
	}
	if strings.Contains(response, "happy") || strings.Contains(response, "😊") {
		return "happy"
	}
	if strings.Contains(response, "shy") || strings.Contains(response, "blush") {
		return "shy"
	}

	return "neutral"
}

// extractTopics identifies key topics mentioned in the response
func (llm *LLMBackend) extractTopics(response string) []string {
	topics := []string{}
	response = strings.ToLower(response)

	// Simple keyword-based topic extraction
	topicKeywords := map[string]string{
		"food":   "food",
		"eat":    "food",
		"hungry": "food",
		"game":   "gaming",
		"play":   "gaming",
		"love":   "romance",
		"heart":  "romance",
		"work":   "work",
		"study":  "study",
		"learn":  "study",
	}

	for keyword, topic := range topicKeywords {
		if strings.Contains(response, keyword) {
			// Avoid duplicates
			found := false
			for _, existing := range topics {
				if existing == topic {
					found = true
					break
				}
			}
			if !found {
				topics = append(topics, topic)
			}
		}
	}

	return topics
}

// createFallbackResponse generates a simple response when LLM generation fails
func (llm *LLMBackend) createFallbackResponse(ctx DialogContext) DialogResponse {
	responses := []string{
		"Hi there! 👋",
		"What's up?",
		"How are you doing?",
		"Nice to see you!",
		"*waves*",
	}

	// Simple selection based on trigger
	var response string
	switch ctx.Trigger {
	case "click":
		response = "Hi there! 👋"
	case "feed":
		response = "Thanks! *nom nom*"
	case "rightclick":
		response = "What's up?"
	default:
		response = responses[int(time.Now().UnixNano())%len(responses)]
	}

	return DialogResponse{
		Text:          response,
		Animation:     "talking",
		Confidence:    0.3, // Low confidence for fallback
		ResponseType:  "fallback",
		EmotionalTone: "neutral",
	}
}

// CanHandle checks if this backend can process the given context
func (llm *LLMBackend) CanHandle(ctx DialogContext) bool {
	llm.mu.RLock()
	defer llm.mu.RUnlock()

	// Can handle if initialized and model is loaded
	return llm.initialized && llm.model != nil
}

// GetBackendInfo returns metadata about this LLM backend implementation
func (llm *LLMBackend) GetBackendInfo() BackendInfo {
	return llm.info
}

// UpdateMemory records interaction outcomes for potential future learning
// Currently a placeholder for future learning implementations
func (llm *LLMBackend) UpdateMemory(ctx DialogContext, response DialogResponse, feedback *UserFeedback) error {
	// Record the interaction for potential future learning
	if feedback != nil {
		llm.contextManager.UpdateFeedback(ctx.InteractionID, feedback.Positive, feedback.Engagement)
	}

	// TODO: Implement actual learning mechanisms (fine-tuning, prompt adaptation, etc.)
	return nil
}

// Close properly shuts down the LLM backend and frees resources
func (llm *LLMBackend) Close() error {
	llm.mu.Lock()
	defer llm.mu.Unlock()

	if llm.model != nil {
		llm.model.Free()
		llm.model = nil
	}

	llm.initialized = false
	return nil
}
