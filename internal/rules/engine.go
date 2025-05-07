package rules

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/yourusername/dcs-ice/pkg/models"
)

// RuleEngine handles rule evaluation
type RuleEngine struct {
	knowledgeBase *ast.KnowledgeBase
	dataContext   *ast.DataContext
	engine        *engine.GruleEngine
}

// NewRuleEngine creates a new rule engine with rules loaded from the specified directory
func NewRuleEngine(ruleDir string) (*RuleEngine, error) {
	knowledgeBase := ast.NewKnowledgeBase("DCS-ICE-Rules")
	dataContext := ast.NewDataContext()
	ruleBuilder := builder.NewRuleBuilder(knowledgeBase)

	// Load rules from directory
	files, err := ioutil.ReadDir(ruleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".grl") {
			filePath := filepath.Join(ruleDir, file.Name())
			ruleFile := pkg.NewFileResource(filePath)
			err := ruleBuilder.BuildRuleFromResource(ruleFile)
			if err != nil {
				return nil, fmt.Errorf("failed to build rule from file %s: %v", filePath, err)
			}
		}
	}

	return &RuleEngine{
		knowledgeBase: knowledgeBase,
		dataContext:   dataContext,
		engine:        engine.NewGruleEngine(),
	}, nil
}

// EvaluateFacts evaluates the provided facts against the loaded rules
func (re *RuleEngine) EvaluateFacts(facts []models.Fact) (*models.RuleResponse, error) {
	// Clear the data context
	re.dataContext = ast.NewDataContext()
	
	// Create a response object
	response := &models.RuleResponse{
		MatchedRules: []string{},
		Actions:      []models.Action{},
	}
	
	// Add the facts and response to the data context
	err := re.dataContext.Add("Facts", facts)
	if err != nil {
		return nil, fmt.Errorf("failed to add facts to data context: %v", err)
	}
	
	err = re.dataContext.Add("Response", response)
	if err != nil {
		return nil, fmt.Errorf("failed to add response to data context: %v", err)
	}
	
	// Execute the rules
	err = re.engine.Execute(re.dataContext, re.knowledgeBase)
	if err != nil {
		return nil, fmt.Errorf("rule execution failed: %v", err)
	}
	
	return response, nil
}
