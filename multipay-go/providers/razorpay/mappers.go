package razorpay

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// rawMapResponse converts bytes to a RawProviderResponse.
func rawMapResponse(b []byte) domain.RawProviderResponse {
	return domain.RawProviderResponse(b)
}

// isNotFoundError checks if an error indicates a "not found" condition.
// This is specific to Razorpay's error response pattern.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found")
}

// PART A: Typed structs for Razorpay API responses

type razorpayItem struct {
	Name        string `json:"name"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

type razorpayPlanResponse struct {
	ID        string            `json:"id"`
	Period    string            `json:"period"`
	Interval  int32             `json:"interval"`
	Item      razorpayItem      `json:"item"`
	Notes     map[string]string `json:"notes"`
	CreatedAt int64             `json:"created_at"`
}

type razorpayPlanCreateRequest struct {
	Period   string            `json:"period"`
	Interval int32             `json:"interval"`
	Item     razorpayItem      `json:"item"`
	Notes    map[string]string `json:"notes,omitempty"`
}

type razorpaySubscriptionResponse struct {
	ID             string            `json:"id"`
	PlanID         string            `json:"plan_id"`
	Status         string            `json:"status"`
	ShortURL       string            `json:"short_url"`
	ChargeAt       int64             `json:"charge_at"`
	StartAt        int64             `json:"start_at"`
	ExpireBy       int64             `json:"expire_by"`
	CustomerID     string            `json:"customer_id"`
	Notes          map[string]string `json:"notes"`
	CreatedAt      int64             `json:"created_at"`
	TotalCount     int32             `json:"total_count"`
	PaidCount      int32             `json:"paid_count"`
	RemainingCount int32             `json:"remaining_count"`
}

type razorpayInvoiceResponse struct {
	ID        string `json:"id"`
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	OrderID   string `json:"order_id"`
	PaidAt    int64  `json:"paid_at"`
	CreatedAt int64  `json:"created_at"`
}

type razorpayInvoiceListResponse struct {
	Items []razorpayInvoiceResponse `json:"items"`
	Count int                       `json:"count"`
}

// PART B: Boundary helpers for encoding/decoding

func decodeResponse[T any](m map[string]any) (*T, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal razorpay response: %w", err)
	}
	out := new(T)
	if err := json.Unmarshal(b, out); err != nil {
		return nil, fmt.Errorf("decode razorpay response: %w", err)
	}
	return out, nil
}

func encodeRequest(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal razorpay request: %w", err)
	}
	m := map[string]any{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("encode razorpay request: %w", err)
	}
	return m, nil
}

func unixPtr(sec int64) *time.Time {
	if sec == 0 {
		return nil
	}
	t := time.Unix(sec, 0).UTC()
	return &t
}

// PART C: Interval type and status helpers

func mapPlanIntervalTypeToRazorpay(it domain.PlanIntervalType) string {
	switch it {
	case domain.PlanIntervalDay:
		return "daily"
	case domain.PlanIntervalWeek:
		return "weekly"
	case domain.PlanIntervalMonth:
		return "monthly"
	case domain.PlanIntervalYear:
		return "yearly"
	default:
		return "monthly"
	}
}

func mapInvoiceStatusToPaymentStatus(status string) domain.SubscriptionPaymentStatus {
	switch strings.ToLower(status) {
	case "paid":
		return domain.SubPaymentStatusSuccess
	case "issued", "partially_paid":
		return domain.SubPaymentStatusPending
	case "cancelled":
		return domain.SubPaymentStatusCancelled
	case "expired":
		return domain.SubPaymentStatusFailed
	default:
		return domain.SubPaymentStatusPending
	}
}

// mapPlanIntervalTypeFromRazorpay maps Razorpay period strings to canonical domain.PlanIntervalType.
// Razorpay uses: "daily", "weekly", "monthly", "yearly"
func mapPlanIntervalTypeFromRazorpay(period string) domain.PlanIntervalType {
	switch strings.ToLower(period) {
	case "daily", "day":
		return domain.PlanIntervalDay
	case "weekly", "week":
		return domain.PlanIntervalWeek
	case "monthly", "month":
		return domain.PlanIntervalMonth
	case "yearly", "year":
		return domain.PlanIntervalYear
	default:
		return domain.PlanIntervalMonth
	}
}

// mapSubscriptionStatusFromRazorpay maps Razorpay subscription status strings to canonical domain.SubscriptionStatus.
// Razorpay subscription statuses: created, issued, authenticated, paused, halted, cancelled, expired, completed
func mapSubscriptionStatusFromRazorpay(status string) domain.SubscriptionStatus {
	switch strings.ToLower(status) {
	case "created":
		return domain.SubscriptionStatusInitialized
	case "issued":
		return domain.SubscriptionStatusPending
	case "authenticated":
		return domain.SubscriptionStatusAuthenticated
	case "active":
		return domain.SubscriptionStatusActive
	case "paused":
		return domain.SubscriptionStatusPaused
	case "halted":
		return domain.SubscriptionStatusHalted
	case "cancelled":
		return domain.SubscriptionStatusCancelled
	case "expired":
		return domain.SubscriptionStatusExpired
	case "completed":
		return domain.SubscriptionStatusCompleted
	default:
		return domain.SubscriptionStatusInitialized
	}
}

// PART E: New struct-based mappers

func mapPlanFromResponse(r *razorpayPlanResponse, rawJSON []byte) *domain.Plan {
	return &domain.Plan{
		PlanID:         r.ID,
		PlanName:       r.Item.Name,
		PlanType:       domain.PlanTypePeriodic,
		Currency:       domain.Currency(r.Item.Currency),
		AmountMinor:    domain.AmountMinor(r.Item.Amount),
		MaxAmountMinor: domain.AmountMinor(r.Item.Amount),
		Interval:       r.Interval,
		IntervalType:   mapPlanIntervalTypeFromRazorpay(r.Period),
		Note:           r.Notes["note"],
		Status:         "",
		Provider:       domain.ProviderRazorpay,
		Raw:            rawMapResponse(rawJSON),
	}
}

func mapSubscriptionFromResponse(r *razorpaySubscriptionResponse, rawJSON []byte) *domain.Subscription {
	return &domain.Subscription{
		SubscriptionID:         r.ID,
		ProviderSubscriptionID: r.ID,
		PlanID:                 r.PlanID,
		Status:                 mapSubscriptionStatusFromRazorpay(r.Status),
		AuthLink:               r.ShortURL,
		ExpiresAt:              unixPtr(r.ExpireBy),
		FirstChargeTime:        unixPtr(r.StartAt),
		NextChargeDate:         unixPtr(r.ChargeAt),
		CustomerEmail:          "",
		CustomerPhone:          "",
		Provider:               domain.ProviderRazorpay,
		Raw:                    rawMapResponse(rawJSON),
	}
}

func mapInvoiceToSubscriptionPayment(inv *razorpayInvoiceResponse, subscriptionID string, rawJSON []byte) *domain.SubscriptionPayment {
	// PaymentType and RetryAttempts: not available from Razorpay invoices. Always zero.
	return &domain.SubscriptionPayment{
		PaymentID:      inv.PaymentID,
		SubscriptionID: subscriptionID,
		AmountMinor:    domain.AmountMinor(inv.Amount),
		Status:         mapInvoiceStatusToPaymentStatus(inv.Status),
		ScheduledDate:  unixPtr(inv.CreatedAt),
		InitiatedDate:  unixPtr(inv.PaidAt),
		Provider:       domain.ProviderRazorpay,
		Raw:            rawMapResponse(rawJSON),
	}
}
