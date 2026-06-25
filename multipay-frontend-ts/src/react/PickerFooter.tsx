import styles from "./styles/components.module.css";

export interface PickerFooterProps {
  footerText?: string;
}

export function PickerFooter({
  footerText,
}: PickerFooterProps): JSX.Element | null {
  if (!footerText) {
    return null;
  }

  return (
    <div className={styles.footer}>
      <p className={styles.footerText}>{footerText}</p>
    </div>
  );
}
