package razorpay

import (
	"context"
	"fmt"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// TestRazorpayParseEvent_SubscriptionEvents tests parseEvent with subscription webhook events.
// This is a white-box test that directly calls the unexported parseEvent function.
// It verifies that the function correctly parses Razorpay subscription webhook payloads
// and maps them to the canonical domain.WebhookEvent types.
func TestRazorpayParseEvent_SubscriptionEvents(t *testing.T) {
	tests := []struct {
		eventName string
		expected  domain.WebhookEventType
	}{
		{"subscription.authenticated", domain.EventSubAuthenticated},
		{"subscription.activated", domain.EventSubActivated},
		{"subscription.charged", domain.EventSubCharged},
		{"subscription.pending", domain.EventSubOnHold},
		{"subscription.paused", domain.EventSubPaused},
		{"subscription.resumed", domain.EventSubResumed},
		{"subscription.cancelled", domain.EventSubCancelled},
		{"subscription.completed", domain.EventSubCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.eventName, func(t *testing.T) {
			// Build a webhook body matching the exact structure parseEvent expects:
			// - Top level: event (string), created_at (int64), payload (map)
			// - payload contains "subscription" key with "entity" wrapper containing subscription data
			body := fmt.Appendf(nil, `{
				"event":"%s",
				"created_at":1625097600,
				"payload":{
					"subscription":{
						"entity":{
							"id":"sub_1",
							"plan_id":"plan_1",
							"status":"active",
							"short_url":"https://rzp.io/i/test",
							"charge_at":1625184000,
							"start_at":1625097600,
							"expire_by":1640745600,
							"customer_id":"cust_1",
							"notes":{},
							"created_at":1625097600,
							"total_count":12,
							"paid_count":0,
							"remaining_count":12
						}
					}
				}
			}`, tt.eventName)

			// Parse the webhook event using the unexported parseEvent function
			event, err := parseEvent(context.Background(), body, nil)

			// Verify no error occurred during parsing
			if err != nil {
				t.Fatalf("failed to parse event %s: %v", tt.eventName, err)
			}

			// Verify the event type matches the expected canonical type
			if event.EventType != tt.expected {
				t.Fatalf("expected EventType=%v for %s, got %v", tt.expected, tt.eventName, event.EventType)
			}

			// Verify the subscription data was extracted and populated
			if event.Subscription == nil {
				t.Fatalf("expected non-nil event.Subscription for %s, got nil", tt.eventName)
			}

			// Verify the subscription ID was correctly parsed from the payload
			if event.Subscription.SubscriptionID != "sub_1" {
				t.Fatalf("expected SubscriptionID='sub_1', got '%s'", event.Subscription.SubscriptionID)
			}

			// Verify the plan ID was correctly parsed
			if event.Subscription.PlanID != "plan_1" {
				t.Fatalf("expected PlanID='plan_1', got '%s'", event.Subscription.PlanID)
			}

			// Verify the provider is set to Razorpay
			if event.Provider != domain.ProviderRazorpay {
				t.Fatalf("expected provider=%s, got %s", domain.ProviderRazorpay, event.Provider)
			}
		})
	}
}

// TestRazorpayParseEvent_RefundEvents tests that Razorpay refund events map to GENERAL event types.
// This verifies that refund.created, refund.processed, refund.failed are treated as general events (G11),
// not EventSubRefund, and are routed to the correct domain event types.
func TestRazorpayParseEvent_RefundEvents(t *testing.T) {
	cases := []struct {
		event    string
		expected domain.WebhookEventType
	}{
		{"refund.created", domain.EventRefundCreated},
		{"refund.processed", domain.EventRefundProcessed},
		{"refund.failed", domain.EventRefundFailed},
	}

	for _, tc := range cases {
		t.Run(tc.event, func(t *testing.T) {
			// Build a minimal webhook payload with the refund event type
			body := fmt.Appendf(nil, `{
				"event":"%s",
				"event_id":"evt_refund_1",
				"created_at":1625097600,
				"payload":{
					"refund":{
						"id":"ref_1",
						"amount":50000
					}
				}
			}`, tc.event)

			// Parse the webhook event using the unexported parseEvent function
			event, err := parseEvent(context.Background(), body, nil)

			// Verify no error occurred during parsing
			if err != nil {
				t.Fatalf("failed to parse event %s: %v", tc.event, err)
			}

			// Verify the event type matches the expected canonical type
			if event.EventType != tc.expected {
				t.Errorf("expected EventType=%v for %s, got %v", tc.expected, tc.event, event.EventType)
			}

			// Verify the provider is set to Razorpay
			if event.Provider != domain.ProviderRazorpay {
				t.Errorf("expected provider=%s, got %s", domain.ProviderRazorpay, event.Provider)
			}
		})
	}
}

// TestRazorpayParseEvent_SubscriptionRefundIsUnknown tests that subscription.refund (bogus entry removed from wave 2)
// now routes to EventUnknown instead of a subscription-specific event.
func TestRazorpayParseEvent_SubscriptionRefundIsUnknown(t *testing.T) {
	// subscription.refund is NOT a real Razorpay event name; was a bogus map entry
	body := []byte(`{
		"event":"subscription.refund",
		"event_id":"evt_sub_ref_1",
		"created_at":1625097600,
		"payload":{}
	}`)

	// Parse the webhook event
	event, err := parseEvent(context.Background(), body, nil)

	// Verify no error occurred during parsing
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the event type is EventUnknown (not mapped in webhookEventMap)
	if event.EventType != domain.EventUnknown {
		t.Errorf("expected EventUnknown, got %s", event.EventType)
	}

	// Verify the provider is set to Razorpay
	if event.Provider != domain.ProviderRazorpay {
		t.Errorf("expected provider=%s, got %s", domain.ProviderRazorpay, event.Provider)
	}
}

// TestRazorpayParseEvent_UnknownEventIsNotError tests that completely unknown event names
// return EventUnknown without error. Unknown events are gracefully handled and routed to the DefaultHandler.
func TestRazorpayParseEvent_UnknownEventIsNotError(t *testing.T) {
	body := []byte(`{
		"event":"some.future.event",
		"event_id":"evt_future_1",
		"created_at":1625097600,
		"payload":{}
	}`)

	// Parse the webhook event
	event, err := parseEvent(context.Background(), body, nil)

	// Verify parseEvent does not error on unknown events
	if err != nil {
		t.Fatalf("parseEvent should not error on unknown events: %v", err)
	}

	// Verify event is not nil
	if event == nil {
		t.Fatal("event must not be nil")
	}

	// Verify the event type is EventUnknown
	if event.EventType != domain.EventUnknown {
		t.Errorf("expected EventUnknown, got %s", event.EventType)
	}

	// Verify the provider is set to Razorpay
	if event.Provider != domain.ProviderRazorpay {
		t.Errorf("expected provider=%s, got %s", domain.ProviderRazorpay, event.Provider)
	}
}

// TestRazorpayParseEvent_D11RawVendorEventType tests that the D11 raw field RawVendorEventType
// is populated with the verbatim Razorpay event name, never hiding vendor data from callers.
func TestRazorpayParseEvent_D11RawVendorEventType(t *testing.T) {
	body := []byte(`{
		"event":"subscription.charged",
		"event_id":"evt_sub_charged_1",
		"created_at":1625097600,
		"payload":{
			"subscription":{
				"entity":{
					"id":"sub_1",
					"plan_id":"plan_1",
					"status":"active",
					"short_url":"https://rzp.io/i/test",
					"charge_at":1625184000,
					"start_at":1625097600,
					"expire_by":1640745600,
					"customer_id":"cust_1",
					"notes":{},
					"created_at":1625097600,
					"total_count":12,
					"paid_count":0,
					"remaining_count":12
				}
			}
		}
	}`)

	// Parse the webhook event
	event, err := parseEvent(context.Background(), body, nil)

	// Verify no error occurred during parsing
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify RawVendorEventType contains the verbatim Razorpay event name
	if event.RawVendorEventType != "subscription.charged" {
		t.Errorf("RawVendorEventType: got %q, want %q", event.RawVendorEventType, "subscription.charged")
	}

	// Verify RawVendorStatus contains the subscription status
	if event.RawVendorStatus != "active" {
		t.Errorf("RawVendorStatus: got %q, want %q", event.RawVendorStatus, "active")
	}

	// Verify the event type is correct
	if event.EventType != domain.EventSubCharged {
		t.Errorf("expected EventSubCharged, got %s", event.EventType)
	}

	// Verify the provider is set to Razorpay
	if event.Provider != domain.ProviderRazorpay {
		t.Errorf("expected provider=%s, got %s", domain.ProviderRazorpay, event.Provider)
	}
}
