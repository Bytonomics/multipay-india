package capabilities

import (
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// SupportMatrix maintains a static registry of which capabilities are supported by each provider.
// It is immutable after NewSupportMatrix() and can be safely shared across goroutines.
type SupportMatrix struct {
	// Matrix maps provider -> (capability -> supported).
	// Each inner map is a snapshot at construction time and never modified.
	matrix map[domain.Provider]map[Capability]bool

	// descriptions maps (provider, capability) -> optional description.
	descriptions map[string]string
}

// NewSupportMatrix constructs a SupportMatrix with capability support data from
// verified SDK coverage (doc 144-vendor-sdk-use-case-comparison.md).
// Both providers are pre-populated with their exact verified capabilities.
func NewSupportMatrix() *SupportMatrix {
	// Cashfree verified capabilities from doc 144 section 2.
	cashfreeSupport := map[Capability]bool{
		// Core shared capabilities
		CapOrderCreate:       true,
		CapOrderFetch:        true,
		CapOrderListPayments: true,
		CapPaymentFetch:      true,
		CapPaymentList:       true,
		CapPaymentPay:        true,
		CapRefundCreate:      true,
		CapRefundFetch:       true,
		CapRefundList:        true,
		CapInstrumentFetch:   true,
		CapInstrumentList:    true,
		CapInstrumentDelete:  true,
		CapPaymentLinkCreate: true,
		CapPaymentLinkFetch:  true,
		CapPaymentLinkCancel: true,
		CapWebhookConsume:    true,

		// Cashfree-specific capabilities
		CapInstrumentCryptogram:     true,
		CapOfferCreate:              true,
		CapOfferFetch:               true,
		CapEligibilityFetch:         true,
		CapPaymentLinkListOrders:    true,
		CapSettlementOrderFetch:     true,
		CapSettlementList:           true,
		CapSettlementReconFetch:     true,
		CapReconFetch:               true,
		CapSubscriptionManualCharge: true,
		// PARTLY VERIFIED (vendor documents checked 2026-10-19). This capability is a library composite:
		// the library computes the prorated delta and raises a manual charge. Cashfree documents raising a
		// charge with POST /pg/subscriptions/pay and payment_type CHARGE for on-demand subscriptions
		// (https://www.cashfree.com/docs/payments/subscription/charge). Cashfree documents that CHANGE_PLAN
		// and PAUSE are not supported for on-demand subscriptions
		// (https://www.cashfree.com/docs/api-reference/payments/latest/subscription/mandate/manage).
		// The documents do not say whether a PERIODIC subscription accepts an extra ad-hoc charge.
		// An agent or developer MUST verify this with Cashfree documents or support before extending this
		// capability or relying on it for periodic subscriptions.
		CapSubscriptionUpgradeProration: true,
		CapSubscriptionEligibility:      true,
		// false means NOT DOCUMENTED, not proven unsupported (checked 2026-10-19). The Manage Subscription
		// CANCEL action lists no cancel-at-cycle-end or effective-date parameter. Re-check the vendor
		// documents before you change this value.
		CapSubscriptionCancelAtCycleEnd: false,
		CapSubscriptionList:             false,
		CapPlanList:                     false,

		// Razorpay-only capabilities (not supported by Cashfree)
		CapOrderUpdate:       false,
		CapOrderList:         false,
		CapPaymentCapture:    false,
		CapRefundUpdate:      false,
		CapCustomerCreate:    false,
		CapCustomerFetch:     false,
		CapCustomerEdit:      false,
		CapCustomerList:      false,
		CapWebhookCreate:     false,
		CapWebhookFetch:      false,
		CapWebhookEdit:       false,
		CapWebhookDelete:     false,
		CapWebhookList:       false,
		CapPaymentLinkUpdate: false,
		CapPaymentLinkNotify: false,
		CapPaymentLinkList:   false,
		CapUPICreate:         false,
		CapVPAValidate:       false,

		// Razorpay-specific settlement capabilities (not supported by Cashfree)
		CapSettlementAll:            false,
		CapSettlementFetch:          false,
		CapSettlementReports:        false,
		CapSettlementOnDemandCreate: false,
		CapSettlementOnDemandFetch:  false,
		CapSettlementOnDemandList:   false,
	}

	// Razorpay verified capabilities from doc 144 section 2.
	razorpaySupport := map[Capability]bool{
		// Core shared capabilities
		CapOrderCreate:       true,
		CapOrderFetch:        true,
		CapOrderListPayments: true,
		CapPaymentFetch:      true,
		CapPaymentList:       true,
		CapPaymentPay:        true,
		CapRefundCreate:      true,
		CapRefundFetch:       true,
		CapRefundList:        true,
		CapInstrumentFetch:   true,
		CapInstrumentList:    true,
		CapInstrumentDelete:  true,
		CapPaymentLinkCreate: true,
		CapPaymentLinkFetch:  true,
		CapPaymentLinkCancel: true,
		CapWebhookConsume:    true,

		// Razorpay-specific capabilities
		CapOrderUpdate:       true,
		CapOrderList:         true,
		CapPaymentCapture:    true,
		CapRefundUpdate:      true,
		CapCustomerCreate:    true,
		CapCustomerFetch:     true,
		CapCustomerEdit:      true,
		CapCustomerList:      true,
		CapWebhookCreate:     true,
		CapWebhookFetch:      true,
		CapWebhookEdit:       true,
		CapWebhookDelete:     true,
		CapWebhookList:       true,
		CapSubscriptionList:  true,
		CapPlanList:          true,
		CapPaymentLinkUpdate: true,
		CapPaymentLinkNotify: true,
		CapPaymentLinkList:   true,
		CapUPICreate:         true,
		CapVPAValidate:       true,

		// Cashfree-specific capabilities (supported by Razorpay)
		CapSubscriptionManualCharge: true,
		// PARTLY VERIFIED (vendor documents checked 2026-10-19). Razorpay documents its own proration for
		// an immediate plan change (Update a Subscription, schedule_change_at "now"; minimum difference 50
		// subunits): https://razorpay.com/docs/api/payments/subscriptions/update-subscription
		// The charge flow that the library uses for Razorpay was NOT compared with the vendor documents.
		// Verify it before you change the Razorpay upgrade flow.
		CapSubscriptionUpgradeProration: true,
		CapSubscriptionEligibility:      false,
		// VERIFIED (vendor documents checked 2026-10-19): Cancel Subscription accepts the boolean body
		// parameter cancel_at_cycle_end (true cancels at the end of the billing cycle):
		// https://razorpay.com/docs/api/payments/subscriptions/cancel-subscription/
		// The adapter forwards it (providers/razorpay/subscriptions.go).
		CapSubscriptionCancelAtCycleEnd: true,

		// Cashfree-only capabilities (not supported by Razorpay)
		CapInstrumentCryptogram:  false,
		CapOfferCreate:           false,
		CapOfferFetch:            false,
		CapEligibilityFetch:      false,
		CapPaymentLinkListOrders: false,

		// Razorpay-specific settlement capabilities
		CapSettlementAll:            true,
		CapSettlementFetch:          true,
		CapSettlementReports:        true,
		CapSettlementOnDemandCreate: true,
		CapSettlementOnDemandFetch:  true,
		CapSettlementOnDemandList:   true,

		// Cashfree-specific (not supported by Razorpay)
		CapSettlementOrderFetch: false,
		CapSettlementList:       false,
		CapSettlementReconFetch: false,
		CapReconFetch:           false,
	}

	m := &SupportMatrix{
		matrix:       make(map[domain.Provider]map[Capability]bool),
		descriptions: make(map[string]string),
	}

	m.matrix[domain.ProviderCashfree] = cashfreeSupport
	m.matrix[domain.ProviderRazorpay] = razorpaySupport

	// Populate optional descriptions for common capabilities.
	m.descriptions[descKey(domain.ProviderCashfree, CapOrderCreate)] = "Create order via PGCreateOrder"
	m.descriptions[descKey(domain.ProviderCashfree, CapPaymentPay)] = "Pay order via PGPayOrder"
	m.descriptions[descKey(domain.ProviderCashfree, CapInstrumentCryptogram)] = "Fetch instrument cryptogram via PGCustomerInstrumentsFetchCryptogram"
	m.descriptions[descKey(domain.ProviderCashfree, CapSubscriptionManualCharge)] = "Trigger manual charge for ON_DEMAND subscription"
	m.descriptions[descKey(domain.ProviderCashfree, CapSubscriptionEligibility)] = "Check subscription eligibility"

	m.descriptions[descKey(domain.ProviderRazorpay, CapOrderCreate)] = "Create order via resources.Order.Create"
	m.descriptions[descKey(domain.ProviderRazorpay, CapOrderUpdate)] = "Update order via resources.Order.Update"
	m.descriptions[descKey(domain.ProviderRazorpay, CapPaymentCapture)] = "Capture payment via resources.Payment.Capture"

	return m
}

// Supports returns true if the provider supports the given capability.
// Returns false if the provider is not registered or does not support the capability.
func (m *SupportMatrix) Supports(provider domain.Provider, cap Capability) bool {
	if m == nil || m.matrix == nil {
		return false
	}
	providerMap, ok := m.matrix[provider]
	if !ok {
		return false
	}
	return providerMap[cap]
}

// Describe returns an optional description of the capability for the provider.
// Returns an empty string if no description is available.
func (m *SupportMatrix) Describe(provider domain.Provider, cap Capability) string {
	if m == nil || m.descriptions == nil {
		return ""
	}
	return m.descriptions[descKey(provider, cap)]
}

// descKey returns a composite key for (provider, capability) -> description lookup.
func descKey(provider domain.Provider, cap Capability) string {
	return string(provider) + ":" + string(cap)
}
