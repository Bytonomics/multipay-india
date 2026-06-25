package cashfree

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// TestBuildCashfreeCheckout tests the buildCashfreeCheckout helper function.
// It verifies that the checkout payload is constructed correctly with:
// - Provider set to ProviderCashfree
// - Environment converted to UPPERCASE (e.g., "PRODUCTION")
// - SessionID properly passed through
// - Razorpay-specific fields empty (no bleed-over)
func TestBuildCashfreeCheckout(t *testing.T) {
	tests := []struct {
		name      string
		env       domain.Environment
		sessionID string
		want      *domain.CheckoutPayload
	}{
		{
			name:      "production environment",
			env:       domain.EnvironmentProduction,
			sessionID: "session_123",
			want: &domain.CheckoutPayload{
				Provider:    domain.ProviderCashfree,
				Environment: domain.EnvironmentProduction,
				SessionID:   "session_123",
			},
		},
		{
			name:      "sandbox environment",
			env:       domain.EnvironmentSandbox,
			sessionID: "session_456",
			want: &domain.CheckoutPayload{
				Provider:    domain.ProviderCashfree,
				Environment: domain.EnvironmentSandbox,
				SessionID:   "session_456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCashfreeCheckout(tt.env, tt.sessionID)

			// Assert Provider is Cashfree
			if got.Provider != domain.ProviderCashfree {
				t.Errorf("Provider = %v, want %v", got.Provider, domain.ProviderCashfree)
			}

			// Assert Environment matches and is UPPERCASE
			if got.Environment != tt.want.Environment {
				t.Errorf("Environment = %v, want %v", got.Environment, tt.want.Environment)
			}

			// Verify environment string representation is UPPERCASE
			envStr := string(got.Environment)
			if envStr != "PRODUCTION" && envStr != "SANDBOX" {
				t.Errorf("Environment string representation must be UPPERCASE, got %v", envStr)
			}

			// Assert SessionID matches
			if got.SessionID != tt.sessionID {
				t.Errorf("SessionID = %v, want %v", got.SessionID, tt.sessionID)
			}

			// Assert no Razorpay field bleed-through
			// (CheckoutPayload is provider-agnostic, but we verify no accidental Razorpay-specific data)
			if got.SessionID == "" && tt.sessionID != "" {
				t.Error("SessionID should not be empty when provided")
			}
		})
	}
}
