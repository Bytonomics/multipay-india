import type { PickerVariantProps } from "../types";
import styles from "./secure-vault.module.css";
import { ProviderSlot } from "../ProviderSlot";

/**
 * SecureVault - V3 Trust & security/banking/vault variant
 *
 * Features:
 * - Framed vault of recessed/embedded selection slots
 * - Selected slot has green inner-glow + shield/verification badge
 * - Lock + circuit motifs for security aesthetic
 * - Dark-optimized design with full light palette support
 * - One slot per AGGREGATOR (visible providers only)
 */
export function SecureVault({
  views,
  selected,
  onSelect,
  theme,
}: PickerVariantProps): JSX.Element {
  return (
    <div className={styles.secureVault} data-theme={theme}>
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
          {views.map((view) => {
            const isSelected = view.id === selected;
            const isDisabled = !view.entry.enabled || view.state.loading;

            return (
              <button
                key={view.id}
                className={`${styles.vaultSlot} ${isSelected ? styles.selected : ""} ${view.state.loading ? styles.loading : ""} ${view.state.error ? styles.error : ""}`}
                onClick={() => {
                  if (view.entry.enabled && !view.state.loading) {
                    void onSelect(view.id);
                  }
                }}
                disabled={isDisabled}
                aria-selected={isSelected}
                aria-busy={view.state.loading}
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
                <span className={styles.providerLabel}>{view.entry.label}</span>

                {/* Shared state rendering */}
                <ProviderSlot
                  view={view}
                  isDisabled={isDisabled}
                  isSelected={isSelected}
                />

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
      </div>
    </div>
  );
}
