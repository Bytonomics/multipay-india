import type { CashfreeCheckoutPayload } from "./types";
import { loadScript } from "./script-loader";

export async function checkoutCashfree(
  payload: CashfreeCheckoutPayload,
): Promise<void> {
  const mode = payload.environment === "PRODUCTION" ? "production" : "sandbox";
  await loadScript("https://sdk.cashfree.com/js/v3/cashfree.js");

  interface CashfreeInstance {
    checkout(_opts: {
      paymentSessionId: string;
      redirectTarget: "_self";
    }): void;
  }
  interface CashfreeGlobal {
    Cashfree(_opts: { mode: "production" | "sandbox" }): CashfreeInstance;
  }
  const cf = (window as unknown as CashfreeGlobal).Cashfree({ mode });
  cf.checkout({
    paymentSessionId: payload.session_id,
    redirectTarget: "_self",
  });
}
