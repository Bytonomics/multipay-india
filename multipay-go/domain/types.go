package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// RawProviderResponse preserves the original provider response for debugging.
type RawProviderResponse json.RawMessage

type CustomerInfo struct {
	CustomerID string `json:"customer_id" pedantigo:"required,minLength=1,maxLength=250"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty" pedantigo:"omitempty,email"`
	Phone      string `json:"phone" pedantigo:"required,minLength=5,maxLength=20"`
}

type Metadata map[string]string

// --- Order types ---

type CreateOrderRequest struct {
	OrderID     string        `json:"order_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	AmountMinor AmountMinor   `json:"amount_minor" pedantigo:"required,gt=0"`
	Currency    Currency      `json:"currency" pedantigo:"required,iso4217"`
	Customer    *CustomerInfo `json:"customer" pedantigo:"required"`
	ReturnURL   string        `json:"return_url" pedantigo:"required,url"`
	NotifyURL   string        `json:"notify_url,omitempty" pedantigo:"omitempty,url"`
	ExpiryTime  *time.Time    `json:"expiry_time,omitempty"`
	Note        string        `json:"note,omitempty" pedantigo:"omitempty,maxLength=500"`
	Metadata    Metadata      `json:"metadata,omitempty"`
}

// Validate enforces presence and non-empty constraints on CreateOrderRequest fields.
// Checks: Customer non-nil, Currency non-empty, ReturnURL non-empty.
func (r *CreateOrderRequest) Validate() error {
	if r.Customer == nil {
		return errors.New("customer is required")
	}
	if r.Currency == "" {
		return errors.New("currency is required")
	}
	if r.ReturnURL == "" {
		return errors.New("return_url is required")
	}
	return nil
}

type Order struct {
	ProviderOrderID string               `json:"provider_order_id"`
	OrderID         string               `json:"order_id"`
	Status          OrderStatus          `json:"status"`
	AmountMinor     AmountMinor          `json:"amount_minor"`
	Currency        Currency             `json:"currency"`
	SessionID       string               `json:"session_id,omitempty"`
	ExpiryTime      *time.Time           `json:"expiry_time,omitempty"`
	CreatedAt       *time.Time           `json:"created_at,omitempty"`
	Customer        *CustomerInfo        `json:"customer,omitempty"`
	Metadata        Metadata             `json:"metadata,omitempty"`
	ProviderDetails *OrderProviderDetail `json:"provider_details,omitempty"`
	Raw             RawProviderResponse  `json:"raw,omitempty"`
	Checkout        *CheckoutPayload     `json:"checkout,omitempty"`
}

// CheckoutPayload is the frontend-bound, provider-agnostic payload that lets a JS client
// drive the vendor's hosted checkout redirect. Discriminated by Provider; only the fields
// for that provider are populated.
type CheckoutPayload struct {
	Provider    Provider    `json:"provider"`
	Environment Environment `json:"environment"` // serialized UPPERCASE (SANDBOX/PRODUCTION)
	// Cashfree-only
	SessionID string `json:"session_id,omitempty"`
	// Razorpay-only
	OrderID     string      `json:"order_id,omitempty"`
	PublicKey   string      `json:"public_key,omitempty"`
	CallbackURL string      `json:"callback_url,omitempty"`
	AmountMinor AmountMinor `json:"amount_minor,omitempty"`
	Currency    Currency    `json:"currency,omitempty"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id" pedantigo:"required,minLength=1"`
}

type ListOrderPaymentsRequest struct {
	OrderID string `json:"order_id" pedantigo:"required,minLength=1"`
}

// --- Payment types ---

type Payment struct {
	ProviderPaymentID string                 `json:"provider_payment_id"`
	OrderID           string                 `json:"order_id,omitempty"`
	Status            PaymentStatus          `json:"status"`
	AmountMinor       AmountMinor            `json:"amount_minor"`
	Currency          Currency               `json:"currency,omitempty"`
	PaymentGroup      string                 `json:"payment_group,omitempty"`
	PaymentMethod     string                 `json:"payment_method,omitempty"`
	PaymentTime       *time.Time             `json:"payment_time,omitempty"`
	CompletionTime    *time.Time             `json:"completion_time,omitempty"`
	IsCaptured        bool                   `json:"is_captured"`
	BankReference     string                 `json:"bank_reference,omitempty"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	ProviderDetails   *PaymentProviderDetail `json:"provider_details,omitempty"`
	Raw               RawProviderResponse    `json:"raw,omitempty"`
}

type GetPaymentRequest struct {
	// OrderID is optional in the canonical contract, but required by Cashfree (enforced in adapter).
	// Razorpay fetches payments by PaymentID alone.
	OrderID   string `json:"order_id,omitempty" pedantigo:"omitempty,minLength=1"`
	PaymentID string `json:"payment_id" pedantigo:"required,minLength=1"`
}

type ListPaymentsRequest struct {
	OrderID string `json:"order_id" pedantigo:"required,minLength=1"`
}

type CapturePaymentRequest struct {
	PaymentID   string      `json:"payment_id" pedantigo:"required,minLength=1"`
	AmountMinor AmountMinor `json:"amount_minor" pedantigo:"required,gt=0"`
	Currency    Currency    `json:"currency" pedantigo:"required,iso4217"`
}

// --- Refund types ---

type CreateRefundRequest struct {
	OrderID     string      `json:"order_id,omitempty" pedantigo:"omitempty,minLength=1"`
	PaymentID   string      `json:"payment_id,omitempty" pedantigo:"omitempty,minLength=1"`
	RefundID    string      `json:"refund_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	AmountMinor AmountMinor `json:"amount_minor" pedantigo:"required,gt=0"`
	Currency    Currency    `json:"currency" pedantigo:"required,iso4217"`
	Reason      string      `json:"reason,omitempty" pedantigo:"omitempty,maxLength=500"`
	Metadata    Metadata    `json:"metadata,omitempty"`
}

// Validate enforces the cross-field rule: at least one of OrderID or PaymentID must be provided.
// Provider-specific refund identification is enforced in adapters: Cashfree uses OrderID, Razorpay uses PaymentID.
func (r *CreateRefundRequest) Validate() error {
	if r.OrderID == "" && r.PaymentID == "" {
		return errors.New("at least one of order_id or payment_id is required")
	}
	return nil
}

type Refund struct {
	ProviderRefundID string                `json:"provider_refund_id"`
	RefundID         string                `json:"refund_id,omitempty"`
	OrderID          string                `json:"order_id,omitempty"`
	PaymentID        string                `json:"payment_id,omitempty"`
	Status           RefundStatus          `json:"status"`
	AmountMinor      AmountMinor           `json:"amount_minor"`
	Currency         Currency              `json:"currency,omitempty"`
	Reason           string                `json:"reason,omitempty"`
	ARN              string                `json:"arn,omitempty"`
	CreatedAt        *time.Time            `json:"created_at,omitempty"`
	ProcessedAt      *time.Time            `json:"processed_at,omitempty"`
	ProviderDetails  *RefundProviderDetail `json:"provider_details,omitempty"`
	Raw              RawProviderResponse   `json:"raw,omitempty"`
}

type GetRefundRequest struct {
	OrderID  string `json:"order_id,omitempty" pedantigo:"omitempty,minLength=1"`
	RefundID string `json:"refund_id" pedantigo:"required,minLength=1"`
}

type ListRefundsRequest struct {
	OrderID string `json:"order_id" pedantigo:"required,minLength=1"`
}

// --- Instrument types ---

type Instrument struct {
	CustomerID      string                    `json:"customer_id"`
	InstrumentID    string                    `json:"instrument_id"`
	InstrumentType  string                    `json:"instrument_type,omitempty"`
	DisplayValue    string                    `json:"display_value,omitempty"`
	Status          string                    `json:"status,omitempty"`
	CreatedAt       *time.Time                `json:"created_at,omitempty"`
	ProviderDetails *InstrumentProviderDetail `json:"provider_details,omitempty"`
	Raw             RawProviderResponse       `json:"raw,omitempty"`
}

type GetInstrumentRequest struct {
	CustomerID   string `json:"customer_id" pedantigo:"required,minLength=1"`
	InstrumentID string `json:"instrument_id" pedantigo:"required,minLength=1"`
}

type ListInstrumentsRequest struct {
	CustomerID string `json:"customer_id" pedantigo:"required,minLength=1"`
}

type DeleteInstrumentRequest struct {
	CustomerID   string `json:"customer_id" pedantigo:"required,minLength=1"`
	InstrumentID string `json:"instrument_id" pedantigo:"required,minLength=1"`
}

// --- Webhook types ---

type WebhookEvent struct {
	Provider     Provider         `json:"provider"`
	AccountID    string           `json:"account_id,omitempty"`
	EventType    WebhookEventType `json:"event_type"`
	EventTime    *time.Time       `json:"event_time,omitempty"`
	Order        *Order           `json:"order,omitempty"`
	Payment      *Payment         `json:"payment,omitempty"`
	Refund       *Refund          `json:"refund,omitempty"`
	Subscription *Subscription    `json:"subscription,omitempty"`
	RawPayload   []byte           `json:"raw_payload,omitempty"`
	DedupeKey    string           `json:"dedupe_key"`

	// D11: raw-passthrough — never hide vendor data from callers
	WebhookURL         string            `json:"webhook_url,omitempty"`
	RawVendorEventType string            `json:"raw_vendor_event_type,omitempty"`
	RawVendorStatus    string            `json:"raw_vendor_status,omitempty"`
	RawHeaders         map[string]string `json:"raw_headers,omitempty"`

	// D12: graceful degradation — parser errors do not abort dispatch
	ParseError string `json:"parse_error,omitempty"`
}

type WebhookEventHandler func(ctx context.Context, event *WebhookEvent) error
