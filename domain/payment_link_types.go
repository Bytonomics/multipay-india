package domain

import "time"

type CreatePaymentLinkRequest struct {
	LinkID         string        `json:"link_id,omitempty"`
	AmountMinor    AmountMinor   `json:"amount_minor"`
	Currency       Currency      `json:"currency"`
	Purpose        string        `json:"purpose"`
	Customer       *CustomerInfo `json:"customer"`
	PartialPayment *bool         `json:"partial_payment,omitempty"`
	ExpiryTime     *time.Time    `json:"expiry_time,omitempty"`
	NotifySMS      *bool         `json:"notify_sms,omitempty"`
	NotifyEmail    *bool         `json:"notify_email,omitempty"`
	ReturnURL      string        `json:"return_url,omitempty"`
	Metadata       Metadata      `json:"metadata,omitempty"`
}

type PaymentLink struct {
	ProviderLinkID  string                     `json:"provider_link_id"`
	LinkID          string                     `json:"link_id"`
	Status          PaymentLinkStatus          `json:"status"`
	AmountMinor     AmountMinor                `json:"amount_minor"`
	AmountPaid      AmountMinor                `json:"amount_paid"`
	Currency        Currency                   `json:"currency"`
	Purpose         string                     `json:"purpose,omitempty"`
	LinkURL         string                     `json:"link_url,omitempty"`
	Customer        *CustomerInfo              `json:"customer,omitempty"`
	CreatedAt       *time.Time                 `json:"created_at,omitempty"`
	ExpiryTime      *time.Time                 `json:"expiry_time,omitempty"`
	Metadata        Metadata                   `json:"metadata,omitempty"`
	ProviderDetails *PaymentLinkProviderDetail `json:"provider_details,omitempty"`
	Raw             RawProviderResponse        `json:"raw,omitempty"`
}

type GetPaymentLinkRequest struct {
	LinkID string `json:"link_id"`
}

type CancelPaymentLinkRequest struct {
	LinkID string `json:"link_id"`
}
