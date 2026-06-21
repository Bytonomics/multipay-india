package domain

// Provider identifies the payment gateway provider.
type Provider string

const (
	ProviderCashfree Provider = "cashfree"
	ProviderRazorpay Provider = "razorpay"
)

// String returns the string representation of the Provider.
func (p Provider) String() string {
	return string(p)
}

// IsValid checks if the Provider value is valid.
func (p Provider) IsValid() bool {
	switch p {
	case ProviderCashfree, ProviderRazorpay:
		return true
	}
	return false
}

// Environment identifies the Cashfree-specific deployment environment.
type Environment string

const (
	EnvironmentSandbox    Environment = "SANDBOX"
	EnvironmentProduction Environment = "PRODUCTION"
)

// String returns the string representation of the Environment.
func (e Environment) String() string {
	return string(e)
}

// IsValid checks if the Environment value is valid.
func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentSandbox, EnvironmentProduction:
		return true
	}
	return false
}

// OrderStatus represents the canonical order lifecycle state.
type OrderStatus string

const (
	OrderCreated   OrderStatus = "created"
	OrderPaid      OrderStatus = "paid"
	OrderExpired   OrderStatus = "expired"
	OrderCancelled OrderStatus = "cancelled"
)

// String returns the string representation of the OrderStatus.
func (o OrderStatus) String() string {
	return string(o)
}

// IsValid checks if the OrderStatus value is valid.
func (o OrderStatus) IsValid() bool {
	switch o {
	case OrderCreated, OrderPaid, OrderExpired, OrderCancelled:
		return true
	}
	return false
}

// PaymentStatus represents the payment state machine.
type PaymentStatus string

const (
	PaymentAuthorized PaymentStatus = "authorized"
	PaymentCaptured   PaymentStatus = "captured"
	PaymentFailed     PaymentStatus = "failed"
	PaymentRefunded   PaymentStatus = "refunded"
)

// String returns the string representation of the PaymentStatus.
func (p PaymentStatus) String() string {
	return string(p)
}

// IsValid checks if the PaymentStatus value is valid.
func (p PaymentStatus) IsValid() bool {
	switch p {
	case PaymentAuthorized, PaymentCaptured, PaymentFailed, PaymentRefunded:
		return true
	}
	return false
}

// RefundStatus represents the refund lifecycle state.
type RefundStatus string

const (
	RefundCreated   RefundStatus = "created"
	RefundProcessed RefundStatus = "processed"
	RefundFailed    RefundStatus = "failed"
	RefundPartial   RefundStatus = "partial"
)

// String returns the string representation of the RefundStatus.
func (r RefundStatus) String() string {
	return string(r)
}

// IsValid checks if the RefundStatus value is valid.
func (r RefundStatus) IsValid() bool {
	switch r {
	case RefundCreated, RefundProcessed, RefundFailed, RefundPartial:
		return true
	}
	return false
}

// WebhookEventType represents events emitted by payment SDKs.
type WebhookEventType string

const (
	EventOrderCreated      WebhookEventType = "order.created"
	EventPaymentAuthorized WebhookEventType = "payment.authorized"
	EventPaymentCaptured   WebhookEventType = "payment.captured"
	EventPaymentFailed     WebhookEventType = "payment.failed"
	EventRefundCreated     WebhookEventType = "refund.created"
	EventRefundProcessed   WebhookEventType = "refund.processed"
)

// String returns the string representation of the WebhookEventType.
func (w WebhookEventType) String() string {
	return string(w)
}

// IsValid checks if the WebhookEventType value is valid.
func (w WebhookEventType) IsValid() bool {
	switch w {
	case EventOrderCreated, EventPaymentAuthorized, EventPaymentCaptured,
		EventPaymentFailed, EventRefundCreated, EventRefundProcessed:
		return true
	}
	return false
}

// Capability represents a feature that may or may not be supported by a payment provider.
type Capability string

// Core Shared Capabilities (supported by both Cashfree and Razorpay)
const (
	CapOrderCreate       Capability = "order_create"
	CapOrderFetch        Capability = "order_fetch"
	CapPaymentFetch      Capability = "payment_fetch"
	CapPaymentList       Capability = "payment_list"
	CapPaymentPay        Capability = "payment_pay"
	CapRefundCreate      Capability = "refund_create"
	CapRefundFetch       Capability = "refund_fetch"
	CapRefundList        Capability = "refund_list"
	CapInstrumentFetch   Capability = "instrument_fetch"
	CapInstrumentList    Capability = "instrument_list"
	CapInstrumentDelete  Capability = "instrument_delete"
	CapPaymentLinkCreate Capability = "payment_link_create"
	CapPaymentLinkFetch  Capability = "payment_link_fetch"
	CapPaymentLinkCancel Capability = "payment_link_cancel"
)

// Cashfree-specific Capabilities
const (
	CapInstrumentCryptogram Capability = "instrument_cryptogram"
	CapOfferCreate          Capability = "offer_create"
	CapOfferFetch           Capability = "offer_fetch"
	CapEligibilityFetch     Capability = "eligibility_fetch"
)

// Razorpay-specific Capabilities
const (
	CapOrderUpdate        Capability = "order_update"
	CapOrderList          Capability = "order_list"
	CapPaymentCapture     Capability = "payment_capture"
	CapRefundUpdate       Capability = "refund_update"
	CapCustomerCreate     Capability = "customer_create"
	CapCustomerFetch      Capability = "customer_fetch"
	CapCustomerEdit       Capability = "customer_edit"
	CapCustomerList       Capability = "customer_list"
	CapWebhookCreate      Capability = "webhook_create"
	CapWebhookFetch       Capability = "webhook_fetch"
	CapWebhookEdit        Capability = "webhook_edit"
	CapWebhookDelete      Capability = "webhook_delete"
	CapWebhookList        Capability = "webhook_list"
	CapSubscriptionCreate Capability = "subscription_create"
	CapSubscriptionFetch  Capability = "subscription_fetch"
	CapSubscriptionList   Capability = "subscription_list"
	CapPlanCreate         Capability = "plan_create"
	CapPlanFetch          Capability = "plan_fetch"
	CapPlanList           Capability = "plan_list"
	CapPaymentLinkUpdate  Capability = "payment_link_update"
	CapPaymentLinkNotify  Capability = "payment_link_notify"
	CapPaymentLinkList    Capability = "payment_link_list"
	CapUPICreate          Capability = "upi_create"
	CapVPAValidate        Capability = "vpa_validate"
)
