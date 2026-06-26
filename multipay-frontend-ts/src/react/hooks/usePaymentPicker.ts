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
    setRuntime((prev) => {
      switch (provider) {
        case Provider.CASHFREE:
          return { ...prev, cashfree: { ...prev.cashfree, loading } };
        case Provider.RAZORPAY:
          return { ...prev, razorpay: { ...prev.razorpay, loading } };
        case Provider.PAYU:
          return { ...prev, payu: { ...prev.payu, loading } };
      }
    });
  }, []);

  const setError = useCallback((provider: Provider, error?: string) => {
    setRuntime((prev) => {
      switch (provider) {
        case Provider.CASHFREE:
          return { ...prev, cashfree: { ...prev.cashfree, error } };
        case Provider.RAZORPAY:
          return { ...prev, razorpay: { ...prev.razorpay, error } };
        case Provider.PAYU:
          return { ...prev, payu: { ...prev.payu, error } };
      }
    });
  }, []);

  const clearError = useCallback((provider: Provider) => {
    setRuntime((prev) => {
      switch (provider) {
        case Provider.CASHFREE:
          return { ...prev, cashfree: { ...prev.cashfree, error: undefined } };
        case Provider.RAZORPAY:
          return { ...prev, razorpay: { ...prev.razorpay, error: undefined } };
        case Provider.PAYU:
          return { ...prev, payu: { ...prev.payu, error: undefined } };
      }
    });
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
