package capabilities

// This file re-exports Capability and constants from domain for convenience.
// The canonical definitions are in domain/enums.go.

import "github.com/Bytonomics/multipay-adapter/domain"

// Capability is re-exported from domain for convenience.
type Capability = domain.Capability

// Re-export core and provider-specific capability constants.
const (
	// Core Shared Capabilities
	CapOrderCreate       = domain.CapOrderCreate
	CapOrderFetch        = domain.CapOrderFetch
	CapPaymentFetch      = domain.CapPaymentFetch
	CapPaymentList       = domain.CapPaymentList
	CapPaymentPay        = domain.CapPaymentPay
	CapRefundCreate      = domain.CapRefundCreate
	CapRefundFetch       = domain.CapRefundFetch
	CapRefundList        = domain.CapRefundList
	CapInstrumentFetch   = domain.CapInstrumentFetch
	CapInstrumentList    = domain.CapInstrumentList
	CapInstrumentDelete  = domain.CapInstrumentDelete
	CapPaymentLinkCreate = domain.CapPaymentLinkCreate
	CapPaymentLinkFetch  = domain.CapPaymentLinkFetch
	CapPaymentLinkCancel = domain.CapPaymentLinkCancel
	// Cashfree-specific
	CapInstrumentCryptogram = domain.CapInstrumentCryptogram
	CapOfferCreate          = domain.CapOfferCreate
	CapOfferFetch           = domain.CapOfferFetch
	CapEligibilityFetch     = domain.CapEligibilityFetch
	// Razorpay-specific
	CapOrderUpdate                   = domain.CapOrderUpdate
	CapOrderList                     = domain.CapOrderList
	CapPaymentCapture                = domain.CapPaymentCapture
	CapRefundUpdate                  = domain.CapRefundUpdate
	CapCustomerCreate                = domain.CapCustomerCreate
	CapCustomerFetch                 = domain.CapCustomerFetch
	CapCustomerEdit                  = domain.CapCustomerEdit
	CapCustomerList                  = domain.CapCustomerList
	CapWebhookCreate                 = domain.CapWebhookCreate
	CapWebhookFetch                  = domain.CapWebhookFetch
	CapWebhookEdit                   = domain.CapWebhookEdit
	CapWebhookDelete                 = domain.CapWebhookDelete
	CapWebhookList                   = domain.CapWebhookList
	CapSubscriptionCreate            = domain.CapSubscriptionCreate
	CapSubscriptionFetch             = domain.CapSubscriptionFetch
	CapSubscriptionList              = domain.CapSubscriptionList
	CapPlanCreate                    = domain.CapPlanCreate
	CapPlanFetch          Capability = "plan_fetch"
	CapPlanList           Capability = "plan_list"
	CapPaymentLinkUpdate  Capability = "payment_link_update"
	CapPaymentLinkNotify  Capability = "payment_link_notify"
	CapPaymentLinkList    Capability = "payment_link_list"
	CapUPICreate          Capability = "upi_create"
	CapVPAValidate        Capability = "vpa_validate"
)
