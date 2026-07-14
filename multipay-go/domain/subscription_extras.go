package domain

// This file holds canonical, provider-neutral mirrors of the richer optional
// structures the Cashfree Create-Subscription / Raise-Charge / Manage-Subscription
// payloads support (authorization details, subscription meta extras, TPV bank
// details, Easy-Split payment splits). They are typed (never map[string]any) so
// the boundary stays strict. Money fields are always minor units (AmountMinor);
// the Cashfree adapter converts to major units using the request currency.

// SubscriptionAuthorizationDetails mirrors cf CreateSubscriptionRequestAuthorizationDetails.
// It controls the mandate-authorization step of a Cashfree subscription. All fields optional.
//   - AuthorizationAmountMinor: the amount to charge/hold during authorization (minor units).
//   - AuthorizationAmountRefund: whether the authorization amount is refunded after mandate setup.
//   - PaymentMethods: the payment methods offered on the mandate-authorization screen
//     (e.g. "upi", "enach", "card", "pnach"). Cashfree-only; Razorpay has no equivalent.
type SubscriptionAuthorizationDetails struct {
	AuthorizationAmountMinor  AmountMinor `json:"authorization_amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
	AuthorizationAmountRefund *bool       `json:"authorization_amount_refund,omitempty"`
	PaymentMethods            []string    `json:"payment_methods,omitempty"`
}

// SubscriptionMeta mirrors the optional cf CreateSubscriptionRequestSubscriptionMeta extras
// beyond ReturnURL (which stays a first-class field on CreateSubscriptionRequest).
//   - NotificationChannel: channels Cashfree uses to notify the customer (e.g. "EMAIL", "SMS").
//   - SessionIDExpiry: ISO-8601 timestamp after which the subscription session id expires.
//
// Cashfree-only; Razorpay has no equivalent.
type SubscriptionMeta struct {
	NotificationChannel []string `json:"notification_channel,omitempty"`
	SessionIDExpiry     string   `json:"session_id_expiry,omitempty" pedantigo:"omitempty,maxLength=64"`
}

// SubscriptionBankDetails mirrors the TPV (Third-Party Validation) bank fields on the
// cf SubscriptionCustomerDetails. Used to pre-bind the customer's bank account for an
// eNACH / physical-NACH mandate. Cashfree-only; Razorpay has no equivalent.
type SubscriptionBankDetails struct {
	AccountHolderName string `json:"account_holder_name,omitempty" pedantigo:"omitempty,maxLength=200"`
	AccountNumber     string `json:"account_number,omitempty" pedantigo:"omitempty,maxLength=50"`
	IFSC              string `json:"ifsc,omitempty" pedantigo:"omitempty,maxLength=20"`
	BankCode          string `json:"bank_code,omitempty" pedantigo:"omitempty,maxLength=20"`
	AccountType       string `json:"account_type,omitempty" pedantigo:"omitempty,maxLength=20"`
}

// SubscriptionPaymentSplit mirrors cf SubscriptionPaymentSplitItem (Easy Split). Each split
// names a vendor and a percentage of the collected amount. Cashfree-only.
type SubscriptionPaymentSplit struct {
	VendorID   string  `json:"vendor_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	Percentage float64 `json:"percentage,omitempty" pedantigo:"omitempty,gte=0,lte=100"`
}

// SubscriptionAddon mirrors a Razorpay subscription addon (item collected upfront during
// authorization). Amount is minor units (Razorpay-native). Razorpay-only.
type SubscriptionAddon struct {
	Name        string      `json:"name,omitempty" pedantigo:"omitempty,maxLength=250"`
	AmountMinor AmountMinor `json:"amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
	Currency    Currency    `json:"currency,omitempty" pedantigo:"omitempty,iso4217"`
	Quantity    int64       `json:"quantity,omitempty" pedantigo:"omitempty,gte=1"`
}
