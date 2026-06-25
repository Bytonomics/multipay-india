import type { CashfreeCheckoutPayload } from "./types";
import { CashfreeMode, Environment } from "./types";
import { loadScript } from "./script-loader";

export async function checkoutCashfree(
  payload: CashfreeCheckoutPayload,
): Promise<void> {
  const mode: CashfreeMode =
    payload.environment === Environment.PRODUCTION
      ? CashfreeMode.PRODUCTION
      : CashfreeMode.SANDBOX;
  await loadScript("https://sdk.cashfree.com/js/v3/cashfree.js");

  interface CashfreeInstance {
    checkout(_opts: {
      paymentSessionId: string;
      redirectTarget: "_self";
    }): void;
  }
  interface CashfreeGlobal {
    Cashfree(_opts: { mode: CashfreeMode }): CashfreeInstance;
  }
  const cf = (window as unknown as CashfreeGlobal).Cashfree({ mode });
  cf.checkout({
    paymentSessionId: payload.session_id,
    redirectTarget: "_self",
  });
}
