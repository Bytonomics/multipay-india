import type { ReactNode } from "react";
import type {
  PickerVariant,
  PickerTheme,
  ResolvedTheme,
  Provider,
} from "../core/types";

// Re-export core types for convenience
export type { PickerVariant, PickerTheme, ResolvedTheme, Provider };

/**
 * Configuration for a single provider entry (Cashfree, Razorpay, or PayU)
 */
export interface ProviderEntry {
  label: string;
  description?: string;
  icon?: ReactNode;
  recommended?: boolean;
  visible: boolean;
  enabled: boolean;
  disabledMessage?: string;
}

/**
 * Provider configuration with fixed named fields (one per provider)
 */
export interface PickerProviders {
  cashfree: ProviderEntry;
  razorpay: ProviderEntry;
  payu: ProviderEntry;
}

/**
 * Runtime state for a single provider
 */
export interface ProviderRuntimeState {
  loading: boolean;
  error?: string;
}

/**
 * Runtime state with fixed named fields (one per provider)
 */
export interface PickerRuntimeState {
  cashfree: ProviderRuntimeState;
  razorpay: ProviderRuntimeState;
  payu: ProviderRuntimeState;
}

/**
 * Payment data for the picker
 */
export interface PaymentData {
  /**
   * Amount in minor units (paisa, cents, etc.)
   */
  amountMinor: number;

  /**
   * ISO 4217 currency code (e.g., 'INR', 'USD')
   */
  currency: string;

  /**
   * Provider configurations
   */
  providers: PickerProviders;

  /**
   * Default selected provider (a canonical Provider enum value)
   */
  default: Provider;
}

/**
 * Branding configuration for picker header and footer
 */
export interface PickerBranding {
  logo?: ReactNode;
  title?: string;
  subtitle?: string;
  footerText?: string;
}

/**
 * Visual appearance configuration for the payment picker
 */
export interface PickerAppearance {
  variant?: PickerVariant;
  theme?: PickerTheme;
  primaryColor?: string;
  borderRadius?: number;
  fontFamily?: string;
  branding?: PickerBranding;
  className?: string;
  taxNote?: string;
}

/**
 * Combined view of a provider entry + its runtime state
 */
export interface PickerProviderView {
  id: Provider;
  entry: ProviderEntry;
  state: ProviderRuntimeState;
}

/**
 * Props consumed by all picker variants
 * Single shared props interface for the four visual variants
 */
export interface PickerVariantProps {
  views: PickerProviderView[];
  selected: Provider;
  onSelect: (provider: Provider) => void | Promise<void>;
  theme: ResolvedTheme;
  formattedTotal: string;
  taxNote: string;
}

/**
 * Props for the PaymentPicker component
 */
export interface PaymentPickerProps {
  /**
   * Payment data (amount, currency, providers, default)
   */
  payment: PaymentData;

  /**
   * Visual appearance configuration
   */
  appearance?: PickerAppearance;

  /**
   * Selection callback (emits canonical Provider enum values)
   */
  onSelect: (provider: Provider) => void | Promise<void>;
}

/**
 * Control methods for programmatic picker control
 */
export interface PickerControls {
  /**
   * Select a provider programmatically
   */
  selectProvider: (providerId: Provider) => void;

  /**
   * Get the currently selected provider
   */
  getSelectedProvider: () => Provider;

  /**
   * Check if a specific provider is currently selected
   */
  isSelected: (providerId: Provider) => boolean;

  /**
   * Enable or disable a specific provider
   */
  setProviderDisabled: (
    providerId: Provider,
    disabled: boolean,
    reason?: string,
  ) => void;

  /**
   * Focus the picker component
   */
  focus: () => void;

  /**
   * Blur the picker component
   */
  blur: () => void;
}
