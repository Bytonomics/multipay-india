import React from "react";
import styles from "./interactive-matrix.module.css";
import type { ProviderOption, PickerProviderId } from "../types";
import { Provider } from "../../core/types";

interface InteractiveMatrixProps {
  providers: ProviderOption[];
  selected?: PickerProviderId;
  onSelect: (_provider: PickerProviderId) => void;
  loadingRecord: Record<Provider, boolean>;
  errorRecord: Record<Provider, string | undefined>;
}

const PROVIDER_DISPLAY: Record<
  Provider,
  { name: string; logo: string; description: string }
> = {
  [Provider.CASHFREE]: {
    name: "Cashfree",
    logo: "CF",
    description: "UPI, Cards, Wallets, Netbanking",
  },
  [Provider.RAZORPAY]: {
    name: "Razorpay",
    logo: "RZ",
    description: "UPI, Cards, Wallets, EMI",
  },
};

export const InteractiveMatrix: React.FC<InteractiveMatrixProps> = ({
  providers,
  selected,
  onSelect,
  loadingRecord,
  errorRecord,
}) => {
  // Filter providers - only show enabled Cashfree and Razorpay
  const visibleProviders = providers.filter(
    (p) => (p.id === Provider.CASHFREE || p.id === Provider.RAZORPAY) && p.enabled,
  );

  // Default select Cashfree if no selection
  React.useEffect(() => {
    if (!selected && visibleProviders.some((p) => p.id === Provider.CASHFREE)) {
      onSelect(Provider.CASHFREE);
    }
  }, [selected, visibleProviders, onSelect]);

  return (
    <div className={styles.container}>
      <div className={styles.grid}>
        {visibleProviders.map((provider) => {
          const providerKey = provider.id as Provider;
          const displayName = PROVIDER_DISPLAY[providerKey];
          const isLoading = loadingRecord[providerKey] || false;
          const hasError = !!errorRecord[providerKey];
          const isSelected = selected === provider.id;
          const isDisabled = isLoading || hasError || !provider.enabled;

          return (
            <button
              key={provider.id}
              role="button"
              className={`${styles.card} ${isSelected ? styles.cardSelected : ""} ${!isSelected && selected ? styles.cardDim : ""} ${isDisabled ? styles.cardDisabled : ""}`}
              onClick={() => !isDisabled && onSelect(provider.id)}
              disabled={!!isDisabled}
              aria-pressed={isSelected}
              aria-disabled={!!isDisabled}
            >
              <div className={styles.cardInner}>
                <div className={styles.logo}>{displayName.logo}</div>
                <div className={styles.providerName}>{displayName.name}</div>
                <div className={styles.description}>
                  {displayName.description}
                </div>

                {isSelected && (
                  <div className={styles.checkmark} aria-hidden="true">
                    <svg width={20} height={20} viewBox="0 0 20 20" fill="none">
                      <path
                        d="M16.6667 5.83334L7.50001 15L3.33334 10.8333"
                        stroke="currentColor"
                        strokeWidth={2}
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </div>
                )}
              </div>

              {isDisabled && (
                <div className={styles.disabledOverlay}>
                  {isLoading && (
                    <div className={styles.spinner} aria-hidden="true">
                      <svg
                        width={16}
                        height={16}
                        viewBox="0 0 16 16"
                        fill="none"
                      >
                        <circle
                          cx={8}
                          cy={8}
                          r={6}
                          stroke="currentColor"
                          strokeWidth={2}
                          strokeDasharray="4 2"
                        />
                      </svg>
                    </div>
                  )}
                  {hasError && <span className={styles.errorIcon}>⚠</span>}
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
};
