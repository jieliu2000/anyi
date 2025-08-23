package agent

import (
	"fmt"
	"sync"
	"time"
)

// AgentMemory defines the interface for agent memory management.
type AgentMemory interface {
	// StoreTask stores a task result in memory
	StoreTask(result *TaskResult)

	// GetTaskHistory returns the complete task execution history
	GetTaskHistory() []*TaskResult

	// GetTask retrieves a specific task by objective
	GetTask(objective string) (*TaskResult, error)

	// Clear clears all memory
	Clear()
}

// MemoryItem represents a single item in agent memory.
type MemoryItem struct {
	Key       string    `mapstructure:"key"`
	Value     any       `mapstructure:"value"`
	Timestamp time.Time `mapstructure:"timestamp"`
}

// SimpleMemory is a simple in-memory implementation of AgentMemory.
type SimpleMemory struct {
	tasks []TaskResult
	items map[string]MemoryItem
	mutex sync.RWMutex
}

// NewSimpleMemory creates a new SimpleMemory instance.
func NewSimpleMemory() *SimpleMemory {
	return &SimpleMemory{
		tasks: make([]TaskResult, 0),
		items: make(map[string]MemoryItem),
	}
}

// StoreTask stores a task result in memory.
func (m *SimpleMemory) StoreTask(result *TaskResult) {
	if result == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if task already exists and update it
	for i, task := range m.tasks {
		if task.Objective == result.Objective {
			m.tasks[i] = *result
			return
		}
	}

	// Add new task
	m.tasks = append(m.tasks, *result)
}

// GetTaskHistory returns the complete task execution history.
func (m *SimpleMemory) GetTaskHistory() []*TaskResult {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]*TaskResult, len(m.tasks))
	for i := range m.tasks {
		// Create a copy to avoid race conditions
		taskCopy := m.tasks[i]
		history[i] = &taskCopy
	}

	return history
}

// GetTask retrieves a specific task by objective.
func (m *SimpleMemory) GetTask(objective string) (*TaskResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, task := range m.tasks {
		if task.Objective == objective {
			// Return a copy to avoid race conditions
			taskCopy := task
			return &taskCopy, nil
		}
	}

	return nil, fmt.Errorf("task with objective '%s' not found", objective)
}

// Clear clears all memory.
func (m *SimpleMemory) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.tasks = make([]TaskResult, 0)
	m.items = make(map[string]MemoryItem)
}

// Store stores a key-value pair in memory.
func (m *SimpleMemory) Store(key string, value any) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items[key] = MemoryItem{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}
	return nil
}

// Retrieve retrieves a value by key from memory.
func (m *SimpleMemory) Retrieve(key string) (any, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return nil, nil
	}
	return item.Value, nil
}

// Search searches for memory items by query (simple implementation).
func (m *SimpleMemory) Search(query string) ([]MemoryItem, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []MemoryItem
	for _, item := range m.items {
		// Simple string matching - can be enhanced with more sophisticated search
		if item.Key == query {
			results = append(results, item)
		}
	}
	return results, nil
}
