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
	SubscriptionID         string             `json:"subscription_id"`
	ProviderSubscriptionID string             `json:"provider_subscription_id"`
	PlanID                 string             `json:"plan_id"`
	Status                 SubscriptionStatus `json:"status"`
	CustomerEmail          string             `json:"customer_email,omitempty"`
	CustomerPhone          string             `json:"customer_phone,omitempty"`
	AuthLink               string             `json:"auth_link,omitempty"`
	// AuthSessionID is an opaque provider SDK-auth session handle for the mandate-authorization step
	// (Cashfree: subscription_session_id, consumed by the Cashfree JS SDK subscriptionsCheckout).
	// Empty for providers that return a redirect URL in AuthLink instead (e.g. Razorpay).
	AuthSessionID string `json:"auth_session_id,omitempty"`
	// Environment is the client environment (SANDBOX/PRODUCTION) the frontend SDK must initialize with.
	Environment     Environment         `json:"environment,omitempty"`
	ExpiresAt       *time.Time          `json:"expires_at,omitempty"`
	FirstChargeTime *time.Time          `json:"first_charge_time,omitempty"`
	NextChargeDate  *time.Time          `json:"next_charge_date,omitempty"`
	Provider        Provider            `json:"provider"`
	Raw             RawProviderResponse `json:"raw_provider_response,omitempty"`
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
	PlanID         string           `json:"plan_id" validate:"required,minLength=1,maxLength=250"`
	PlanName       string           `json:"plan_name" validate:"required,minLength=1,maxLength=250"`
	PlanType       PlanType         `json:"plan_type" validate:"required,oneof=PERIODIC ON_DEMAND"`
	MaxAmountMinor AmountMinor      `json:"max_amount_minor" validate:"required,gt=0"`
	Currency       Currency         `json:"currency" validate:"required,iso4217"`
	AmountMinor    AmountMinor      `json:"amount_minor,omitempty" validate:"skip_unless=PlanType PERIODIC,required,gt=0"`
	Interval       int32            `json:"interval,omitempty" validate:"skip_unless=PlanType PERIODIC,required,gte=1"`
	IntervalType   PlanIntervalType `json:"interval_type,omitempty" validate:"skip_unless=PlanType PERIODIC,required,oneof=DAY WEEK MONTH YEAR"`
	MaxCycles      int32            `json:"max_cycles,omitempty" validate:"omitempty,gte=0"`
	Note           string           `json:"note,omitempty" validate:"omitempty,maxLength=500"`
	// Description populates the Razorpay plan item description (item.description). Cashfree has
	// no per-plan description field (it uses PlanNote, mapped from Note). Optional.
	Description string `json:"description,omitempty" validate:"omitempty,maxLength=500"`
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
	PlanID string `json:"plan_id" validate:"required,minLength=1"`
}

// CreateSubscriptionRequest represents a request to create a new subscription.
type CreateSubscriptionRequest struct {
	SubscriptionID string             `json:"subscription_id" validate:"required,minLength=1,maxLength=250"`
	PlanID         string             `json:"plan_id,omitempty" validate:"omitempty,minLength=1"`
	PlanDetails    *CreatePlanRequest `json:"plan_details,omitempty"`
	// CustomerEmail is optional in the canonical contract but required by Cashfree provider adapter.
	// The adapter enforces this requirement; validation is provider-specific, not checked here.
	CustomerEmail   string     `json:"customer_email,omitempty" validate:"omitempty,email"`
	CustomerPhone   string     `json:"customer_phone" validate:"required,minLength=5,maxLength=20"`
	CustomerName    string     `json:"customer_name,omitempty" validate:"omitempty,maxLength=200"`
	ReturnURL       string     `json:"return_url" validate:"required,url"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	FirstChargeTime *time.Time `json:"first_charge_time,omitempty"`
	// FirstChargeWithMandate: collect the first billing period immediately at signup
	// (right after mandate authorization). Provider-neutral; each adapter maps it to its own mechanism.
	FirstChargeWithMandate bool `json:"first_charge_with_mandate,omitempty"`

	// RecurringAmountMinor/RecurringInterval/RecurringIntervalType: recurring-schedule hint used by
	// adapters that must compute the first-charge/start offset when FirstChargeWithMandate is true and an
	// existing PlanID is used. Cashfree ignores these (its plan drives the schedule); Razorpay needs them
	// (addon amount + start_at = now+interval).
	RecurringAmountMinor  AmountMinor      `json:"recurring_amount_minor,omitempty"  validate:"omitempty,gte=0"`
	RecurringInterval     int32            `json:"recurring_interval,omitempty"      validate:"omitempty,gte=1"`
	RecurringIntervalType PlanIntervalType `json:"recurring_interval_type,omitempty" validate:"omitempty,oneof=DAY WEEK MONTH YEAR"`
	// RecurringCurrency is the ISO-4217 currency for the recurring charge and the first-period addon.
	// It is threaded on EVERY create-subscription request (both Cashfree and Razorpay) to keep the two
	// provider flows in sync. Cashfree ignores it (the provider plan drives currency); Razorpay uses it
	// as the currency of the first-period addon when FirstChargeWithMandate is true.
	RecurringCurrency Currency          `json:"recurring_currency,omitempty" validate:"omitempty,iso4217"`
	Tags              map[string]string `json:"tags,omitempty" validate:"omitempty,maxItems=10"`

	// TotalCount is the number of billing cycles the customer will be charged. Razorpay
	// treats this as MANDATORY (unless end_at is supplied); the Razorpay adapter sends it
	// UNCONDITIONALLY, preferring this field and falling back to PlanDetails.MaxCycles. When
	// zero and no inline plan bounds cycles, the mapper omits it. Cashfree ignores it (the
	// plan's own max_cycles governs cycle count). Optional in the canonical contract.
	TotalCount int32 `json:"total_count,omitempty" validate:"omitempty,gte=0"`
	// Quantity multiplies the plan amount per invoice (Razorpay-only, defaults to 1).
	Quantity int32 `json:"quantity,omitempty" validate:"omitempty,gte=1"`
	// OfferID links a Razorpay offer to the subscription (Razorpay-only).
	OfferID string `json:"offer_id,omitempty" validate:"omitempty,maxLength=250"`
	// Addons are items collected upfront during authorization (Razorpay-only).
	Addons []SubscriptionAddon `json:"addons,omitempty"`

	// AuthorizationDetails controls the Cashfree mandate-authorization step (Cashfree-only).
	AuthorizationDetails *SubscriptionAuthorizationDetails `json:"authorization_details,omitempty"`
	// Meta carries Cashfree subscription_meta extras beyond ReturnURL (Cashfree-only).
	Meta *SubscriptionMeta `json:"subscription_meta,omitempty"`
	// BankDetails pre-binds the customer's bank account for a TPV eNACH mandate (Cashfree-only).
	BankDetails *SubscriptionBankDetails `json:"customer_bank_details,omitempty"`
	// PaymentSplits configure Cashfree Easy Split for the subscription (Cashfree-only).
	PaymentSplits []SubscriptionPaymentSplit `json:"subscription_payment_splits,omitempty"`

	// CustomerNotify controls whether Razorpay notifies the customer (emails/SMSes the auth link).
	// Razorpay-only. The adapter forwards exactly what the caller sets and imposes NO default:
	// nil ⇒ customer_notify is omitted (Razorpay applies its own default). The caller (e.g. the
	// cloud) is responsible for choosing the value. Cashfree ignores it.
	CustomerNotify *bool `json:"customer_notify,omitempty"`
	// CfOrderID forwards Cashfree's cf_order_id on the create-subscription request (present on the
	// vendored Cashfree SDK struct; attaches the subscription to a pre-created Cashfree order).
	// Cashfree-only; ignored by Razorpay. Empty ⇒ not sent.
	CfOrderID string `json:"cf_order_id,omitempty" validate:"omitempty,maxLength=250"`
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
	if r.FirstChargeWithMandate && r.FirstChargeTime != nil {
		return errors.New("first_charge_with_mandate and first_charge_time are mutually exclusive")
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
	SubscriptionID string `json:"subscription_id" validate:"required,minLength=1"`
}

// CancelSubscriptionRequest represents a request to cancel a subscription.
type CancelSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" validate:"required,minLength=1"`
	// CancelAtCycleEnd is Razorpay's cancel_at_cycle_end (Boolean per Razorpay docs):
	// nil/false = cancel immediately, true = cancel at the end of the current billing cycle.
	// Razorpay-only — Cashfree's Manage Subscription CANCEL is immediate-only and ignores it.
	CancelAtCycleEnd *bool `json:"cancel_at_cycle_end,omitempty"`
}

// PauseSubscriptionRequest represents a request to pause a subscription.
type PauseSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" validate:"required,minLength=1"`
	// PauseAt is Razorpay's pause_at. The ONLY value Razorpay accepts is "now" (pause immediately);
	// Validate() rejects anything else. Razorpay-only — Cashfree pause is action-based and ignores it.
	PauseAt string `json:"pause_at,omitempty" validate:"omitempty"`
}

// Validate rejects any pause_at value other than "now" (the only value the vendor accepts).
func (r *PauseSubscriptionRequest) Validate() error {
	if r.PauseAt != "" && r.PauseAt != "now" {
		return fmt.Errorf("pause_at must be \"now\" if set, got %q", r.PauseAt)
	}
	return nil
}

// ResumeSubscriptionRequest represents a request to resume a subscription.
type ResumeSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id" validate:"required,minLength=1"`
	// NextScheduledTime is REQUIRED by Cashfree's Manage Subscription ACTIVATE action
	// (action_details.next_scheduled_time) — it is the date the resumed subscription's next
	// charge is scheduled. Cashfree rejects an ACTIVATE with no action_details. Razorpay's
	// Resume has no equivalent (it resumes on its own schedule), so the Razorpay adapter
	// ignores this field. Optional in the canonical contract because it is Cashfree-specific;
	// the Cashfree adapter forwards it when set.
	NextScheduledTime *time.Time `json:"next_scheduled_time,omitempty"`
	// ResumeAt is Razorpay's resume_at. The ONLY value Razorpay accepts is "now" (resume immediately);
	// Validate() rejects anything else. Razorpay-only — Cashfree resume is action-based and ignores it.
	ResumeAt string `json:"resume_at,omitempty" validate:"omitempty"`
}

// Validate rejects any resume_at value other than "now" (the only value the vendor accepts).
func (r *ResumeSubscriptionRequest) Validate() error {
	if r.ResumeAt != "" && r.ResumeAt != "now" {
		return fmt.Errorf("resume_at must be \"now\" if set, got %q", r.ResumeAt)
	}
	return nil
}

// ChangePlanRequest represents a request to change the plan of a subscription.
type ChangePlanRequest struct {
	SubscriptionID string           `json:"subscription_id" validate:"required,minLength=1"`
	NewPlanID      string           `json:"new_plan_id" validate:"required,minLength=1"`
	ScheduleAt     ScheduleChangeAt `json:"schedule_at,omitempty" validate:"omitempty,oneof=NOW CYCLE_END"`
	// The fields below are Razorpay Update-Subscription optional params. Cashfree's Manage
	// Subscription CHANGE_PLAN action supports only plan_id, so the Cashfree adapter ignores them.
	// OfferID links a Razorpay offer to the subscription.
	OfferID string `json:"offer_id,omitempty" validate:"omitempty,maxLength=250"`
	// Quantity multiplies the plan charge per invoice (Razorpay, defaults to 1).
	Quantity int32 `json:"quantity,omitempty" validate:"omitempty,gte=1"`
	// RemainingCount updates the subscription's total_count (Razorpay).
	RemainingCount int32 `json:"remaining_count,omitempty" validate:"omitempty,gte=0"`
	// StartAt is the new start date for the subscription (Razorpay, Unix seconds).
	StartAt *time.Time `json:"start_at,omitempty"`
	// CustomerNotify controls whether Razorpay notifies the customer (nil = use Razorpay default).
	CustomerNotify *bool `json:"customer_notify,omitempty"`
}

// GetSubscriptionPaymentsRequest represents a request to get payments for a subscription.
type GetSubscriptionPaymentsRequest struct {
	SubscriptionID string `json:"subscription_id" validate:"required,minLength=1"`
}

// UpgradeStrategy represents the strategy used for upgrading a subscription.
type UpgradeStrategy string

const (
	UpgradeReauthProrated  UpgradeStrategy = "REAUTH_PRORATED"
	UpgradeNativeImmediate UpgradeStrategy = "NATIVE_IMMEDIATE"
	UpgradeCycleEnd        UpgradeStrategy = "CYCLE_END"
)

// PlanChangeKind represents the type of plan change operation.
type PlanChangeKind string

const (
	PlanChangeKindCreate           PlanChangeKind = "CREATE"
	PlanChangeKindUpgradeSameCycle PlanChangeKind = "UPGRADE_SAME_CYCLE"
	PlanChangeKindUpgradeCross     PlanChangeKind = "UPGRADE_CROSS_CYCLE"
	PlanChangeKindDowngrade        PlanChangeKind = "DOWNGRADE"
)

// PlanChangePreviewRequest is the PURE, generic input to PreviewPlanChange. The CALLER (e.g. the
// studio control plane) decides the Kind and passes already-resolved amounts in minor units. This
// type carries NO plan keys, no cycle strings, no DB concepts — only the money-math primitives.
type PlanChangePreviewRequest struct {
	Kind               PlanChangeKind `json:"kind" validate:"required,oneof=CREATE UPGRADE_SAME_CYCLE UPGRADE_CROSS_CYCLE DOWNGRADE"`
	CurrentAmountMinor int64          `json:"current_amount_minor" validate:"gte=0"`
	NewAmountMinor     int64          `json:"new_amount_minor" validate:"gte=0"`
	RemainingDays      int64          `json:"remaining_days" validate:"gte=0"`
	CurrentCycleDays   int64          `json:"current_cycle_days" validate:"gte=0"`
	Currency           string         `json:"currency" validate:"omitempty"`
}

// Validate enforces the money-math preconditions.
func (r *PlanChangePreviewRequest) Validate() error {
	switch r.Kind {
	case PlanChangeKindCreate, PlanChangeKindUpgradeSameCycle, PlanChangeKindUpgradeCross, PlanChangeKindDowngrade:
	default:
		return fmt.Errorf("invalid plan change kind %q", r.Kind)
	}
	if r.RemainingDays < 0 {
		return errors.New("remaining_days must be >= 0")
	}
	if r.CurrentCycleDays > 0 && r.RemainingDays > r.CurrentCycleDays {
		return errors.New("remaining_days must be within [0, current_cycle_days]")
	}
	return nil
}

// PlanChangeQuote is the PURE money breakdown returned by PreviewPlanChange. All amounts are minor
// units (int64). RecurringEffective is "IMMEDIATE" (new price starts now) or "CYCLE_END" (new price
// starts at the next renewal). No dates/names/phone — the caller derives those.
type PlanChangeQuote struct {
	Kind                PlanChangeKind `json:"kind"`
	ChargeNowMinor      int64          `json:"charge_now_minor"`
	ProratedCreditMinor int64          `json:"prorated_credit_minor"`
	NewRecurringMinor   int64          `json:"new_recurring_minor"`
	RecurringEffective  string         `json:"recurring_effective"`
}

// UpgradeSubscriptionRequest represents a request to upgrade an existing subscription to a new plan.
type UpgradeSubscriptionRequest struct {
	SubscriptionID    string      `json:"subscription_id" validate:"required,minLength=1"`
	NewSubscriptionID string      `json:"new_subscription_id" validate:"required,minLength=1"`
	CurrentPlanID     string      `json:"current_plan_id" validate:"required,minLength=1"`
	NewPlanID         string      `json:"new_plan_id" validate:"required,minLength=1"`
	OldAmountMinor    AmountMinor `json:"old_amount_minor" validate:"required,gt=0"`
	NewAmountMinor    AmountMinor `json:"new_amount_minor" validate:"required,gt=0"`
	Currency          Currency    `json:"currency" validate:"required,iso4217"`
	RemainingDays     int         `json:"remaining_days" validate:"required,gte=0"`
	CycleDays         int         `json:"cycle_days" validate:"required,gt=0"`
	CustomerEmail     string      `json:"customer_email" validate:"required,email"`
	CustomerPhone     string      `json:"customer_phone" validate:"required,minLength=5,maxLength=20"`
	CustomerName      string      `json:"customer_name" validate:"omitempty,maxLength=200"`
	ReturnURL         string      `json:"return_url" validate:"required,url"`
	CrossCycle        bool        `json:"cross_cycle"`
	// NewRecurringInterval and NewRecurringIntervalType describe the NEW plan's recurring cadence.
	// They are REQUIRED when CrossCycle is true: a cross-cycle upgrade charges a full new-cycle amount
	// up front, so the new mandate's first auto-charge must be one NEW interval out, not the remainder
	// of the old cycle. Ignored when CrossCycle is false.
	NewRecurringInterval     int32            `json:"new_recurring_interval" validate:"omitempty,gte=1"`
	NewRecurringIntervalType PlanIntervalType `json:"new_recurring_interval_type" validate:"omitempty,oneof=DAY WEEK MONTH YEAR"`
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
	if r.CrossCycle {
		if r.NewRecurringInterval < 1 {
			return errors.New("new_recurring_interval is required and must be >= 1 when cross_cycle is true")
		}
		if r.NewRecurringIntervalType == "" {
			return errors.New("new_recurring_interval_type is required when cross_cycle is true")
		}
	}
	return nil
}

// UpgradeResult represents the result of an upgrade operation.
type UpgradeResult struct {
	Strategy                UpgradeStrategy `json:"strategy"`
	ProratedAmountMinor     AmountMinor     `json:"prorated_amount_minor"`
	RequiresReauthorization bool            `json:"requires_reauthorization"`
	AuthLink                string          `json:"auth_link,omitempty"`
	AuthSessionID           string          `json:"auth_session_id,omitempty"`
	Environment             Environment     `json:"environment,omitempty"`
	NewSubscriptionID       string          `json:"new_subscription_id"`
	RecurringEffective      string          `json:"recurring_effective"`
}

// FinalizeUpgradeRequest represents a request to finalize an upgrade operation.
type FinalizeUpgradeRequest struct {
	NewSubscriptionID string `json:"new_subscription_id" validate:"required,minLength=1"`
	OldSubscriptionID string `json:"old_subscription_id" validate:"required,minLength=1"`
	// PaymentRef, ProratedAmountMinor and Currency are required ONLY when the library must raise the
	// proration charge itself, i.e. when ProrationCollectedExternally is false. The tags are omitempty
	// because the requirement is conditional; presence is enforced in Validate().
	PaymentRef          string      `json:"payment_ref" validate:"omitempty,minLength=1"`
	ProratedAmountMinor AmountMinor `json:"prorated_amount_minor" validate:"omitempty,gte=0"`
	Currency            Currency    `json:"currency" validate:"omitempty,iso4217"`
	// ProrationCollectedExternally declares that the caller already collected the prorated delta
	// out-of-band (for example a one-time Order settled through Orders().CreateOrder). When true the
	// library performs ONLY the provider transition — it does NOT raise a charge on the new mandate.
	// The default (false) preserves the historical charge-then-cancel behaviour.
	ProrationCollectedExternally bool `json:"proration_collected_externally"`
}

// Validate enforces presence of mandatory fields and the conditional proration fields.
func (r *FinalizeUpgradeRequest) Validate() error {
	if r.NewSubscriptionID == "" {
		return errors.New("new_subscription_id is required")
	}
	if r.OldSubscriptionID == "" {
		return errors.New("old_subscription_id is required")
	}
	if !r.ProrationCollectedExternally {
		if r.PaymentRef == "" {
			return errors.New("payment_ref is required when proration is not collected externally")
		}
		if r.ProratedAmountMinor <= 0 {
			return errors.New("prorated_amount_minor must be > 0 when proration is not collected externally")
		}
		if r.Currency == "" {
			return errors.New("currency is required when proration is not collected externally")
		}
	}
	return nil
}

// ChargeSubscriptionRequest represents a request to perform an on-demand charge on a subscription.
type ChargeSubscriptionRequest struct {
	SubscriptionID string      `json:"subscription_id" validate:"required,minLength=1"`
	PaymentRef     string      `json:"payment_ref" validate:"required,minLength=1"`
	AmountMinor    AmountMinor `json:"amount_minor" validate:"required,gt=0"`
	Currency       Currency    `json:"currency" validate:"required,iso4217"`
	Remarks        string      `json:"remarks,omitempty" validate:"omitempty,maxLength=500"`
	// PaymentScheduleDate future-dates the charge (Cashfree payment_schedule_date, date-only
	// "YYYY-MM-DD"). When zero, Cashfree charges immediately. Cashfree-only; Razorpay's
	// CreateAddon has no scheduled-date concept.
	PaymentScheduleDate *time.Time `json:"payment_schedule_date,omitempty"`
	// Description populates the addon item description on Razorpay (item.description).
	// Cashfree uses PaymentRemarks (mapped from Remarks) instead.
	Description string `json:"description,omitempty" validate:"omitempty,maxLength=500"`
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
