package razorpay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// rzRoundTripFunc implements http.RoundTripper for mocking HTTP calls
type rzRoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper interface
func (f rzRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestBuildRazorpayCheckout(t *testing.T) {
	// Setup test config
	cfg := &Config{
		Environment: domain.EnvironmentSandbox,
		Key:         "test_public_key",
	}

	// Setup test request
	req := &domain.CreateOrderRequest{
		OrderID:     "order_123",
		ReturnURL:   "https://example.com/callback",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		Metadata:    domain.Metadata{"key": "value"},
	}

	// Setup test order
	order := &domain.Order{
		ProviderOrderID: "rzp_order_456",
		OrderID:         "order_123",
		AmountMinor:     domain.AmountMinor(50000),
		Currency:        domain.Currency("INR"),
		Status:          domain.OrderCreated,
	}

	// Call the helper function
	checkout := buildRazorpayCheckout(cfg, req, order)

	// Assert all fields match expected values
	if checkout.Provider != domain.ProviderRazorpay {
		t.Errorf("expected Provider %s, got %s", domain.ProviderRazorpay, checkout.Provider)
	}

	if checkout.Environment != domain.EnvironmentSandbox {
		t.Errorf("expected Environment %s, got %s", domain.EnvironmentSandbox, checkout.Environment)
	}

	if checkout.OrderID != "rzp_order_456" {
		t.Errorf("expected OrderID %s, got %s", "rzp_order_456", checkout.OrderID)
	}

	if checkout.PublicKey != "test_public_key" {
		t.Errorf("expected PublicKey %s, got %s", "test_public_key", checkout.PublicKey)
	}

	if checkout.CallbackURL != "https://example.com/callback" {
		t.Errorf("expected CallbackURL %s, got %s", "https://example.com/callback", checkout.CallbackURL)
	}

	if checkout.AmountMinor != domain.AmountMinor(50000) {
		t.Errorf("expected AmountMinor %d, got %d", 50000, checkout.AmountMinor)
	}

	if checkout.Currency != domain.Currency("INR") {
		t.Errorf("expected Currency %s, got %s", domain.Currency("INR"), checkout.Currency)
	}

	// Assert SessionID is empty (not used by Razorpay)
	if checkout.SessionID != "" {
		t.Errorf("expected SessionID to be empty, got %s", checkout.SessionID)
	}
}

func TestCreateOrder_PopulatesCheckout(t *testing.T) {
	// Create mock HTTP client that intercepts Razorpay SDK calls
	mockHTTPClient := &http.Client{
		Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Construct proper razorpayOrderResponse struct (strong typing, no maps)
			mockOrder := &razorpayOrderResponse{
				ID:         "order_1A2B3C4D5E6F",
				Entity:     "order",
				Receipt:    "order_123",
				Amount:     50000,
				Currency:   "INR",
				Status:     "created",
				CreatedAt:  1234567890,
				OfferID:    "",
				AmountPaid: 0,
				AmountDue:  50000,
				Attempts:   0,
			}

			// Marshal to JSON
			jsonData, err := json.Marshal(mockOrder)
			if err != nil {
				return nil, err
			}

			header := make(http.Header)
			header.Set("Content-Type", "application/json")
			return &http.Response{
				Status:        "200 OK",
				StatusCode:    200,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        header,
				Body:          io.NopCloser(strings.NewReader(string(jsonData))),
				ContentLength: int64(len(jsonData)),
			}, nil
		}),
	}

	// Setup test config with mocked HTTP client
	cfg := &Config{
		Key:         "rzp_test_1234567890",
		Secret:      "test_secret",
		Environment: domain.EnvironmentSandbox,
		AccountID:   "test_account",
		HTTPClient:  mockHTTPClient,
	}

	// Create adapter
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Setup test request
	req := &domain.CreateOrderRequest{
		OrderID:     "order_123",
		ReturnURL:   "https://example.com/callback",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		Metadata:    domain.Metadata{"key": "value"},
		Customer: &domain.CustomerInfo{
			CustomerID: "cust_123",
			Email:      "test@example.com",
			Phone:      "+919876543210",
		},
	}

	// Call CreateOrder
	order, err := adapter.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	// Assert checkout is populated
	if order.Checkout == nil {
		t.Fatal("expected order.Checkout to be populated, got nil")
	}

	// Assert Provider is ProviderRazorpay
	if order.Checkout.Provider != domain.ProviderRazorpay {
		t.Errorf("expected Provider %s, got %s", domain.ProviderRazorpay, order.Checkout.Provider)
	}

	// Assert Environment is UPPERCASE (SANDBOX)
	if order.Checkout.Environment != domain.EnvironmentSandbox {
		t.Errorf("expected Environment %s, got %s", domain.EnvironmentSandbox, order.Checkout.Environment)
	}

	// Assert OrderID (ProviderOrderID from response)
	if order.Checkout.OrderID != "order_1A2B3C4D5E6F" {
		t.Errorf("expected OrderID %s, got %s", "order_1A2B3C4D5E6F", order.Checkout.OrderID)
	}

	// Assert PublicKey
	if order.Checkout.PublicKey != "rzp_test_1234567890" {
		t.Errorf("expected PublicKey %s, got %s", "rzp_test_1234567890", order.Checkout.PublicKey)
	}

	// Assert CallbackURL
	if order.Checkout.CallbackURL != "https://example.com/callback" {
		t.Errorf("expected CallbackURL %s, got %s", "https://example.com/callback", order.Checkout.CallbackURL)
	}

	// Assert AmountMinor
	if order.Checkout.AmountMinor != domain.AmountMinor(50000) {
		t.Errorf("expected AmountMinor %d, got %d", 50000, order.Checkout.AmountMinor)
	}

	// Assert Currency
	if order.Checkout.Currency != domain.Currency("INR") {
		t.Errorf("expected Currency %s, got %s", domain.Currency("INR"), order.Checkout.Currency)
	}

	// Assert SessionID is empty (not used by Razorpay)
	if order.Checkout.SessionID != "" {
		t.Errorf("expected SessionID to be empty, got %s", order.Checkout.SessionID)
	}
}
