import type { PickerBranding } from "./types";
import styles from "./styles/components.module.css";

export interface PickerHeaderProps {
  branding?: PickerBranding;
}

export function PickerHeader({
  branding,
}: PickerHeaderProps): JSX.Element | null {
  if (!branding) {
    return null;
  }

  const { logo, title, subtitle } = branding;

  return (
    <div className={styles.header}>
      {logo && <div className={styles.logo}>{logo}</div>}
      {title && <h3 className={styles.title}>{title}</h3>}
      {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
    </div>
  );
}
