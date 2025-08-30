package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CharacterAssetIntegrator automatically adds LLM configuration to existing character files
type CharacterAssetIntegrator struct {
	assetsPath    string
	backupEnabled bool
	dryRun        bool
}

// CharacterJSON represents the structure of character configuration files
type CharacterJSON struct {
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Dialogs       []Dialog             `json:"dialogs,omitempty"`
	DialogBackend *DialogBackendConfig `json:"dialogBackend,omitempty"`
	// Include other fields as map to preserve existing data
	Extra map[string]interface{} `json:"-"`
}

// Dialog represents a dialog configuration
type Dialog struct {
	Trigger   string   `json:"trigger"`
	Responses []string `json:"responses"`
	Animation string   `json:"animation,omitempty"`
	Cooldown  int      `json:"cooldown,omitempty"`
}

// DialogBackendConfig represents the dialog backend configuration
type DialogBackendConfig struct {
	Enabled             bool                       `json:"enabled"`
	DefaultBackend      string                     `json:"defaultBackend"`
	FallbackChain       []string                   `json:"fallbackChain,omitempty"`
	ConfidenceThreshold float64                    `json:"confidenceThreshold"`
	MemoryEnabled       bool                       `json:"memoryEnabled"`
	LearningEnabled     bool                       `json:"learningEnabled"`
	DebugMode           bool                       `json:"debugMode,omitempty"`
	Backends            map[string]json.RawMessage `json:"backends"`
}

// LLMBackendConfig represents LLM-specific backend configuration
type LLMBackendConfig struct {
	ModelPath        string            `json:"modelPath"`
	MaxTokens        int               `json:"maxTokens"`
	Temperature      float32           `json:"temperature"`
	TopP             float32           `json:"topP"`
	ContextSize      int               `json:"contextSize"`
	Threads          int               `json:"threads"`
	MarkovConfig     MarkovChainConfig `json:"markov_chain"`
	MaxHistoryLength int               `json:"maxHistoryLength"`
	TimeoutMs        int               `json:"timeoutMs"`
	FallbackEnabled  bool              `json:"fallbackEnabled"`
}

// MarkovChainConfig represents Markov chain configuration for personality
type MarkovChainConfig struct {
	ChainOrder      int      `json:"chainOrder"`
	MinWords        int      `json:"minWords"`
	MaxWords        int      `json:"maxWords"`
	TemperatureMin  float64  `json:"temperatureMin"`
	TemperatureMax  float64  `json:"temperatureMax"`
	UsePersonality  bool     `json:"usePersonality"`
	TrainingData    []string `json:"trainingData"`
	FallbackPhrases []string `json:"fallbackPhrases"`
}

// NewCharacterAssetIntegrator creates a new integrator instance
func NewCharacterAssetIntegrator(assetsPath string) *CharacterAssetIntegrator {
	return &CharacterAssetIntegrator{
		assetsPath:    assetsPath,
		backupEnabled: true,
		dryRun:        false,
	}
}

// SetDryRun enables or disables dry run mode
func (c *CharacterAssetIntegrator) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// SetBackupEnabled enables or disables backup creation
func (c *CharacterAssetIntegrator) SetBackupEnabled(enabled bool) {
	c.backupEnabled = enabled
}

// IntegrateAll processes all character.json files in the assets directory
func (c *CharacterAssetIntegrator) IntegrateAll() error {
	charactersPath := filepath.Join(c.assetsPath, "characters")

	if _, err := os.Stat(charactersPath); os.IsNotExist(err) {
		return fmt.Errorf("characters directory not found: %s", charactersPath)
	}

	results := c.processAllCharacterFiles(charactersPath)
	c.printIntegrationSummary(results)

	return nil
}

// processAllCharacterFiles walks through and processes all character.json files
func (c *CharacterAssetIntegrator) processAllCharacterFiles(charactersPath string) *integrationResults {
	results := &integrationResults{
		processedFiles: []string{},
		skippedFiles:   []string{},
		errorFiles:     []string{},
	}

	filepath.WalkDir(charactersPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "character.json") {
			return nil
		}

		c.processFileAndTrackResults(path, results)
		return nil
	})

	return results
}

// integrationResults tracks the results of processing character files
type integrationResults struct {
	processedFiles []string
	skippedFiles   []string
	errorFiles     []string
}

// processFileAndTrackResults processes a single file and tracks the result
func (c *CharacterAssetIntegrator) processFileAndTrackResults(path string, results *integrationResults) {
	fmt.Printf("Processing: %s\n", path)

	err := c.processCharacterFile(path)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		results.errorFiles = append(results.errorFiles, path)
	} else {
		results.processedFiles = append(results.processedFiles, path)
	}
}

// printIntegrationSummary prints a summary of the integration results
func (c *CharacterAssetIntegrator) printIntegrationSummary(results *integrationResults) {
	fmt.Printf("\n=== Integration Summary ===\n")
	fmt.Printf("Processed: %d files\n", len(results.processedFiles))
	fmt.Printf("Skipped: %d files\n", len(results.skippedFiles))
	fmt.Printf("Errors: %d files\n", len(results.errorFiles))

	if len(results.errorFiles) > 0 {
		fmt.Printf("\nFiles with errors:\n")
		for _, file := range results.errorFiles {
			fmt.Printf("  - %s\n", file)
		}
	}
}

// processCharacterFile processes a single character.json file
func (c *CharacterAssetIntegrator) processCharacterFile(filePath string) error {
	rawData, originalData, err := c.readAndParseCharacterFile(filePath)
	if err != nil {
		return err
	}

	if hasLLMConfig(rawData) {
		fmt.Printf("  Already has LLM config, skipping\n")
		return nil
	}

	err = c.updateCharacterWithLLMConfig(filePath, rawData, originalData)
	if err != nil {
		return err
	}

	return nil
}

// readAndParseCharacterFile reads and parses a character JSON file
func (c *CharacterAssetIntegrator) readAndParseCharacterFile(filePath string) (map[string]interface{}, []byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return rawData, data, nil
}

// updateCharacterWithLLMConfig adds LLM configuration and updates the file
func (c *CharacterAssetIntegrator) updateCharacterWithLLMConfig(filePath string, rawData map[string]interface{}, originalData []byte) error {
	personalityData := extractPersonalityData(rawData)
	addLLMConfiguration(rawData, personalityData)

	err := c.createBackupIfEnabled(filePath, originalData)
	if err != nil {
		return err
	}

	return c.writeUpdatedCharacterFile(filePath, rawData)
}

// createBackupIfEnabled creates a backup file if backup is enabled
func (c *CharacterAssetIntegrator) createBackupIfEnabled(filePath string, originalData []byte) error {
	if c.backupEnabled && !c.dryRun {
		backupPath := filePath + ".backup"
		if err := os.WriteFile(backupPath, originalData, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("  Created backup: %s\n", backupPath)
	}
	return nil
}

// writeUpdatedCharacterFile writes the updated character data to file
func (c *CharacterAssetIntegrator) writeUpdatedCharacterFile(filePath string, rawData map[string]interface{}) error {
	if !c.dryRun {
		updatedData, err := json.MarshalIndent(rawData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal updated JSON: %w", err)
		}

		if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
			return fmt.Errorf("failed to write updated file: %w", err)
		}

		fmt.Printf("  Updated with LLM configuration\n")
	} else {
		fmt.Printf("  Would add LLM configuration (dry run)\n")
	}

	return nil
}

// hasLLMConfig checks if the character already has LLM configuration
func hasLLMConfig(data map[string]interface{}) bool {
	dialogBackend, exists := data["dialogBackend"]
	if !exists {
		return false
	}

	if backendMap, ok := dialogBackend.(map[string]interface{}); ok {
		if backends, exists := backendMap["backends"]; exists {
			if backendsMap, ok := backends.(map[string]interface{}); ok {
				_, hasLLM := backendsMap["llm"]
				return hasLLM
			}
		}
	}

	return false
}

// extractDialogResponses extracts training data from existing dialog responses
func extractDialogResponses(data map[string]interface{}) []string {
	var trainingData []string

	dialogList := getDialogList(data)
	if dialogList == nil {
		return trainingData
	}

	return extractResponsesFromDialogs(dialogList)
}

// getDialogList retrieves and validates the dialogs array from character data
func getDialogList(data map[string]interface{}) []interface{} {
	dialogs, exists := data["dialogs"]
	if !exists {
		return nil
	}

	dialogList, ok := dialogs.([]interface{})
	if !ok {
		return nil
	}

	return dialogList
}

// extractResponsesFromDialogs extracts response strings from dialog objects
func extractResponsesFromDialogs(dialogList []interface{}) []string {
	var trainingData []string

	for _, dialog := range dialogList {
		dialogResponses := extractResponsesFromDialog(dialog)
		trainingData = append(trainingData, dialogResponses...)
	}

	return trainingData
}

// extractResponsesFromDialog extracts response strings from a single dialog object
func extractResponsesFromDialog(dialog interface{}) []string {
	var responses []string

	dialogMap, ok := dialog.(map[string]interface{})
	if !ok {
		return responses
	}

	responseList := getResponseList(dialogMap)
	if responseList == nil {
		return responses
	}

	return convertResponsesToStrings(responseList)
}

// getResponseList retrieves and validates the responses array from dialog
func getResponseList(dialogMap map[string]interface{}) []interface{} {
	responses, exists := dialogMap["responses"]
	if !exists {
		return nil
	}

	responseList, ok := responses.([]interface{})
	if !ok {
		return nil
	}

	return responseList
}

// convertResponsesToStrings converts response interfaces to valid string responses
func convertResponsesToStrings(responseList []interface{}) []string {
	var responses []string

	for _, response := range responseList {
		responseStr, ok := response.(string)
		if ok && len(responseStr) > 0 {
			responses = append(responses, responseStr)
		}
	}

	return responses
}

// extractMarkovTrainingData extracts existing Markov training data from dialog backend
func extractMarkovTrainingData(data map[string]interface{}) []string {
	var trainingData []string

	markovConfig := extractMarkovConfig(data)
	if markovConfig == nil {
		return trainingData
	}

	return extractTrainingDataFromMarkovConfig(markovConfig)
}

// extractMarkovConfig retrieves the Markov configuration from dialog backend data
func extractMarkovConfig(data map[string]interface{}) map[string]interface{} {
	dialogBackend := getDialogBackend(data)
	if dialogBackend == nil {
		return nil
	}

	backends := getBackendsConfig(dialogBackend)
	if backends == nil {
		return nil
	}

	markovConfig, exists := backends["markov_chain"]
	if !exists {
		return nil
	}

	markovMap, ok := markovConfig.(map[string]interface{})
	if !ok {
		return nil
	}

	return markovMap
}

// getDialogBackend extracts and validates the dialog backend configuration
func getDialogBackend(data map[string]interface{}) map[string]interface{} {
	dialogBackend, exists := data["dialogBackend"]
	if !exists {
		return nil
	}

	backendMap, ok := dialogBackend.(map[string]interface{})
	if !ok {
		return nil
	}

	return backendMap
}

// getBackendsConfig extracts and validates the backends configuration
func getBackendsConfig(dialogBackend map[string]interface{}) map[string]interface{} {
	backends, exists := dialogBackend["backends"]
	if !exists {
		return nil
	}

	backendsMap, ok := backends.(map[string]interface{})
	if !ok {
		return nil
	}

	return backendsMap
}

// extractTrainingDataFromMarkovConfig extracts training data strings from Markov config
func extractTrainingDataFromMarkovConfig(markovConfig map[string]interface{}) []string {
	var trainingData []string

	training, exists := markovConfig["trainingData"]
	if !exists {
		return trainingData
	}

	trainingList, ok := training.([]interface{})
	if !ok {
		return trainingData
	}

	for _, item := range trainingList {
		itemStr, ok := item.(string)
		if ok && len(itemStr) > 0 {
			trainingData = append(trainingData, itemStr)
		}
	}

	return trainingData
}

// generateDefaultTrainingData creates default training data based on character name
func generateDefaultTrainingData(data map[string]interface{}) []string {
	name := "Companion"
	if nameVal, exists := data["name"]; exists {
		if nameStr, ok := nameVal.(string); ok {
			name = nameStr
		}
	}

	return []string{
		fmt.Sprintf("Hello! I'm %s, and I'm happy to see you! 😊", name),
		"How are you doing today? I hope you're having a wonderful time!",
		"Thanks for spending time with me! Your company means everything.",
		"I'm here whenever you need a friend or just want to chat!",
		"Let's make today amazing together! What would you like to do?",
	}
}

// limitTrainingDataSize ensures training data doesn't exceed the specified limit
func limitTrainingDataSize(trainingData []string, maxSize int) []string {
	if len(trainingData) > maxSize {
		return trainingData[:maxSize]
	}
	return trainingData
}

// extractPersonalityData extracts training data from existing Markov configuration
func extractPersonalityData(data map[string]interface{}) []string {
	// Try extracting from dialog responses first
	trainingData := extractDialogResponses(data)

	// If no data found, try extracting from Markov backend
	if len(trainingData) == 0 {
		trainingData = extractMarkovTrainingData(data)
	}

	// If still no data found, generate default training data
	if len(trainingData) == 0 {
		trainingData = generateDefaultTrainingData(data)
	}

	// Limit to first 5 entries for personality extraction
	return limitTrainingDataSize(trainingData, 5)
}

// addLLMConfiguration adds LLM backend configuration to the character data
func addLLMConfiguration(data map[string]interface{}, personalityData []string) {
	llmConfig := createLLMBackendConfig(personalityData)
	llmConfigJSON, _ := json.Marshal(llmConfig)

	dialogBackend := getOrCreateDialogBackend(data)
	setDialogBackendDefaults(dialogBackend)

	backends := getOrCreateBackends(dialogBackend)
	addLLMBackendToConfig(backends, llmConfigJSON)

	dialogBackend["backends"] = backends
	data["dialogBackend"] = dialogBackend
}

// createLLMBackendConfig creates a new LLM backend configuration with personality data
func createLLMBackendConfig(personalityData []string) LLMBackendConfig {
	return LLMBackendConfig{
		ModelPath:        "/models/tinyllama-1.1b-q4.gguf",
		MaxTokens:        50,
		Temperature:      0.8,
		TopP:             0.9,
		ContextSize:      2048,
		Threads:          4,
		MaxHistoryLength: 5,
		TimeoutMs:        2000,
		FallbackEnabled:  true,
		MarkovConfig: MarkovChainConfig{
			ChainOrder:     2,
			MinWords:       3,
			MaxWords:       12,
			TemperatureMin: 0.4,
			TemperatureMax: 0.8,
			UsePersonality: true,
			TrainingData:   personalityData,
			FallbackPhrases: []string{
				"Hi there! 👋",
				"How can I help you?",
				"Thanks for being here with me!",
			},
		},
	}
}

// getOrCreateDialogBackend retrieves or creates the dialog backend configuration
func getOrCreateDialogBackend(data map[string]interface{}) map[string]interface{} {
	var dialogBackend map[string]interface{}
	if existing, exists := data["dialogBackend"]; exists {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			dialogBackend = existingMap
		} else {
			dialogBackend = make(map[string]interface{})
		}
	} else {
		dialogBackend = make(map[string]interface{})
	}
	return dialogBackend
}

// setDialogBackendDefaults sets default values for dialog backend configuration
func setDialogBackendDefaults(dialogBackend map[string]interface{}) {
	if _, exists := dialogBackend["enabled"]; !exists {
		dialogBackend["enabled"] = true
	}
	if _, exists := dialogBackend["defaultBackend"]; !exists {
		dialogBackend["defaultBackend"] = "llm"
	}
	if _, exists := dialogBackend["fallbackChain"]; !exists {
		dialogBackend["fallbackChain"] = []string{"markov_chain", "simple_random"}
	}
	if _, exists := dialogBackend["confidenceThreshold"]; !exists {
		dialogBackend["confidenceThreshold"] = 0.6
	}
	if _, exists := dialogBackend["memoryEnabled"]; !exists {
		dialogBackend["memoryEnabled"] = true
	}
	if _, exists := dialogBackend["learningEnabled"]; !exists {
		dialogBackend["learningEnabled"] = false
	}
}

// getOrCreateBackends retrieves or creates the backends configuration
func getOrCreateBackends(dialogBackend map[string]interface{}) map[string]interface{} {
	var backends map[string]interface{}
	if existing, exists := dialogBackend["backends"]; exists {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			backends = existingMap
		} else {
			backends = make(map[string]interface{})
		}
	} else {
		backends = make(map[string]interface{})
	}
	return backends
}

// addLLMBackendToConfig adds the LLM backend configuration to the backends map
func addLLMBackendToConfig(backends map[string]interface{}, llmConfigJSON []byte) {
	var llmConfigMap map[string]interface{}
	json.Unmarshal(llmConfigJSON, &llmConfigMap)
	backends["llm"] = llmConfigMap
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <assets_path> [--dry-run] [--no-backup]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nThis tool automatically adds LLM backend configuration to existing character.json files.\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --dry-run     Show what would be changed without modifying files\n")
		fmt.Fprintf(os.Stderr, "  --no-backup   Don't create backup files (.backup)\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s ./assets --dry-run\n", os.Args[0])
		os.Exit(1)
	}

	assetsPath := os.Args[1]
	integrator := NewCharacterAssetIntegrator(assetsPath)

	// Parse command line options
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--dry-run":
			integrator.SetDryRun(true)
		case "--no-backup":
			integrator.SetBackupEnabled(false)
		default:
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", os.Args[i])
			os.Exit(1)
		}
	}

	if integrator.dryRun {
		fmt.Println("=== DRY RUN MODE - No files will be modified ===")
	}

	fmt.Printf("Character Asset LLM Integration Tool\n")
	fmt.Printf("Assets path: %s\n", assetsPath)
	fmt.Printf("Backup enabled: %v\n", integrator.backupEnabled)
	fmt.Println()

	err := integrator.IntegrateAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Integration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nIntegration completed successfully!")
}
