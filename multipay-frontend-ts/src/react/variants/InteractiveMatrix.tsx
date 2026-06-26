import React from "react";
import styles from "./interactive-matrix.module.css";
import type { PickerVariantProps } from "../types";
import { ProviderSlot } from "../ProviderSlot";

export const InteractiveMatrix: React.FC<PickerVariantProps> = ({
  views,
  selected,
  onSelect,
}) => {
  return (
    <div className={styles.container}>
      <div className={styles.grid}>
        {views.map((view) => {
          const isSelected = view.id === selected;
          const isDisabled = !view.entry.enabled || view.state.loading;

          return (
            <button
              key={view.id}
              role="button"
              className={`${styles.card} ${isSelected ? styles.cardSelected : ""} ${!isSelected && selected ? styles.cardDim : ""} ${isDisabled ? styles.cardDisabled : ""}`}
              onClick={() => {
                if (!isDisabled) {
                  void onSelect(view.id);
                }
              }}
              disabled={isDisabled}
              aria-pressed={isSelected}
              aria-disabled={isDisabled}
            >
              <div className={styles.cardInner}>
                <div className={styles.logo}>{view.entry.icon}</div>
                <div className={styles.providerName}>{view.entry.label}</div>
                {view.entry.description && (
                  <div className={styles.description}>
                    {view.entry.description}
                  </div>
                )}

                <ProviderSlot
                  view={view}
                  isDisabled={isDisabled}
                  isSelected={isSelected}
                />

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
                  {view.state.loading && (
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
                  {view.state.error && (
                    <span className={styles.errorIcon}>⚠</span>
                  )}
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
};
