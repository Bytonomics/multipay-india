package domain

// --- Order Provider Details ---

type OrderProviderDetail struct {
	Cashfree *CashfreeOrderDetail `json:"cashfree,omitempty"`
	Razorpay *RazorpayOrderDetail `json:"razorpay,omitempty"`
}

type CashfreeOrderDetail struct {
	CfOrderID   string                `json:"cf_order_id"`
	Entity      string                `json:"entity,omitempty"`
	OrderNote   string                `json:"order_note,omitempty"`
	OrderSplits []CashfreeVendorSplit `json:"order_splits,omitempty"`
	OrderMeta   *CashfreeOrderMeta    `json:"order_meta,omitempty"`
}

type RazorpayOrderDetail struct {
	Entity     string `json:"entity,omitempty"`
	Receipt    string `json:"receipt,omitempty"`
	OfferID    string `json:"offer_id,omitempty"`
	AmountPaid int64  `json:"amount_paid"`
	AmountDue  int64  `json:"amount_due"`
	Attempts   int64  `json:"attempts"`
}

// --- Payment Provider Details ---

type PaymentProviderDetail struct {
	Cashfree *CashfreePaymentDetail `json:"cashfree,omitempty"`
	Razorpay *RazorpayPaymentDetail `json:"razorpay,omitempty"`
}

type CashfreePaymentDetail struct {
	CfPaymentID    string                       `json:"cf_payment_id"`
	Entity         string                       `json:"entity,omitempty"`
	OrderAmount    float64                      `json:"order_amount,omitempty"`
	OrderCurrency  string                       `json:"order_currency,omitempty"`
	PaymentMessage string                       `json:"payment_message,omitempty"`
	AuthID         string                       `json:"auth_id,omitempty"`
	ErrorDetails   *CashfreePaymentErrorDetails `json:"error_details,omitempty"`
}

type RazorpayPaymentDetail struct {
	Entity         string                `json:"entity,omitempty"`
	Description    string                `json:"description,omitempty"`
	Email          string                `json:"email,omitempty"`
	Contact        string                `json:"contact,omitempty"`
	Fee            int64                 `json:"fee,omitempty"`
	Tax            int64                 `json:"tax,omitempty"`
	AmountRefunded int64                 `json:"amount_refunded,omitempty"`
	RefundStatus   string                `json:"refund_status,omitempty"`
	International  bool                  `json:"international,omitempty"`
	CardID         string                `json:"card_id,omitempty"`
	Bank           string                `json:"bank,omitempty"`
	VPA            string                `json:"vpa,omitempty"`
	Wallet         string                `json:"wallet,omitempty"`
	AcquirerData   *RazorpayAcquirerData `json:"acquirer_data,omitempty"`
	ErrorSource    string                `json:"error_source,omitempty"`
	ErrorStep      string                `json:"error_step,omitempty"`
	ErrorReason    string                `json:"error_reason,omitempty"`
}

// --- Refund Provider Details ---

type RefundProviderDetail struct {
	Cashfree *CashfreeRefundDetail `json:"cashfree,omitempty"`
	Razorpay *RazorpayRefundDetail `json:"razorpay,omitempty"`
}

type CashfreeRefundDetail struct {
	CfRefundID        string                `json:"cf_refund_id"`
	CfPaymentID       string                `json:"cf_payment_id"`
	Entity            string                `json:"entity,omitempty"`
	RefundCharge      float64               `json:"refund_charge,omitempty"`
	RefundType        string                `json:"refund_type,omitempty"`
	RefundMode        string                `json:"refund_mode,omitempty"`
	StatusDescription string                `json:"status_description,omitempty"`
	RefundSpeed       *CashfreeRefundSpeed  `json:"refund_speed,omitempty"`
	RefundSplits      []CashfreeVendorSplit `json:"refund_splits,omitempty"`
	ChargesCurrency   string                `json:"charges_currency,omitempty"`
	ForexRate         float64               `json:"forex_conversion_rate,omitempty"`
	ForexCharge       float64               `json:"forex_conversion_handling_charge,omitempty"`
	ForexTax          float64               `json:"forex_conversion_handling_tax,omitempty"`
	ProviderMetadata  map[string]any        `json:"provider_metadata,omitempty"`
}

type RazorpayRefundDetail struct {
	Entity         string                `json:"entity,omitempty"`
	Receipt        string                `json:"receipt,omitempty"`
	SpeedRequested string                `json:"speed_requested,omitempty"`
	SpeedProcessed string                `json:"speed_processed,omitempty"`
	BatchID        string                `json:"batch_id,omitempty"`
	AcquirerData   *RazorpayAcquirerData `json:"acquirer_data,omitempty"`
}

// --- Instrument Provider Details ---

type InstrumentProviderDetail struct {
	Cashfree *CashfreeInstrumentDetail `json:"cashfree,omitempty"`
	Razorpay *RazorpayInstrumentDetail `json:"razorpay,omitempty"`
}

type CashfreeInstrumentDetail struct {
	AfaReference   string                  `json:"afa_reference,omitempty"`
	InstrumentUID  string                  `json:"instrument_uid,omitempty"`
	InstrumentMeta *CashfreeInstrumentMeta `json:"instrument_meta,omitempty"`
}

type RazorpayInstrumentDetail struct {
	Entity           string `json:"entity,omitempty"`
	Token            string `json:"token,omitempty"`
	MaxPaymentAmount int64  `json:"max_payment_amount,omitempty"`
	ExpiredAt        int64  `json:"expired_at,omitempty"`
	Compliant        bool   `json:"compliant,omitempty"`
}

// --- PaymentLink Provider Details ---

type PaymentLinkProviderDetail struct {
	Cashfree *CashfreePaymentLinkDetail `json:"cashfree,omitempty"`
	Razorpay *RazorpayPaymentLinkDetail `json:"razorpay,omitempty"`
}

type CashfreePaymentLinkDetail struct {
	CfLinkID         string                `json:"cf_link_id"`
	PartialPayments  bool                  `json:"partial_payments,omitempty"`
	MinPartialAmount float64               `json:"min_partial_amount,omitempty"`
	AutoReminders    bool                  `json:"auto_reminders,omitempty"`
	LinkQRCode       string                `json:"link_qrcode,omitempty"`
	OrderSplits      []CashfreeVendorSplit `json:"order_splits,omitempty"`
}

type RazorpayPaymentLinkDetail struct {
	Entity          string `json:"entity,omitempty"`
	Description     string `json:"description,omitempty"`
	CallbackURL     string `json:"callback_url,omitempty"`
	CallbackMethod  string `json:"callback_method,omitempty"`
	ReminderEnable  bool   `json:"reminder_enable,omitempty"`
	PaymentsCount   int64  `json:"payments_count,omitempty"`
	FirstMinPartial int64  `json:"first_min_partial_amount,omitempty"`
}

// --- Shared sub-types (Cashfree) ---

type CashfreeVendorSplit struct {
	VendorID   string  `json:"vendor_id,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

type CashfreeOrderMeta struct {
	ReturnURL      string `json:"return_url,omitempty"`
	NotifyURL      string `json:"notify_url,omitempty"`
	PaymentMethods string `json:"payment_methods,omitempty"`
}

type CashfreePaymentErrorDetails struct {
	ErrorCode           string `json:"error_code,omitempty"`
	ErrorDescription    string `json:"error_description,omitempty"`
	ErrorReason         string `json:"error_reason,omitempty"`
	ErrorSource         string `json:"error_source,omitempty"`
	ErrorCodeRaw        string `json:"error_code_raw,omitempty"`
	ErrorDescriptionRaw string `json:"error_description_raw,omitempty"`
	ErrorSubcodeRaw     string `json:"error_subcode_raw,omitempty"`
}

type CashfreeRefundSpeed struct {
	Requested string `json:"requested,omitempty"`
	Accepted  string `json:"accepted,omitempty"`
	Processed string `json:"processed,omitempty"`
	Message   string `json:"message,omitempty"`
}

type CashfreeInstrumentMeta struct {
	CardNetwork  string `json:"card_network,omitempty"`
	CardBankName string `json:"card_bank_name,omitempty"`
	CardCountry  string `json:"card_country,omitempty"`
	CardType     string `json:"card_type,omitempty"`
	CardSubType  string `json:"card_sub_type,omitempty"`
}

// --- Shared sub-types (Razorpay) ---

type RazorpayAcquirerData struct {
	BankTransactionID string `json:"bank_transaction_id,omitempty"`
	AuthCode          string `json:"auth_code,omitempty"`
	RRN               string `json:"rrn,omitempty"`
}
