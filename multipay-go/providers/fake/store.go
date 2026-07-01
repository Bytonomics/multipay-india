// Package fake provides a vendor-free test harness for the multipay-go library.
// It exposes a single FakeAdapter (implementing ports.ProviderAdapter) plus an
// in-memory WebhookStore, a delay Scheduler, and a Harness that wires a real
// client.NewClient around the fake so ALL library flows can be exercised without
// ever hitting a real vendor (Cashfree/Razorpay).
package fake

import (
	"context"
	"sync"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// compile-time interface check
var _ ports.WebhookStore = (*InMemoryWebhookStore)(nil)

// InMemoryWebhookStore is an in-memory ports.WebhookStore for tests. It records
// every raw payload and tracks processed dedupe keys so real webhook dedupe
// semantics are exercised. Safe for concurrent use.
type InMemoryWebhookStore struct {
	mu        sync.Mutex
	processed map[string]bool
	Payloads  [][]byte
}

// NewInMemoryWebhookStore returns a ready-to-use in-memory webhook store.
func NewInMemoryWebhookStore() *InMemoryWebhookStore {
	return &InMemoryWebhookStore{processed: make(map[string]bool)}
}

func (s *InMemoryWebhookStore) key(provider domain.Provider, accountID, dedupeKey string) string {
	return string(provider) + "|" + accountID + "|" + dedupeKey
}

// StoreRawPayload appends the raw payload to the in-memory log.
func (s *InMemoryWebhookStore) StoreRawPayload(ctx context.Context, provider domain.Provider, accountID string, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, len(payload))
	copy(cp, payload)
	s.Payloads = append(s.Payloads, cp)
	return nil
}

// IsDuplicate reports whether an event with the given dedupe key was already marked processed.
func (s *InMemoryWebhookStore) IsDuplicate(ctx context.Context, provider domain.Provider, accountID string, dedupeKey string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.processed[s.key(provider, accountID, dedupeKey)], nil
}

// MarkProcessed records the event's dedupe key as processed.
func (s *InMemoryWebhookStore) MarkProcessed(ctx context.Context, provider domain.Provider, accountID string, dedupeKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processed[s.key(provider, accountID, dedupeKey)] = true
	return nil
}
