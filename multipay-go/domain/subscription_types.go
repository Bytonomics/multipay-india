package domain

import (
	"errors"
	"fmt"
	"time"
)

// PlanType represents the type of plan.
type PlanType string

const (
	PlanTypePeriodic PlanType = "PERIODIC"
	PlanTypeOnDemand PlanType = "ON_DEMAND"
)

// PlanIntervalType represents the interval unit for periodic plans.
type PlanIntervalType string

const (
	PlanIntervalDay   PlanIntervalType = "DAY"
	PlanIntervalWeek  PlanIntervalType = "WEEK"
	PlanIntervalMonth PlanIntervalType = "MONTH"
	PlanIntervalYear  PlanIntervalType = "YEAR"
)

// SubscriptionStatus represents the lifecycle state of a subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusInitialized         SubscriptionStatus = "INITIALIZED"
	SubscriptionStatusBankApprovalPending SubscriptionStatus = "BANK_APPROVAL_PENDING"
	SubscriptionStatusAuthenticated       SubscriptionStatus = "AUTHENTICATED"
	SubscriptionStatusActive              SubscriptionStatus = "ACTIVE"
	SubscriptionStatusPending             SubscriptionStatus = "PENDING"
	SubscriptionStatusOnHold              SubscriptionStatus = "ON_HOLD"
	SubscriptionStatusHalted              SubscriptionStatus = "HALTED"
	SubscriptionStatusPaused              SubscriptionStatus = "PAUSED"
	SubscriptionStatusCustomerPaused      SubscriptionStatus = "CUSTOMER_PAUSED"
	SubscriptionStatusCancelled           SubscriptionStatus = "CANCELLED"
	SubscriptionStatusCustomerCancelled   SubscriptionStatus = "CUSTOMER_CANCELLED"
	SubscriptionStatusCompleted           SubscriptionStatus = "COMPLETED"
	SubscriptionStatusExpired             SubscriptionStatus = "EXPIRED"
)

// ScheduleChangeAt represents when a plan change should take effect.
type ScheduleChangeAt string

const (
	ScheduleChangeNow      ScheduleChangeAt = "NOW"
	ScheduleChangeCycleEnd ScheduleChangeAt = "CYCLE_END"
)

// SubscriptionPaymentStatus represents the status of a subscription payment.
type SubscriptionPaymentStatus string

const (
	SubPaymentStatusScheduled SubscriptionPaymentStatus = "SCHEDULED"
	SubPaymentStatusPending   SubscriptionPaymentStatus = "PENDING"
	SubPaymentStatusSuccess   SubscriptionPaymentStatus = "SUCCESS"
	SubPaymentStatusFailed    SubscriptionPaymentStatus = "FAILED"
	SubPaymentStatusCancelled SubscriptionPaymentStatus = "CANCELLED"
)

// SubscriptionPaymentType represents the type of subscription payment.
type SubscriptionPaymentType string

const (
	SubPaymentTypeAuth   SubscriptionPaymentType = "AUTH"
	SubPaymentTypeCharge SubscriptionPaymentType = "CHARGE"
)

// --- Subscription Response Types ---

// Plan represents a subscription plan.
type Plan struct {
	PlanID         string              `json:"plan_id"`
	PlanName       string              `json:"plan_name"`
	PlanType       PlanType            `json:"plan_type"`
	Currency       Currency            `json:"currency"`
	AmountMinor    AmountMinor         `json:"amount_minor"`
	MaxAmountMinor AmountMinor         `json:"max_amount_minor"`
	Interval       int32               `json:"interval"`
	IntervalType   PlanIntervalType    `json:"interval_type"`
	MaxCycles      int32               `json:"max_cycles,omitempty"`
	Status         string              `json:"status,omitempty"`
	Note           string              `json:"note,omitempty"`
	Provider       Provider            `json:"provider"`
	Raw            RawProviderResponse `json:"raw_provider_response,omitempty"`
}

// Subscription represents a subscription.
type Subscription struct {
	SubscriptionID         string              `json:"subscription_id"`
	ProviderSubscriptionID string              `json:"provider_subscription_id"`
	PlanID                 string              `json:"plan_id"`
	Status                 SubscriptionStatus  `json:"status"`
	CustomerEmail          string              `json:"customer_email,omitempty"`
	CustomerPhone          string              `json:"customer_phone,omitempty"`
	AuthLink               string              `json:"auth_link,omitempty"`
	ExpiresAt              *time.Time          `json:"expires_at,omitempty"`
	FirstChargeTime        *time.Time          `json:"first_charge_time,omitempty"`
	NextChargeDate         *time.Time          `json:"next_charge_date,omitempty"`
	Provider               Provider            `json:"provider"`
	Raw                    RawProviderResponse `json:"raw_provider_response,omitempty"`
}

// SubscriptionPayment represents a subscription payment.
type SubscriptionPayment struct {
	PaymentID      string                    `json:"payment_id"`
	SubscriptionID string                    `json:"subscription_id"`
	AmountMinor    AmountMinor               `json:"amount_minor"`
	Status         SubscriptionPaymentStatus `json:"status"`
	PaymentType    SubscriptionPaymentType   `json:"payment_type,omitempty"`
	ScheduledDate  *time.Time                `json:"scheduled_date,omitempty"`
	InitiatedDate  *time.Time                `json:"initiated_date,omitempty"`
	RetryAttempts  int                       `json:"retry_attempts,omitempty"`
	Provider       Provider                  `json:"provider"`
	Raw            RawProviderResponse       `json:"raw_provider_response,omitempty"`
}

// --- Subscription Request Types ---

// CreatePlanRequest represents a request to create a new plan.
type CreatePlanRequest struct {
	PlanID         string           `json:"plan_id" pedantigo:"required,minLength=1,maxLength=250"`
	PlanName       string           `json:"plan_name" pedantigo:"required,minLength=1,maxLength=250"`
	PlanType       PlanType         `json:"plan_type" pedantigo:"required,oneof=PERIODIC ON_DEMAND"`
	MaxAmountMinor AmountMinor      `json:"max_amount_minor" pedantigo:"required,gt=0"`
	Currency       Currency         `json:"currency" pedantigo:"required,iso4217"`
	AmountMinor    AmountMinor      `json:"amount_minor,omitempty" pedantigo:"skip_unless=PlanType PERIODIC,required,gt=0"`
	Interval       int32            `json:"interval,omitempty" pedantigo:"skip_unless=PlanType PERIODIC,required,gte=1"`
	IntervalType   PlanIntervalType `json:"interval_type,omitempty" pedantigo:"skip_unless=PlanType PERIODIC,required,oneof=DAY WEEK MONTH YEAR"`
	MaxCycles      int32            `json:"max_cycles,omitempty" pedantigo:"omitempty,gte=0"`
	Note           string           `json:"note,omitempty" pedantigo:"omitempty,maxLength=500"`
}

// Validate enforces presence of the mandatory plan fields (pedantigo's Validate() does not
// enforce the `required` tag), plus the PERIODIC-only conditional fields. ON_DEMAND plans
// legitimately omit AmountMinor/Interval/IntervalType. Currency is always mandatory: providers
// (e.g. Cashfree) require plan_currency and reject a blank value, and the amount conversion to
// major units depends on the currency's ISO-4217 exponent.
func (r *CreatePlanRequest) Validate() error {
	if r.PlanID == "" {
		return errors.New("plan_id is required")
	}
	if r.PlanName == "" {
		return errors.New("plan_name is required")
	}
	if r.PlanType == "" {
		return errors.New("plan_type is required")
	}
	if r.MaxAmountMinor <= 0 {
		return errors.New("max_amount_minor must be greater than 0")
	}
	if r.Currency == "" {
		return errors.New("currency is required")
	}
	if r.PlanType == PlanTypePeriodic {
		if r.AmountMinor <= 0 {
			return errors.New("amount_minor is required and must be greater than 0 for PERIODIC plans")
		}
		if r.Interval < 1 {
			return errors.New("interval must be at least 1 for PERIODIC plans")
		}
		if r.IntervalType == "" {
			return errors.New("interval_type is required for PERIODIC plans")
		}
	}
	return nil
}

// GetPlanRequest represents a request to get a plan.
type GetPlanRequest struct {
	PlanID string `json:"plan_id" pedantigo:"required,minLength=1"`
}

// CreateSubscriptionRequest represents a request to create a new subscription.
type CreateSubscriptionRequest struct {
	SubscriptionID string             `json:"subscription_id" pedantigo:"required,minLength=1,maxLength=250"`
	PlanID         string             `json:"plan_id,omitempty" pedantigo:"omitempty,minLength=1"`
	PlanDetails    *CreatePlanRequest `json:"plan_details,omitempty"`
	// CustomerEmail is optional in the canonical contract but required by Cashfree provider adapter.
	// The adapter enforces this requirement; validation is provider-specific, not checked here.
	CustomerEmail   string            `json:"customer_email,omitempty" pedantigo:"omitempty,email"`
	CustomerPhone   string            `json:"customer_phone" pedantigo:"required,minLength=5,maxLength=20"`
	CustomerName    string            `json:"customer_name,omitempty" pedantigo:"omitempty,maxLength=200"`
	ReturnURL       string            `json:"return_url" pedantigo:"required,url"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"`
	FirstChargeTime *time.Time        `json:"first_charge_time,omitempty"`
	Tags            map[string]string `json:"tags,omitempty" pedantigo:"omitempty,maxItems=10"`
}

// Validate enforces cross-field rules:
// 1. Exactly one of PlanID or PlanDetails must be provided (mutually exclusive XOR)
// 2. ReturnURL must not be empty (required by payment providers)
// 3. When PlanDetails is provided, all nested required fields are validated:
//   - PlanID, PlanName, PlanType, MaxAmountMinor (always required)
//   - For PERIODIC plans: AmountMinor, Interval, IntervalType (required)
//
// 4. CustomerEmail is NOT validated here; Cashfree enforces it separately as a provider-specific requirement
func (r *CreateSubscriptionRequest) Validate() error {
	if r.PlanID == "" && r.PlanDetails == nil {
		return errors.New("exactly one of plan_id or plan_details is required")
	}
	if r.PlanID != "" && r.PlanDetails != nil {
		return errors.New("plan_id and plan_details are mutually exclusive")
	}
	if r.ReturnURL == "" {
		return errors.New("return_url is required and must not be empty")
	}
	// Inline plan details are validated by the canonical CreatePlanRequest.Validate()
	// (single source of truth) rather than re-implementing the rules here.
	if r.PlanDetails != nil {
		if err := r.PlanDetails.Validate(); err != nil {
			return fmt.Errorf("plan_details.%w", err)
		}
	}
	return nil
}

// GetSubscriptionRequest represents a request to get a subscription.
type GetSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" pedantigo:"required,minLength=1"`
}

// CancelSubscriptionRequest represents a request to cancel a subscription.
type CancelSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" pedantigo:"required,minLength=1"`
}

// PauseSubscriptionRequest represents a request to pause a subscription.
type PauseSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" pedantigo:"required,minLength=1"`
}

// ResumeSubscriptionRequest represents a request to resume a subscription.
type ResumeSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" pedantigo:"required,minLength=1"`
}

// ChangePlanRequest represents a request to change the plan of a subscription.
type ChangePlanRequest struct {
	SubscriptionID string           `json:"subscription_id" pedantigo:"required,minLength=1"`
	NewPlanID      string           `json:"new_plan_id" pedantigo:"required,minLength=1"`
	ScheduleAt     ScheduleChangeAt `json:"schedule_at,omitempty" pedantigo:"omitempty,oneof=NOW CYCLE_END"`
}

// GetSubscriptionPaymentsRequest represents a request to get payments for a subscription.
type GetSubscriptionPaymentsRequest struct {
	SubscriptionID string `json:"subscription_id" pedantigo:"required,minLength=1"`
}

// UpgradeStrategy represents the strategy used for upgrading a subscription.
type UpgradeStrategy string

const (
	UpgradeReauthProrated  UpgradeStrategy = "REAUTH_PRORATED"
	UpgradeNativeImmediate UpgradeStrategy = "NATIVE_IMMEDIATE"
	UpgradeCycleEnd        UpgradeStrategy = "CYCLE_END"
)

// UpgradeSubscriptionRequest represents a request to upgrade an existing subscription to a new plan.
type UpgradeSubscriptionRequest struct {
	SubscriptionID    string      `json:"subscription_id" pedantigo:"required,minLength=1"`
	NewSubscriptionID string      `json:"new_subscription_id" pedantigo:"required,minLength=1"`
	CurrentPlanID     string      `json:"current_plan_id" pedantigo:"required,minLength=1"`
	NewPlanID         string      `json:"new_plan_id" pedantigo:"required,minLength=1"`
	OldAmountMinor    AmountMinor `json:"old_amount_minor" pedantigo:"required,gt=0"`
	NewAmountMinor    AmountMinor `json:"new_amount_minor" pedantigo:"required,gt=0"`
	Currency          Currency    `json:"currency" pedantigo:"required,iso4217"`
	RemainingDays     int         `json:"remaining_days" pedantigo:"required,gte=0"`
	CycleDays         int         `json:"cycle_days" pedantigo:"required,gt=0"`
	CustomerEmail     string      `json:"customer_email" pedantigo:"required,email"`
	CustomerPhone     string      `json:"customer_phone" pedantigo:"required,minLength=5,maxLength=20"`
	CustomerName      string      `json:"customer_name" pedantigo:"omitempty,maxLength=200"`
	ReturnURL         string      `json:"return_url" pedantigo:"required,url"`
}

// Validate enforces presence of mandatory fields and cross-field constraints.
func (r *UpgradeSubscriptionRequest) Validate() error {
	if r.SubscriptionID == "" {
		return errors.New("subscription_id is required")
	}
	if r.NewSubscriptionID == "" {
		return errors.New("new_subscription_id is required")
	}
	if r.NewPlanID == "" {
		return errors.New("new_plan_id is required")
	}
	if r.CycleDays <= 0 {
		return errors.New("cycle_days must be > 0")
	}
	if r.RemainingDays < 0 || r.RemainingDays > r.CycleDays {
		return errors.New("remaining_days must be within [0, cycle_days]")
	}
	if r.ReturnURL == "" {
		return errors.New("return_url is required")
	}
	return nil
}

// UpgradeResult represents the result of an upgrade operation.
type UpgradeResult struct {
	Strategy                UpgradeStrategy `json:"strategy"`
	ProratedAmountMinor     AmountMinor     `json:"prorated_amount_minor"`
	RequiresReauthorization bool            `json:"requires_reauthorization"`
	AuthLink                string          `json:"auth_link,omitempty"`
	NewSubscriptionID       string          `json:"new_subscription_id"`
	RecurringEffective      string          `json:"recurring_effective"`
}

// FinalizeUpgradeRequest represents a request to finalize an upgrade operation.
type FinalizeUpgradeRequest struct {
	NewSubscriptionID   string      `json:"new_subscription_id" pedantigo:"required,minLength=1"`
	OldSubscriptionID   string      `json:"old_subscription_id" pedantigo:"required,minLength=1"`
	PaymentRef          string      `json:"payment_ref" pedantigo:"required,minLength=1"`
	ProratedAmountMinor AmountMinor `json:"prorated_amount_minor" pedantigo:"required,gt=0"`
	Currency            Currency    `json:"currency" pedantigo:"required,iso4217"`
}

// Validate enforces presence of mandatory fields.
func (r *FinalizeUpgradeRequest) Validate() error {
	if r.NewSubscriptionID == "" {
		return errors.New("new_subscription_id is required")
	}
	if r.PaymentRef == "" {
		return errors.New("payment_ref is required")
	}
	return nil
}

// ChargeSubscriptionRequest represents a request to perform an on-demand charge on a subscription.
type ChargeSubscriptionRequest struct {
	SubscriptionID string      `json:"subscription_id" pedantigo:"required,minLength=1"`
	PaymentRef     string      `json:"payment_ref" pedantigo:"required,minLength=1"`
	AmountMinor    AmountMinor `json:"amount_minor" pedantigo:"required,gt=0"`
	Currency       Currency    `json:"currency" pedantigo:"required,iso4217"`
	Remarks        string      `json:"remarks,omitempty" pedantigo:"omitempty,maxLength=500"`
}

// Validate enforces presence of mandatory fields.
func (r *ChargeSubscriptionRequest) Validate() error {
	if r.SubscriptionID == "" {
		return errors.New("subscription_id is required")
	}
	if r.PaymentRef == "" {
		return errors.New("payment_ref is required")
	}
	if r.AmountMinor <= 0 {
		return errors.New("amount_minor must be > 0")
	}
	return nil
}
