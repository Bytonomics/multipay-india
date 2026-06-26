package domain

import (
	"errors"
	"time"
)

type CreatePaymentLinkRequest struct {
	LinkID         string        `json:"link_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	AmountMinor    AmountMinor   `json:"amount_minor" pedantigo:"required,gt=0"`
	Currency       Currency      `json:"currency" pedantigo:"required,iso4217"`
	Purpose        string        `json:"purpose" pedantigo:"required,minLength=1,maxLength=500"`
	Customer       *CustomerInfo `json:"customer" pedantigo:"required"`
	PartialPayment *bool         `json:"partial_payment,omitempty"`
	ExpiryTime     *time.Time    `json:"expiry_time,omitempty"`
	NotifySMS      *bool         `json:"notify_sms,omitempty"`
	NotifyEmail    *bool         `json:"notify_email,omitempty"`
	ReturnURL      string        `json:"return_url,omitempty" pedantigo:"omitempty,url"`
	Metadata       Metadata      `json:"metadata,omitempty"`
}

// Validate enforces presence of the customer (pedantigo's Validate() does not enforce the
// `required` tag on a pointer field) and the Cashfree-mandatory customer phone. ReturnURL is
// intentionally optional here — recovery links are paid from an email and reconcile via webhook,
// so they do not require a redirect.
func (r *CreatePaymentLinkRequest) Validate() error {
	if r.Customer == nil {
		return errors.New("customer is required")
	}
	if r.Customer.Phone == "" {
		return errors.New("customer.phone is required")
	}
	return nil
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
	LinkID string `json:"link_id" pedantigo:"required,minLength=1"`
}

type CancelPaymentLinkRequest struct {
	LinkID string `json:"link_id" pedantigo:"required,minLength=1"`
}
