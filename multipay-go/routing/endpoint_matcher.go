package routing

import (
	"strings"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// EndpointMatcher parses HTTP webhook paths and extracts provider and accountID.
// It validates the path format and ensures the provider is valid.
type EndpointMatcher struct {
	basePath string
}

// NewEndpointMatcher creates a new EndpointMatcher with the given base path.
// Example: basePath = "/webhooks"
func NewEndpointMatcher(basePath string) *EndpointMatcher {
	return &EndpointMatcher{
		basePath: basePath,
	}
}

// Match parses the HTTP path and extracts the provider and accountID.
// The expected path format is: {basePath}/{provider}/{accountID}
// Example: "/webhooks/cashfree/acct_123"
//
// It returns:
//   - (provider, accountID, true) on successful match
//   - ("", "", false) on parse failure or invalid format
//
// Validation includes:
//   - Exact path format with 3 parts (after basePath split)
//   - Non-empty accountID
//   - Valid provider (checked via Provider.IsValid())
func (m *EndpointMatcher) Match(path string) (domain.Provider, string, bool) {
	// Remove basePath prefix if present
	if !strings.HasPrefix(path, m.basePath) {
		return "", "", false
	}

	// Remove the basePath and leading slash
	remainingPath := strings.TrimPrefix(path, m.basePath)
	remainingPath = strings.TrimPrefix(remainingPath, "/")

	// Split into parts: provider and accountID
	parts := strings.Split(remainingPath, "/")

	// Validate we have exactly 2 parts: provider and accountID
	if len(parts) != 2 {
		return "", "", false
	}

	providerStr := parts[0]
	accountID := parts[1]

	// Validate accountID is non-empty
	if accountID == "" {
		return "", "", false
	}

	// Validate provider
	provider := domain.Provider(providerStr)
	if !provider.IsValid() {
		return "", "", false
	}

	return provider, accountID, true
}
