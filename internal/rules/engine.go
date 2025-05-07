// File: internal/rules/engine.go
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/bass4/dcs-ice/internal/actions"
	"github.com/bass4/dcs-ice/pkg/models"
)

const (
	// KnowledgeBaseName is the name of the knowledge base
	KnowledgeBaseName = "DCS-ICE-Rules"
	
	// KnowledgeBaseVersion is the version of the knowledge base
	KnowledgeBaseVersion = "0.0.1"
)

// RuleEngine handles rule evaluation with optimized processing
type RuleEngine struct {
	knowledgeLibrary *ast.KnowledgeLibrary
	engine           *engine.GruleEngine
	rulePath         string
	mutex            sync.RWMutex
	actionManager    *actions.Manager
}

// NewRuleEngine creates a new rule engine with rules loaded from the specified directory
func NewRuleEngine(ruleDir string, actionManager *actions.Manager) (*RuleEngine, error) {
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	
	// Create the rule engine with optimized settings
	gruleEngine := engine.NewGruleEngine()

	re := &RuleEngine{
		knowledgeLibrary: knowledgeLibrary,
		engine:           gruleEngine,
		rulePath:         ruleDir,
		actionManager:    actionManager,
	}

	// Load initial rules
	if err := re.LoadRules(); err != nil {
		return nil, err
	}

	return re, nil
}

// LoadRules loads all rule files from the rule directory
func (re *RuleEngine) LoadRules() error {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	// Create a new rule builder
	ruleBuilder := builder.NewRuleBuilder(re.knowledgeLibrary)

	// Load rules from directory
	files, err := os.ReadDir(re.rulePath)
	if err != nil {
		return fmt.Errorf("failed to read rules directory: %v", err)
	}

	ruleCount := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".grl") {
			filePath := filepath.Join(re.rulePath, file.Name())
			ruleFile := pkg.NewFileResource(filePath)
			err := ruleBuilder.BuildRuleFromResource(KnowledgeBaseName, KnowledgeBaseVersion, ruleFile)
			if err != nil {
				return fmt.Errorf("failed to build rule from file %s: %v", filePath, err)
			}
			ruleCount++
		}
	}

	if ruleCount == 0 {
		return fmt.Errorf("no rule files (.grl) found in directory: %s", re.rulePath)
	}

	return nil
}

// EvaluateFacts evaluates the provided facts against the loaded rules
// and returns matched rules and actions without blocking
func (re *RuleEngine) EvaluateFacts(facts []models.Fact) (*models.RuleResponse, error) {
	re.mutex.RLock()
	defer re.mutex.RUnlock()

	// Get the knowledge base
	kb := re.knowledgeLibrary.GetKnowledgeBase(KnowledgeBaseName, KnowledgeBaseVersion)
	
	// Create a new data context for this evaluation
	dataContext := ast.NewDataContext()
	
	// Create a response object to store results
	response := &models.RuleResponse{
		MatchedRules: []string{},
		Actions:      []models.Action{},
	}
	
	// Create a fact context to hold all the necessary data
	factContext := &models.FactContext{
		Facts:    facts,
		Response: response,
	}
	
	// Add the fact context to the data context
	if err := dataContext.Add("FactContext", factContext); err != nil {
		return nil, fmt.Errorf("failed to add fact context to data context: %v", err)
	}
	
	// Add the action manager to the data context
	if err := dataContext.Add("ActionManager", re.actionManager); err != nil {
		return nil, fmt.Errorf("failed to add action manager to data context: %v", err)
	}
	
	// Add helper functions to the data context
	helpers := NewHelpers()
	if err := dataContext.Add("Helpers", helpers); err != nil {
		return nil, fmt.Errorf("failed to add helpers to data context: %v", err)
	}
	
	// Execute the rules with optimized settings
	err := re.engine.Execute(dataContext, kb)
	if err != nil {
		return nil, fmt.Errorf("rule execution failed: %v", err)
	}
	
	return response, nil
}

// ReloadRules reloads rules from the rule directory
func (re *RuleEngine) ReloadRules() error {
	return re.LoadRules()
}
