import styles from "./styles/components.module.css";

export interface ErrorBannerProps {
  message: string;
  onRetry: () => void;
}

export function ErrorBanner({
  message,
  onRetry,
}: ErrorBannerProps): JSX.Element {
  return (
    <div className={styles.errorBanner} role="alert" aria-live="assertive">
      <p className={styles.errorMessage}>{message}</p>
      <button
        type="button"
        className={styles.retryButton}
        onClick={onRetry}
        aria-label="Retry connection"
      >
        Try Again
      </button>
    </div>
  );
}
