package rules

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
	"strings"

    "github.com/hyperjumptech/grule-rule-engine/ast"
    "github.com/hyperjumptech/grule-rule-engine/builder"
    "github.com/hyperjumptech/grule-rule-engine/engine"
    "github.com/hyperjumptech/grule-rule-engine/pkg"

    "github.com/bass4/dcs-ice/pkg/models"
)

// BattleEngine constants
const (
    // BattleKnowledgeBaseName is the name of the knowledge base for the battle engine
    BattleKnowledgeBaseName = "DCS-ICE-Battle-Rules"
    
    // BattleKnowledgeBaseVersion is the version of the knowledge base for the battle engine
    BattleKnowledgeBaseVersion = "0.0.1"
)

// BattleEngine handles rule evaluation for battlefield events
type BattleEngine struct {
    knowledgeLibrary *ast.KnowledgeLibrary
    engine           *engine.GruleEngine
    rulePath         string
    mutex            sync.RWMutex
}

// NewBattleEngine creates a new battle rule engine
func NewBattleEngine(ruleDir string) (*BattleEngine, error) {
    knowledgeLibrary := ast.NewKnowledgeLibrary()
    gruleEngine := engine.NewGruleEngine()
    be := &BattleEngine{
        knowledgeLibrary: knowledgeLibrary,
        engine:           gruleEngine,
        rulePath:         ruleDir,
    }
    
    // Load initial rules
    if err := be.LoadRules(); err != nil {
        return nil, err
    }
    
    return be, nil
}

// LoadRules loads all rule files from the rule directory
func (be *BattleEngine) LoadRules() error {
    be.mutex.Lock()
    defer be.mutex.Unlock()
    
    // Create a new rule builder
    ruleBuilder := builder.NewRuleBuilder(be.knowledgeLibrary)
    
    // Load rules from directory
    files, err := os.ReadDir(be.rulePath)
    if err != nil {
        return fmt.Errorf("failed to read rules directory: %v", err)
    }
    
    ruleCount := 0
    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".grl" {
            filePath := filepath.Join(be.rulePath, file.Name())
            ruleFile := pkg.NewFileResource(filePath)
            err := ruleBuilder.BuildRuleFromResource(BattleKnowledgeBaseName, BattleKnowledgeBaseVersion, ruleFile)
            if err != nil {
                return fmt.Errorf("failed to build rule from file %s: %v", filePath, err)
            }
            ruleCount++
        }
    }
    
    if ruleCount == 0 {
        return fmt.Errorf("no rule files (.grl) found in directory: %s", be.rulePath)
    }
    
    return nil
}



// ReloadRules reloads rules from the rule directory
func (be *BattleEngine) ReloadRules() error {
    return be.LoadRules()
}

// EvaluateFacts evaluates the provided facts against the loaded rules
func (be *BattleEngine) EvaluateFacts(facts []models.Fact) (*models.BattleContext, error) {
    be.mutex.RLock()
    defer be.mutex.RUnlock()
    
    fmt.Println("Starting fact evaluation with", len(facts), "facts")
    for i, fact := range facts {
        fmt.Printf("Fact %d: Type=%s, Value=%s, Zone=%s\n", i, fact.Type, fact.Value, fact.Zone)
    }
    
    // Get the knowledge base
    kb := be.knowledgeLibrary.GetKnowledgeBase(BattleKnowledgeBaseName, BattleKnowledgeBaseVersion)
    
    // Create a battle context with the facts
    battleContext := models.NewBattleContext(facts)
    
    // Create a new data context for this evaluation
    dataContext := ast.NewDataContext()
    
    // Add the battle context to the data context
    if err := dataContext.Add("Battle", battleContext); err != nil {
        return nil, fmt.Errorf("failed to add battle context to data context: %v", err)
    }
    
    fmt.Println("Executing rules with MaxCycle=1")
    
    // Force single cycle execution
    be.engine.MaxCycle = 1
    
    // Execute the rules - ignore "Max cycle reached" error since we expect it
    err := be.engine.Execute(dataContext, kb)
    if err != nil && !strings.Contains(err.Error(), "Max cycle reached") {
        return nil, fmt.Errorf("rule execution failed: %v", err)
    }
    
    // Check for actions
    actions := battleContext.GetActions()
    fmt.Println("Rule execution generated", len(actions), "actions")
    for i, action := range actions {
        fmt.Printf("Action %d: Type=%s, SubType=%s, Zone=%s\n", i, action.Type, action.SubType, action.Zone)
    }
    
    return battleContext, nil
}