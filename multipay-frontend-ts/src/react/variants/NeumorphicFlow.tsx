import React from "react";
import type { PickerVariantProps } from "../types";
import { Provider } from "../../core/types";
import styles from "./neumorphic-flow.module.css";
import { ProviderSlot } from "../ProviderSlot";

export const NeumorphicFlow: React.FC<PickerVariantProps> = ({
  views,
  selected,
  onSelect,
}) => {
  if (views.length === 0) {
    return null;
  }

  // Get the selected view
  const selectedView = views.find((v) => v.id === selected);
  const isLoading = selectedView?.state.loading || false;
  const hasError = !!selectedView?.state.error;

  const handleSelect = (providerId: Provider): void => {
    if (!isLoading) {
      void onSelect(providerId);
    }
  };

  return (
    <div className={styles.neumorphicFlow} data-theme="light">
      {/* One-Tap Pay Primary Action */}
      <button
        type="button"
        className={styles.payButton}
        disabled={
          selectedView
            ? selectedView.state.loading || !!selectedView.state.error
            : false
        }
        onClick={() => selected && handleSelect(selected)}
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
          {views.map((view) => {
            const isSelected = view.id === selected;
            const isDisabled = !view.entry.enabled;

            return (
              <button
                key={view.id}
                type="button"
                className={`${styles.segment} ${isSelected ? styles.segmentActive : ""} ${isDisabled ? styles.segmentDisabled : ""}`}
                disabled={isDisabled || isLoading}
                onClick={() => handleSelect(view.id)}
                role="radio"
                aria-checked={isSelected}
                aria-disabled={isDisabled}
                tabIndex={isSelected ? 0 : -1}
              >
                <span className={styles.segmentContent}>
                  {view.entry.icon && (
                    <span className={styles.segmentIcon}>
                      {view.entry.icon}
                    </span>
                  )}
                  <span className={styles.segmentLabel}>
                    {view.entry.label}
                  </span>
                </span>

                <ProviderSlot
                  view={view}
                  isDisabled={isDisabled}
                  isSelected={isSelected}
                />
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};
