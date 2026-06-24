package capabilities

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// Validator checks whether a capability is supported by a provider.
// It wraps a SupportMatrix and provides early validation before adapter dispatch.
type Validator struct {
	matrix *SupportMatrix
}

// NewValidator constructs a Validator with a capability matrix.
// The matrix must be non-nil.
func NewValidator(matrix *SupportMatrix) *Validator {
	return &Validator{
		matrix: matrix,
	}
}

// RequireCapability returns nil if the provider supports the capability,
// or a domain.CapabilityError if unsupported.
// This method provides early validation before any SDK call.
func (v *Validator) RequireCapability(ctx context.Context, provider domain.Provider, cap Capability) error {
	if v.matrix == nil {
		return domain.NewCapabilityError(provider, string(cap), "validator matrix is nil")
	}

	if !v.matrix.Supports(provider, cap) {
		msg := fmt.Sprintf("provider %s does not support capability %s", provider, cap)
		description := v.matrix.Describe(provider, cap)
		if description != "" {
			msg = fmt.Sprintf("%s (%s)", msg, description)
		}
		return domain.NewCapabilityError(provider, string(cap), msg)
	}

	return nil
}
