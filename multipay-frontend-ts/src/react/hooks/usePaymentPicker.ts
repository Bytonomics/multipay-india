import { useCallback, useState } from "react";
import { Provider } from "../../core/types";

interface ProviderState {
  loading: boolean;
  error?: string;
}

interface PaymentPickerState {
  loadingRecord: Record<Provider, boolean>;
  errorRecord: Record<Provider, string | undefined>;
}

export function usePaymentPicker(): {
  loadingRecord: PaymentPickerState["loadingRecord"];
  errorRecord: PaymentPickerState["errorRecord"];
  controls: {
    setLoading: (_provider: Provider, _loading: boolean) => void;
    setError: (_provider: Provider, _error?: string) => void;
    clearError: (_provider: Provider) => void;
  };
} {
  const [providerStates, setProviderStates] = useState<
    Map<Provider, ProviderState>
  >(() => {
    const initialStates = new Map<Provider, ProviderState>();
    initialStates.set(Provider.CASHFREE, { loading: false });
    initialStates.set(Provider.RAZORPAY, { loading: false });
    return initialStates;
  });

  const setLoading = useCallback(
    (provider: Provider, loading: boolean) => {
      setProviderStates((prev) => {
        const next = new Map(prev);
        const existing = next.get(provider) || { loading: false };
        next.set(provider, { ...existing, loading });
        return next;
      });
    },
    [],
  );

  const setError = useCallback((provider: Provider, error?: string) => {
    setProviderStates((prev) => {
      const next = new Map(prev);
      const existing = next.get(provider) || { loading: false };
      next.set(provider, { ...existing, error });
      return next;
    });
  }, []);

  const clearError = useCallback((provider: Provider) => {
    setProviderStates((prev) => {
      const next = new Map(prev);
      const existing = next.get(provider);
      if (existing) {
        next.set(provider, { ...existing, error: undefined });
      }
      return next;
    });
  }, []);

  // Derive loadingRecord and errorRecord from providerStates
  const loadingRecord: Record<Provider, boolean> = {
    [Provider.CASHFREE]: providerStates.get(Provider.CASHFREE)?.loading || false,
    [Provider.RAZORPAY]: providerStates.get(Provider.RAZORPAY)?.loading || false,
  };

  const errorRecord: Record<Provider, string | undefined> = {
    [Provider.CASHFREE]: providerStates.get(Provider.CASHFREE)?.error,
    [Provider.RAZORPAY]: providerStates.get(Provider.RAZORPAY)?.error,
  };

  return {
    loadingRecord,
    errorRecord,
    controls: {
      setLoading,
      setError,
      clearError,
    },
  };
}
