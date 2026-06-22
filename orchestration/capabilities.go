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
func (s *CapabilityService) Supports(ctx context.Context, provider domain.Provider, cap capabilities.Capability) bool {
	if s.supportMatrix == nil {
		return false
	}
	return s.supportMatrix.Supports(provider, cap)
}

// AllCapabilities returns all capabilities supported by the given provider.
// Returns an empty slice if the provider is not registered.
// The context is accepted for future instrumentation but is not used in the current implementation.
func (s *CapabilityService) AllCapabilities(ctx context.Context, provider domain.Provider) []capabilities.Capability {
	if s.supportMatrix == nil {
		return []capabilities.Capability{}
	}

	// Manually collect all known capabilities for the provider
	// by querying the SupportMatrix for each standard capability.
	// This is a simple implementation that covers all capabilities defined in domain.
	allCaps := []capabilities.Capability{
		// Core shared capabilities
		capabilities.CapOrderCreate,
		capabilities.CapOrderFetch,
		capabilities.CapPaymentFetch,
		capabilities.CapPaymentList,
		capabilities.CapPaymentPay,
		capabilities.CapRefundCreate,
		capabilities.CapRefundFetch,
		capabilities.CapRefundList,
		capabilities.CapInstrumentFetch,
		capabilities.CapInstrumentList,
		capabilities.CapInstrumentDelete,
		capabilities.CapPaymentLinkCreate,
		capabilities.CapPaymentLinkFetch,
		capabilities.CapPaymentLinkCancel,

		// Cashfree-specific capabilities
		capabilities.CapInstrumentCryptogram,
		capabilities.CapOfferCreate,
		capabilities.CapOfferFetch,
		capabilities.CapEligibilityFetch,

		// Razorpay-specific capabilities
		capabilities.CapOrderUpdate,
		capabilities.CapOrderList,
		capabilities.CapPaymentCapture,
		capabilities.CapRefundUpdate,
		capabilities.CapCustomerCreate,
		capabilities.CapCustomerFetch,
		capabilities.CapCustomerEdit,
		capabilities.CapCustomerList,
		capabilities.CapWebhookCreate,
		capabilities.CapWebhookFetch,
		capabilities.CapWebhookEdit,
		capabilities.CapWebhookDelete,
		capabilities.CapWebhookList,
		capabilities.CapSubscriptionCreate,
		capabilities.CapSubscriptionFetch,
		capabilities.CapSubscriptionList,
		capabilities.CapPlanCreate,
		capabilities.CapPlanFetch,
		capabilities.CapPlanList,
		capabilities.CapPaymentLinkUpdate,
		capabilities.CapPaymentLinkNotify,
		capabilities.CapPaymentLinkList,
		capabilities.CapUPICreate,
		capabilities.CapVPAValidate,
	}

	// Filter to only those supported by this provider
	var supported []capabilities.Capability
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
func (s *CapabilityService) Describe(ctx context.Context, provider domain.Provider, cap capabilities.Capability) string {
	if s.supportMatrix == nil {
		return ""
	}
	return s.supportMatrix.Describe(provider, cap)
}
