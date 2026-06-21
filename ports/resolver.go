package ports

import (
	"fmt"
	"sync"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// ProviderResolver resolves a payment provider to its corresponding adapter.
type ProviderResolver interface {
	// Resolve returns the adapter for a given payment provider.
	// Returns ErrProviderNotFound if the provider is not registered.
	Resolve(provider domain.Provider) (ProviderAdapter, error)
}

// ProviderRegistry is a registry that manages provider adapters.
// It provides thread-safe registration and resolution of payment provider adapters.
type ProviderRegistry struct {
	adapters map[domain.Provider]ProviderAdapter
	mu       sync.RWMutex
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		adapters: make(map[domain.Provider]ProviderAdapter),
	}
}

// Register registers a provider adapter.
// The adapter must be non-nil. Returns an error if registration fails.
func (r *ProviderRegistry) Register(provider domain.Provider, adapter ProviderAdapter) error {
	if !provider.IsValid() {
		return fmt.Errorf("invalid provider: %s", provider)
	}

	if adapter == nil {
		return fmt.Errorf("adapter cannot be nil for provider: %s", provider)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.adapters[provider] = adapter
	return nil
}

// Resolve resolves a provider to its registered adapter.
// Returns domain.ErrProviderNotFound if the provider is not registered.
func (r *ProviderRegistry) Resolve(provider domain.Provider) (ProviderAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[provider]
	if !exists {
		return nil, domain.ErrProviderNotFound
	}

	return adapter, nil
}

// All returns all registered adapters for a given provider.
// If the provider is not found, returns an empty slice.
// Note: This method is typically used for informational purposes or debugging.
func (r *ProviderRegistry) All() map[domain.Provider]ProviderAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[domain.Provider]ProviderAdapter, len(r.adapters))
	for k, v := range r.adapters {
		result[k] = v
	}
	return result
}
