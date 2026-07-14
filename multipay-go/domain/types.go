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
	Name       string `json:"name,omitempty" pedantigo:"omitempty,maxLength=200"`
	Email      string `json:"email,omitempty" pedantigo:"omitempty,email"`
	Phone      string `json:"phone" pedantigo:"required,minLength=5,maxLength=20"`

	// --- Cashfree TPV (Third-Party Validation) bank fields (Cashfree-only; ignored by Razorpay) ---
	// BankAccountNumber → cf CustomerDetails.customer_bank_account_number / link customer_bank_account_number.
	BankAccountNumber string `json:"bank_account_number,omitempty" pedantigo:"omitempty,maxLength=50"`
	// BankIFSC → cf customer_bank_ifsc.
	BankIFSC string `json:"bank_ifsc,omitempty" pedantigo:"omitempty,maxLength=20"`
	// BankCode → cf customer_bank_code (net-banking code). Cashfree wire type is numeric.
	BankCode int32 `json:"bank_code,omitempty" pedantigo:"omitempty"`
	// BankAccountHolderName → cf link customer_bank_acoount_holder_name (note the vendor's
	// misspelling on the wire) and subscription customer_bank_account_holder_name. Cashfree-only;
	// ignored by Razorpay.
	BankAccountHolderName string `json:"bank_account_holder_name,omitempty" pedantigo:"omitempty,maxLength=250"`
	// UID → cf customer_uid (Cashfree customer identifier). Order-only.
	UID string `json:"uid,omitempty" pedantigo:"omitempty,maxLength=250"`
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

	// --- Cashfree-only optional fields (ignored by the Razorpay adapter) ---
	// CartDetails → cf CreateOrderRequest.cart_details.
	CartDetails *CartDetails `json:"cart_details,omitempty"`
	// Terminal → cf CreateOrderRequest.terminal (softPOS terminal binding).
	Terminal *TerminalDetails `json:"terminal,omitempty"`
	// OrderSplits → cf CreateOrderRequest.order_splits (Easy Split).
	OrderSplits []VendorSplit `json:"order_splits,omitempty"`
	// Products → cf CreateOrderRequest.products (One-Click-Checkout / Verify-and-Pay toggles).
	Products *OrderProducts `json:"products,omitempty"`
	// PaymentMethods → cf order_meta.payment_methods (comma-separated allowed modes, e.g. "cc,dc,upi").
	PaymentMethods string `json:"payment_methods,omitempty" pedantigo:"omitempty,maxLength=500"`
	// PaymentMethodsFilters → cf order_meta.payment_methods_filters (card bin/scheme/bank/suffix filters).
	PaymentMethodsFilters *OrderPaymentMethodsFilters `json:"payment_methods_filters,omitempty"`
	// OfferFilters → cf order_meta.offer_filters (allow/deny offer ids).
	OfferFilters *OfferFilters `json:"offer_filters,omitempty"`
	// UpiAppPriority → cf order_meta.upi_app_priority (ordered UPI-app hint list).
	UpiAppPriority []string `json:"upi_app_priority,omitempty"`

	// --- Razorpay-only optional fields (ignored by the Cashfree adapter) ---
	// PartialPayment enables customer partial payment on the order (rzp partial_payment).
	PartialPayment *bool `json:"partial_payment,omitempty"`
	// FirstPaymentMinAmount is the minimum first installment when partial payment is on.
	// Minor units (rzp first_payment_min_amount). Only meaningful with PartialPayment=true.
	FirstPaymentMinAmount AmountMinor `json:"first_payment_min_amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
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

	// RefundSpeed selects how fast the provider settles the refund. The two vendors
	// spell the values differently, so this carries the CANONICAL casing and each
	// adapter maps it to its wire value:
	//   Cashfree refund_speed: STANDARD | INSTANT
	//   Razorpay speed:        normal   | optimum
	// Empty means "provider default" (STANDARD / normal). The Cashfree adapter forwards
	// STANDARD/INSTANT verbatim; the Razorpay adapter maps STANDARD→normal, INSTANT→optimum.
	RefundSpeed RefundSpeed `json:"refund_speed,omitempty" pedantigo:"omitempty,oneof=STANDARD INSTANT"`

	// RefundSplits reverses an Easy-Split order proportionally across vendors.
	// Cashfree-only (cf refund_splits); ignored by the Razorpay adapter.
	RefundSplits []RefundSplit `json:"refund_splits,omitempty"`
}

// RefundSplit mirrors cf OrderCreateRefundRequestRefundSplitsInner. VendorID is
// mandatory on the vendor struct; Amount is minor units.
type RefundSplit struct {
	VendorID string            `json:"vendor_id" pedantigo:"required,maxLength=250"`
	Amount   AmountMinor       `json:"amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
	Tags     map[string]string `json:"tags,omitempty"`
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

// ListRefundsRequest asks a provider for every refund issued against ONE transaction.
// The two providers scope a refund list differently, so the request carries BOTH ids and
// each adapter uses only the one its provider requires. At least one of OrderID or
// PaymentID must be set (enforced by Validate()).
type ListRefundsRequest struct {
	// OrderID — CASHFREE only. Cashfree lists refunds under the ORDER:
	//   GET /pg/orders/{order_id}/refunds  (SDK: PGOrderFetchRefunds).
	// Set this for Cashfree. It is ignored by the Razorpay adapter.
	OrderID string `json:"order_id,omitempty" pedantigo:"omitempty,minLength=1"`
	// PaymentID — RAZORPAY only. Razorpay lists refunds under the captured PAYMENT:
	//   GET /v1/payments/{payment_id}/refunds  (SDK: Payment.FetchMultipleRefund).
	// Set this for Razorpay. It is ignored by the Cashfree adapter.
	PaymentID string `json:"payment_id,omitempty" pedantigo:"omitempty,minLength=1"`
}

// Validate enforces that at least one provider identifier is present. pedantigo's Validate()
// does not enforce field presence, so this cross-field rule is checked explicitly here.
func (r *ListRefundsRequest) Validate() error {
	if r.OrderID == "" && r.PaymentID == "" {
		return errors.New("at least one of order_id (Cashfree) or payment_id (Razorpay) is required")
	}
	return nil
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
