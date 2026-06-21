package orchestration

import (
	"context"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
)

// CapabilityService provides thin wrappers around SupportMatrix for capability queries.
// It is a simple orchestration layer that delegates all capability logic to SupportMatrix.
type CapabilityService struct {
	supportMatrix *capabilities.SupportMatrix
}

// NewCapabilityService creates a new CapabilityService with the given SupportMatrix.
func NewCapabilityService(supportMatrix *capabilities.SupportMatrix) *CapabilityService {
	return &CapabilityService{
		supportMatrix: supportMatrix,
	}
}

// Supports checks if a provider supports the given capability.
// Returns true if supported, false if the provider is not found or doesn't support the capability.
// The context is accepted for future instrumentation but is not used in the current implementation.
func (s *CapabilityService) Supports(ctx context.Context, provider domain.Provider, cap domain.Capability) bool {
	if s.supportMatrix == nil {
		return false
	}
	return s.supportMatrix.Supports(provider, cap)
}

// AllCapabilities returns all capabilities supported by the given provider.
// Returns an empty slice if the provider is not registered.
// The context is accepted for future instrumentation but is not used in the current implementation.
func (s *CapabilityService) AllCapabilities(ctx context.Context, provider domain.Provider) []domain.Capability {
	if s.supportMatrix == nil {
		return []domain.Capability{}
	}

	// Manually collect all known capabilities for the provider
	// by querying the SupportMatrix for each standard capability.
	// This is a simple implementation that covers all capabilities defined in domain.
	allCaps := []domain.Capability{
		// Core shared capabilities
		domain.CapOrderCreate,
		domain.CapOrderFetch,
		domain.CapPaymentFetch,
		domain.CapPaymentList,
		domain.CapPaymentPay,
		domain.CapRefundCreate,
		domain.CapRefundFetch,
		domain.CapRefundList,
		domain.CapInstrumentFetch,
		domain.CapInstrumentList,
		domain.CapInstrumentDelete,
		domain.CapPaymentLinkCreate,
		domain.CapPaymentLinkFetch,
		domain.CapPaymentLinkCancel,

		// Cashfree-specific capabilities
		domain.CapInstrumentCryptogram,
		domain.CapOfferCreate,
		domain.CapOfferFetch,
		domain.CapEligibilityFetch,

		// Razorpay-specific capabilities
		domain.CapOrderUpdate,
		domain.CapOrderList,
		domain.CapPaymentCapture,
		domain.CapRefundUpdate,
		domain.CapCustomerCreate,
		domain.CapCustomerFetch,
		domain.CapCustomerEdit,
		domain.CapCustomerList,
		domain.CapWebhookCreate,
		domain.CapWebhookFetch,
		domain.CapWebhookEdit,
		domain.CapWebhookDelete,
		domain.CapWebhookList,
		domain.CapSubscriptionCreate,
		domain.CapSubscriptionFetch,
		domain.CapSubscriptionList,
		domain.CapPlanCreate,
		domain.CapPlanFetch,
		domain.CapPlanList,
		domain.CapPaymentLinkUpdate,
		domain.CapPaymentLinkNotify,
		domain.CapPaymentLinkList,
		domain.CapUPICreate,
		domain.CapVPAValidate,
	}

	// Filter to only those supported by this provider
	var supported []domain.Capability
	for _, cap := range allCaps {
		if s.supportMatrix.Supports(provider, cap) {
			supported = append(supported, cap)
		}
	}

	return supported
}

// Describe returns a human-readable description of a capability for the given provider.
// Returns an empty string if no description is available or the provider is not found.
// The context is accepted for future instrumentation but is not used in the current implementation.
func (s *CapabilityService) Describe(ctx context.Context, provider domain.Provider, cap domain.Capability) string {
	if s.supportMatrix == nil {
		return ""
	}
	return s.supportMatrix.Describe(provider, cap)
}
