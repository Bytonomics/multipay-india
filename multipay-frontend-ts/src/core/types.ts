/**
 * Payment providers supported by MultiPay India
 */
export enum Provider {
  CASHFREE = "cashfree",
  RAZORPAY = "razorpay",
  PAYU = "payu",
}

/**
 * Environment configuration for payment providers
 */
export enum Environment {
  SANDBOX = "SANDBOX",
  PRODUCTION = "PRODUCTION",
}

/**
 * Cashfree JS SDK mode configuration
 */
export enum CashfreeMode {
  PRODUCTION = "production",
  SANDBOX = "sandbox",
}

/**
 * Picker visual variants
 */
export enum PickerVariant {
  DYNAMIC_STACK = "dynamic-stack",
  INTERACTIVE_MATRIX = "interactive-matrix",
  SECURE_VAULT = "secure-vault",
  NEUMORPHIC_FLOW = "neumorphic-flow",
}

/**
 * Picker theme selection (user input, can be auto)
 */
export enum PickerTheme {
  LIGHT = "light",
  DARK = "dark",
  AUTO = "auto",
}

/**
 * Resolved theme (the actual applied data-theme value, never AUTO)
 */
export enum ResolvedTheme {
  LIGHT = "light",
  DARK = "dark",
}

/**
 * Cashfree-specific checkout payload
 */
export interface CashfreeCheckoutPayload {
  provider: Provider.CASHFREE;
  environment: Environment;
  session_id: string;
}

/**
 * Razorpay-specific checkout payload
 */
export interface RazorpayCheckoutPayload {
  provider: Provider.RAZORPAY;
  environment: Environment;
  order_id: string;
  public_key: string;
  callback_url: string;
  amount_minor: number;
  currency: string;
}

/**
 * Checkout payload union type - supports multiple provider-specific formats
 */
export type CheckoutPayload = CashfreeCheckoutPayload | RazorpayCheckoutPayload;

// --- Subscription mandate-authorization payloads (mirror of CheckoutPayload) ---
// The consumer builds one of these from the backend subscription response and calls
// MultiPay.authorizeSubscription(). All provider differences (JS-SDK session vs redirect URL)
// are hidden inside the library.
export interface CashfreeSubscriptionAuthorizationPayload {
  provider: Provider.CASHFREE;
  environment: Environment;
  auth_session_id: string;
}

export interface RazorpaySubscriptionAuthorizationPayload {
  provider: Provider.RAZORPAY;
  environment: Environment;
  auth_link: string;
}

export type SubscriptionAuthorizationPayload =
  | CashfreeSubscriptionAuthorizationPayload
  | RazorpaySubscriptionAuthorizationPayload;

/**
 * Razorpay form fields for POST-based redirect
 */
export interface RazorpayFormFields {
  key_id: string;
  order_id: string;
  amount: string;
  currency: string;
  callback_url: string;
}

/**
 * Upgrade strategy for subscription upgrades
 */
export enum UpgradeStrategy {
  REAUTH_PRORATED = "REAUTH_PRORATED",
  NATIVE_IMMEDIATE = "NATIVE_IMMEDIATE",
  CYCLE_END = "CYCLE_END",
}

/**
 * When recurring charges should take effect
 */
export enum RecurringEffective {
  IMMEDIATE = "IMMEDIATE",
  CYCLE_END = "CYCLE_END",
}

/**
 * Interval unit for periodic plans. Wire values must match the Go port exactly.
 */
export enum PlanIntervalType {
  DAY = "DAY",
  WEEK = "WEEK",
  MONTH = "MONTH",
  YEAR = "YEAR",
}

/**
 * Request to upgrade an existing subscription to a new plan
 */
export interface UpgradeSubscriptionRequest {
  subscription_id: string;
  new_subscription_id: string;
  current_plan_id: string;
  new_plan_id: string;
  old_amount_minor: number;
  new_amount_minor: number;
  currency: string;
  remaining_days: number;
  cycle_days: number;
  customer_email: string;
  customer_phone: string;
  customer_name?: string;
  return_url: string;
  cross_cycle?: boolean;
  /**
   * The NEW plan's recurring cadence. REQUIRED when cross_cycle is true: a cross-cycle upgrade charges a
   * full new-cycle amount up front, so the new mandate's first auto-charge must be one NEW interval out
   * rather than at the end of the old cycle. Ignored when cross_cycle is false.
   */
  new_recurring_interval?: number;
  new_recurring_interval_type?: PlanIntervalType;
}

/**
 * Result of an upgrade operation
 */
export interface UpgradeResult {
  strategy: UpgradeStrategy;
  prorated_amount_minor: number;
  requires_reauthorization: boolean;
  auth_link?: string;
  new_subscription_id: string;
  recurring_effective: RecurringEffective;
}

/**
 * Request to finalize an upgrade operation
 */
export interface FinalizeUpgradeRequest {
  new_subscription_id: string;
  old_subscription_id: string;
  /**
   * payment_ref, prorated_amount_minor and currency are required ONLY when the library must raise the
   * proration charge itself, i.e. when proration_collected_externally is false or omitted.
   */
  payment_ref?: string;
  prorated_amount_minor?: number;
  currency?: string;
  /**
   * Declares that the caller already collected the prorated delta out-of-band (for example a one-time
   * Order settled through the orders API). When true the library performs ONLY the provider transition —
   * it does NOT raise a charge on the new mandate. Omitting it preserves the historical
   * charge-then-cancel behaviour.
   */
  proration_collected_externally?: boolean;
}

/**
 * Request to perform an on-demand charge on a subscription
 */
export interface ChargeSubscriptionRequest {
  subscription_id: string;
  payment_ref: string;
  amount_minor: number;
  currency: string;
  remarks?: string;
  /** Future-dated charge (Cashfree payment_schedule_date, date-only "YYYY-MM-DD"). Cashfree-only. */
  payment_schedule_date?: string;
  /** Addon item description (Razorpay item.description). Razorpay-only. */
  description?: string;
}

/**
 * Plan change kind represents the type of plan change operation
 */
export type PlanChangeKind =
  | "CREATE"
  | "UPGRADE_SAME_CYCLE"
  | "UPGRADE_CROSS_CYCLE"
  | "DOWNGRADE";

/**
 * Plan change quote represents the cost breakdown for a plan change
 */
export interface PlanChangeQuote {
  kind: PlanChangeKind;
  charge_now_minor: number;
  prorated_credit_minor: number;
  new_recurring_minor: number;
  recurring_effective: string;
}
