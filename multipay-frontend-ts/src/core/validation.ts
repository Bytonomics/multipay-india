import type {
  CheckoutPayload,
  SubscriptionAuthorizationPayload,
} from "./types";
import { Provider, Environment } from "./types";
import { MultiPayError, ErrorCodes } from "./errors";

function validateEnvironment(environment: unknown): boolean {
  return Object.values(Environment).includes(environment as Environment);
}

export function validatePayload(payload: CheckoutPayload): void {
  switch (payload.provider) {
    case Provider.CASHFREE:
      if (!payload.session_id)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "session_id is required for Cashfree",
        );
      if (!validateEnvironment(payload.environment))
        throw MultiPayError.withCode(
          ErrorCodes.INVALID_ENVIRONMENT,
          "environment must be SANDBOX or PRODUCTION",
        );
      break;
    case Provider.RAZORPAY:
      if (!payload.order_id)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "order_id is required for Razorpay",
        );
      if (!payload.public_key)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "public_key is required for Razorpay",
        );
      if (!payload.callback_url)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "callback_url is required for Razorpay",
        );
      if (payload.amount_minor === undefined || payload.amount_minor === null)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "amount_minor is required for Razorpay",
        );
      if (payload.amount_minor <= 0)
        throw MultiPayError.withCode(
          ErrorCodes.INVALID_PAYLOAD,
          "amount_minor must be positive",
        );
      if (!payload.currency)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "currency is required for Razorpay",
        );
      if (!validateEnvironment(payload.environment))
        throw MultiPayError.withCode(
          ErrorCodes.INVALID_ENVIRONMENT,
          "environment must be SANDBOX or PRODUCTION",
        );
      break;
    default:
      throw MultiPayError.withCode(
        ErrorCodes.INVALID_PROVIDER,
        `Provider "${(payload as { provider: string }).provider}" is not yet supported`,
      );
  }
}

export function validateSubscriptionPayload(
  payload: SubscriptionAuthorizationPayload,
): void {
  switch (payload.provider) {
    case Provider.CASHFREE:
      if (!payload.auth_session_id)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "auth_session_id is required for Cashfree subscription authorization",
        );
      if (!validateEnvironment(payload.environment))
        throw MultiPayError.withCode(
          ErrorCodes.INVALID_ENVIRONMENT,
          "environment must be SANDBOX or PRODUCTION",
        );
      break;
    case Provider.RAZORPAY:
      if (!payload.auth_link)
        throw MultiPayError.withCode(
          ErrorCodes.MISSING_REQUIRED_FIELD,
          "auth_link is required for Razorpay subscription authorization",
        );
      if (!validateEnvironment(payload.environment))
        throw MultiPayError.withCode(
          ErrorCodes.INVALID_ENVIRONMENT,
          "environment must be SANDBOX or PRODUCTION",
        );
      break;
    default:
      throw MultiPayError.withCode(
        ErrorCodes.INVALID_PROVIDER,
        `Provider "${(payload as { provider: string }).provider}" is not yet supported`,
      );
  }
}
