package main

import (
	"sync"
)

// StateManager tracks variant availability across polls
type StateManager struct {
	mu    sync.RWMutex
	state map[int64]bool // variantID -> isAvailable
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		state: make(map[int64]bool),
	}
}

// CheckAndUpdate checks for changes and updates state
// Returns list of stock changes detected
func (sm *StateManager) CheckAndUpdate(products []Product) []StockChange {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var changes []StockChange

	for _, product := range products {
		for _, variant := range product.Variants {
			wasAvailable, exists := sm.state[variant.ID]

			// Detect change
			if variant.Available != wasAvailable {
				changes = append(changes, StockChange{
					ProductID:     product.ID,
					ProductTitle:  product.Title,
					ProductHandle: product.Handle,
					VariantID:     variant.ID,
					VariantTitle:  variant.Title,
					VariantPrice:  variant.Price,
					VariantSKU:    variant.SKU,
					WasAvailable:  wasAvailable,
					IsAvailable:   variant.Available,
				})
			}

			// Update state
			sm.state[variant.ID] = variant.Available

			// If this is the first time seeing this variant and it's available,
			// don't count it as a "new" stock (only notify on transitions)
			if !exists && variant.Available {
				// Remove from changes if it was just added
				if len(changes) > 0 && changes[len(changes)-1].VariantID == variant.ID {
					changes = changes[:len(changes)-1]
				}
			}
		}
	}

	return changes
}

// GetState returns current availability for a variant
func (sm *StateManager) GetState(variantID int64) (available bool, exists bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	available, exists = sm.state[variantID]
	return
}

// Reset clears all state
func (sm *StateManager) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state = make(map[int64]bool)
}

// Size returns the number of tracked variants
func (sm *StateManager) Size() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.state)
}
