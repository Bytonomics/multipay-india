package razorpay

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

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
