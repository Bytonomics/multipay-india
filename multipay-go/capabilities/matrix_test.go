package capabilities

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// TestCapabilityMatrix_CancelAtCycleEndEntry verifies the capability matrix entry
// for CapSubscriptionCancelAtCycleEnd: Razorpay=true, Cashfree=false.
func TestCapabilityMatrix_CancelAtCycleEndEntry(t *testing.T) {
	matrix := NewSupportMatrix()

	// Razorpay should support cancel at cycle end
	if !matrix.Supports(domain.ProviderRazorpay, CapSubscriptionCancelAtCycleEnd) {
		t.Errorf("expected Razorpay to support CapSubscriptionCancelAtCycleEnd")
	}

	// Cashfree should NOT support cancel at cycle end (from domain/capabilities/enums.go line 55)
	if matrix.Supports(domain.ProviderCashfree, CapSubscriptionCancelAtCycleEnd) {
		t.Errorf("expected Cashfree to NOT support CapSubscriptionCancelAtCycleEnd")
	}
}

// TestCapabilityMatrix_UpgradeProrateEntry verifies the capability matrix entry
// for CapSubscriptionUpgradeProration: both Razorpay and Cashfree support it.
func TestCapabilityMatrix_UpgradeProrateEntry(t *testing.T) {
	matrix := NewSupportMatrix()

	// Both providers should support upgrade proration
	if !matrix.Supports(domain.ProviderRazorpay, CapSubscriptionUpgradeProration) {
		t.Errorf("expected Razorpay to support CapSubscriptionUpgradeProration")
	}

	if !matrix.Supports(domain.ProviderCashfree, CapSubscriptionUpgradeProration) {
		t.Errorf("expected Cashfree to support CapSubscriptionUpgradeProration")
	}
}

// TestCapabilityMatrix_UnregisteredProvider returns false for unknown providers.
func TestCapabilityMatrix_UnregisteredProvider(t *testing.T) {
	matrix := NewSupportMatrix()

	// Unknown provider should return false
	result := matrix.Supports(domain.Provider("UNKNOWN_PROVIDER"), CapOrderCreate)
	if result {
		t.Errorf("expected false for unregistered provider, got true")
	}
}

// TestCapabilityMatrix_Describe returns optional descriptions for capabilities.
func TestCapabilityMatrix_Describe(t *testing.T) {
	matrix := NewSupportMatrix()

	desc := matrix.Describe(domain.ProviderCashfree, CapOrderCreate)
	if desc == "" {
		t.Errorf("expected non-empty description for Cashfree CapOrderCreate")
	}

	// Capability without a description should return empty string
	descEmpty := matrix.Describe(domain.ProviderRazorpay, CapOrderFetch)
	if descEmpty != "" {
		t.Errorf("expected empty description for capability without entry, got %s", descEmpty)
	}
}
