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
  payment_ref: string;
  prorated_amount_minor: number;
  currency: string;
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
}
