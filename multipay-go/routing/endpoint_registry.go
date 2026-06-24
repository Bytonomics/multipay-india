package routing

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

const maxEndpoints = 40

// EndpointRegistry manages webhook handlers registered for specific provider+accountID combinations.
// It is thread-safe for concurrent access using sync.RWMutex.
type EndpointRegistry struct {
	mu       sync.RWMutex
	handlers map[string]map[string]domain.WebhookEventHandler // provider -> accountID -> handler
}

// NewEndpointRegistry creates a new EndpointRegistry instance.
func NewEndpointRegistry() *EndpointRegistry {
	return &EndpointRegistry{
		handlers: make(map[string]map[string]domain.WebhookEventHandler),
	}
}

// Register registers a webhook handler for a specific provider+accountID combination.
// It returns an error if a handler is already registered for that provider+accountID combo
// or if the total endpoint count would exceed the maximum.
func (r *EndpointRegistry) Register(provider domain.Provider, accountID string, handler domain.WebhookEventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	if accountID == "" {
		return errors.New("accountID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	providerStr := provider.String()

	// Count total registered endpoints across all providers
	total := 0
	for _, accounts := range r.handlers {
		total += len(accounts)
	}
	if total >= maxEndpoints {
		return fmt.Errorf("endpoint registration limit reached (%d); maximum %d webhook endpoints are supported", total, maxEndpoints)
	}

	// Check if provider entry exists
	if _, ok := r.handlers[providerStr]; !ok {
		r.handlers[providerStr] = make(map[string]domain.WebhookEventHandler)
	}

	// Check if handler already registered for this provider+accountID
	if _, exists := r.handlers[providerStr][accountID]; exists {
		return fmt.Errorf("handler already registered for provider %s, accountID %s", providerStr, accountID)
	}

	r.handlers[providerStr][accountID] = handler
	return nil
}

// Lookup returns the webhook handler for a specific provider+accountID combination.
// It returns ErrProviderNotFound if no handler is registered.
func (r *EndpointRegistry) Lookup(provider domain.Provider, accountID string) (domain.WebhookEventHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerStr := provider.String()

	// Check if provider exists
	providerHandlers, ok := r.handlers[providerStr]
	if !ok {
		return nil, domain.ErrProviderNotFound
	}

	// Check if accountID exists for this provider
	handler, ok := providerHandlers[accountID]
	if !ok {
		return nil, domain.ErrProviderNotFound
	}

	return handler, nil
}

// LookupAll returns a list of all registered accountIDs for a given provider.
// It returns an empty slice if the provider has no registered handlers.
func (r *EndpointRegistry) LookupAll(provider domain.Provider) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerStr := provider.String()

	providerHandlers, ok := r.handlers[providerStr]
	if !ok {
		return []string{}
	}

	// Collect all accountIDs for this provider
	accountIDs := make([]string, 0, len(providerHandlers))
	for accountID := range providerHandlers {
		accountIDs = append(accountIDs, accountID)
	}

	return accountIDs
}
