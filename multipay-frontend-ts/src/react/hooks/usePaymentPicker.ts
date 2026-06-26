import { useCallback, useState } from "react";
import { Provider } from "../../core/types";
import type { PickerRuntimeState } from "../types";

export interface PaymentPickerControls {
  setLoading: (provider: Provider, loading: boolean) => void;
  setError: (provider: Provider, error?: string) => void;
  clearError: (provider: Provider) => void;
}

export function usePaymentPicker(): {
  runtime: PickerRuntimeState;
  controls: PaymentPickerControls;
} {
  const [runtime, setRuntime] = useState<PickerRuntimeState>(() => ({
    cashfree: { loading: false },
    razorpay: { loading: false },
    payu: { loading: false },
  }));

  const setLoading = useCallback((provider: Provider, loading: boolean) => {
    setRuntime((prev) => ({
      ...prev,
      [provider]: { ...prev[provider as keyof PickerRuntimeState], loading },
    }));
  }, []);

  const setError = useCallback((provider: Provider, error?: string) => {
    setRuntime((prev) => ({
      ...prev,
      [provider]: { ...prev[provider as keyof PickerRuntimeState], error },
    }));
  }, []);

  const clearError = useCallback((provider: Provider) => {
    setRuntime((prev) => ({
      ...prev,
      [provider]: {
        ...prev[provider as keyof PickerRuntimeState],
        error: undefined,
      },
    }));
  }, []);

  return {
    runtime,
    controls: {
      setLoading,
      setError,
      clearError,
    },
  };
}
