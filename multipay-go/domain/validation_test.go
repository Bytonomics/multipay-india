package domain

import (
	"testing"
)

func TestCreateOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			req:     &CreateOrderRequest{Customer: &CustomerInfo{}, Currency: "INR", ReturnURL: "https://example.com/return"},
			wantErr: false,
		},
		{
			name:    "missing customer",
			req:     &CreateOrderRequest{Currency: "INR", ReturnURL: "https://example.com/return"},
			wantErr: true,
			errMsg:  "customer is required",
		},
		{
			name:    "missing currency",
			req:     &CreateOrderRequest{Customer: &CustomerInfo{}, ReturnURL: "https://example.com/return"},
			wantErr: true,
			errMsg:  "currency is required",
		},
		{
			name:    "missing return_url",
			req:     &CreateOrderRequest{Customer: &CustomerInfo{}, Currency: "INR"},
			wantErr: true,
			errMsg:  "return_url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestCreateSubscriptionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateSubscriptionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with plan_id",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				PlanID:         "plan123",
				CustomerPhone:  "9876543210",
				ReturnURL:      "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "valid with inline plan_details",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				CustomerPhone:  "9876543210",
				ReturnURL:      "https://example.com/return",
				PlanDetails: &CreatePlanRequest{
					PlanID:         "plan123",
					PlanName:       "test plan",
					PlanType:       PlanTypePeriodic,
					Currency:       "INR",
					AmountMinor:    50000,
					MaxAmountMinor: 100000,
					Interval:       1,
					IntervalType:   PlanIntervalMonth,
				},
			},
			wantErr: false,
		},
		{
			name: "missing return_url",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				PlanID:         "plan123",
				CustomerPhone:  "9876543210",
			},
			wantErr: true,
			errMsg:  "return_url is required and must not be empty",
		},
		{
			name: "missing both plan_id and plan_details",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				CustomerPhone:  "9876543210",
				ReturnURL:      "https://example.com/return",
			},
			wantErr: true,
			errMsg:  "exactly one of plan_id or plan_details is required",
		},
		{
			name: "both plan_id and plan_details provided",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				PlanID:         "plan123",
				CustomerPhone:  "9876543210",
				ReturnURL:      "https://example.com/return",
				PlanDetails: &CreatePlanRequest{
					PlanID:         "plan456",
					PlanName:       "test",
					PlanType:       PlanTypePeriodic,
					Currency:       "INR",
					AmountMinor:    50000,
					MaxAmountMinor: 100000,
					Interval:       1,
					IntervalType:   PlanIntervalMonth,
				},
			},
			wantErr: true,
			errMsg:  "plan_id and plan_details are mutually exclusive",
		},
		{
			name: "plan_details missing required fields",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				CustomerPhone:  "9876543210",
				ReturnURL:      "https://example.com/return",
				PlanDetails: &CreatePlanRequest{
					PlanID:   "plan123",
					PlanName: "",
					PlanType: PlanTypePeriodic,
				},
			},
			wantErr: true,
			errMsg:  "plan_details.plan_name is required",
		},
		{
			name: "periodic plan missing amount",
			req: &CreateSubscriptionRequest{
				SubscriptionID: "sub123",
				CustomerPhone:  "9876543210",
				ReturnURL:      "https://example.com/return",
				PlanDetails: &CreatePlanRequest{
					PlanID:         "plan123",
					PlanName:       "test plan",
					PlanType:       PlanTypePeriodic,
					Currency:       "INR",
					AmountMinor:    0,
					MaxAmountMinor: 100000,
					Interval:       1,
					IntervalType:   PlanIntervalMonth,
				},
			},
			wantErr: true,
			errMsg:  "plan_details.amount_minor is required and must be greater than 0 for PERIODIC plans",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestListRefundsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *ListRefundsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid with order_id (Cashfree)",
			req:     &ListRefundsRequest{OrderID: "order_123"},
			wantErr: false,
		},
		{
			name:    "valid with payment_id (Razorpay)",
			req:     &ListRefundsRequest{PaymentID: "pay_123"},
			wantErr: false,
		},
		{
			name:    "valid with both ids",
			req:     &ListRefundsRequest{OrderID: "order_123", PaymentID: "pay_123"},
			wantErr: false,
		},
		{
			name:    "missing both order_id and payment_id",
			req:     &ListRefundsRequest{},
			wantErr: true,
			errMsg:  "at least one of order_id (Cashfree) or payment_id (Razorpay) is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
