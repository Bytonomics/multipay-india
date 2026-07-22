import { describe, it, expect } from "vitest";
import {
  validatePayload,
  validateSubscriptionPayload,
} from "../src/core/validation";
import { MultiPayError } from "../src/core/errors";
import type {
  CheckoutPayload,
  SubscriptionAuthorizationPayload,
} from "../src/core/types";
import { Provider, Environment } from "../src/core/types";

describe("validatePayload", () => {
  describe("Cashfree validation", () => {
    it("should throw when session_id is missing", () => {
      const payload = {
        provider: Provider.CASHFREE,
        environment: Environment.PRODUCTION,
        // session_id is missing
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "session_id is required for Cashfree",
      );
    });

    it("should not throw for valid Cashfree payload", () => {
      const payload: CheckoutPayload = {
        provider: Provider.CASHFREE,
        session_id: "session_abc123",
        environment: Environment.PRODUCTION,
      };

      expect(() => validatePayload(payload)).not.toThrow();
    });

    it("should not throw for valid Cashfree payload matching golden vector", () => {
      const payload: CheckoutPayload = {
        provider: Provider.CASHFREE,
        session_id: "session_abc123",
        environment: Environment.PRODUCTION,
      };

      expect(() => validatePayload(payload)).not.toThrow();
    });
  });

  describe("Razorpay validation", () => {
    it("should throw when order_id is missing", () => {
      const payload = {
        provider: Provider.RAZORPAY,
        public_key: "public_key_123",
        callback_url: "https://example.com/callback",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
        // order_id is missing
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "order_id is required for Razorpay",
      );
    });

    it("should throw when public_key is missing", () => {
      const payload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        callback_url: "https://example.com/callback",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
        // public_key is missing
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "public_key is required for Razorpay",
      );
    });

    it("should throw when callback_url is missing", () => {
      const payload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        public_key: "public_key_123",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
        // callback_url is missing
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "callback_url is required for Razorpay",
      );
    });

    it("should throw when currency is missing", () => {
      const payload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        public_key: "public_key_123",
        callback_url: "https://example.com/callback",
        amount_minor: 50000,
        environment: Environment.PRODUCTION,
        // currency is missing
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "currency is required for Razorpay",
      );
    });

    it("should throw when amount_minor is missing", () => {
      const payload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        public_key: "public_key_123",
        callback_url: "https://example.com/callback",
        currency: "INR",
        environment: Environment.PRODUCTION,
        // amount_minor is missing
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "amount_minor is required for Razorpay",
      );
    });

    it("should throw when amount_minor is zero", () => {
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        public_key: "public_key_123",
        callback_url: "https://example.com/callback",
        amount_minor: 0,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "amount_minor must be positive",
      );
    });

    it("should throw when amount_minor is negative", () => {
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        public_key: "public_key_123",
        callback_url: "https://example.com/callback",
        amount_minor: -100,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        "amount_minor must be positive",
      );
    });

    it("should not throw for valid Razorpay payload", () => {
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        public_key: "public_key_123",
        callback_url: "https://example.com/callback",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      expect(() => validatePayload(payload)).not.toThrow();
    });

    it("should not throw for valid Razorpay payload matching golden vector", () => {
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_RZP123",
        public_key: "rzp_live_xxx",
        callback_url: "https://api.smriti.ai/v1/payments/callback/razorpay",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      expect(() => validatePayload(payload)).not.toThrow();
    });
  });

  describe("Unsupported provider validation", () => {
    it("should throw for unsupported provider", () => {
      const payload = {
        provider: "unsupported" as unknown as
          | "unsupported"
          | "cashfree"
          | "razorpay",
        order_id: "order_123",
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        'Provider "unsupported" is not yet supported',
      );
    });

    it("should throw for stripe provider (not yet supported)", () => {
      const payload = {
        provider: "stripe" as unknown as "stripe" | "cashfree" | "razorpay",
        order_id: "order_123",
      } as CheckoutPayload;

      expect(() => validatePayload(payload)).toThrow(MultiPayError);
      expect(() => validatePayload(payload)).toThrow(
        'Provider "stripe" is not yet supported',
      );
    });
  });
});

describe("validateSubscriptionPayload", () => {
  it("throws when auth_session_id missing for Cashfree", () => {
    expect(() =>
      validateSubscriptionPayload({
        provider: Provider.CASHFREE,
        environment: Environment.PRODUCTION,
      } as SubscriptionAuthorizationPayload),
    ).toThrow(
      "auth_session_id is required for Cashfree subscription authorization",
    );
  });
  it("throws when auth_link missing for Razorpay", () => {
    expect(() =>
      validateSubscriptionPayload({
        provider: Provider.RAZORPAY,
        environment: Environment.SANDBOX,
      } as SubscriptionAuthorizationPayload),
    ).toThrow("auth_link is required for Razorpay subscription authorization");
  });
  it("passes for valid cashfree/razorpay payloads", () => {
    expect(() =>
      validateSubscriptionPayload({
        provider: Provider.CASHFREE,
        environment: Environment.PRODUCTION,
        auth_session_id: "s",
      }),
    ).not.toThrow();
    expect(() =>
      validateSubscriptionPayload({
        provider: Provider.RAZORPAY,
        environment: Environment.SANDBOX,
        auth_link: "u",
      }),
    ).not.toThrow();
  });
});
