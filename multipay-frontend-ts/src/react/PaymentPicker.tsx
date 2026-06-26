import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
  useCallback,
} from "react";
import type {
  PaymentPickerProps,
  PickerControls,
  PickerProviderView,
  ProviderEntry,
  ProviderRuntimeState,
} from "./types";
import { Provider, PickerVariant, ResolvedTheme } from "../core/types";
import { PickerTheme } from "../core/types";
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
 * Provider order for building the picker view array.
 * Controls the order in which providers are rendered.
 */
const PROVIDER_ORDER: readonly Provider[] = [
  Provider.CASHFREE,
  Provider.RAZORPAY,
  Provider.PAYU,
];

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
 *     providers: {
 *       cashfree: { label: 'Cashfree', visible: true, enabled: true },
 *       razorpay: { label: 'Razorpay', visible: true, enabled: true },
 *       payu: { label: 'PayU', visible: false, enabled: false },
 *     },
 *     default: Provider.CASHFREE,
 *   }}
 *   appearance={{
 *     variant: PickerVariant.INTERACTIVE_MATRIX,
 *     theme: PickerTheme.AUTO,
 *   }}
 *   onSelect={(provider) => {
 *     // Handle provider selection
 *   }}
 * />
 * ```
 */
export const PaymentPicker = forwardRef<PickerControls, PaymentPickerProps>(
  (props, ref) => {
    const { payment, appearance = {}, onSelect } = props;
    const pickerRootRef = useRef<HTMLDivElement>(null);

    // Validate required payment fields
    if (!payment.amountMinor || payment.amountMinor <= 0) {
      throw new Error(
        "payment.amountMinor is required and must be greater than 0",
      );
    }
    if (!payment.currency) {
      throw new Error("payment.currency is required");
    }

    // Apply defaults to appearance
    const variant = appearance.variant ?? PickerVariant.INTERACTIVE_MATRIX;
    const theme = appearance.theme ?? PickerTheme.AUTO;
    const branding = appearance.branding;
    const className = appearance.className;

    // Build PickerProviderView[] from named-field providers + runtime state
    const { runtime, controls } = usePaymentPicker();

    const getProviderEntry = useCallback(
      (id: Provider): ProviderEntry => {
        switch (id) {
          case Provider.CASHFREE:
            return payment.providers.cashfree;
          case Provider.RAZORPAY:
            return payment.providers.razorpay;
          case Provider.PAYU:
            return payment.providers.payu;
        }
      },
      [payment.providers],
    );

    const getRuntimeState = (id: Provider): ProviderRuntimeState => {
      switch (id) {
        case Provider.CASHFREE:
          return runtime.cashfree;
        case Provider.RAZORPAY:
          return runtime.razorpay;
        case Provider.PAYU:
          return runtime.payu;
      }
    };

    const views: PickerProviderView[] = PROVIDER_ORDER.map((id) => ({
      id,
      entry: getProviderEntry(id),
      state: getRuntimeState(id) || { loading: false },
    })).filter((view) => view.entry.visible);

    // Local state for provider selection (canonical Provider enum value)
    const [selected, setSelected] = useState<Provider>(() => {
      // Validate default is enabled and visible
      const defaultEntry = payment.default
        ? getProviderEntry(payment.default)
        : null;
      const isValidDefault =
        payment.default &&
        defaultEntry &&
        defaultEntry.visible &&
        defaultEntry.enabled;

      if (isValidDefault) {
        return payment.default;
      }

      // Emit warning if default was provided but invalid/disabled/absent
      if (payment.default && !isValidDefault) {
        console.warn(
          `[PaymentPicker] Invalid/disabled/absent default provider: ${payment.default}. Falling back to first enabled provider.`,
        );
      }

      // Fallback to first visible & enabled provider
      const firstValid = views.find((v) => v.entry.enabled);
      return firstValid?.id || views[0]?.id || Provider.CASHFREE;
    });

    // Resolve theme: handle 'auto' with media query
    const [currentTheme, setCurrentTheme] = useState<ResolvedTheme>(() => {
      if (theme === PickerTheme.AUTO) {
        return window.matchMedia("(prefers-color-scheme: dark)").matches
          ? ResolvedTheme.DARK
          : ResolvedTheme.LIGHT;
      }
      return theme === PickerTheme.DARK
        ? ResolvedTheme.DARK
        : ResolvedTheme.LIGHT;
    });

    // Update theme when 'auto' changes based on system preference
    useEffect(() => {
      if (theme !== PickerTheme.AUTO) {
        const resolved =
          theme === PickerTheme.DARK ? ResolvedTheme.DARK : ResolvedTheme.LIGHT;
        setCurrentTheme(resolved);
        return;
      }

      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
      const handleChange = (e: MediaQueryListEvent): void => {
        setCurrentTheme(e.matches ? ResolvedTheme.DARK : ResolvedTheme.LIGHT);
      };

      mediaQuery.addEventListener("change", handleChange);
      return () => mediaQuery.removeEventListener("change", handleChange);
    }, [theme]);

    // Format total amount using ISO 4217 exponent
    const formattedTotal = formatMinor(payment.amountMinor, payment.currency);

    // Tax note text with default (PRE-TAX ONLY: final taxes added at checkout by vendor)
    const taxNoteText =
      appearance.taxNote ??
      "Final taxes are added at checkout by the payment provider.";

    // Handle provider selection
    const handleProviderSelect = useCallback(
      async (selectedProvider: Provider) => {
        const entry = getProviderEntry(selectedProvider);

        // Guard: reject if not enabled
        if (!entry.enabled) {
          return;
        }

        setSelected(selectedProvider);
        controls.setLoading(selectedProvider, true);

        try {
          await onSelect(selectedProvider);
        } catch (_error) {
          controls.setError(
            selectedProvider,
            _error instanceof Error ? _error.message : "Payment failed",
          );
        } finally {
          controls.setLoading(selectedProvider, false);
        }
      },
      [getProviderEntry, onSelect, controls],
    );

    // Imperative API via ref
    useImperativeHandle(
      ref,
      () => ({
        selectProvider: (provider: Provider) => {
          void handleProviderSelect(provider);
        },
        getSelectedProvider: () => selected,
        isSelected: (provider: Provider) => selected === provider,
        focus: () => {
          const firstCard = pickerRootRef.current?.querySelector(
            '[role="button"]',
          ) as HTMLElement | null;
          firstCard?.focus();
        },
        blur: () => {
          const focused = document.activeElement as HTMLElement;
          focused?.blur();
        },
      }),
      [selected, handleProviderSelect],
    );

    // Render variant with shared PickerVariantProps
    const variantProps = {
      views,
      selected,
      onSelect: handleProviderSelect,
      theme: currentTheme,
      formattedTotal,
      taxNote: taxNoteText,
    };

    let variantComponent: JSX.Element;

    switch (variant) {
      case PickerVariant.DYNAMIC_STACK:
        variantComponent = <DynamicStack {...variantProps} />;
        break;

      case PickerVariant.INTERACTIVE_MATRIX:
        variantComponent = <InteractiveMatrix {...variantProps} />;
        break;

      case PickerVariant.SECURE_VAULT:
        variantComponent = <SecureVault {...variantProps} />;
        break;

      case PickerVariant.NEUMORPHIC_FLOW:
        variantComponent = <NeumorphicFlow {...variantProps} />;
        break;

      default: {
        // Exhaustive check: TypeScript will error if any variant is missing
        const _exhaustiveCheck: never = variant;
        return _exhaustiveCheck;
      }
    }

    return (
      <div
        ref={pickerRootRef}
        className={`mpay-picker ${styles.picker} ${className || ""}`}
        data-theme={currentTheme}
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
        <div className={styles.providersSection}>{variantComponent}</div>

        {/* Footer with branding */}
        {branding && branding.footerText && (
          <PickerFooter footerText={branding.footerText} />
        )}

        {/* Loading overlay */}
        {(runtime.cashfree.loading ||
          runtime.razorpay.loading ||
          runtime.payu.loading) && (
          <LoadingOverlay
            provider={
              views.find((v) => v.id === selected)?.entry.label ||
              "payment provider"
            }
          />
        )}

        {/* Error banner */}
        {(runtime.cashfree.error ||
          runtime.razorpay.error ||
          runtime.payu.error) && (
          <ErrorBanner
            message={
              runtime.cashfree.error ||
              runtime.razorpay.error ||
              runtime.payu.error ||
              "Payment failed"
            }
            onRetry={() => {
              if (runtime.cashfree.error) {
                void handleProviderSelect(Provider.CASHFREE);
              } else if (runtime.razorpay.error) {
                void handleProviderSelect(Provider.RAZORPAY);
              } else if (runtime.payu.error) {
                void handleProviderSelect(Provider.PAYU);
              }
            }}
          />
        )}
      </div>
    );
  },
);

PaymentPicker.displayName = "PaymentPicker";
