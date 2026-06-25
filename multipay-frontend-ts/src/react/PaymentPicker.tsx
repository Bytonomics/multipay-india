import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useState,
  useCallback,
} from "react";
import type { PaymentPickerProps, PickerProviderId } from "./types";
import { Provider } from "../core/types";
import { PickerHeader } from "./PickerHeader";
import { PickerFooter } from "./PickerFooter";
import { LoadingOverlay } from "./LoadingOverlay";
import { ErrorBanner } from "./ErrorBanner";
import { usePaymentPicker } from "./hooks/usePaymentPicker";
import { formatMinor } from "./utils/formatMinor";
import { DynamicStack } from "./variants/DynamicStack";
import { InteractiveMatrix } from "./variants/InteractiveMatrix";
import { SecureVault } from "./variants/SecureVault";
import { NeumorphicFlow } from "./variants/NeumorphicFlow";
import styles from "./styles/picker.module.css";
import "./styles/variables.css";

/**
 * PaymentPicker Component
 *
 * A comprehensive payment provider selection component with multiple visual variants,
 * theme support, and built-in loading/error states.
 *
 * @example
 * ```tsx
 * <PaymentPicker
 *   payment={{
 *     amountMinor: 50000,
 *     currency: 'INR',
 *     providers: [
 *       { id: 'cashfree', label: 'Cashfree', enabled: true },
 *       { id: 'razorpay', label: 'Razorpay', enabled: true },
 *     ],
 *     defaultSelected: 'cashfree',
 *   }}
 *   appearance={{
 *     variant: 'interactive-matrix',
 *     theme: 'auto',
 *   }}
 *   onSelect={(provider) => {
 *     // Handle provider selection
 *   }}
 * />
 * ```
 */
export const PaymentPicker = forwardRef<
  {
    selectProvider: (_providerId: PickerProviderId) => void;
    getSelectedProvider: () => PickerProviderId;
    isSelected: (_providerId: PickerProviderId) => boolean;
    setProviderDisabled: (
      _providerId: PickerProviderId,
      _disabled: boolean,
      _reason?: string,
    ) => void;
    focus: () => void;
    blur: () => void;
  },
  PaymentPickerProps
>((props, ref) => {
  const { payment, appearance = {}, onSelect } = props;

  // Apply defaults to appearance
  const variant = appearance.variant || "interactive-matrix";
  const theme = appearance.theme || "auto";
  const branding = appearance.branding;
  const className = appearance.className;
  const taxNote = appearance.taxNote;

  // Validate required payment fields
  if (
    typeof payment.amountMinor !== "number" ||
    isNaN(payment.amountMinor) ||
    !payment.currency ||
    !payment.providers ||
    payment.providers.length === 0
  ) {
    throw new Error(
      "PaymentPicker requires payment.amountMinor (valid number), payment.currency, and payment.providers (non-empty array)",
    );
  }

  // Local state for provider selection
  const [selectedProvider, setSelectedProvider] = useState<PickerProviderId>(
    () => {
      // Default to first enabled provider or defaultSelected if valid
      const defaultProvider = payment.defaultSelected;
      const isValidDefault =
        defaultProvider &&
        payment.providers.some(
          (p) => p.id === defaultProvider && p.enabled !== false,
        );

      if (isValidDefault) {
        return defaultProvider as PickerProviderId;
      }

      // Fallback to first enabled provider
      const firstEnabled = payment.providers.find((p) => p.enabled !== false);
      return firstEnabled?.id || payment.providers[0]?.id || "multipay_default";
    },
  );

  // Local state for provider options (to support dynamic disabling)
  const [providerOptions, setProviderOptions] = useState(payment.providers);

  // Theme state for 'auto' mode
  const [currentTheme, setCurrentTheme] = useState<"light" | "dark">(() => {
    if (theme === "auto") {
      return window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light";
    }
    return theme;
  });

  // Handle theme changes for 'auto' mode
  useEffect(() => {
    if (theme !== "auto") return;

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = (_e: MediaQueryListEvent): void => {
      setCurrentTheme(_e.matches ? "dark" : "light");
    };

    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, [theme]);

  // Get loading and error states from hook
  const { loadingRecord, errorRecord, controls } = usePaymentPicker();

  // Filter providers - only show enabled ones
  const visibleProviders = providerOptions.filter((provider) => {
    // Cashfree and Razorpay are shown by default
    if (provider.id === Provider.CASHFREE || provider.id === Provider.RAZORPAY) {
      return provider.enabled !== false;
    }
    // PayU and others are NOT shown by default
    return false;
  });

  // Handle provider selection with guard
  const handleProviderSelect = useCallback(
    async (providerId: PickerProviderId) => {
      const provider = providerOptions.find((p) => p.id === providerId);

      // Guard: Only allow selection of enabled canonical providers
      if (!provider || provider.enabled === false) {
        return;
      }

      // Only emit canonical Provider types (cashfree, razorpay), not placeholders like payu
      const canonicalProviders: Provider[] = [
        Provider.CASHFREE,
        Provider.RAZORPAY,
      ];
      if (!canonicalProviders.includes(providerId as Provider)) {
        return;
      }

      setSelectedProvider(providerId);

      // Set loading state (providerId is guaranteed to be Provider after validation above)
      controls.setLoading(providerId as Provider, true);

      try {
        await onSelect(providerId as Provider);
      } catch (error) {
        controls.setError(
          providerId as Provider,
          error instanceof Error ? error.message : "Payment failed",
        );
      } finally {
        controls.setLoading(providerId as Provider, false);
      }
    },
    [providerOptions, onSelect, controls],
  );

  // Imperative API via ref
  useImperativeHandle(
    ref,
    () => ({
      selectProvider: (providerId: PickerProviderId) => {
        void handleProviderSelect(providerId);
      },
      getSelectedProvider: () => selectedProvider,
      isSelected: (providerId: PickerProviderId) =>
        selectedProvider === providerId,
      setProviderDisabled: (
        providerId: PickerProviderId,
        disabled: boolean,
        reason?: string,
      ) => {
        setProviderOptions((prev) =>
          prev.map((p) =>
            p.id === providerId
              ? {
                  ...p,
                  enabled: !disabled,
                  disabledReason: disabled ? reason : undefined,
                }
              : p,
          ),
        );
      },
      focus: () => {
        // Focus the first provider card
        const firstCard = document.querySelector(
          `[role="button"]`,
        ) as HTMLElement;
        firstCard?.focus();
      },
      blur: () => {
        // Blur the focused element
        const focused = document.activeElement as HTMLElement;
        focused?.blur();
      },
    }),
    [selectedProvider, handleProviderSelect],
  );

  // Determine which variant component to render
  const renderVariant = (): JSX.Element => {
    const commonProps = {
      providers: visibleProviders,
      selected: selectedProvider,
      onSelect: handleProviderSelect,
      loadingRecord,
      errorRecord,
    };

    switch (variant) {
      case "dynamic-stack":
        return <DynamicStack {...commonProps} loadingRecord={loadingRecord} errorRecord={errorRecord} />;
      case "interactive-matrix":
        return <InteractiveMatrix {...commonProps} loadingRecord={loadingRecord} errorRecord={errorRecord} />;
      case "secure-vault":
        return <SecureVault {...commonProps} loadingRecord={loadingRecord} errorRecord={errorRecord} />;
      case "neumorphic-flow":
        return <NeumorphicFlow {...commonProps} loadingRecord={loadingRecord} errorRecord={errorRecord} formattedTotal={formattedTotal} taxNote={taxNoteText} />;
      default:
        return <InteractiveMatrix {...commonProps} loadingRecord={loadingRecord} errorRecord={errorRecord} />;
    }
  };

  // Format total amount
  const formattedTotal = formatMinor(payment.amountMinor, payment.currency);

  // Tax note text
  const taxNoteText = taxNote || "Total amount inclusive of all taxes";

  return (
    <div
      className={`${styles.picker} ${className || ""}`}
      data-theme={currentTheme}
      style={{
        fontFamily: appearance.fontFamily || "var(--mpay-font-family, inherit)",
      }}
      role="region"
      aria-label="Payment provider selection"
    >
      {/* Header with branding */}
      {branding && <PickerHeader branding={branding} />}

      {/* Total amount display */}
      <div className={styles.totalSection}>
        <div className={styles.totalLabel}>Total Amount</div>
        <div className={styles.totalAmount}>{formattedTotal}</div>
        <div className={styles.taxNote}>{taxNoteText}</div>
      </div>

      {/* Provider grid/stack based on variant */}
      <div className={styles.providersSection}>{renderVariant()}</div>

      {/* Footer with branding */}
      {branding && branding.footerText && (
        <PickerFooter footerText={branding.footerText} />
      )}

      {/* Loading overlay */}
      {Object.values(loadingRecord).some((loading) => loading) && (
        <LoadingOverlay
          provider={
            visibleProviders.find((p) => p.id === selectedProvider)?.label ||
            "payment provider"
          }
        />
      )}

      {/* Error banner */}
      {Object.values(errorRecord).some((error) => error) && (
        <ErrorBanner
          message={
            Object.values(errorRecord).find((error) => error) ||
            "Payment failed"
          }
          onRetry={() => {
            const providerWithErrors = Object.entries(errorRecord)
              .filter(([, error]) => error)
              .map(([providerId]) => providerId);

            if (providerWithErrors.length > 0) {
              void handleProviderSelect(
                providerWithErrors[0] as PickerProviderId,
              );
            }
          }}
        />
      )}
    </div>
  );
});

PaymentPicker.displayName = "PaymentPicker";
