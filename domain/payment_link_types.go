package domain

import "time"

// CreatePaymentLinkRequest represents a request to create a new payment link.
// Payment links allow customers to make payments through a shareable URL.
type CreatePaymentLinkRequest struct {
	// Amount is the payment amount in paisa/cents (minor units).
	// Must be greater than 0.
	Amount int64 `validate:"required,min=1"`

	// Currency is the ISO 4217 currency code (e.g., "INR", "USD").
	Currency string `validate:"required"`

	// Description is an optional description for the payment link.
	Description string `validate:"omitempty"`

	// Notes is optional provider-specific metadata stored as key-value pairs.
	Notes map[string]string `validate:"omitempty"`

	// NotifyEmail is an optional email address to send payment notifications.
	NotifyEmail string `validate:"omitempty,email"`

	// NotifyPhone is an optional phone number to send payment notifications.
	NotifyPhone string `validate:"omitempty"`

	// ExpiresAt is an optional expiration time for the payment link.
	ExpiresAt *time.Time `validate:"omitempty"`
}

// PaymentLinkResponse represents a payment link returned from the payment gateway.
// It contains the complete state of a created or retrieved payment link.
type PaymentLinkResponse struct {
	// ID is the unique identifier for this payment link, assigned by the gateway.
	ID string

	// URL is the shareable payment link that customers can use to make payments.
	URL string

	// Amount is the payment amount in paisa/cents (minor units).
	Amount int64

	// Currency is the ISO 4217 currency code.
	Currency string

	// Description is the payment link description.
	Description string

	// Status is the current status of the payment link (e.g., "created", "paid", "cancelled").
	Status string

	// NotifyEmail is the email address for payment notifications, if configured.
	NotifyEmail string

	// NotifyPhone is the phone number for payment notifications, if configured.
	NotifyPhone string

	// ExpiresAt is the expiration time of the payment link, if set.
	ExpiresAt *time.Time

	// CreatedAt is the timestamp when the payment link was created.
	CreatedAt time.Time

	// UpdatedAt is the timestamp when the payment link was last updated.
	UpdatedAt time.Time
}

// GetPaymentLinkRequest represents a request to retrieve an existing payment link.
type GetPaymentLinkRequest struct {
	// LinkID is the unique identifier of the payment link to retrieve.
	// Must be non-empty.
	LinkID string `validate:"required"`
}

// CancelPaymentLinkRequest represents a request to cancel an existing payment link.
type CancelPaymentLinkRequest struct {
	// LinkID is the unique identifier of the payment link to cancel.
	// Must be non-empty.
	LinkID string `validate:"required"`
}

// UpdatePaymentLinkRequest represents a request to update an existing payment link.
// Currently used by Razorpay for updating link attributes.
type UpdatePaymentLinkRequest struct {
	// LinkID is the unique identifier of the payment link to update.
	// Must be non-empty.
	LinkID string `validate:"required"`

	// Description is the updated description for the payment link (optional).
	Description string `validate:"omitempty"`

	// Notes is the updated provider-specific metadata (optional).
	Notes map[string]string `validate:"omitempty"`
}

// PaymentLinkEvent represents an event emitted by the payment gateway related to a payment link.
// Events track state transitions in the payment link lifecycle.
type PaymentLinkEvent struct {
	// ID is the unique identifier for this event.
	ID string

	// EventType is the type of event (e.g., "created", "paid", "cancelled", "expired").
	EventType string

	// Amount is the payment amount in paisa/cents (minor units).
	Amount int64

	// Status is the current status of the payment link after this event.
	Status string

	// URL is the shareable payment link URL.
	URL string

	// CreatedAt is the timestamp when the event occurred.
	CreatedAt time.Time
}
