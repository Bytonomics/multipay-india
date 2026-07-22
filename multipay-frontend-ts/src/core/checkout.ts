import type {
  CheckoutPayload,
  SubscriptionAuthorizationPayload,
} from "./types";
import { Provider } from "./types";
import { validatePayload, validateSubscriptionPayload } from "./validation";
import { checkoutCashfree, authorizeSubscriptionCashfree } from "./cashfree";
import { checkoutRazorpay, authorizeSubscriptionRazorpay } from "./razorpay";
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

  async authorizeSubscription(
    payload: SubscriptionAuthorizationPayload,
  ): Promise<void> {
    validateSubscriptionPayload(payload);
    switch (payload.provider) {
      case Provider.CASHFREE:
        await authorizeSubscriptionCashfree(payload);
        break;
      case Provider.RAZORPAY:
        authorizeSubscriptionRazorpay(payload);
        break;
      default:
        throw new MultiPayError("Unsupported provider");
    }
  }
}
