package razorpay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// newRzCaptureAdapter returns a Razorpay adapter whose transport captures the
// outbound JSON body into *captured and replies with the given response map.
func newRzCaptureAdapter(t *testing.T, captured *map[string]any, resp map[string]any) *Adapter {
	t.Helper()
	client := &http.Client{
		Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, captured); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}
			jsonData, err := json.Marshal(resp)
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
		Key:         "rzp_mock_testonly",
		Secret:      "test_secret",
		Environment: domain.EnvironmentSandbox,
		HTTPClient:  client,
	}
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	return adapter
}

// TestCreateOrder_ForwardsPartialPayment asserts partial_payment and
// first_payment_min_amount reach the Razorpay order body.
func TestCreateOrder_ForwardsPartialPayment(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "order_1", "entity": "order", "amount": 50000, "currency": "INR", "status": "created",
	})

	yes := true
	req := &domain.CreateOrderRequest{
		OrderID:               "o1",
		AmountMinor:           50000,
		Currency:              domain.Currency("INR"),
		ReturnURL:             "https://example.com/return",
		Customer:              &domain.CustomerInfo{CustomerID: "c1", Phone: "+919876543210"},
		PartialPayment:        &yes,
		FirstPaymentMinAmount: 10000,
	}

	if _, err := adapter.CreateOrder(context.Background(), req); err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if pp, ok := captured["partial_payment"].(bool); !ok || !pp {
		t.Errorf("partial_payment not forwarded: %v", captured["partial_payment"])
	}
	// JSON numbers decode to float64.
	if amt, ok := captured["first_payment_min_amount"].(float64); !ok || amt != 10000 {
		t.Errorf("first_payment_min_amount not forwarded: %v", captured["first_payment_min_amount"])
	}
}

// TestCancelSubscription_ForwardsCancelAtCycleEnd asserts the optional cancel_at_cycle_end body
// param reaches Razorpay's Cancel Subscription request (was previously dropped — nil data map).
func TestCancelSubscription_ForwardsCancelAtCycleEnd(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "sub_1", "plan_id": "plan_1", "status": "cancelled",
	})

	atCycleEnd := true
	req := &domain.CancelSubscriptionRequest{SubscriptionID: "sub_1", CancelAtCycleEnd: &atCycleEnd}

	if _, err := adapter.CancelSubscription(context.Background(), req); err != nil {
		t.Fatalf("CancelSubscription returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if v, ok := captured["cancel_at_cycle_end"].(bool); !ok || !v {
		t.Errorf("cancel_at_cycle_end not forwarded: %v", captured["cancel_at_cycle_end"])
	}
}

// TestPauseSubscription_ForwardsPauseAt asserts pause_at reaches Razorpay's Pause request.
func TestPauseSubscription_ForwardsPauseAt(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{"id": "sub_1", "status": "paused"})
	req := &domain.PauseSubscriptionRequest{SubscriptionID: "sub_1", PauseAt: "now"}
	if _, err := adapter.PauseSubscription(context.Background(), req); err != nil {
		t.Fatalf("PauseSubscription returned error: %v", err)
	}
	if captured == nil || captured["pause_at"] != "now" {
		t.Errorf("pause_at not forwarded: %v", captured["pause_at"])
	}
}

// TestResumeSubscription_ForwardsResumeAt asserts resume_at reaches Razorpay's Resume request.
func TestResumeSubscription_ForwardsResumeAt(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{"id": "sub_1", "status": "active"})
	req := &domain.ResumeSubscriptionRequest{SubscriptionID: "sub_1", ResumeAt: "now"}
	if _, err := adapter.ResumeSubscription(context.Background(), req); err != nil {
		t.Fatalf("ResumeSubscription returned error: %v", err)
	}
	if captured == nil || captured["resume_at"] != "now" {
		t.Errorf("resume_at not forwarded: %v", captured["resume_at"])
	}
}

// TestCreateSubscription_SendsTotalCountUnconditionally asserts the ★ correctness fix: on the
// existing-PlanID path, total_count comes from the top-level canonical TotalCount (Razorpay
// treats total_count as mandatory), and quantity/offer_id/addons are forwarded too.
func TestCreateSubscription_SendsTotalCountUnconditionally(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "sub_1", "plan_id": "plan_1", "status": "created",
	})

	req := &domain.CreateSubscriptionRequest{
		SubscriptionID: "sub_1",
		PlanID:         "plan_1", // existing-PlanID path (no inline PlanDetails)
		CustomerEmail:  "test@example.com",
		CustomerPhone:  "+919876543210",
		ReturnURL:      "https://example.com/return",
		TotalCount:     12,
		Quantity:       2,
		OfferID:        "offer_1",
		Addons: []domain.SubscriptionAddon{
			{Name: "Setup fee", AmountMinor: 50000, Currency: domain.Currency("INR")},
		},
	}

	if _, err := adapter.CreateSubscription(context.Background(), req); err != nil {
		t.Fatalf("CreateSubscription returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if tc, ok := captured["total_count"].(float64); !ok || tc != 12 {
		t.Errorf("total_count not forwarded from canonical TotalCount (want 12): %v", captured["total_count"])
	}
	if q, ok := captured["quantity"].(float64); !ok || q != 2 {
		t.Errorf("quantity not forwarded: %v", captured["quantity"])
	}
	if oid, ok := captured["offer_id"].(string); !ok || oid != "offer_1" {
		t.Errorf("offer_id not forwarded: %v", captured["offer_id"])
	}
	addons, ok := captured["addons"].([]any)
	if !ok || len(addons) != 1 {
		t.Fatalf("addons not forwarded: %v", captured["addons"])
	}
	addon0, ok := addons[0].(map[string]any)
	if !ok {
		t.Fatalf("addon[0] wrong shape: %v", addons[0])
	}
	item, ok := addon0["item"].(map[string]any)
	if !ok || item["name"] != "Setup fee" {
		t.Errorf("addon item.name not forwarded: %v", addon0["item"])
	}
	if amt, ok := item["amount"].(float64); !ok || amt != 50000 {
		t.Errorf("addon item.amount not forwarded (minor units): %v", item["amount"])
	}
}

// TestCreateSubscription_TotalCountFallsBackToMaxCycles asserts that when the top-level
// TotalCount is zero, the inline plan's MaxCycles supplies total_count.
func TestCreateSubscription_TotalCountFallsBackToMaxCycles(t *testing.T) {
	var captured map[string]any
	// Inline-plan path makes TWO calls: Plan.Create then Subscription.Create. Reply the same
	// map to both; the capture keeps the last (subscription) body.
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "plan_or_sub", "status": "created",
	})

	req := &domain.CreateSubscriptionRequest{
		SubscriptionID: "sub_2",
		PlanDetails: &domain.CreatePlanRequest{
			PlanID:         "plan_2",
			PlanName:       "Plan 2",
			PlanType:       domain.PlanTypePeriodic,
			Currency:       domain.Currency("INR"),
			AmountMinor:    50000,
			MaxAmountMinor: 100000,
			Interval:       1,
			IntervalType:   domain.PlanIntervalMonth,
			MaxCycles:      6,
		},
		CustomerEmail: "test@example.com",
		CustomerPhone: "+919876543210",
		ReturnURL:     "https://example.com/return",
		// TotalCount intentionally zero → falls back to MaxCycles.
	}

	if _, err := adapter.CreateSubscription(context.Background(), req); err != nil {
		t.Fatalf("CreateSubscription returned error: %v", err)
	}
	if tc, ok := captured["total_count"].(float64); !ok || tc != 6 {
		t.Errorf("total_count did not fall back to MaxCycles (want 6): %v", captured["total_count"])
	}
}

// TestChangePlan_ForwardsUpdateFields asserts offer_id, quantity, remaining_count, start_at and
// customer_notify reach the Razorpay Update Subscription body.
func TestChangePlan_ForwardsUpdateFields(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "sub_1", "plan_id": "plan_new", "status": "active",
	})

	notify := false
	startAt := time.Unix(1893456000, 0) // fixed future timestamp
	req := &domain.ChangePlanRequest{
		SubscriptionID: "sub_1",
		NewPlanID:      "plan_new",
		OfferID:        "offer_9",
		Quantity:       3,
		RemainingCount: 5,
		StartAt:        &startAt,
		CustomerNotify: &notify,
	}

	if _, err := adapter.ChangePlan(context.Background(), req); err != nil {
		t.Fatalf("ChangePlan returned error: %v", err)
	}
	if oid, ok := captured["offer_id"].(string); !ok || oid != "offer_9" {
		t.Errorf("offer_id not forwarded: %v", captured["offer_id"])
	}
	if q, ok := captured["quantity"].(float64); !ok || q != 3 {
		t.Errorf("quantity not forwarded: %v", captured["quantity"])
	}
	if rc, ok := captured["remaining_count"].(float64); !ok || rc != 5 {
		t.Errorf("remaining_count not forwarded: %v", captured["remaining_count"])
	}
	if sa, ok := captured["start_at"].(float64); !ok || int64(sa) != 1893456000 {
		t.Errorf("start_at not forwarded: %v", captured["start_at"])
	}
	if cn, ok := captured["customer_notify"].(bool); !ok || cn != false {
		t.Errorf("customer_notify not forwarded: %v", captured["customer_notify"])
	}
}

// TestCreatePlan_ForwardsDescription asserts CreatePlanRequest.Description reaches the Razorpay
// plan item.description.
func TestCreatePlan_ForwardsDescription(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "plan_1", "period": "monthly", "interval": 1,
	})

	req := &domain.CreatePlanRequest{
		PlanID:         "plan_1",
		PlanName:       "Premium",
		PlanType:       domain.PlanTypePeriodic,
		Currency:       domain.Currency("INR"),
		AmountMinor:    50000,
		MaxAmountMinor: 100000,
		Interval:       1,
		IntervalType:   domain.PlanIntervalMonth,
		Description:    "Premium monthly plan",
	}

	if _, err := adapter.CreatePlan(context.Background(), req); err != nil {
		t.Fatalf("CreatePlan returned error: %v", err)
	}
	item, ok := captured["item"].(map[string]any)
	if !ok {
		t.Fatalf("item not present: %v", captured)
	}
	if item["description"] != "Premium monthly plan" {
		t.Errorf("item.description not forwarded: %v", item["description"])
	}
}

// TestChargeSubscription_ForwardsAddonDescription asserts ChargeSubscriptionRequest.Description
// reaches the Razorpay addon item.description.
func TestChargeSubscription_ForwardsAddonDescription(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "inv_1", "payment_id": "pay_1", "amount": 20000, "currency": "INR", "status": "paid",
	})

	req := &domain.ChargeSubscriptionRequest{
		SubscriptionID: "sub_1",
		PaymentRef:     "charge_1",
		AmountMinor:    20000,
		Currency:       domain.Currency("INR"),
		Remarks:        "One-time charge",
		Description:    "Extra usage",
	}

	if _, err := adapter.ChargeSubscription(context.Background(), req); err != nil {
		t.Fatalf("ChargeSubscription returned error: %v", err)
	}
	item, ok := captured["item"].(map[string]any)
	if !ok {
		t.Fatalf("item not present: %v", captured)
	}
	if item["description"] != "Extra usage" {
		t.Errorf("addon item.description not forwarded: %v", item["description"])
	}
	if item["name"] != "One-time charge" {
		t.Errorf("addon item.name not forwarded from Remarks: %v", item["name"])
	}
}

// TestCreateRefund_ForwardsSpeed asserts the canonical RefundSpeed maps to
// Razorpay's speed wire value (INSTANT → optimum).
func TestCreateRefund_ForwardsSpeed(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "refund_1", "status": "processed", "amount": 25000, "currency": "INR",
	})

	req := &domain.CreateRefundRequest{
		PaymentID:   "pay_1",
		AmountMinor: 25000,
		Currency:    domain.Currency("INR"),
		RefundSpeed: domain.RefundSpeedInstant,
	}

	if _, err := adapter.CreateRefund(context.Background(), req); err != nil {
		t.Fatalf("CreateRefund returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if speed, ok := captured["speed"].(string); !ok || speed != "optimum" {
		t.Errorf("speed not mapped/forwarded (want optimum): %v", captured["speed"])
	}
}

// TestCreatePaymentLink_ForwardsReminderAndPartial asserts reminder_enable and
// first_min_partial_amount reach the Razorpay payment-link body.
func TestCreatePaymentLink_ForwardsReminderAndPartial(t *testing.T) {
	var captured map[string]any
	adapter := newRzCaptureAdapter(t, &captured, map[string]any{
		"id": "plink_1", "amount": 50000, "currency": "INR", "status": "created", "short_url": "https://rzp.io/x",
	})

	yes := true
	req := &domain.CreatePaymentLinkRequest{
		AmountMinor:      50000,
		Currency:         domain.Currency("INR"),
		Purpose:          "test",
		Customer:         &domain.CustomerInfo{CustomerID: "c1", Phone: "+919876543210"},
		PartialPayment:   &yes,
		MinPartialAmount: 10000,
		AutoReminders:    &yes,
	}

	if _, err := adapter.CreatePaymentLink(context.Background(), req); err != nil {
		t.Fatalf("CreatePaymentLink returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("request was not captured")
	}
	if re, ok := captured["reminder_enable"].(bool); !ok || !re {
		t.Errorf("reminder_enable not forwarded: %v", captured["reminder_enable"])
	}
	if amt, ok := captured["first_min_partial_amount"].(float64); !ok || amt != 10000 {
		t.Errorf("first_min_partial_amount not forwarded: %v", captured["first_min_partial_amount"])
	}
	if ap, ok := captured["accept_partial"].(bool); !ok || !ap {
		t.Errorf("accept_partial not forwarded: %v", captured["accept_partial"])
	}
}
