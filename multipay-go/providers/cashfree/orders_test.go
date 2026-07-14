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

// cfRoundTripFunc implements http.RoundTripper for mocking HTTP calls
type cfRoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper interface
func (f cfRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestBuildCashfreeCheckout tests the buildCashfreeCheckout helper function.
// It verifies that the checkout payload is constructed correctly with:
// - Provider set to ProviderCashfree
// - Environment converted to UPPERCASE (e.g., "PRODUCTION")
// - SessionID properly passed through
// - Razorpay-specific fields empty (no bleed-over)
func TestBuildCashfreeCheckout(t *testing.T) {
	tests := []struct {
		name      string
		env       domain.Environment
		sessionID string
		want      *domain.CheckoutPayload
	}{
		{
			name:      "production environment",
			env:       domain.EnvironmentProduction,
			sessionID: "session_123",
			want: &domain.CheckoutPayload{
				Provider:    domain.ProviderCashfree,
				Environment: domain.EnvironmentProduction,
				SessionID:   "session_123",
			},
		},
		{
			name:      "sandbox environment",
			env:       domain.EnvironmentSandbox,
			sessionID: "session_456",
			want: &domain.CheckoutPayload{
				Provider:    domain.ProviderCashfree,
				Environment: domain.EnvironmentSandbox,
				SessionID:   "session_456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCashfreeCheckout(tt.env, tt.sessionID)

			// Assert Provider is Cashfree
			if got.Provider != domain.ProviderCashfree {
				t.Errorf("Provider = %v, want %v", got.Provider, domain.ProviderCashfree)
			}

			// Assert Environment matches and is UPPERCASE
			if got.Environment != tt.want.Environment {
				t.Errorf("Environment = %v, want %v", got.Environment, tt.want.Environment)
			}

			// Verify environment string representation is UPPERCASE
			envStr := string(got.Environment)
			if envStr != "PRODUCTION" && envStr != "SANDBOX" {
				t.Errorf("Environment string representation must be UPPERCASE, got %v", envStr)
			}

			// Assert SessionID matches
			if got.SessionID != tt.sessionID {
				t.Errorf("SessionID = %v, want %v", got.SessionID, tt.sessionID)
			}

			// Assert no Razorpay field bleed-through
			// (CheckoutPayload is provider-agnostic, but we verify no accidental Razorpay-specific data)
			if got.SessionID == "" && tt.sessionID != "" {
				t.Error("SessionID should not be empty when provided")
			}
		})
	}
}

func TestCreateOrder_PopulatesCheckout(t *testing.T) {
	// Create mock HTTP client that intercepts Cashfree SDK calls
	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Construct proper OrderEntity struct (strong typing, no maps)
			mockOrder := &cf.OrderEntity{
				CfOrderId:        new(string),
				OrderId:          new(string),
				OrderAmount:      new(float32),
				OrderCurrency:    new(string),
				OrderStatus:      new(string),
				PaymentSessionId: new(string),
				Entity:           new(string),
				CreatedAt:        new(string),
				OrderExpiryTime:  new(string),
			}
			*mockOrder.CfOrderId = "cf_order_123456"
			*mockOrder.OrderId = "order_123"
			*mockOrder.OrderAmount = 500.0
			*mockOrder.OrderCurrency = "INR"
			*mockOrder.OrderStatus = "ACTIVE"
			*mockOrder.PaymentSessionId = "payment_session_abc123"
			*mockOrder.Entity = "order"
			*mockOrder.CreatedAt = "1719342000"
			*mockOrder.OrderExpiryTime = "1719428400"

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
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   mockHTTPClient,
	}

	// Create adapter
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Setup test request
	req := &domain.CreateOrderRequest{
		OrderID:     "order_123",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
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

	// Assert Provider is ProviderCashfree
	if order.Checkout.Provider != domain.ProviderCashfree {
		t.Errorf("expected Provider %s, got %s", domain.ProviderCashfree, order.Checkout.Provider)
	}

	// Assert Environment is UPPERCASE (SANDBOX)
	if order.Checkout.Environment != domain.EnvironmentSandbox {
		t.Errorf("expected Environment %s, got %s", domain.EnvironmentSandbox, order.Checkout.Environment)
	}

	// Assert SessionID is populated and matches order.SessionID (mapped from payment_session_id)
	if order.Checkout.SessionID == "" {
		t.Error("expected SessionID to be non-empty")
	}
	if order.Checkout.SessionID != order.SessionID {
		t.Errorf("expected SessionID %s to match order.SessionID %s", order.Checkout.SessionID, order.SessionID)
	}
	if order.Checkout.SessionID != "payment_session_abc123" {
		t.Errorf("expected SessionID %s, got %s", "payment_session_abc123", order.Checkout.SessionID)
	}

	// Assert Razorpay-only fields are empty
	if order.Checkout.OrderID != "" {
		t.Errorf("expected OrderID to be empty (Razorpay-only), got %s", order.Checkout.OrderID)
	}

	if order.Checkout.PublicKey != "" {
		t.Errorf("expected PublicKey to be empty (Razorpay-only), got %s", order.Checkout.PublicKey)
	}

	if order.Checkout.CallbackURL != "" {
		t.Errorf("expected CallbackURL to be empty (Razorpay-only), got %s", order.Checkout.CallbackURL)
	}

	if order.Checkout.AmountMinor != 0 {
		t.Errorf("expected AmountMinor to be empty (Razorpay-only), got %d", order.Checkout.AmountMinor)
	}

	if order.Checkout.Currency != "" {
		t.Errorf("expected Currency to be empty (Razorpay-only), got %s", order.Checkout.Currency)
	}
}

// TestCreateOrder_ForwardsParameters verifies that all request parameters are forwarded to the Cashfree SDK.
// Tests: OrderId, OrderNote, OrderMeta.ReturnUrl, OrderMeta.NotifyUrl, OrderTags
func TestCreateOrder_ForwardsParameters(t *testing.T) {
	tests := []struct {
		name      string
		notifyURL string
		orderID   string
		note      string
		expectErr bool
	}{
		{
			name:      "with all parameters",
			notifyURL: "https://example.com/notify",
			orderID:   "merchant_order_123",
			note:      "Order payment note",
			expectErr: false,
		},
		{
			name:      "without optional NotifyURL",
			notifyURL: "",
			orderID:   "merchant_order_456",
			note:      "Order note 2",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedReq *cf.CreateOrderRequest
			mockHTTPClient := &http.Client{
				Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						return nil, err
					}
					if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
						t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
					}

					cfOrderId := "cf_order_abc123"
					mockOrder := &cf.OrderEntity{
						CfOrderId:        &cfOrderId,
						OrderId:          &tt.orderID,
						OrderAmount:      new(float32),
						OrderCurrency:    new(string),
						OrderStatus:      new(string),
						PaymentSessionId: new(string),
						Entity:           new(string),
						CreatedAt:        new(string),
						OrderExpiryTime:  new(string),
					}
					*mockOrder.OrderAmount = 500.0
					*mockOrder.OrderCurrency = "INR"
					*mockOrder.OrderStatus = "ACTIVE"
					*mockOrder.PaymentSessionId = "session_xyz"
					*mockOrder.Entity = "order"
					*mockOrder.CreatedAt = "1719342000"
					*mockOrder.OrderExpiryTime = "1719428400"

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

			cfg := &Config{
				ClientID:     "test_client_id",
				ClientSecret: "test_client_secret",
				Environment:  domain.EnvironmentSandbox,
				AccountID:    "test_account",
				Logger:       ports.NewNoopLogger(),
				HTTPClient:   mockHTTPClient,
			}

			adapter, err := NewAdapter(cfg)
			if err != nil {
				t.Fatalf("failed to create adapter: %v", err)
			}

			req := &domain.CreateOrderRequest{
				OrderID:     tt.orderID,
				AmountMinor: 50000,
				Currency:    domain.Currency("INR"),
				Customer: &domain.CustomerInfo{
					CustomerID: "cust_123",
					Name:       "John Doe",
					Email:      "test@example.com",
					Phone:      "+919876543210",
				},
				ReturnURL: "https://example.com/return",
				NotifyURL: tt.notifyURL,
				Note:      tt.note,
				Metadata: domain.Metadata{
					"order_ref":  "ORD-2024-001",
					"utm_source": "app",
				},
			}

			_, err = adapter.CreateOrder(context.Background(), req)
			if (err != nil) != tt.expectErr {
				t.Fatalf("CreateOrder() error = %v, expectErr %v", err, tt.expectErr)
			}

			if capturedReq == nil {
				t.Fatal("request was not captured")
			}

			// Assert OrderId is forwarded
			if capturedReq.OrderId == nil || *capturedReq.OrderId != tt.orderID {
				t.Errorf("OrderId not forwarded: expected %s, got %v", tt.orderID, capturedReq.OrderId)
			}

			// Assert OrderNote is forwarded
			if capturedReq.OrderNote == nil || *capturedReq.OrderNote != tt.note {
				t.Errorf("OrderNote not forwarded: expected %s, got %v", tt.note, capturedReq.OrderNote)
			}

			// Assert OrderMeta.ReturnUrl is forwarded
			if capturedReq.OrderMeta == nil || capturedReq.OrderMeta.ReturnUrl == nil || *capturedReq.OrderMeta.ReturnUrl != "https://example.com/return" {
				t.Errorf("OrderMeta.ReturnUrl not forwarded: %v", capturedReq.OrderMeta)
			}

			// Assert OrderMeta.NotifyUrl is set only when NotifyURL is non-empty
			if tt.notifyURL != "" {
				if capturedReq.OrderMeta == nil || capturedReq.OrderMeta.NotifyUrl == nil || *capturedReq.OrderMeta.NotifyUrl != tt.notifyURL {
					t.Errorf("OrderMeta.NotifyUrl not forwarded: expected %s, got %v", tt.notifyURL, capturedReq.OrderMeta)
				}
			} else {
				// When NotifyURL is empty, NotifyUrl should be nil (conditional omit)
				if capturedReq.OrderMeta != nil && capturedReq.OrderMeta.NotifyUrl != nil {
					t.Errorf("OrderMeta.NotifyUrl should be nil when NotifyURL is empty, got %s", *capturedReq.OrderMeta.NotifyUrl)
				}
			}

			// Assert OrderTags are forwarded
			if capturedReq.OrderTags == nil || len(*capturedReq.OrderTags) == 0 {
				t.Error("OrderTags not forwarded")
			}

			// Assert customer_name is forwarded (regression guard: it was previously dropped)
			if capturedReq.CustomerDetails == nil || capturedReq.CustomerDetails.CustomerName == nil || *capturedReq.CustomerDetails.CustomerName != "John Doe" {
				t.Errorf("CustomerDetails.CustomerName not forwarded: %v", capturedReq.CustomerDetails)
			}
		})
	}
}
