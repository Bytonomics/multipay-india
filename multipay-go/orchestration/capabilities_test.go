package orchestration

import (
	"context"
	"slices"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// allKnownCapabilities lists EVERY Capability constant from capabilities/enums.go.
// When you add a Capability constant, add it here AND to allCaps in AllCapabilities; this test fails when they drift.
var allKnownCapabilities = []capabilities.Capability{
	// Core shared capabilities (16)
	capabilities.CapOrderCreate,
	capabilities.CapOrderFetch,
	capabilities.CapOrderListPayments,
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
	capabilities.CapWebhookConsume,

	// Cashfree-specific capabilities (13)
	capabilities.CapInstrumentCryptogram,
	capabilities.CapOfferCreate,
	capabilities.CapOfferFetch,
	capabilities.CapEligibilityFetch,
	capabilities.CapPaymentLinkListOrders,
	capabilities.CapSettlementOrderFetch,
	capabilities.CapSettlementList,
	capabilities.CapSettlementReconFetch,
	capabilities.CapReconFetch,
	capabilities.CapSubscriptionManualCharge,
	capabilities.CapSubscriptionUpgradeProration,
	capabilities.CapSubscriptionEligibility,
	capabilities.CapSubscriptionCancelAtCycleEnd,

	// Razorpay-specific capabilities (20)
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
	capabilities.CapSubscriptionList,
	capabilities.CapPlanList,
	capabilities.CapPaymentLinkUpdate,
	capabilities.CapPaymentLinkNotify,
	capabilities.CapPaymentLinkList,
	capabilities.CapUPICreate,
	capabilities.CapVPAValidate,

	// Razorpay-specific settlement capabilities (6)
	capabilities.CapSettlementAll,
	capabilities.CapSettlementFetch,
	capabilities.CapSettlementReports,
	capabilities.CapSettlementOnDemandCreate,
	capabilities.CapSettlementOnDemandFetch,
	capabilities.CapSettlementOnDemandList,
}

// TestCapabilityService_AllCapabilities_MatchesSupportMatrix verifies that AllCapabilities
// agrees with SupportMatrix.Supports for every known capability, and never returns duplicates.
func TestCapabilityService_AllCapabilities_MatchesSupportMatrix(t *testing.T) {
	ctx := context.Background()
	svc := NewCapabilityService(capabilities.NewSupportMatrix())

	tests := []struct {
		name     string
		provider domain.Provider
	}{
		{name: "cashfree", provider: domain.ProviderCashfree},
		{name: "razorpay", provider: domain.ProviderRazorpay},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			all := svc.AllCapabilities(ctx, tc.provider)

			seen := make(map[capabilities.Capability]bool, len(all))
			for _, c := range all {
				if seen[c] {
					t.Errorf("provider %q: AllCapabilities returned duplicate capability %q", tc.provider, c)
				}
				seen[c] = true
			}

			for _, c := range allKnownCapabilities {
				got := slices.Contains(all, c)
				want := svc.Supports(ctx, tc.provider, c)
				if got != want {
					t.Errorf("provider %q, capability %q: AllCapabilities contains=%t, but Supports=%t",
						tc.provider, c, got, want)
				}
			}
		})
	}
}

// TestCapabilityService_AllCapabilities_SubscriptionCapabilities verifies the subscription
// capability split per provider, as pinned in capabilities/matrix.go.
func TestCapabilityService_AllCapabilities_SubscriptionCapabilities(t *testing.T) {
	ctx := context.Background()
	svc := NewCapabilityService(capabilities.NewSupportMatrix())

	tests := []struct {
		name        string
		provider    domain.Provider
		wantPresent []capabilities.Capability
		wantAbsent  []capabilities.Capability
	}{
		{
			name:     "cashfree",
			provider: domain.ProviderCashfree,
			wantPresent: []capabilities.Capability{
				capabilities.CapSubscriptionManualCharge,
				capabilities.CapSubscriptionUpgradeProration,
				capabilities.CapSubscriptionEligibility,
			},
			wantAbsent: []capabilities.Capability{
				capabilities.CapSubscriptionCancelAtCycleEnd,
				capabilities.CapSubscriptionList,
				capabilities.CapPlanList,
			},
		},
		{
			name:     "razorpay",
			provider: domain.ProviderRazorpay,
			wantPresent: []capabilities.Capability{
				capabilities.CapSubscriptionManualCharge,
				capabilities.CapSubscriptionUpgradeProration,
				capabilities.CapSubscriptionCancelAtCycleEnd,
				capabilities.CapSubscriptionList,
				capabilities.CapPlanList,
			},
			wantAbsent: []capabilities.Capability{
				capabilities.CapSubscriptionEligibility,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.AllCapabilities(ctx, tc.provider)

			for _, c := range tc.wantPresent {
				if !slices.Contains(got, c) {
					t.Errorf("provider %q: expected AllCapabilities to contain %q, but it is missing",
						tc.provider, c)
				}
			}
			for _, c := range tc.wantAbsent {
				if slices.Contains(got, c) {
					t.Errorf("provider %q: expected AllCapabilities to NOT contain %q, but it is present",
						tc.provider, c)
				}
			}
		})
	}
}

// TestCapabilityService_AllCapabilities_EmptyCases verifies the empty results for an unregistered
// provider and for a service with no SupportMatrix.
func TestCapabilityService_AllCapabilities_EmptyCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		svc  *CapabilityService
		// provider is the provider queried.
		provider domain.Provider
		// requireNonNil is true when the implementation must return a non-nil empty slice.
		requireNonNil bool
	}{
		{
			name:     "unknown provider",
			svc:      NewCapabilityService(capabilities.NewSupportMatrix()),
			provider: domain.Provider("unknown"),
		},
		{
			name:          "nil matrix",
			svc:           &CapabilityService{},
			provider:      domain.ProviderCashfree,
			requireNonNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.svc.AllCapabilities(ctx, tc.provider)

			if len(got) != 0 {
				t.Errorf("provider %q: expected empty AllCapabilities, got %d entries", tc.provider, len(got))
			}
			if tc.requireNonNil && got == nil {
				t.Errorf("provider %q: expected non-nil empty slice, got nil", tc.provider)
			}
		})
	}
}
