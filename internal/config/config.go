// internal/config/config.go
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all configuration options for the application
type Config struct {
	// Server settings
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	
	// Rules settings
	RulesDirs     []string `json:"rules_dirs"`
	RulesFiles    []string `json:"rules_files"`
	
	// Logging settings
	LogLevel      string   `json:"log_level"`
	LogFile       string   `json:"log_file"`
	
	// Additional settings
	MaxCycles     uint64      `json:"max_cycles"`
	ConfigFile    string   // Not stored in JSON, used for command line only
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Host:       "0.0.0.0",
		Port:       8080,
		RulesDirs:  []string{"config/rules"},
		RulesFiles: []string{},
		LogLevel:   "info",
		LogFile:    "",  // Empty means stdout
		MaxCycles:  5,
	}
}

// LoadConfig loads configuration with the following precedence:
// 1. Command-line arguments (highest)
// 2. Environment variables
// 3. Configuration file
// 4. Default values (lowest)
func LoadConfig() (*Config, error) {
	// Start with defaults
	config := DefaultConfig()
	
	// Parse command-line flags but don't fail on unknown flags yet
	cmdConfig := flag.NewFlagSet("config", flag.ContinueOnError)
	cmdConfig.SetOutput(os.Stdout)
	
	// Define flags but don't process them yet
	configFile := cmdConfig.String("config", "", "Path to configuration file")
	// Just parse the config flag first to see if we need to load a config file
	cmdConfig.Parse(os.Args[1:])
	
	// Load from config file if specified
	if *configFile != "" {
		config.ConfigFile = *configFile
		if err := config.loadFromFile(*configFile); err != nil {
			return nil, fmt.Errorf("error loading config file: %v", err)
		}
	}
	
	// Load from environment variables
	config.loadFromEnv()
	
	// Reset the flags parser to process all flags
	cmdConfig = flag.NewFlagSet("config", flag.ExitOnError)
	
	// Server settings
	cmdHost := cmdConfig.String("host", config.Host, "Host to listen on")
	cmdPort := cmdConfig.Int("port", config.Port, "Port to listen on")
	
	// Rules settings
	cmdRulesDirs := cmdConfig.String("rules-dirs", strings.Join(config.RulesDirs, ","), "Comma-separated list of rules directories")
	cmdRulesFiles := cmdConfig.String("rules-files", strings.Join(config.RulesFiles, ","), "Comma-separated list of specific rule files")
	
	// Logging settings
	cmdLogLevel := cmdConfig.String("log-level", config.LogLevel, "Log level (debug, info, warn, error)")
	cmdLogFile := cmdConfig.String("log-file", config.LogFile, "Log file (empty for stdout)")
	
	// Additional settings
	cmdMaxCycles := cmdConfig.Uint64("max-cycles", config.MaxCycles, "Maximum rule execution cycles")
	
	// Redundant config file flag
	cmdConfig.String("config", config.ConfigFile, "Path to configuration file")
	
	// Parse command line arguments, overriding previous values
	cmdConfig.Parse(os.Args[1:])
	
	// Apply command line values if explicitly provided
	if cmdConfig.Lookup("host").Value.String() != config.Host {
		config.Host = *cmdHost
	}
	if cmdConfig.Lookup("port").Value.String() != fmt.Sprintf("%d", config.Port) {
		config.Port = *cmdPort
	}
	if cmdConfig.Lookup("rules-dirs").Value.String() != strings.Join(config.RulesDirs, ",") {
		config.RulesDirs = splitAndTrim(*cmdRulesDirs)
	}
	if cmdConfig.Lookup("rules-files").Value.String() != strings.Join(config.RulesFiles, ",") {
		config.RulesFiles = splitAndTrim(*cmdRulesFiles)
	}
	if cmdConfig.Lookup("log-level").Value.String() != config.LogLevel {
		config.LogLevel = *cmdLogLevel
	}
	if cmdConfig.Lookup("log-file").Value.String() != config.LogFile {
		config.LogFile = *cmdLogFile
	}
	if cmdConfig.Lookup("max-cycles").Value.String() != fmt.Sprintf("%d", config.MaxCycles) {
		config.MaxCycles = *cmdMaxCycles
	}
	
	return config, validateConfig(config)
}

// loadFromFile loads configuration from a JSON file
func (c *Config) loadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	return decoder.Decode(c)
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	// Helper function to get env var if it exists
	getEnv := func(key, defaultValue string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return defaultValue
	}
	
	// Server settings
	if host := getEnv("DCS_ICE_HOST", ""); host != "" {
		c.Host = host
	}
	if port := getEnv("DCS_ICE_PORT", ""); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}
	
	// Rules settings
	if rulesDirs := getEnv("DCS_ICE_RULES_DIRS", ""); rulesDirs != "" {
		c.RulesDirs = splitAndTrim(rulesDirs)
	}
	if rulesFiles := getEnv("DCS_ICE_RULES_FILES", ""); rulesFiles != "" {
		c.RulesFiles = splitAndTrim(rulesFiles)
	}
	
	// Logging settings
	if logLevel := getEnv("DCS_ICE_LOG_LEVEL", ""); logLevel != "" {
		c.LogLevel = logLevel
	}
	if logFile := getEnv("DCS_ICE_LOG_FILE", ""); logFile != "" {
		c.LogFile = logFile
	}
	
	// Additional settings
	if maxCycles := getEnv("DCS_ICE_MAX_CYCLES", ""); maxCycles != "" {
		if mc, err := strconv.ParseUint(maxCycles,10,64); err == nil {
			c.MaxCycles = mc
		}
	}
}

// splitAndTrim splits a comma-separated string and trims spaces
func splitAndTrim(s string) []string {
	if s == "" {
		return []string{}
	}
	
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

// validateConfig ensures the configuration is valid
func validateConfig(c *Config) error {
	// Validate rules directories
	for _, dir := range c.RulesDirs {
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("rules directory does not exist: %s", dir)
			}
			return fmt.Errorf("error accessing rules directory %s: %v", dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("specified rules path is not a directory: %s", dir)
		}
	}
	
	// Validate rules files
	for _, file := range c.RulesFiles {
		info, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("rules file does not exist: %s", file)
			}
			return fmt.Errorf("error accessing rules file %s: %v", file, err)
		}
		if info.IsDir() {
			return fmt.Errorf("specified rules file is a directory: %s", file)
		}
		if filepath.Ext(file) != ".grl" {
			return fmt.Errorf("rules file does not have .grl extension: %s", file)
		}
	}
	
	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	
	// Validate port range
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	
	// Validate max cycles
	if c.MaxCycles < 1 {
		return fmt.Errorf("max cycles must be at least 1")
	}
	
	return nil
}
