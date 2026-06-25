package razorpay

import (
	"encoding/json"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

const planFixtureJSON = `{"id":"plan_00000000000001","entity":"plan","interval":1,"period":"weekly","item":{"id":"item_1","name":"Test plan - Weekly","description":"d","amount":69900,"unit_amount":69900,"currency":"INR"},"notes":{"note":"hello"},"created_at":1580219935}`

const subFixtureJSON = `{"id":"sub_00000000000001","entity":"subscription","plan_id":"plan_00000000000001","status":"created","charge_at":1580453311,"start_at":1580626111,"end_at":1583433000,"total_count":6,"paid_count":0,"created_at":1580280581,"expire_by":1580626111,"short_url":"https://rzp.io/i/z3b1R61A9","remaining_count":5}`

func TestMapPlanFromResponse_ItemNesting(t *testing.T) {
	var plan razorpayPlanResponse
	if err := json.Unmarshal([]byte(planFixtureJSON), &plan); err != nil {
		t.Fatalf("failed to unmarshal plan fixture: %v", err)
	}

	// Marshal plan to bytes for mapper
	planBytes, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("failed to marshal plan: %v", err)
	}

	p := mapPlanFromResponse(&plan, planBytes)

	if p.PlanName != "Test plan - Weekly" {
		t.Fatalf("expected PlanName='Test plan - Weekly', got '%s'", p.PlanName)
	}
	if int64(p.AmountMinor) != 69900 {
		t.Fatalf("expected AmountMinor=69900, got %d", int64(p.AmountMinor))
	}
	if string(p.Currency) != "INR" {
		t.Fatalf("expected Currency='INR', got '%s'", string(p.Currency))
	}
	if p.IntervalType != domain.PlanIntervalWeek {
		t.Fatalf("expected IntervalType=PlanIntervalWeek, got %v", p.IntervalType)
	}
}

func TestMapPlanFromResponse_NotesFromMap(t *testing.T) {
	r := &razorpayPlanResponse{
		Notes: map[string]string{"note": "hello"},
		Item:  razorpayItem{Name: "x"},
	}

	// Marshal plan to bytes for mapper
	rawBytes, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal plan: %v", err)
	}

	p := mapPlanFromResponse(r, rawBytes)

	if p.Note != "hello" {
		t.Fatalf("expected Note='hello', got '%s'", p.Note)
	}
}

func TestMapSubscriptionFromResponse_AuthLinkAndCharge(t *testing.T) {
	var sub razorpaySubscriptionResponse
	if err := json.Unmarshal([]byte(subFixtureJSON), &sub); err != nil {
		t.Fatalf("failed to unmarshal subscription fixture: %v", err)
	}

	// Marshal subscription to bytes for mapper
	subBytes, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("failed to marshal subscription: %v", err)
	}

	s := mapSubscriptionFromResponse(&sub, subBytes)

	if s.AuthLink != "https://rzp.io/i/z3b1R61A9" {
		t.Fatalf("expected AuthLink='https://rzp.io/i/z3b1R61A9', got '%s'", s.AuthLink)
	}

	if s.PlanID != "plan_00000000000001" {
		t.Fatalf("expected PlanID='plan_00000000000001', got '%s'", s.PlanID)
	}

	if s.NextChargeDate == nil {
		t.Fatalf("expected non-nil NextChargeDate, got nil")
	}
	if s.NextChargeDate.Unix() != 1580453311 {
		t.Fatalf("expected NextChargeDate.Unix()=1580453311, got %d", s.NextChargeDate.Unix())
	}

	if s.Status != domain.SubscriptionStatusInitialized {
		t.Fatalf("expected Status=SubscriptionStatusInitialized (from 'created'), got %v", s.Status)
	}
}
