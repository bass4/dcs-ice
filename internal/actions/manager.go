// File: internal/actions/manager.go
package actions

import (
	"sync"
	"time"

	"github.com/bass4/dcs-ice/pkg/models"
)

// Manager handles action creation and processing
type Manager struct {
	actionHistory []models.Action
	observers     []chan<- models.Action
	mutex         sync.RWMutex
}

// NewManager creates a new action manager
func NewManager() *Manager {
	return &Manager{
		actionHistory: make([]models.Action, 0),
		observers:     make([]chan<- models.Action, 0),
	}
}

// CreateAction creates a new action and adds it to the response
// This method is exposed to the rule engine
func (m *Manager) CreateAction(actionType, target string, parameters map[string]string) models.Action {
	action := models.Action{
		Type:       actionType,
		Target:     target,
		Parameters: parameters,
		Timestamp:  time.Now().UnixNano() / int64(time.Millisecond),
	}

	m.mutex.Lock()
	m.actionHistory = append(m.actionHistory, action)
	m.mutex.Unlock()

	// Notify observers
	m.notifyObservers(action)

	return action
}

// GetRecentActions returns the most recent actions
func (m *Manager) GetRecentActions(limit int) []models.Action {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	historyLength := len(m.actionHistory)
	if limit <= 0 || limit > historyLength {
		limit = historyLength
	}

	start := historyLength - limit
	if start < 0 {
		start = 0
	}

	result := make([]models.Action, limit)
	copy(result, m.actionHistory[start:])

	return result
}

// RegisterObserver registers a channel to receive notifications of new actions
func (m *Manager) RegisterObserver(observer chan<- models.Action) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.observers = append(m.observers, observer)
}

// UnregisterObserver removes a channel from the observers list
func (m *Manager) UnregisterObserver(observer chan<- models.Action) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, ch := range m.observers {
		if ch == observer {
			m.observers = append(m.observers[:i], m.observers[i+1:]...)
			break
		}
	}
}

// notifyObservers sends the action to all registered observers
func (m *Manager) notifyObservers(action models.Action) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, observer := range m.observers {
		select {
		case observer <- action:
			// Action sent successfully
		default:
			// Channel is blocked, skip it
		}
	}
}
