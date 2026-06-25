import type { CheckoutPayload } from "./types";
import { Provider } from "./types";
import { validatePayload } from "./validation";
import { checkoutCashfree } from "./cashfree";
import { checkoutRazorpay } from "./razorpay";
import { MultiPayError } from "./errors";

export class MultiPay {
  async checkout(payload: CheckoutPayload): Promise<void> {
    validatePayload(payload);
    switch (payload.provider) {
      case Provider.CASHFREE:
        await checkoutCashfree(payload);
        break;
      case Provider.RAZORPAY:
        checkoutRazorpay(payload);
        break;
      default:
        throw new MultiPayError("Unsupported provider");
    }
  }
}
