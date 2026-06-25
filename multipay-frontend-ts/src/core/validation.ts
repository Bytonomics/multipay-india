import type { CheckoutPayload } from "./types";
import { Provider } from "./types";
import { MultiPayError } from "./errors";

export function validatePayload(payload: CheckoutPayload): void {
  switch (payload.provider) {
    case Provider.CASHFREE:
      if (!payload.session_id)
        throw new MultiPayError("session_id is required for Cashfree");
      break;
    case Provider.RAZORPAY:
      if (!payload.order_id)
        throw new MultiPayError("order_id is required for Razorpay");
      if (!payload.public_key)
        throw new MultiPayError("public_key is required for Razorpay");
      if (!payload.callback_url)
        throw new MultiPayError("callback_url is required for Razorpay");
      if (payload.amount_minor === undefined || payload.amount_minor === null)
        throw new MultiPayError("amount_minor is required for Razorpay");
      if (payload.amount_minor <= 0)
        throw new MultiPayError("amount_minor must be positive");
      if (!payload.currency)
        throw new MultiPayError("currency is required for Razorpay");
      break;
    default:
      throw new MultiPayError(
        `Provider "${(payload as { provider: string }).provider}" is not yet supported`,
      );
  }
}
