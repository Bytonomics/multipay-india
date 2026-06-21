package domain

import "time"

// Currency represents ISO 4217 three-letter currency codes.
// Any valid ISO 4217 code can be used; the library does not restrict to a predefined set.
// Examples: "INR", "USD", "EUR", "GBP", "AED", "AUD", "CAD", "CHF", "CNY", "JPY", etc.
type Currency string

// CustomerInfo represents customer information embedded in requests.
type CustomerInfo struct {
	// CustomerID is a unique identifier for the customer.
	CustomerID string `json:"customer_id" db:"customer_id"`

	// Email is the customer's email address.
	Email string `json:"email" db:"email"`

	// Phone is the customer's phone number.
	Phone string `json:"phone" db:"phone"`
}

// Metadata represents provider-specific metadata as a string map.
type Metadata map[string]string

// CreateOrderRequest represents a request to create an order.
type CreateOrderRequest struct {
	// Amount is the order amount in paisa/cents (int64).
	Amount int64 `json:"amount" validate:"required,gt=0"`

	// Currency is the currency code for the order.
	Currency Currency `json:"currency" validate:"required"`

	// Receipt is a reference number for internal tracking.
	Receipt string `json:"receipt"`

	// Notes are optional notes for the order.
	Notes string `json:"notes"`

	// Customer contains customer information for the order.
	Customer CustomerInfo `json:"customer"`

	// Metadata contains provider-specific metadata.
	Metadata Metadata `json:"metadata,omitempty"`
}

// Order represents a created order.
type Order struct {
	// ID is the unique order identifier.
	ID string `json:"id" db:"id"`

	// Amount is the order amount in paisa/cents (int64).
	Amount int64 `json:"amount" db:"amount"`

	// Currency is the currency code.
	Currency Currency `json:"currency" db:"currency"`

	// Receipt is the reference number.
	Receipt string `json:"receipt" db:"receipt"`

	// Status is the current order status.
	Status OrderStatus `json:"status" db:"status"`

	// Notes are associated notes.
	Notes string `json:"notes" db:"notes"`

	// CreatedAt is when the order was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetOrderRequest represents a request to retrieve an order.
type GetOrderRequest struct {
	// OrderID is the unique order identifier.
	OrderID string `json:"order_id" validate:"required"`
}

// GetPaymentRequest represents a request to retrieve a payment.
type GetPaymentRequest struct {
	// OrderID is the order identifier.
	OrderID string `json:"order_id" validate:"required"`

	// PaymentID is the payment identifier.
	PaymentID string `json:"payment_id" validate:"required"`
}

// CapturePaymentRequest represents a request to capture an authorized payment.
type CapturePaymentRequest struct {
	// PaymentID is the payment identifier.
	PaymentID string `json:"payment_id" validate:"required"`

	// Amount is the capture amount in paisa/cents (int64).
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

// Payment represents a payment for an order.
type Payment struct {
	// ID is the unique payment identifier.
	ID string `json:"id" db:"id"`

	// OrderID is the associated order identifier.
	OrderID string `json:"order_id" db:"order_id"`

	// Amount is the payment amount in paisa/cents (int64).
	Amount int64 `json:"amount" db:"amount"`

	// Status is the current payment status.
	Status PaymentStatus `json:"status" db:"status"`

	// Method is the payment method used (e.g., "card", "upi", "netbanking").
	Method string `json:"method" db:"method"`

	// CreatedAt is when the payment was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateRefundRequest represents a request to create a refund.
type CreateRefundRequest struct {
	// OrderID is the order identifier for the refund.
	OrderID string `json:"order_id" validate:"required"`

	// Amount is the refund amount in paisa/cents (int64). Optional; if omitted, full refund is assumed.
	Amount int64 `json:"amount" validate:"omitempty,gt=0"`

	// Notes are optional notes for the refund.
	Notes string `json:"notes"`
}

// Refund represents a refund for an order.
type Refund struct {
	// ID is the unique refund identifier.
	ID string `json:"id" db:"id"`

	// OrderID is the associated order identifier.
	OrderID string `json:"order_id" db:"order_id"`

	// Amount is the refund amount in paisa/cents (int64).
	Amount int64 `json:"amount" db:"amount"`

	// Status is the current refund status.
	Status RefundStatus `json:"status" db:"status"`

	// CreatedAt is when the refund was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetRefundRequest represents a request to retrieve a refund.
type GetRefundRequest struct {
	// OrderID is the order identifier.
	OrderID string `json:"order_id" validate:"required"`

	// RefundID is the refund identifier.
	RefundID string `json:"refund_id" validate:"required"`
}

// GetInstrumentRequest represents a request to retrieve an instrument.
type GetInstrumentRequest struct {
	// CustomerID is the customer identifier.
	CustomerID string `json:"customer_id" validate:"required"`

	// InstrumentID is the instrument identifier.
	InstrumentID string `json:"instrument_id" validate:"required"`
}

// Instrument represents a payment instrument (e.g., card, token).
type Instrument struct {
	// ID is the unique instrument identifier.
	ID string `json:"id" db:"id"`

	// CustomerID is the associated customer identifier.
	CustomerID string `json:"customer_id" db:"customer_id"`

	// Type is the instrument type (e.g., "card", "upi", "netbanking").
	Type string `json:"type" db:"type"`

	// CreatedAt is when the instrument was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WebhookEvent represents a webhook event from a payment provider.
type WebhookEvent struct {
	// ID is the unique event identifier.
	ID string `json:"id"`

	// EventType is the type of webhook event.
	EventType WebhookEventType `json:"event_type"`

	// Provider is the payment provider that sent the event.
	Provider string `json:"provider"`

	// Data contains provider-specific event data.
	Data map[string]interface{} `json:"data"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}
