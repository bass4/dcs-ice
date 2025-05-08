// internal/rules/rule_engine.go
package rules

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/hyperjumptech/grule-rule-engine/ast"
    "github.com/hyperjumptech/grule-rule-engine/builder"
    "github.com/hyperjumptech/grule-rule-engine/engine"
    "github.com/hyperjumptech/grule-rule-engine/pkg"

    "github.com/bass4/dcs-ice/pkg/models"
)

const (
    KnowledgeBaseName    = "DCS-ICE-Rules"
    KnowledgeBaseVersion = "0.0.1"
)

// RuleEngine handles rule evaluation
type RuleEngine struct {
    knowledgeLibrary *ast.KnowledgeLibrary
    engine           *engine.GruleEngine
    rulePath         string
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(ruleDir string) (*RuleEngine, error) {
    knowledgeLibrary := ast.NewKnowledgeLibrary()
    gruleEngine := engine.NewGruleEngine()
    re := &RuleEngine{
        knowledgeLibrary: knowledgeLibrary,
        engine:           gruleEngine,
        rulePath:         ruleDir,
    }
    
    // Load rules
    if err := re.LoadRules(); err != nil {
        return nil, err
    }
    
    return re, nil
}

// LoadRules loads rules from the specified directory
func (re *RuleEngine) LoadRules() error {
    ruleBuilder := builder.NewRuleBuilder(re.knowledgeLibrary)
    
    files, err := os.ReadDir(re.rulePath)
    if err != nil {
        return fmt.Errorf("failed to read rules directory: %v", err)
    }
    
    ruleCount := 0
    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".grl" {
            filePath := filepath.Join(re.rulePath, file.Name())
            ruleFile := pkg.NewFileResource(filePath)
            
            fmt.Printf("Loading rule file: %s\n", filePath)
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
    
    fmt.Printf("Loaded %d rule files\n", ruleCount)
    return nil
}

// ProcessMessage processes a DCS message through the rules engine
func (re *RuleEngine) ProcessMessage(message *models.Message) ([]models.Action, error) {
    fmt.Printf("Processing message: Event=%s, Zone=%s\n", message.Event, message.Zone)
    
    // Get the knowledge base
    kb := re.knowledgeLibrary.GetKnowledgeBase(KnowledgeBaseName, KnowledgeBaseVersion)
    
    // Create an ActionCollector to store actions
    actionCollector := models.NewActionCollector()
    
    // Create data context
    dataContext := ast.NewDataContext()
    if err := dataContext.Add("Message", message); err != nil {
        return nil, fmt.Errorf("failed to add message to data context: %v", err)
    }
    if err := dataContext.Add("Actions", actionCollector); err != nil {
        return nil, fmt.Errorf("failed to add action collector to data context: %v", err)
    }
    
    // Set max cycle to 1 to avoid loops
    re.engine.MaxCycle = 1
    
    // Execute rules - ignore max cycle error
    err := re.engine.Execute(dataContext, kb)
    if err != nil {
        fmt.Printf("Rule execution warning: %v\n", err)
    }
    
    actions := actionCollector.GetActions()
    fmt.Printf("Generated %d actions\n", len(actions))
    for i, action := range actions {
        fmt.Printf("Action %d: Type=%s, SubType=%s, Zone=%s\n", i, action.Type, action.SubType, action.Zone)
    }
    
    return actions, nil
}

// ProcessMessages processes multiple DCS messages through the rules engine
func (re *RuleEngine) ProcessMessages(messages []*models.Message) ([]models.Action, error) {
    fmt.Printf("Processing %d messages\n", len(messages))
    
    for i, msg := range messages {
        fmt.Printf("Message %d: Event=%s, Zone=%s\n", i, msg.Event, msg.Zone)
    }
    
    // Create a message collection
    messageCollection := models.NewMessageCollection()
    for _, msg := range messages {
        messageCollection.AddMessage(msg)
    }
    
    // Get the knowledge base
    kb := re.knowledgeLibrary.GetKnowledgeBase(KnowledgeBaseName, KnowledgeBaseVersion)
    
    // Create an ActionCollector to store actions
    actionCollector := models.NewActionCollector()
    
    // Create data context
    dataContext := ast.NewDataContext()
    if err := dataContext.Add("Messages", messageCollection); err != nil {
        return nil, fmt.Errorf("failed to add message collection to data context: %v", err)
    }
    if err := dataContext.Add("Actions", actionCollector); err != nil {
        return nil, fmt.Errorf("failed to add action collector to data context: %v", err)
    }
    
    // Set max cycle to 1 to avoid loops
    re.engine.MaxCycle = 1
    
    // Execute rules - ignore max cycle error
    err := re.engine.Execute(dataContext, kb)
    if err != nil {
        fmt.Printf("Rule execution warning: %v\n", err)
    }
    
    actions := actionCollector.GetActions()
    fmt.Printf("Generated %d actions\n", len(actions))
    for i, action := range actions {
        fmt.Printf("Action %d: Type=%s, SubType=%s, Zone=%s\n", i, action.Type, action.SubType, action.Zone)
    }
    
    return actions, nil
}