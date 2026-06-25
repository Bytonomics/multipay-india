import React from "react";
import type { ProviderOption, PickerProviderId } from "../types";
import { Provider } from "../../core/types";
import styles from "./neumorphic-flow.module.css";

export interface NeumorphicFlowProps {
  providers: ProviderOption[];
  selected?: PickerProviderId;
  onSelect: (_provider: PickerProviderId) => void;
  loadingRecord?: Record<Provider, boolean>;
  errorRecord?: Record<Provider, string | undefined>;
  formattedTotal: string;
  taxNote: string;
}

export const NeumorphicFlow: React.FC<NeumorphicFlowProps> = ({
  providers,
  selected,
  onSelect,
  loadingRecord = {} as Record<Provider, boolean>,
  errorRecord = {} as Record<Provider, string | undefined>,
  formattedTotal,
  taxNote,
}) => {
  // Filter providers: only show aggregators (Cashfree, Razorpay)
  const visibleProviders = providers.filter(
    (p) => (p.id === Provider.CASHFREE || p.id === Provider.RAZORPAY) && p.enabled,
  );

  if (visibleProviders.length === 0) {
    return null;
  }

  // Default to first provider if none selected
  const activeProvider = selected || visibleProviders[0].id;
  const activeProviderKey = activeProvider as Provider;
  const isLoading = loadingRecord[activeProviderKey] || false;
  const hasError = !!errorRecord[activeProviderKey];

  const handleSelect = (providerId: PickerProviderId): void => {
    if (!isLoading) {
      onSelect(providerId);
    }
  };

  return (
    <div className={styles.neumorphicFlow} data-theme="light">
      {/* Recessed Total Pill */}
      <div className={styles.totalPill}>
        <span className={styles.totalLabel}>Total</span>
        <span className={styles.totalAmount}>{formattedTotal}</span>
      </div>

      {/* One-Tap Pay Primary Action */}
      <button
        type="button"
        className={styles.payButton}
        disabled={isLoading || hasError}
        onClick={() => handleSelect(activeProvider)}
      >
        <span className={styles.payButtonText}>
          {isLoading ? "Processing..." : hasError ? "Retry" : "One-Tap  Pay"}
        </span>
        {!isLoading && !hasError && (
          <span className={styles.payButtonArrow}>→</span>
        )}
      </button>

      {/* Recessed Segmented Toggle for Aggregator Selection */}
      <div
        className={styles.toggleContainer}
        role="radiogroup"
        aria-label="Select payment provider"
      >
        <div className={styles.toggleLabel}>pay via</div>
        <div className={styles.toggleSegments}>
          {visibleProviders.map((provider) => {
            const providerId = provider.id;
            const isSelected = providerId === activeProvider;
            const isDisabled = !provider.enabled;

            return (
              <button
                key={providerId}
                type="button"
                className={`${styles.segment} ${isSelected ? styles.segmentActive : ""} ${isDisabled ? styles.segmentDisabled : ""}`}
                disabled={isDisabled || isLoading}
                onClick={() => handleSelect(providerId)}
                role="radio"
                aria-checked={isSelected}
                aria-disabled={isDisabled}
                tabIndex={isSelected ? 0 : -1}
              >
                <span className={styles.segmentContent}>
                  {provider.icon && (
                    <span className={styles.segmentIcon}>
                      {provider.icon as React.ReactNode}
                    </span>
                  )}
                  <span className={styles.segmentLabel}>{provider.label}</span>
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Tax Disclaimer */}
      <div className={styles.taxDisclaimer}>
        {taxNote || "Final taxes added at checkout (vendor)"}
      </div>
    </div>
  );
};
