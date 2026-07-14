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

	// MinPartialAmount is the minimum first installment when partial payments are
	// enabled. Minor units. Cashfree link_minimum_partial_amount / Razorpay first_min_partial_amount.
	MinPartialAmount AmountMinor `json:"min_partial_amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
	// AutoReminders enables provider-driven payment reminders.
	// Cashfree link_auto_reminders / Razorpay reminder_enable.
	AutoReminders *bool `json:"auto_reminders,omitempty"`
	// NotifyURL is the server-to-server notification (webhook) URL for the link.
	// Cashfree link_meta.notify_url only (Razorpay reconciles the link via account webhooks).
	NotifyURL string `json:"notify_url,omitempty" pedantigo:"omitempty,url"`
	// UpiIntent, when set (Cashfree accepts "true"), makes the link open the UPI-intent
	// flow on mobile. Cashfree link_meta.upi_intent only.
	UpiIntent string `json:"upi_intent,omitempty" pedantigo:"omitempty,maxLength=20"`
	// PaymentMethods restricts allowed modes on the link (comma-separated, e.g. "upi,cc").
	// Cashfree link_meta.payment_methods only.
	PaymentMethods string `json:"payment_methods,omitempty" pedantigo:"omitempty,maxLength=500"`
	// OrderSplits splits the collected amount across Easy-Split vendors.
	// Cashfree order_splits only; ignored by the Razorpay adapter.
	OrderSplits []VendorSplit `json:"order_splits,omitempty"`
	// UpiLink, when true, makes Razorpay create a UPI-only payment link (Razorpay upi_link).
	// Razorpay-only; ignored by the Cashfree adapter. Distinct from UpiIntent (Cashfree flow hint).
	// The adapter forwards exactly what the caller sets and imposes no default (nil ⇒ not sent).
	UpiLink *bool `json:"upi_link,omitempty"`
	// EnableInvoice toggles Cashfree invoice generation for the link (Cashfree enable_invoice).
	// Cashfree-only; ignored by Razorpay. nil ⇒ not sent (Cashfree applies its own default).
	EnableInvoice *bool `json:"enable_invoice,omitempty"`
	// Subscription attaches a subscription mandate to the Cashfree link (Cashfree link `subscription`).
	// Cashfree-only; ignored by Razorpay.
	Subscription *LinkSubscription `json:"subscription,omitempty"`
}

// LinkSubscription mirrors Cashfree's link `subscription` object: it attaches a subscription
// mandate to a payment link. Cashfree-only. PlanDetails REUSES the canonical CreatePlanRequest
// (same field set as Cashfree's link plan) rather than defining a duplicate plan type.
type LinkSubscription struct {
	// SubscriptionID → cf subscription.subscription_id.
	SubscriptionID string `json:"subscription_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	// AuthorizationAmountMinor → cf subscription.authorization_amount (minor units; the adapter
	// converts to major via the link currency).
	AuthorizationAmountMinor AmountMinor `json:"authorization_amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
	// AuthorizationAmountRefund → cf subscription.authorization_amount_refund.
	AuthorizationAmountRefund *bool `json:"authorization_amount_refund,omitempty"`
	// ExpiryTime → cf subscription.subscription_expiry_time (ISO 8601).
	ExpiryTime *time.Time `json:"expiry_time,omitempty"`
	// FirstChargeTime → cf subscription.subscription_first_charge_time (ISO 8601).
	FirstChargeTime *time.Time `json:"first_charge_time,omitempty"`
	// PlanDetails → cf subscription.plan_details (cf.CreateLinkPlanRequest). Reuses CreatePlanRequest.
	PlanDetails *CreatePlanRequest `json:"plan_details,omitempty"`
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
