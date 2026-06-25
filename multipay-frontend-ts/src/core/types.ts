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
  order_id: string;
  session_id: string;
  environment: Environment;
  amount: number;
  currency: string;
  customer_id?: string;
  customer_phone?: string;
  customer_email?: string;
}

/**
 * Razorpay-specific checkout payload
 */
export interface RazorpayCheckoutPayload {
  provider: Provider.RAZORPAY;
  order_id: string;
  key_id: string;
  public_key: string;
  amount_minor: number;
  currency: string;
  environment: Environment;
  customer_id?: string;
  customer_phone?: string;
  customer_email?: string;
  callback_url?: string;
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
