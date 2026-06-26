import type { KeyboardEvent, ReactNode } from "react";
import type { PickerProviderView } from "./types";
import { Provider } from "../core/types";
import styles from "./styles/card.module.css";

interface ProviderCardProps {
  view: PickerProviderView;
  selected: boolean;
  onClick: () => void;
  onRetry?: () => void;
}

export function ProviderCard({
  view,
  selected,
  onClick,
  onRetry,
}: ProviderCardProps): JSX.Element {
  const { id, entry, state } = view;
  const loading = state.loading;
  const error = state.error;
  const isDisabled = !entry.enabled || loading;

  const handleClick = (): void => {
    if (!isDisabled && !loading) {
      onClick();
    }
  };

  const handleKeyDown = (e: KeyboardEvent): void => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleClick();
    }
  };

  // Built-in SVG icons for providers when entry.icon is not provided
  const renderIcon = (): ReactNode => {
    if (entry.icon) {
      return entry.icon;
    }

    switch (id) {
      case Provider.CASHFREE:
        return (
          <svg
            className={styles.providerIcon}
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <rect width="24" height="24" rx="4" fill="#0066FF" />
            <path
              d="M7 12h10M7 8h10M7 16h6"
              stroke="white"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
        );
      case Provider.RAZORPAY:
        return (
          <svg
            className={styles.providerIcon}
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <rect width="24" height="24" rx="4" fill="#0B4A7B" />
            <path
              d="M12 7v10M7 12h10"
              stroke="white"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
        );
      case Provider.PAYU:
        return (
          <svg
            className={styles.providerIcon}
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <rect width="24" height="24" rx="4" fill="#2C3E50" />
            <path
              d="M12 8l4 4-4 4M8 12h8"
              stroke="white"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      default:
        return null;
    }
  };

  return (
    <div
      className={`${styles.card} ${selected ? styles.selected : ""} ${
        isDisabled ? styles.disabled : ""
      } ${loading ? styles.loading : ""} ${error ? styles.error : ""}`}
      role="button"
      tabIndex={isDisabled ? -1 : 0}
      aria-pressed={selected}
      aria-disabled={isDisabled}
      aria-label={entry.label}
      aria-describedby={error ? `${styles.error}-${id}` : undefined}
      onClick={!isDisabled ? handleClick : undefined}
      onKeyDown={!isDisabled ? handleKeyDown : undefined}
    >
      {/* ENABLED IDLE State - default */}
      {!loading && !error && !isDisabled && (
        <>
          <div className={styles.iconContainer}>
            {entry.icon || renderIcon()}
            {entry.recommended && (
              <span className={styles.recommendedBadge}>Recommended</span>
            )}
          </div>
          <div className={styles.content}>
            <div className={styles.label}>{entry.label}</div>
            {entry.description && (
              <div className={styles.description}>{entry.description}</div>
            )}
          </div>
          {selected && (
            <div className={styles.checkmark}>
              <svg
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                aria-hidden="true"
              >
                <circle cx="12" cy="12" r="10" fill="currentColor" />
                <path
                  d="M8 12l2 2 4-4"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </div>
          )}
        </>
      )}

      {/* DISABLED State */}
      {isDisabled && !loading && (
        <>
          <div className={styles.iconContainer}>
            {entry.icon || renderIcon()}
          </div>
          <div className={styles.content}>
            <div className={styles.label}>{entry.label}</div>
            {entry.description && (
              <div className={styles.description}>{entry.description}</div>
            )}
            {entry.disabledMessage && (
              <div className={styles.disabledMessage}>
                {entry.disabledMessage}
              </div>
            )}
          </div>
        </>
      )}

      {/* LOADING State */}
      {loading && !error && (
        <>
          <div className={styles.iconContainer}>
            {entry.icon || renderIcon()}
          </div>
          <div className={styles.content}>
            <div className={styles.label}>{entry.label}</div>
            <div className={styles.loadingText}>
              <span className={styles.spinner} aria-hidden="true" />
              Redirecting to {entry.label}...
            </div>
          </div>
        </>
      )}

      {/* ERROR State */}
      {error && !loading && (
        <>
          <div className={styles.iconContainer}>
            {entry.icon || renderIcon()}
          </div>
          <div className={styles.content}>
            <div className={styles.label}>{entry.label}</div>
            <div
              id={`${styles.error}-${id}`}
              className={styles.errorMessage}
              role="alert"
              aria-live="polite"
            >
              {error}
            </div>
            {onRetry && (
              <button
                className={styles.retryButton}
                onClick={(e) => {
                  e.stopPropagation();
                  onRetry();
                }}
                type="button"
                aria-label={`Retry ${entry.label}`}
              >
                Retry
              </button>
            )}
          </div>
        </>
      )}
    </div>
  );
}
