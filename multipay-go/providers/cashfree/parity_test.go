package cashfree

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// newCaptureAdapter returns an adapter whose HTTP transport unmarshals the
// outbound request body into *target and replies with the given JSON body.
func newCaptureAdapter(t *testing.T, target any, respBody string) *Adapter {
	t.Helper()
	client := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, target); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(respBody)),
			}, nil
		}),
	}
	cfg := &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   client,
	}
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	return adapter
}

// TestCreateOrder_ForwardsRichFields asserts every new Create-Order optional
// field is copied into the Cashfree request: customer TPV bank fields, order_meta
// (payment_methods / filters / offer_filters / upi_app_priority), cart_details,
// terminal, order_splits and products.
func TestCreateOrder_ForwardsRichFields(t *testing.T) {
	orderResp := `{"cf_order_id":"cf_1","order_id":"o1","order_amount":500.0,"order_currency":"INR","order_status":"ACTIVE","payment_session_id":"sess","entity":"order"}`
	var captured *cf.CreateOrderRequest
	adapter := newCaptureAdapter(t, &captured, orderResp)

	req := &domain.CreateOrderRequest{
		OrderID:     "o1",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		ReturnURL:   "https://example.com/return",
		Customer: &domain.CustomerInfo{
			CustomerID:        "cust_1",
			Phone:             "+919876543210",
			BankAccountNumber: "1234567890",
			BankIFSC:          "HDFC0000123",
			BankCode:          3333,
			UID:               "uid_1",
		},
		PaymentMethods: "cc,dc,upi",
		PaymentMethodsFilters: &domain.OrderPaymentMethodsFilters{
			Methods: &domain.CardFilter{Action: "ALLOW", Values: []string{"credit_card"}},
			Filters: &domain.CardBinFilters{
				CardBins:        []int64{424242},
				CardSchemes:     []string{"VISA"},
				CardSuffix:      []int64{1111},
				CardIssuingBank: []string{"HDFC"},
			},
		},
		OfferFilters:   &domain.OfferFilters{Action: "ALLOW", Values: []string{"offer_1"}},
		UpiAppPriority: []string{"gpay", "phonepe"},
		CartDetails: &domain.CartDetails{
			CustomerNote:   "note",
			ShippingCharge: 10000, // ₹100
			CartName:       "cart",
			ShippingAddress: &domain.CartAddress{
				FullName: "John", Country: "IN", City: "Pune", State: "MH",
				Pincode: "411001", Address1: "a1", Address2: "a2",
			},
			CartItems: []domain.CartItem{
				{
					ItemID:            "i1",
					ItemName:          "Widget",
					OriginalUnitPrice: 20000, // ₹200
					ItemQuantity:      2,
					ItemCurrency:      domain.Currency("INR"),
				},
			},
		},
		Terminal:    &domain.TerminalDetails{TerminalID: "term_1", TerminalType: "SPOS", TerminalPhoneNo: "+911111111111"},
		OrderSplits: []domain.VendorSplit{{VendorID: "vendor_1", Amount: 10000, Tags: map[string]string{"k": "v"}}},
		Products:    &domain.OrderProducts{OneClickCheckout: &domain.OrderProductDetail{}},
	}

	// The outbound request is captured before the SDK decodes the response, so a
	// response-decode error is irrelevant here — assert only on the captured body.
	adapter.CreateOrder(context.Background(), req) //nolint:errcheck // capture-only test; response decode error is irrelevant
	if captured == nil {
		t.Fatal("request was not captured")
	}

	// Customer TPV bank fields.
	cd := captured.CustomerDetails
	if cd == nil {
		t.Fatal("customer_details missing")
	}
	if cd.CustomerBankAccountNumber == nil || *cd.CustomerBankAccountNumber != "1234567890" {
		t.Error("customer_bank_account_number not forwarded")
	}
	if cd.CustomerBankIfsc == nil || *cd.CustomerBankIfsc != "HDFC0000123" {
		t.Error("customer_bank_ifsc not forwarded")
	}
	if cd.CustomerBankCode == nil || *cd.CustomerBankCode != 3333 {
		t.Error("customer_bank_code not forwarded")
	}
	if cd.CustomerUid == nil || *cd.CustomerUid != "uid_1" {
		t.Error("customer_uid not forwarded")
	}

	// order_meta extras.
	om := captured.OrderMeta
	if om == nil {
		t.Fatal("order_meta missing")
	}
	if om.PaymentMethods != "cc,dc,upi" {
		t.Errorf("payment_methods not forwarded: %v", om.PaymentMethods)
	}
	if om.PaymentMethodsFilters == nil || om.PaymentMethodsFilters.Filters == nil {
		t.Fatal("payment_methods_filters not forwarded")
	}
	pmf := om.PaymentMethodsFilters
	if pmf.Methods == nil || pmf.Methods.Action == nil || *pmf.Methods.Action != "ALLOW" {
		t.Error("payment_methods_filters.methods not forwarded")
	}
	if len(pmf.Filters.CardBins) != 1 || pmf.Filters.CardBins[0] != 424242 {
		t.Errorf("card_bins not forwarded: %v", pmf.Filters.CardBins)
	}
	if len(pmf.Filters.CardSuffix) != 1 || pmf.Filters.CardSuffix[0] != 1111 {
		t.Errorf("card_suffix not forwarded: %v", pmf.Filters.CardSuffix)
	}
	if len(pmf.Filters.CardSchemes) != 1 || pmf.Filters.CardSchemes[0] != "VISA" {
		t.Errorf("card_schemes not forwarded: %v", pmf.Filters.CardSchemes)
	}
	if len(pmf.Filters.CardIssuingBank) != 1 || pmf.Filters.CardIssuingBank[0] != "HDFC" {
		t.Errorf("card_issuing_bank not forwarded: %v", pmf.Filters.CardIssuingBank)
	}
	if om.OfferFilters == nil || om.OfferFilters.Action == nil || *om.OfferFilters.Action != "ALLOW" {
		t.Error("offer_filters not forwarded")
	}
	upiPriority, ok := om.UpiAppPriority.([]interface{})
	if !ok || len(upiPriority) != 2 {
		t.Errorf("upi_app_priority not forwarded: %v", om.UpiAppPriority)
	}

	// cart_details.
	if captured.CartDetails == nil {
		t.Fatal("cart_details not forwarded")
	}
	cart := captured.CartDetails
	if cart.ShippingCharge == nil || *cart.ShippingCharge != 100.0 {
		t.Errorf("cart shipping_charge not converted to major: %v", cart.ShippingCharge)
	}
	if cart.CustomerShippingAddress == nil || cart.CustomerShippingAddress.Pincode == nil || *cart.CustomerShippingAddress.Pincode != "411001" {
		t.Error("cart shipping address not forwarded")
	}
	if len(cart.CartItems) != 1 {
		t.Fatalf("cart_items not forwarded: %v", cart.CartItems)
	}
	item := cart.CartItems[0]
	if item.ItemOriginalUnitPrice == nil || *item.ItemOriginalUnitPrice != 200.0 {
		t.Errorf("cart item original_unit_price not converted: %v", item.ItemOriginalUnitPrice)
	}
	if item.ItemQuantity == nil || *item.ItemQuantity != 2 {
		t.Errorf("cart item quantity not forwarded: %v", item.ItemQuantity)
	}

	// terminal.
	if captured.Terminal == nil || captured.Terminal.TerminalType != "SPOS" {
		t.Error("terminal not forwarded")
	}
	if captured.Terminal != nil && (captured.Terminal.TerminalId == nil || *captured.Terminal.TerminalId != "term_1") {
		t.Error("terminal_id not forwarded")
	}

	// order_splits.
	if len(captured.OrderSplits) != 1 {
		t.Fatalf("order_splits not forwarded: %v", captured.OrderSplits)
	}
	split := captured.OrderSplits[0]
	if split.VendorId == nil || *split.VendorId != "vendor_1" {
		t.Error("order_split vendor_id not forwarded")
	}
	if split.Amount == nil || *split.Amount != 100.0 {
		t.Errorf("order_split amount not converted to major: %v", split.Amount)
	}

	// products.
	if captured.Products == nil || captured.Products.OneClickCheckout == nil {
		t.Error("products.one_click_checkout not forwarded")
	}
}

// TestCreateRefund_ForwardsSpeedAndSplits asserts refund_speed and refund_splits
// reach the Cashfree refund request.
func TestCreateRefund_ForwardsSpeedAndSplits(t *testing.T) {
	var captured *cf.OrderCreateRefundRequest
	adapter := newCaptureAdapter(t, &captured, `{"refund_id":"r1"}`)

	req := &domain.CreateRefundRequest{
		OrderID:     "order_1",
		RefundID:    "r1",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		RefundSpeed: domain.RefundSpeedInstant,
		RefundSplits: []domain.RefundSplit{
			{VendorID: "vendor_1", Amount: 10000, Tags: map[string]string{"k": "v"}},
		},
	}

	createRefund(context.Background(), adapter, req) //nolint:errcheck // capture-only test; response decode error is irrelevant
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if captured.RefundSpeed == nil || *captured.RefundSpeed != "INSTANT" {
		t.Errorf("refund_speed not forwarded: %v", captured.RefundSpeed)
	}
	if len(captured.RefundSplits) != 1 {
		t.Fatalf("refund_splits not forwarded: %v", captured.RefundSplits)
	}
	rs := captured.RefundSplits[0]
	if rs.VendorId != "vendor_1" {
		t.Errorf("refund_split vendor_id not forwarded: %v", rs.VendorId)
	}
	if rs.Amount == nil || *rs.Amount != 100.0 {
		t.Errorf("refund_split amount not converted to major: %v", rs.Amount)
	}
}

// TestCreatePaymentLink_ForwardsRichFields asserts the new Cashfree link fields
// are copied: min_partial_amount, auto_reminders, order_splits, customer bank
// fields, and link_meta (notify_url / upi_intent / payment_methods).
func TestCreatePaymentLink_ForwardsRichFields(t *testing.T) {
	var captured *cf.CreateLinkRequest
	adapter := newCaptureAdapter(t, &captured, `{"link_id":"l1","link_status":"ACTIVE"}`)

	yes := true
	req := &domain.CreatePaymentLinkRequest{
		LinkID:      "l1",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		Purpose:     "test",
		Customer: &domain.CustomerInfo{
			CustomerID:        "cust_1",
			Phone:             "+919876543210",
			BankAccountNumber: "1234567890",
			BankIFSC:          "HDFC0000123",
			BankCode:          3333,
		},
		MinPartialAmount: 10000, // ₹100
		AutoReminders:    &yes,
		NotifyURL:        "https://example.com/notify",
		UpiIntent:        "true",
		PaymentMethods:   "upi,cc",
		ReturnURL:        "https://example.com/return",
		OrderSplits:      []domain.VendorSplit{{VendorID: "vendor_1", Amount: 10000}},
	}

	createPaymentLink(context.Background(), adapter, req) //nolint:errcheck // capture-only test; response decode error is irrelevant
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if captured.LinkMinimumPartialAmount == nil || *captured.LinkMinimumPartialAmount != 100.0 {
		t.Errorf("link_minimum_partial_amount not converted/forwarded: %v", captured.LinkMinimumPartialAmount)
	}
	if captured.LinkAutoReminders == nil || !*captured.LinkAutoReminders {
		t.Error("link_auto_reminders not forwarded")
	}
	if len(captured.OrderSplits) != 1 || captured.OrderSplits[0].VendorId == nil || *captured.OrderSplits[0].VendorId != "vendor_1" {
		t.Errorf("order_splits not forwarded: %v", captured.OrderSplits)
	}
	if captured.CustomerDetails.CustomerBankAccountNumber == nil || *captured.CustomerDetails.CustomerBankAccountNumber != "1234567890" {
		t.Error("link customer_bank_account_number not forwarded")
	}
	if captured.CustomerDetails.CustomerBankCode == nil || *captured.CustomerDetails.CustomerBankCode != 3333 {
		t.Error("link customer_bank_code not forwarded")
	}
	if captured.LinkMeta == nil {
		t.Fatal("link_meta not forwarded")
	}
	lm := captured.LinkMeta
	if lm.NotifyUrl == nil || *lm.NotifyUrl != "https://example.com/notify" {
		t.Errorf("link_meta.notify_url not forwarded: %v", lm.NotifyUrl)
	}
	if lm.UpiIntent == nil || *lm.UpiIntent != "true" {
		t.Errorf("link_meta.upi_intent not forwarded: %v", lm.UpiIntent)
	}
	if lm.PaymentMethods == nil || *lm.PaymentMethods != "upi,cc" {
		t.Errorf("link_meta.payment_methods not forwarded: %v", lm.PaymentMethods)
	}
	if lm.ReturnUrl == nil || *lm.ReturnUrl != "https://example.com/return" {
		t.Errorf("link_meta.return_url not forwarded: %v", lm.ReturnUrl)
	}
}
