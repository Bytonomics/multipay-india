import React from "react";
import type { PickerVariantProps, PickerProviderView } from "../types";
import styles from "./dynamic-stack.module.css";
import { ProviderSlot } from "../ProviderSlot";

export const DynamicStack: React.FC<PickerVariantProps> = ({
  views,
  selected,
  onSelect,
  theme,
}) => {
  const [showAlternatives, setShowAlternatives] = React.useState(false);

  if (!views || views.length === 0) {
    return null;
  }

  // Primary card = views[0], alternatives = views.slice(1)
  const primaryView = views[0];
  const alternativeViews = views.slice(1);

  return (
    <div className={styles.dynamicStack} data-theme={theme}>
      {/* Optimized Route Card - Primary Card */}
      <button
        role="button"
        type="button"
        className={`${styles.primaryCard} ${primaryView.id === selected ? styles.primaryCardSelected : ""}`}
        onClick={() => {
          if (primaryView.entry.enabled) {
            void onSelect(primaryView.id);
          }
        }}
        disabled={!primaryView.entry.enabled}
        aria-pressed={primaryView.id === selected}
      >
        <div className={styles.primaryCardContent}>
          <div className={styles.providerHeader}>
            <div className={styles.iconWrapper}>{primaryView.entry.icon}</div>
            <div className={styles.providerInfo}>
              <h3 className={styles.providerLabel}>
                {primaryView.entry.label}
              </h3>
              {primaryView.entry.description && (
                <p className={styles.providerDescription}>
                  {primaryView.entry.description}
                </p>
              )}
            </div>
            {primaryView.entry.recommended && (
              <span className={styles.recommendedBadge}>Recommended</span>
            )}
          </div>

          <ProviderSlot
            view={primaryView}
            isDisabled={!primaryView.entry.enabled}
            isSelected={primaryView.id === selected}
          />
        </div>

        {/* Selection indicator */}
        {primaryView.id === selected && (
          <div className={styles.selectionIndicator}>
            <div className={styles.glow}></div>
            <div className={styles.checkmark}>✓</div>
          </div>
        )}
      </button>

      {/* Alternative Options Accordion */}
      {alternativeViews.length > 0 && (
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
              {alternativeViews.map((view: PickerProviderView) => (
                <button
                  key={view.id}
                  role="button"
                  type="button"
                  className={`${styles.alternativeCard} ${view.id === selected ? styles.alternativeCardSelected : ""}`}
                  onClick={() => {
                    if (view.entry.enabled) {
                      void onSelect(view.id);
                    }
                  }}
                  disabled={!view.entry.enabled}
                  aria-pressed={view.id === selected}
                >
                  <div className={styles.alternativeCardContent}>
                    <div className={styles.iconWrapper}>{view.entry.icon}</div>
                    <div className={styles.providerInfo}>
                      <h4 className={styles.providerLabel}>
                        {view.entry.label}
                      </h4>
                      {view.entry.description && (
                        <p className={styles.providerDescription}>
                          {view.entry.description}
                        </p>
                      )}
                    </div>
                    {view.entry.recommended && (
                      <span className={styles.recommendedBadge}>
                        Recommended
                      </span>
                    )}
                  </div>

                  <ProviderSlot
                    view={view}
                    isDisabled={!view.entry.enabled}
                    isSelected={view.id === selected}
                  />

                  {view.id === selected && (
                    <div className={styles.selectionIndicator}>
                      <div className={styles.glow}></div>
                      <div className={styles.checkmark}>✓</div>
                    </div>
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
          <button type="button" className={styles.ctaButton} disabled={false}>
            Continue
          </button>
        </div>
      )}
    </div>
  );
};
