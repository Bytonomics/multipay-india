import type { ProviderOption, PickerProviderId } from "../types";
import { Provider } from "../../core/types";
import styles from "./secure-vault.module.css";

interface SecureVaultProps {
  providers: ProviderOption[];
  selected?: PickerProviderId;
  onSelect: (_provider: PickerProviderId) => void;
  loadingRecord?: Record<Provider, boolean>;
  errorRecord?: Record<Provider, string | undefined>;
  formattedTotal?: string;
  taxNote?: string;
}

/**
 * SecureVault - V3 Trust & security/banking/vault variant
 *
 * Features:
 * - Framed vault of recessed/embedded selection slots
 * - Selected slot has green inner-glow + shield/verification badge
 * - Lock + circuit motifs for security aesthetic
 * - Dark-optimized design with full light palette support
 * - One slot per AGGREGATOR (Cashfree default-selected + Razorpay; PayU NOT shown)
 */
export function SecureVault({
  providers,
  selected,
  onSelect,
  loadingRecord = {} as Record<Provider, boolean>,
  errorRecord = {} as Record<Provider, string | undefined>,
  formattedTotal,
  taxNote,
}: SecureVaultProps): JSX.Element {
  // Filter to only show cashfree and razorpay (PayU is excluded)
  const visibleProviders = providers.filter(
    (p) => (p.id === Provider.CASHFREE || p.id === Provider.RAZORPAY) && p.enabled,
  );

  // Default to cashfree if nothing selected
  const selectedProvider = selected || Provider.CASHFREE;

  return (
    <div className={styles.secureVault}>
      <div className={styles.vaultFrame}>
        {/* Vault header with lock motif */}
        <div className={styles.vaultHeader}>
          <svg
            className={styles.lockIcon}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <rect x="5" y="11" width="14" height="10" rx="2" />
            <path d="M12 7V3M8 7V5a4 4 0 018 0v2" />
          </svg>
          <span className={styles.headerTitle}>Secure Payment Gateway</span>
          <div className={styles.circuitDecoration}>
            <span className={styles.circuitDot} />
            <span className={styles.circuitDot} />
            <span className={styles.circuitDot} />
          </div>
        </div>

        {/* Vault slots container */}
        <div className={styles.vaultSlots}>
          {visibleProviders.map((provider) => {
            const isSelected = provider.id === selectedProvider;
            const providerKey = provider.id as Provider;
            const isLoading = loadingRecord[providerKey];
            const hasError = !!errorRecord[providerKey];

            return (
              <button
                key={provider.id}
                className={`${styles.vaultSlot} ${isSelected ? styles.selected : ""} ${isLoading ? styles.loading : ""} ${hasError ? styles.error : ""}`}
                onClick={() => onSelect(provider.id)}
                disabled={isLoading}
                aria-selected={isSelected}
                aria-busy={isLoading}
              >
                {/* Selection glow */}
                {isSelected && (
                  <div className={styles.selectionGlow} aria-hidden="true">
                    <div className={styles.glowInner} />
                  </div>
                )}

                {/* Shield badge for selected provider */}
                {isSelected && (
                  <div className={styles.shieldBadge} aria-hidden="true">
                    <svg
                      viewBox="0 0 24 24"
                      fill="currentColor"
                      className={styles.shieldIcon}
                    >
                      <path d="M12 2L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-3z" />
                      <path
                        d="m9 12 2 2 4-4"
                        fill="none"
                        stroke="white"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </div>
                )}

                {/* Provider label */}
                <span className={styles.providerLabel}>
                  {provider.id === Provider.CASHFREE ? "Cashfree" : "Razorpay"}
                </span>

                {/* Loading indicator */}
                {isLoading && (
                  <div className={styles.slotLoading} aria-hidden="true">
                    <span className={styles.spinner} />
                  </div>
                )}

                {/* Error indicator */}
                {hasError && !isLoading && (
                  <div className={styles.errorIndicator} aria-hidden="true">
                    <span>!</span>
                  </div>
                )}

                {/* Circuit pattern decoration */}
                <div className={styles.circuitPattern} aria-hidden="true">
                  <svg
                    viewBox="0 0 40 40"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="1"
                  >
                    <path d="M2 20h8M30 20h8M20 2v8M20 30v8" />
                    <circle cx="20" cy="20" r="4" />
                  </svg>
                </div>
              </button>
            );
          })}
        </div>

        {/* Vault footer with total and tax note */}
        {(formattedTotal || taxNote) && (
          <div className={styles.vaultFooter}>
            {formattedTotal && (
              <div className={styles.totalDisplay}>
                <span className={styles.totalLabel}>Total Amount</span>
                <span className={styles.totalValue}>{formattedTotal}</span>
              </div>
            )}
            {taxNote && (
              <div className={styles.taxNote}>
                <span className={styles.taxNoteIcon}>ℹ</span>
                <span>{taxNote}</span>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
