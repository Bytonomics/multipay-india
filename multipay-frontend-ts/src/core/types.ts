/**
 * Payment providers supported by MultiPay India
 */
export enum Provider {
  CASHFREE = "cashfree",
  RAZORPAY = "razorpay",
}

/**
 * Environment configuration for payment providers
 */
export enum Environment {
  SANDBOX = "SANDBOX",
  PRODUCTION = "PRODUCTION",
}

/**
 * Payment providers supported by MultiPay India (type alias for backward compatibility)
 */
export type ProviderType = `${Provider}`;

/**
 * Environment configuration for payment providers (type alias for backward compatibility)
 */
export type EnvironmentType = `${Environment}`;

/**
 * Provider status for UI components
 */
export type ProviderStatus = "available" | "loading" | "error" | "disabled";

/**
 * Payment provider for UI components (alias for backward compatibility)
 */
export type PaymentProvider = Provider;

/**
 * Picker provider identifiers - maps to internal provider selection
 */
export type PickerProviderId = Provider | "multipay_default";

/**
 * Picker visual variants
 */
export type PickerVariant =
  | "dynamic-stack"
  | "interactive-matrix"
  | "secure-vault"
  | "neumorphic-flow";

/**
 * Picker theme options
 */
export type PickerTheme = "light" | "dark" | "auto";

/**
 * Checkout payload union type - supports multiple provider-specific formats
 */
export type CheckoutPayload =
  | {
      provider: Provider.CASHFREE;
      order_id: string;
      session_id: string;
      environment: Environment;
      amount: number;
      currency: string;
      customer_id?: string;
      customer_phone?: string;
      customer_email?: string;
      metadata?: Record<string, string>;
    }
  | {
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
      metadata?: Record<string, string>;
    };

/**
 * Cashfree-specific checkout payload type
 */
export type CashfreeCheckoutPayload = Extract<
  CheckoutPayload,
  { provider: Provider.CASHFREE }
>;

/**
 * Razorpay-specific checkout payload type
 */
export type RazorpayCheckoutPayload = Extract<
  CheckoutPayload,
  { provider: Provider.RAZORPAY }
>;

/**
 * Payment data returned after successful checkout
 */
export interface CheckoutResultData {
  provider: Provider;
  transaction_id: string;
  order_id: string;
  amount: number;
  currency: string;
  status: "success" | "failed" | "pending";
  metadata?: Record<string, string>;
  timestamp: string;
}
