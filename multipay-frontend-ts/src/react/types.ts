import type { ReactNode } from "react";
import type {
  PickerProviderId,
  PickerVariant,
  PickerTheme,
  Provider,
} from "../core/types";

// Re-export core types for convenience
export type { PickerProviderId, PickerVariant, PickerTheme, Provider };

/**
 * Branding configuration for picker header and footer
 * React-specific type that lives in the react module
 */
export interface PickerBranding {
  logo?: ReactNode;
  title?: string;
  subtitle?: string;
  footerText?: string;
}

/**
 * Provider option for the payment picker UI
 * React-specific type with icon support
 */
export interface ProviderOption {
  id: PickerProviderId;
  label: string;
  description?: string;
  icon?: ReactNode;
  recommended?: boolean;
  enabled: boolean;
  disabledReason?: string;
}

/**
 * Payment data for the picker
 * React-specific type with provider options
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
   * Available provider options
   */
  providers: ProviderOption[];

  /**
   * Default selected provider ID
   */
  defaultSelected?: PickerProviderId;
}

/**
 * Visual appearance configuration for the payment picker
 * React-specific type with branding
 */
export interface PickerAppearance {
  variant?: PickerVariant;
  theme?: PickerTheme;
  primaryColor?: string;
  borderRadius?: number;
  fontFamily?: string;
  customStyles?: Record<string, string>;
  branding?: PickerBranding;
  className?: string;
  taxNote?: string;
}

/**
 * Props for the PaymentPicker component
 * React-specific component props
 */
export interface PaymentPickerProps {
  /**
   * Payment data (amount, currency, providers)
   */
  payment: PaymentData;

  /**
   * Visual appearance configuration
   */
  appearance?: PickerAppearance;

  /**
   * Selection callback (only emits canonical Provider types)
   */
  onSelect: (_provider: Provider) => void | Promise<void>;
}

/**
 * Control methods for programmatic picker control
 * React-specific imperative API
 */
export interface PickerControls {
  /**
   * Select a provider programmatically
   */
  selectProvider: (_providerId: PickerProviderId) => void;

  /**
   * Get the currently selected provider
   */
  getSelectedProvider: () => PickerProviderId;

  /**
   * Check if a specific provider is currently selected
   */
  isSelected: (_providerId: PickerProviderId) => boolean;

  /**
   * Enable or disable a specific provider option
   */
  setProviderDisabled: (
    _providerId: PickerProviderId,
    _disabled: boolean,
    _reason?: string,
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
