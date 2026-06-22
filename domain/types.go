package domain

import (
	"context"
	"encoding/json"
	"time"
)

// RawProviderResponse preserves the original provider response for debugging.
type RawProviderResponse json.RawMessage

type CustomerInfo struct {
	CustomerID string `json:"customer_id"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone"`
}

type Metadata map[string]string

// --- Order types ---

type CreateOrderRequest struct {
	Provider    Provider      `json:"provider"`
	OrderID     string        `json:"order_id,omitempty"`
	AmountMinor AmountMinor   `json:"amount_minor"`
	Currency    Currency      `json:"currency"`
	Customer    *CustomerInfo `json:"customer"`
	ReturnURL   string        `json:"return_url,omitempty"`
	NotifyURL   string        `json:"notify_url,omitempty"`
	ExpiryTime  *time.Time    `json:"expiry_time,omitempty"`
	Note        string        `json:"note,omitempty"`
	Metadata    Metadata      `json:"metadata,omitempty"`
}

type Order struct {
	ProviderOrderID string              `json:"provider_order_id"`
	OrderID         string              `json:"order_id"`
	Status          OrderStatus         `json:"status"`
	AmountMinor     AmountMinor         `json:"amount_minor"`
	Currency        Currency            `json:"currency"`
	SessionID       string              `json:"session_id,omitempty"`
	ExpiryTime      *time.Time          `json:"expiry_time,omitempty"`
	CreatedAt       *time.Time          `json:"created_at,omitempty"`
	Customer        *CustomerInfo       `json:"customer,omitempty"`
	Metadata        Metadata            `json:"metadata,omitempty"`
	Raw             RawProviderResponse `json:"raw,omitempty"`
}

type GetOrderRequest struct {
	Provider Provider `json:"provider"`
	OrderID  string   `json:"order_id"`
}

type ListOrderPaymentsRequest struct {
	Provider Provider `json:"provider"`
	OrderID  string   `json:"order_id"`
}

// --- Payment types ---

type Payment struct {
	ProviderPaymentID string              `json:"provider_payment_id"`
	OrderID           string              `json:"order_id,omitempty"`
	Status            PaymentStatus       `json:"status"`
	AmountMinor       AmountMinor         `json:"amount_minor"`
	Currency          Currency            `json:"currency,omitempty"`
	PaymentGroup      string              `json:"payment_group,omitempty"`
	PaymentMethod     string              `json:"payment_method,omitempty"`
	PaymentTime       *time.Time          `json:"payment_time,omitempty"`
	CompletionTime    *time.Time          `json:"completion_time,omitempty"`
	IsCaptured        bool                `json:"is_captured"`
	BankReference     string              `json:"bank_reference,omitempty"`
	ErrorCode         string              `json:"error_code,omitempty"`
	ErrorMessage      string              `json:"error_message,omitempty"`
	Raw               RawProviderResponse `json:"raw,omitempty"`
}

type GetPaymentRequest struct {
	Provider  Provider `json:"provider"`
	OrderID   string   `json:"order_id,omitempty"`
	PaymentID string   `json:"payment_id"`
}

type ListPaymentsRequest struct {
	Provider Provider `json:"provider"`
	OrderID  string   `json:"order_id"`
}

type CapturePaymentRequest struct {
	Provider    Provider    `json:"provider"`
	PaymentID   string      `json:"payment_id"`
	AmountMinor AmountMinor `json:"amount_minor"`
	Currency    Currency    `json:"currency"`
}

// --- Refund types ---

type CreateRefundRequest struct {
	Provider    Provider    `json:"provider"`
	OrderID     string      `json:"order_id,omitempty"`
	PaymentID   string      `json:"payment_id,omitempty"`
	RefundID    string      `json:"refund_id,omitempty"`
	AmountMinor AmountMinor `json:"amount_minor"`
	Reason      string      `json:"reason,omitempty"`
	Metadata    Metadata    `json:"metadata,omitempty"`
}

type Refund struct {
	ProviderRefundID string              `json:"provider_refund_id"`
	RefundID         string              `json:"refund_id,omitempty"`
	OrderID          string              `json:"order_id,omitempty"`
	PaymentID        string              `json:"payment_id,omitempty"`
	Status           RefundStatus        `json:"status"`
	AmountMinor      AmountMinor         `json:"amount_minor"`
	Currency         Currency            `json:"currency,omitempty"`
	Reason           string              `json:"reason,omitempty"`
	ARN              string              `json:"arn,omitempty"`
	CreatedAt        *time.Time          `json:"created_at,omitempty"`
	ProcessedAt      *time.Time          `json:"processed_at,omitempty"`
	Raw              RawProviderResponse `json:"raw,omitempty"`
}

type GetRefundRequest struct {
	Provider Provider `json:"provider"`
	OrderID  string   `json:"order_id,omitempty"`
	RefundID string   `json:"refund_id"`
}

type ListRefundsRequest struct {
	Provider Provider `json:"provider"`
	OrderID  string   `json:"order_id"`
}

// --- Instrument types ---

type Instrument struct {
	CustomerID     string              `json:"customer_id"`
	InstrumentID   string              `json:"instrument_id"`
	InstrumentType string              `json:"instrument_type,omitempty"`
	DisplayValue   string              `json:"display_value,omitempty"`
	Status         string              `json:"status,omitempty"`
	CreatedAt      *time.Time          `json:"created_at,omitempty"`
	Raw            RawProviderResponse `json:"raw,omitempty"`
}

type GetInstrumentRequest struct {
	Provider     Provider `json:"provider"`
	CustomerID   string   `json:"customer_id"`
	InstrumentID string   `json:"instrument_id"`
}

type ListInstrumentsRequest struct {
	Provider   Provider `json:"provider"`
	CustomerID string   `json:"customer_id"`
}

type DeleteInstrumentRequest struct {
	Provider     Provider `json:"provider"`
	CustomerID   string   `json:"customer_id"`
	InstrumentID string   `json:"instrument_id"`
}

// --- Webhook types ---

type WebhookEvent struct {
	Provider   Provider         `json:"provider"`
	AccountID  string           `json:"account_id,omitempty"`
	EventType  WebhookEventType `json:"event_type"`
	EventTime  *time.Time       `json:"event_time,omitempty"`
	Order      *Order           `json:"order,omitempty"`
	Payment    *Payment         `json:"payment,omitempty"`
	Refund     *Refund          `json:"refund,omitempty"`
	RawPayload []byte           `json:"raw_payload,omitempty"`
	DedupeKey  string           `json:"dedupe_key"`
}

type WebhookMountOptions struct {
	BasePath string
	Handlers map[WebhookEventType]WebhookEventHandler
}

type WebhookEventHandler func(ctx context.Context, event *WebhookEvent) error
