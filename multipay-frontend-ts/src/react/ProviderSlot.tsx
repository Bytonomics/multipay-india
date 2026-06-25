import React from "react";
import type { PickerProviderView } from "./types";
import styles from "./provider-slot.module.css";

interface ProviderSlotProps {
  view: PickerProviderView;
  isDisabled: boolean;
  isSelected: boolean;
}

export const ProviderSlot: React.FC<ProviderSlotProps> = ({
  view,
  isDisabled,
  isSelected,
}) => {
  return (
    <>
      {isDisabled && view.entry.disabledMessage && (
        <div className={styles.disabledMessage}>
          {view.entry.disabledMessage}
        </div>
      )}

      {view.state.loading && (
        <div className={styles.loadingIndicator} aria-hidden="true">
          <span className={styles.spinner} />
        </div>
      )}

      {view.state.error && (
        <div className={styles.errorIndicator} aria-hidden="true">
          <span>!</span>
        </div>
      )}

      {isSelected && (
        <div className={styles.selectionIndicator} aria-hidden="true">
          ✓
        </div>
      )}
    </>
  );
};
