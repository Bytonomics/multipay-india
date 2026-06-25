import React from "react";
import { Provider } from "../../core/types";
import type { ProviderOption, PickerProviderId } from "../types";
import styles from "./dynamic-stack.module.css";

export interface DynamicStackProps {
  providers: ProviderOption[];
  selected?: PickerProviderId;
  onSelect: (_provider: PickerProviderId) => void;
  loadingRecord?: Record<Provider, boolean>;
  errorRecord?: Record<Provider, string | undefined>;
}

export const DynamicStack: React.FC<DynamicStackProps> = ({
  providers,
  selected,
  onSelect,
  loadingRecord = {} as Record<Provider, boolean>,
  errorRecord = {} as Record<Provider, string | undefined>,
}) => {
  const [showAlternatives, setShowAlternatives] = React.useState(false);

  // Filter providers: only show aggregators (Cashfree, Razorpay)
  const visibleProviders = providers.filter(
    (p) => (p.id === Provider.CASHFREE || p.id === Provider.RAZORPAY) && p.enabled,
  );

  // Primary provider (first in list) - default selected
  const primaryProvider = visibleProviders[0];
  // Alternative providers (rest)
  const alternativeProviders = visibleProviders.slice(1);

  if (!primaryProvider) {
    return null;
  }

  const isSelected = (providerId: PickerProviderId): boolean =>
    providerId === selected;
  const currentSelected = selected || primaryProvider.id;
  const currentSelectedProvider = currentSelected as Provider;
  const isLoading = loadingRecord[currentSelectedProvider] || false;
  const hasError = !!errorRecord[currentSelectedProvider];

  return (
    <div className={styles.dynamicStack} data-theme="light">
      {/* Optimized Route Card - Glowing Selector */}
      <button
        role="button"
        type="button"
        className={`${styles.primaryCard} ${isSelected(primaryProvider.id) ? styles.primaryCardSelected : ""}`}
        onClick={() => primaryProvider.enabled && onSelect(primaryProvider.id)}
        disabled={!primaryProvider.enabled}
        aria-pressed={isSelected(primaryProvider.id)}
      >
        <div className={styles.primaryCardContent}>
          <div className={styles.providerHeader}>
            <div className={styles.iconWrapper}>{primaryProvider.icon}</div>
            <div className={styles.providerInfo}>
              <h3 className={styles.providerLabel}>{primaryProvider.label}</h3>
              {primaryProvider.description && (
                <p className={styles.providerDescription}>
                  {primaryProvider.description}
                </p>
              )}
            </div>
            {primaryProvider.recommended && (
              <span className={styles.recommendedBadge}>Recommended</span>
            )}
          </div>

          {(!primaryProvider.enabled || primaryProvider.disabledReason) && (
            <p className={styles.disabledMessage}>
              {primaryProvider.disabledReason || "Disabled"}
            </p>
          )}

          {isSelected(primaryProvider.id) && (
            <div className={styles.selectedDetails}>
              {/* Success rate if available */}
              {(primaryProvider as ProviderOption & { successRate?: string })
                .successRate && (
                <div className={styles.successRate}>
                  <span className={styles.successRateLabel}>Success Rate:</span>
                  <span className={styles.successRateValue}>
                    {
                      (
                        primaryProvider as ProviderOption & {
                          successRate?: string;
                        }
                      ).successRate
                    }
                  </span>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Selection indicator */}
        {isSelected(primaryProvider.id) && (
          <div className={styles.selectionIndicator}>
            <div className={styles.glow}></div>
            <div className={styles.checkmark}>✓</div>
          </div>
        )}
      </button>

      {/* Alternative Options Accordion */}
      {alternativeProviders.length > 0 && (
        <div className={styles.alternativesSection}>
          <button
            role="button"
            type="button"
            className={styles.accordionTrigger}
            onClick={() => setShowAlternatives(!showAlternatives)}
            aria-expanded={showAlternatives}
          >
            <span className={styles.accordionText}>
              {showAlternatives ? "Hide" : "Show"} Alternative Options
            </span>
            <span
              className={`${styles.accordionIcon} ${showAlternatives ? styles.accordionIconOpen : ""}`}
            >
              ▼
            </span>
          </button>

          {showAlternatives && (
            <div className={styles.alternativesGrid}>
              {alternativeProviders.map((provider) => (
                <button
                  key={provider.id}
                  role="button"
                  type="button"
                  className={`${styles.alternativeCard} ${isSelected(provider.id as Provider) ? styles.alternativeCardSelected : ""}`}
                  onClick={() => provider.enabled && onSelect(provider.id)}
                  disabled={!provider.enabled}
                  aria-pressed={isSelected(provider.id as Provider)}
                >
                  <div className={styles.alternativeCardContent}>
                    <div className={styles.iconWrapper}>{provider.icon}</div>
                    <div className={styles.providerInfo}>
                      <h4 className={styles.providerLabel}>{provider.label}</h4>
                      {provider.description && (
                        <p className={styles.providerDescription}>
                          {provider.description}
                        </p>
                      )}
                    </div>
                    {provider.recommended && (
                      <span className={styles.recommendedBadge}>
                        Recommended
                      </span>
                    )}
                  </div>

                  {isSelected(provider.id as Provider) && (
                    <div className={styles.selectionIndicator}>
                      <div className={styles.glow}></div>
                      <div className={styles.checkmark}>✓</div>
                    </div>
                  )}

                  {(!provider.enabled || provider.disabledReason) && (
                    <p className={styles.disabledMessage}>
                      {provider.disabledReason || "Disabled"}
                    </p>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* CTA Section */}
      {selected && (
        <div className={styles.ctaSection}>
          <button
            type="button"
            className={styles.ctaButton}
            disabled={isLoading || hasError}
          >
            {isLoading ? "Processing..." : hasError ? "Retry" : "Continue"}
          </button>
        </div>
      )}
    </div>
  );
};
