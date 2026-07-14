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

// TestCreatePaymentLink_ForwardsParameters verifies that createPaymentLink forwards all parameters to Cashfree,
// including customer details, link_id, link_meta (conditional), link_notify, and link_notes.
func TestCreatePaymentLink_ForwardsParameters(t *testing.T) {
	tests := []struct {
		name           string
		linkID         string
		returnURL      string
		partialPayment *bool
		metadata       domain.Metadata
		expectLinkMeta bool // whether LinkMeta should be set
		expectLinkID   bool // whether LinkId should be set
	}{
		{
			name:           "with all parameters",
			linkID:         "link_123",
			returnURL:      "https://example.com/return",
			partialPayment: &[]bool{true}[0],
			metadata:       domain.Metadata{"order_ref": "ORD-1"},
			expectLinkMeta: true,
			expectLinkID:   true,
		},
		{
			name:           "with minimal parameters (no LinkID, no ReturnURL, no Metadata)",
			linkID:         "",
			returnURL:      "",
			partialPayment: nil,
			metadata:       nil,
			expectLinkMeta: false,
			expectLinkID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedReq *cf.CreateLinkRequest
			mockHTTPClient := &http.Client{
				Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						return nil, err
					}
					if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
						t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
					}

					linkId := "link_123"
					status := "ACTIVE"
					amount := float32(500.0)
					currency := "INR"
					mockLink := &cf.LinkEntity{
						LinkId:       &linkId,
						LinkStatus:   &status,
						LinkAmount:   &amount,
						LinkCurrency: &currency,
					}
					jsonData, err := json.Marshal(mockLink)
					if err != nil {
						return nil, err
					}
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(string(jsonData))),
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

			req := &domain.CreatePaymentLinkRequest{
				LinkID:         tt.linkID,
				AmountMinor:    50000,
				Currency:       domain.Currency("INR"),
				Purpose:        "Payment",
				PartialPayment: tt.partialPayment,
				Metadata:       tt.metadata,
				Customer: &domain.CustomerInfo{
					CustomerID: "cust_456",
					Name:       "John Doe",
					Email:      "john@example.com",
					Phone:      "+919876543210",
				},
				ReturnURL: tt.returnURL,
				NotifySMS: &[]bool{true}[0],
			}

			createPaymentLink(context.Background(), adapter, req)

			if capturedReq == nil {
				t.Fatal("request was not captured")
			}

			// Assert CustomerDetails are always forwarded
			if capturedReq.CustomerDetails.CustomerEmail == nil || *capturedReq.CustomerDetails.CustomerEmail != "john@example.com" {
				t.Error("customer email not forwarded")
			}
			if capturedReq.CustomerDetails.CustomerName == nil || *capturedReq.CustomerDetails.CustomerName != "John Doe" {
				t.Error("customer name not forwarded")
			}

			// Assert LinkId is set only when LinkID is non-empty
			if tt.expectLinkID {
				if capturedReq.LinkId == nil || *capturedReq.LinkId != "link_123" {
					t.Errorf("LinkId not forwarded: expected link_123, got %v", capturedReq.LinkId)
				}
			} else {
				if capturedReq.LinkId != nil {
					t.Errorf("LinkId should be nil when LinkID is empty, got %v", *capturedReq.LinkId)
				}
			}

			// Assert LinkPartialPayments is forwarded
			if tt.partialPayment != nil {
				if capturedReq.LinkPartialPayments == nil || *capturedReq.LinkPartialPayments != true {
					t.Errorf("LinkPartialPayments not forwarded: expected true, got %v", capturedReq.LinkPartialPayments)
				}
			}

			// Assert LinkMeta.ReturnUrl is set only when ReturnURL is non-empty
			if tt.expectLinkMeta {
				if capturedReq.LinkMeta == nil || capturedReq.LinkMeta.ReturnUrl == nil || *capturedReq.LinkMeta.ReturnUrl != "https://example.com/return" {
					t.Errorf("LinkMeta.ReturnUrl not forwarded: %v", capturedReq.LinkMeta)
				}
			} else {
				// When ReturnURL is empty, LinkMeta should be nil (conditional omit)
				if capturedReq.LinkMeta != nil {
					t.Errorf("LinkMeta should be nil when ReturnURL is empty, got %v", capturedReq.LinkMeta)
				}
			}

			// Assert LinkNotes is set only when Metadata is non-empty
			if len(tt.metadata) > 0 {
				if capturedReq.LinkNotes == nil || (*capturedReq.LinkNotes)["order_ref"] != "ORD-1" {
					t.Errorf("LinkNotes not forwarded: expected {order_ref:ORD-1}, got %v", capturedReq.LinkNotes)
				}
			} else {
				if capturedReq.LinkNotes != nil {
					t.Errorf("LinkNotes should be nil when Metadata is empty, got %v", capturedReq.LinkNotes)
				}
			}

			// Assert LinkNotify is always forwarded when NotifySMS is set
			if capturedReq.LinkNotify == nil {
				t.Error("link_notify not forwarded")
			}
		})
	}
}

// TestCreatePaymentLink_ForwardsInvoiceSubscriptionHolderName verifies createPaymentLink forwards
// enable_invoice, the nested link subscription (with reused CreatePlanRequest plan_details), and the
// customer_bank_acoount_holder_name (vendor's misspelling) to Cashfree.
func TestCreatePaymentLink_ForwardsInvoiceSubscriptionHolderName(t *testing.T) {
	var capturedReq *cf.CreateLinkRequest
	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}
			linkID := "link_1"
			status := "ACTIVE"
			amount := float32(500.0)
			currency := "INR"
			mockLink := &cf.LinkEntity{LinkId: &linkID, LinkStatus: &status, LinkAmount: &amount, LinkCurrency: &currency}
			jsonData, merr := json.Marshal(mockLink)
			if merr != nil {
				return nil, merr
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, nil
		}),
	}
	adapter, err := NewAdapter(&Config{ClientID: "test_client_id", ClientSecret: "test_client_secret", Environment: domain.EnvironmentSandbox, AccountID: "test_account", Logger: ports.NewNoopLogger(), HTTPClient: mockHTTPClient})
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	enableInvoice := true
	req := &domain.CreatePaymentLinkRequest{
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		Purpose:     "Payment",
		Customer: &domain.CustomerInfo{
			CustomerID:            "cust_1",
			Phone:                 "+919876543210",
			BankAccountHolderName: "Jane Doe",
		},
		EnableInvoice: &enableInvoice,
		Subscription: &domain.LinkSubscription{
			SubscriptionID:           "linksub_1",
			AuthorizationAmountMinor: 100000,
			PlanDetails: &domain.CreatePlanRequest{
				PlanID:         "plan_1",
				PlanName:       "Pro",
				PlanType:       domain.PlanTypePeriodic,
				Currency:       domain.Currency("INR"),
				MaxAmountMinor: 100000,
				AmountMinor:    50000,
				Interval:       1,
				IntervalType:   domain.PlanIntervalMonth,
			},
		},
	}

	createPaymentLink(context.Background(), adapter, req)

	if capturedReq == nil {
		t.Fatal("request was not captured")
	}
	// enable_invoice forwarded
	if capturedReq.EnableInvoice == nil || !*capturedReq.EnableInvoice {
		t.Errorf("expected enable_invoice=true, got %v", capturedReq.EnableInvoice)
	}
	// customer_bank_acoount_holder_name forwarded (vendor misspelling)
	if capturedReq.CustomerDetails.CustomerBankAcoountHolderName == nil || *capturedReq.CustomerDetails.CustomerBankAcoountHolderName != "Jane Doe" {
		t.Errorf("expected customer_bank_acoount_holder_name=Jane Doe, got %v", capturedReq.CustomerDetails.CustomerBankAcoountHolderName)
	}
	// subscription forwarded with reused plan_details (amount converted minor→major)
	if capturedReq.Subscription == nil {
		t.Fatal("subscription not forwarded")
	}
	if capturedReq.Subscription.SubscriptionId == nil || *capturedReq.Subscription.SubscriptionId != "linksub_1" {
		t.Errorf("expected subscription.subscription_id=linksub_1, got %v", capturedReq.Subscription.SubscriptionId)
	}
	if capturedReq.Subscription.AuthorizationAmount == nil || *capturedReq.Subscription.AuthorizationAmount != float32(1000.0) {
		t.Errorf("expected subscription.authorization_amount=1000.00 (100000 minor INR), got %v", capturedReq.Subscription.AuthorizationAmount)
	}
	if capturedReq.Subscription.PlanDetails == nil || capturedReq.Subscription.PlanDetails.PlanId != "plan_1" {
		t.Errorf("expected subscription.plan_details.plan_id=plan_1, got %v", capturedReq.Subscription.PlanDetails)
	}
	if capturedReq.Subscription.PlanDetails != nil && capturedReq.Subscription.PlanDetails.PlanMaxAmount != float32(1000.0) {
		t.Errorf("expected plan_max_amount=1000.00 (100000 minor INR), got %v", capturedReq.Subscription.PlanDetails.PlanMaxAmount)
	}
}
