import styles from "./styles/components.module.css";

export interface LoadingOverlayProps {
  provider: string;
}

export function LoadingOverlay({ provider }: LoadingOverlayProps): JSX.Element {
  return (
    <div className={styles.loadingOverlay} role="status" aria-live="polite">
      <div className={styles.spinner} aria-hidden="true" />
      <p className={styles.loadingText}>Connecting to {provider}...</p>
    </div>
  );
}
