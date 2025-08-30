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

	var processedFiles []string
	var skippedFiles []string
	var errorFiles []string

	err := filepath.WalkDir(charactersPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "character.json") {
			return nil
		}

		fmt.Printf("Processing: %s\n", path)

		err = c.processCharacterFile(path)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			errorFiles = append(errorFiles, path)
		} else {
			processedFiles = append(processedFiles, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk characters directory: %w", err)
	}

	// Print summary
	fmt.Printf("\n=== Integration Summary ===\n")
	fmt.Printf("Processed: %d files\n", len(processedFiles))
	fmt.Printf("Skipped: %d files\n", len(skippedFiles))
	fmt.Printf("Errors: %d files\n", len(errorFiles))

	if len(errorFiles) > 0 {
		fmt.Printf("\nFiles with errors:\n")
		for _, file := range errorFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	return nil
}

// processCharacterFile processes a single character.json file
func (c *CharacterAssetIntegrator) processCharacterFile(filePath string) error {
	// Read existing file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse as generic JSON to preserve all fields
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if already has LLM configuration
	if hasLLMConfig(rawData) {
		fmt.Printf("  Already has LLM config, skipping\n")
		return nil
	}

	// Extract existing personality data for LLM configuration
	personalityData := extractPersonalityData(rawData)

	// Add LLM configuration
	addLLMConfiguration(rawData, personalityData)

	// Create backup if enabled
	if c.backupEnabled && !c.dryRun {
		backupPath := filePath + ".backup"
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("  Created backup: %s\n", backupPath)
	}

	// Write updated file
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

// extractPersonalityData extracts training data from existing Markov configuration
func extractPersonalityData(data map[string]interface{}) []string {
	var trainingData []string

	// Check for existing dialog responses
	if dialogs, exists := data["dialogs"]; exists {
		if dialogList, ok := dialogs.([]interface{}); ok {
			for _, dialog := range dialogList {
				if dialogMap, ok := dialog.(map[string]interface{}); ok {
					if responses, exists := dialogMap["responses"]; exists {
						if responseList, ok := responses.([]interface{}); ok {
							for _, response := range responseList {
								if responseStr, ok := response.(string); ok && len(responseStr) > 0 {
									trainingData = append(trainingData, responseStr)
								}
							}
						}
					}
				}
			}
		}
	}

	// Check for existing Markov training data in dialog backend
	if dialogBackend, exists := data["dialogBackend"]; exists {
		if backendMap, ok := dialogBackend.(map[string]interface{}); ok {
			if backends, exists := backendMap["backends"]; exists {
				if backendsMap, ok := backends.(map[string]interface{}); ok {
					if markovConfig, exists := backendsMap["markov_chain"]; exists {
						if markovMap, ok := markovConfig.(map[string]interface{}); ok {
							if training, exists := markovMap["trainingData"]; exists {
								if trainingList, ok := training.([]interface{}); ok {
									for _, item := range trainingList {
										if itemStr, ok := item.(string); ok && len(itemStr) > 0 {
											trainingData = append(trainingData, itemStr)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// If no training data found, use default based on character name
	if len(trainingData) == 0 {
		name := "Companion"
		if nameVal, exists := data["name"]; exists {
			if nameStr, ok := nameVal.(string); ok {
				name = nameStr
			}
		}

		trainingData = []string{
			fmt.Sprintf("Hello! I'm %s, and I'm happy to see you! 😊", name),
			"How are you doing today? I hope you're having a wonderful time!",
			"Thanks for spending time with me! Your company means everything.",
			"I'm here whenever you need a friend or just want to chat!",
			"Let's make today amazing together! What would you like to do?",
		}
	}

	// Limit to first 5 entries for personality extraction
	if len(trainingData) > 5 {
		trainingData = trainingData[:5]
	}

	return trainingData
}

// addLLMConfiguration adds LLM backend configuration to the character data
func addLLMConfiguration(data map[string]interface{}, personalityData []string) {
	// Create LLM backend configuration
	llmConfig := LLMBackendConfig{
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

	// Serialize LLM config to RawMessage
	llmConfigJSON, _ := json.Marshal(llmConfig)

	// Get or create dialog backend configuration
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

	// Set dialog backend defaults if not present
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

	// Get or create backends configuration
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

	// Add LLM backend configuration
	var llmConfigMap map[string]interface{}
	json.Unmarshal(llmConfigJSON, &llmConfigMap)
	backends["llm"] = llmConfigMap

	dialogBackend["backends"] = backends
	data["dialogBackend"] = dialogBackend
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
