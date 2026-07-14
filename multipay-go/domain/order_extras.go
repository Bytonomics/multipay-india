package domain

// This file holds canonical, provider-neutral mirrors of the richer optional
// structures the Cashfree Create-Order / Create-Link payloads support (cart,
// terminal, splits, payment-method filters, offers, products). They are typed
// (never map[string]any) so the boundary stays strict. Money fields are always
// minor units (AmountMinor); the Cashfree adapter converts to major units.

// CartAddress mirrors cf CartAddress (shipping / billing address on a cart).
type CartAddress struct {
	FullName string `json:"full_name,omitempty" pedantigo:"omitempty,maxLength=200"`
	Country  string `json:"country,omitempty" pedantigo:"omitempty,maxLength=100"`
	City     string `json:"city,omitempty" pedantigo:"omitempty,maxLength=100"`
	State    string `json:"state,omitempty" pedantigo:"omitempty,maxLength=100"`
	Pincode  string `json:"pincode,omitempty" pedantigo:"omitempty,maxLength=20"`
	Address1 string `json:"address_1,omitempty" pedantigo:"omitempty,maxLength=500"`
	Address2 string `json:"address_2,omitempty" pedantigo:"omitempty,maxLength=500"`
}

// CartItem mirrors cf CartItem. Prices are minor units.
type CartItem struct {
	ItemID              string      `json:"item_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	ItemName            string      `json:"item_name,omitempty" pedantigo:"omitempty,maxLength=250"`
	ItemDescription     string      `json:"item_description,omitempty" pedantigo:"omitempty,maxLength=1000"`
	ItemTags            []string    `json:"item_tags,omitempty"`
	ItemDetailsURL      string      `json:"item_details_url,omitempty" pedantigo:"omitempty,url"`
	ItemImageURL        string      `json:"item_image_url,omitempty" pedantigo:"omitempty,url"`
	OriginalUnitPrice   AmountMinor `json:"original_unit_price_minor,omitempty" pedantigo:"omitempty,gte=0"`
	DiscountedUnitPrice AmountMinor `json:"discounted_unit_price_minor,omitempty" pedantigo:"omitempty,gte=0"`
	ItemCurrency        Currency    `json:"item_currency,omitempty" pedantigo:"omitempty,iso4217"`
	ItemQuantity        int64       `json:"item_quantity,omitempty" pedantigo:"omitempty,gte=0"`
	ItemVariantID       string      `json:"item_variant_id,omitempty" pedantigo:"omitempty,maxLength=250"`
}

// CartDetails mirrors cf CartDetails. ShippingCharge is minor units.
type CartDetails struct {
	CustomerNote    string       `json:"customer_note,omitempty" pedantigo:"omitempty,maxLength=1000"`
	ShippingCharge  AmountMinor  `json:"shipping_charge_minor,omitempty" pedantigo:"omitempty,gte=0"`
	CartName        string       `json:"cart_name,omitempty" pedantigo:"omitempty,maxLength=250"`
	ShippingAddress *CartAddress `json:"shipping_address,omitempty"`
	BillingAddress  *CartAddress `json:"billing_address,omitempty"`
	CartItems       []CartItem   `json:"cart_items,omitempty"`
}

// TerminalDetails mirrors cf TerminalDetails (softPOS terminal binding).
// TerminalType is the only mandatory field on the vendor struct.
type TerminalDetails struct {
	TerminalID      string `json:"terminal_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	TerminalType    string `json:"terminal_type,omitempty" pedantigo:"omitempty,maxLength=50"`
	TerminalPhoneNo string `json:"terminal_phone_no,omitempty" pedantigo:"omitempty,maxLength=20"`
	TerminalName    string `json:"terminal_name,omitempty" pedantigo:"omitempty,maxLength=250"`
	CfTerminalID    int64  `json:"cf_terminal_id,omitempty" pedantigo:"omitempty,gte=0"`
	TerminalAddress string `json:"terminal_address,omitempty" pedantigo:"omitempty,maxLength=500"`
	TerminalNote    string `json:"terminal_note,omitempty" pedantigo:"omitempty,maxLength=500"`
}

// VendorSplit mirrors cf VendorSplit (Easy Split). Amount is minor units.
// A split names EITHER a flat amount OR a percentage (per Cashfree).
type VendorSplit struct {
	VendorID   string            `json:"vendor_id,omitempty" pedantigo:"omitempty,maxLength=250"`
	Amount     AmountMinor       `json:"amount_minor,omitempty" pedantigo:"omitempty,gte=0"`
	Percentage float64           `json:"percentage,omitempty" pedantigo:"omitempty,gte=0,lte=100"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// OrderProducts mirrors cf Products (product toggles). A non-nil pointer means "enabled".
type OrderProducts struct {
	OneClickCheckout *OrderProductDetail `json:"one_click_checkout,omitempty"`
	VerifyPay        *OrderProductDetail `json:"verify_pay,omitempty"`
}

// OrderProductDetail mirrors cf ProductDetails (currently just an enabled toggle).
type OrderProductDetail struct {
	Enabled bool `json:"enabled,omitempty"`
}

// CardFilter mirrors the ALLOW/values card filter shape cf uses for bins,
// schemes, suffixes and issuing banks under payment_methods_filters.filters.
type CardFilter struct {
	Action string   `json:"action,omitempty" pedantigo:"omitempty,maxLength=20"`
	Values []string `json:"values,omitempty"`
}

// OrderPaymentMethodsFilters mirrors cf OrderMetaPaymentMethodsFilters.
type OrderPaymentMethodsFilters struct {
	Methods *CardFilter     `json:"methods,omitempty"`
	Filters *CardBinFilters `json:"filters,omitempty"`
}

// CardBinFilters mirrors cf OrderPaymentMethodFilters — the card bin/scheme/
// bank/suffix allow-lists. Bins and suffixes are numeric on the Cashfree wire
// (int32); schemes and issuing banks are strings. Canonical mirrors those types.
type CardBinFilters struct {
	CardBins        []int64  `json:"card_bins,omitempty"`
	CardSchemes     []string `json:"card_schemes,omitempty"`
	CardSuffix      []int64  `json:"card_suffix,omitempty"`
	CardIssuingBank []string `json:"card_issuing_bank,omitempty"`
}

// OfferFilters mirrors cf OrderMetaOfferFilters (allow/deny offer ids).
type OfferFilters struct {
	Action string   `json:"action,omitempty" pedantigo:"omitempty,maxLength=20"`
	Values []string `json:"values,omitempty"`
}
