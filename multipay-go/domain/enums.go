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

// AmountMinor represents monetary amounts in minor units (paisa/cents/fils).
// The conversion factor depends on the ISO 4217 currency exponent:
//   - Exponent 0 (JPY, KRW, VND): 1 minor = 1 major
//   - Exponent 2 (INR, USD, EUR): 100 minor = 1 major
//   - Exponent 3 (BHD, KWD, OMR): 1000 minor = 1 major
type AmountMinor int64

// Currency represents an ISO 4217 three-letter currency code.
// All ISO 4217 codes are valid (e.g., "INR", "USD", "EUR", "JPY", "BHD", "KWD").
// The library uses github.com/bojanz/currency for minor unit lookups.
// Do NOT restrict to a hardcoded set — both Cashfree (140+) and Razorpay (130+)
// support a wide range of international currencies.
type Currency string

// String returns the string representation.
func (c Currency) String() string {
	return string(c)
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

	// Subscription webhook events
	EventSubAuthenticated WebhookEventType = "subscription.authenticated"
	EventSubActivated     WebhookEventType = "subscription.activated"
	EventSubCharged       WebhookEventType = "subscription.charged"
	EventSubPaymentFailed WebhookEventType = "subscription.payment_failed"
	EventSubHalted        WebhookEventType = "subscription.halted"
	EventSubOnHold        WebhookEventType = "subscription.on_hold"
	EventSubPaused        WebhookEventType = "subscription.paused"
	EventSubResumed       WebhookEventType = "subscription.resumed"
	EventSubCancelled     WebhookEventType = "subscription.cancelled"
	EventSubCompleted     WebhookEventType = "subscription.completed"
	EventSubUpdated       WebhookEventType = "subscription.updated"
	EventSubCardExpiring  WebhookEventType = "subscription.card_expiring"
	EventSubRefund        WebhookEventType = "subscription.refund"

	// EventUnknown is emitted when the vendor event cannot be classified. Routes to DefaultHandler.
	EventUnknown WebhookEventType = "unknown"

	// EventOrderExpired: Cashfree ORDER.EXPIRED
	EventOrderExpired WebhookEventType = "order.expired"

	// EventRefundFailed: Cashfree REFUND.FAILED; Razorpay refund.failed
	EventRefundFailed WebhookEventType = "refund.failed"

	// EventSubCardExpired: Cashfree SUBSCRIPTION_STATUS_CHANGED status=CARD_EXPIRED
	EventSubCardExpired WebhookEventType = "subscription.card_expired"

	// EventSubExpired: Cashfree SUBSCRIPTION_STATUS_CHANGED status=EXPIRED or LINK_EXPIRED (merged)
	EventSubExpired WebhookEventType = "subscription.expired"

	// EventSubBankApprovalPending: Cashfree SUBSCRIPTION_STATUS_CHANGED status=BANK_APPROVAL_PENDING
	EventSubBankApprovalPending WebhookEventType = "subscription.bank_approval_pending"

	// EventSubPreDebitNotice: Cashfree SUBSCRIPTION_PAYMENT_NOTIFICATION_INITIATED
	EventSubPreDebitNotice WebhookEventType = "subscription.pre_debit_notice"

	// EventSubPaymentCancelled: Cashfree SUBSCRIPTION_PAYMENT_CANCELLED
	EventSubPaymentCancelled WebhookEventType = "subscription.payment_cancelled"
)

// String returns the string representation of the WebhookEventType.
func (w WebhookEventType) String() string {
	return string(w)
}

// IsValid checks if the WebhookEventType value is valid.
func (w WebhookEventType) IsValid() bool {
	switch w {
	case EventOrderCreated, EventPaymentAuthorized, EventPaymentCaptured,
		EventPaymentFailed, EventRefundCreated, EventRefundProcessed,
		EventSubAuthenticated, EventSubActivated, EventSubCharged, EventSubPaymentFailed,
		EventSubHalted, EventSubOnHold, EventSubPaused,
		EventSubResumed, EventSubCancelled, EventSubCompleted, EventSubUpdated,
		EventSubCardExpiring, EventSubRefund,
		EventUnknown, EventOrderExpired, EventRefundFailed,
		EventSubCardExpired, EventSubExpired, EventSubBankApprovalPending,
		EventSubPreDebitNotice, EventSubPaymentCancelled:
		return true
	}
	return false
}

// PaymentLinkStatus represents the payment link lifecycle.
type PaymentLinkStatus string

const (
	PaymentLinkStatusActive        PaymentLinkStatus = "active"
	PaymentLinkStatusPaid          PaymentLinkStatus = "paid"
	PaymentLinkStatusExpired       PaymentLinkStatus = "expired"
	PaymentLinkStatusCancelled     PaymentLinkStatus = "cancelled"
	PaymentLinkStatusPartiallyPaid PaymentLinkStatus = "partially_paid"
)

// String returns the string representation of the PaymentLinkStatus.
func (p PaymentLinkStatus) String() string {
	return string(p)
}

// IsValid checks if the PaymentLinkStatus value is valid.
func (p PaymentLinkStatus) IsValid() bool {
	switch p {
	case PaymentLinkStatusActive, PaymentLinkStatusPaid, PaymentLinkStatusExpired, PaymentLinkStatusCancelled, PaymentLinkStatusPartiallyPaid:
		return true
	}
	return false
}
