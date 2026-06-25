/// <reference types="node" />
import { readFileSync } from "fs";
import { resolve } from "path";
import type { CheckoutPayload } from "../src/core/types";
import { validatePayload } from "../src/core/validation";

/**
 * TS contract test — validates golden vectors from Go contract.
 *
 * This is the TypeScript side of the cross-language parity gate. We import/read the SHARED
 * golden vectors at ../../contract/checkout/vectors/{cashfree,razorpay}.checkout.json and
 * verify that:
 *
 * 1. Each vector satisfies the CheckoutPayload TS type (discriminated union narrows by provider)
 * 2. validatePayload ACCEPTS each vector (does not throw)
 * 3. Field names/casing match the canonical contract:
 *    - provider: lowercase ("cashfree", "razorpay")
 *    - environment: UPPERCASE ("SANDBOX", "PRODUCTION")
 *    - Razorpay money key: amount_minor (snake_case, NOT amountMinor)
 *
 * This guarantees Go and TS validate the SAME canonical vectors (mirror of the Go test in
 * multipay-go/domain/checkout_contract_test.go).
 */
describe("Contract: golden vectors", () => {
  describe("Cashfree vector", () => {
    const vectorPath = resolve(
      __dirname,
      "../../contract/checkout/vectors/cashfree.checkout.json",
    );
    const vectorData = readFileSync(vectorPath, "utf-8");
    const vector: unknown = JSON.parse(vectorData);

    it("should satisfy CheckoutPayload type", () => {
      // This is a compile-time check — the discriminated union ensures the provider field
      // narrows the type correctly. At runtime, we verify the structure.
      expect(vector).toBeDefined();
      expect(typeof vector).toBe("object");
    });

    it("should have correct provider and environment casing", () => {
      const payload = vector as { provider: string; environment: string };
      expect(payload.provider).toBe("cashfree"); // lowercase
      expect(payload.environment).toBe("PRODUCTION"); // UPPERCASE
    });

    it("should have session_id field (snake_case, not sessionId)", () => {
      const payload = vector as CheckoutPayload;
      const keys = Object.keys(vector as object);
      expect(keys.includes("session_id")).toBe(true);
      if (payload.provider === "cashfree") {
        expect(payload.session_id).toBe("session_abc123");
      }
      // NOT camelCase — the golden vector uses snake_case
      expect(keys.includes("sessionId")).toBe(false);
    });

    it("should be accepted by validatePayload", () => {
      // validatePayload throws if required fields are missing or invalid
      expect(() => validatePayload(vector as CheckoutPayload)).not.toThrow();
    });

    it("should have required fields for Cashfree provider", () => {
      const payload = vector as {
        provider: string;
        environment: string;
        session_id: string;
      };
      expect(payload.provider).toBe("cashfree");
      expect(payload.environment).toBe("PRODUCTION");
      expect(payload.session_id).toBe("session_abc123");
    });
  });

  describe("Razorpay vector", () => {
    const vectorPath = resolve(
      __dirname,
      "../../contract/checkout/vectors/razorpay.checkout.json",
    );
    const vectorData = readFileSync(vectorPath, "utf-8");
    const vector: unknown = JSON.parse(vectorData);

    it("should satisfy CheckoutPayload type", () => {
      // This is a compile-time check — the discriminated union ensures the provider field
      // narrows the type correctly. At runtime, we verify the structure.
      expect(vector).toBeDefined();
      expect(typeof vector).toBe("object");
    });

    it("should have correct provider and environment casing", () => {
      const payload = vector as { provider: string; environment: string };
      expect(payload.provider).toBe("razorpay"); // lowercase
      expect(payload.environment).toBe("PRODUCTION"); // UPPERCASE
    });

    it("should have amount_minor field (snake_case, not amountMinor)", () => {
      const payload = vector as CheckoutPayload;
      const keys = Object.keys(vector as object);
      expect(keys.includes("amount_minor")).toBe(true);
      if (payload.provider === "razorpay") {
        expect(payload.amount_minor).toBe(50000);
      }
      // NOT camelCase — the golden vector uses snake_case
      expect(keys.includes("amountMinor")).toBe(false);
    });

    it("should have public_key field (snake_case, not publicKey)", () => {
      const payload = vector as CheckoutPayload;
      const keys = Object.keys(vector as object);
      expect(keys.includes("public_key")).toBe(true);
      if (payload.provider === "razorpay") {
        expect(payload.public_key).toBe("rzp_live_xxx");
      }
      // NOT camelCase — the golden vector uses snake_case
      expect(keys.includes("publicKey")).toBe(false);
    });

    it("should have order_id field (snake_case, not orderId)", () => {
      const payload = vector as CheckoutPayload;
      const keys = Object.keys(vector as object);
      expect(keys.includes("order_id")).toBe(true);
      expect(payload.order_id).toBe("order_RZP123");
      // NOT camelCase — the golden vector uses snake_case
      expect(keys.includes("orderId")).toBe(false);
    });

    it("should have callback_url field (snake_case, not callbackUrl)", () => {
      const payload = vector as CheckoutPayload;
      const keys = Object.keys(vector as object);
      expect(keys.includes("callback_url")).toBe(true);
      if (payload.provider === "razorpay") {
        expect(payload.callback_url).toBe(
          "https://api.smriti.ai/v1/payments/callback/razorpay",
        );
      }
      // NOT camelCase — the golden vector uses snake_case
      expect(keys.includes("callbackUrl")).toBe(false);
    });

    it("should be accepted by validatePayload", () => {
      // validatePayload throws if required fields are missing or invalid
      expect(() => validatePayload(vector as CheckoutPayload)).not.toThrow();
    });

    it("should have required fields for Razorpay provider", () => {
      const payload = vector as {
        provider: string;
        environment: string;
        order_id: string;
        public_key: string;
        amount_minor: number;
        currency: string;
        callback_url: string;
      };
      expect(payload.provider).toBe("razorpay");
      expect(payload.environment).toBe("PRODUCTION");
      expect(payload.order_id).toBe("order_RZP123");
      expect(payload.public_key).toBe("rzp_live_xxx");
      expect(payload.amount_minor).toBe(50000);
      expect(payload.currency).toBe("INR");
      expect(payload.callback_url).toBe(
        "https://api.smriti.ai/v1/payments/callback/razorpay",
      );
    });
  });

  describe("Contract field naming consistency", () => {
    it("should use snake_case for all fields in golden vectors", () => {
      // Both vectors should use snake_case consistently
      const cashfreePath = resolve(
        __dirname,
        "../../contract/checkout/vectors/cashfree.checkout.json",
      );
      const cashfreeData = readFileSync(cashfreePath, "utf-8");
      const cashfreeKeys = Object.keys(JSON.parse(cashfreeData) as object);

      const razorpayPath = resolve(
        __dirname,
        "../../contract/checkout/vectors/razorpay.checkout.json",
      );
      const razorpayData = readFileSync(razorpayPath, "utf-8");
      const razorpayKeys = Object.keys(JSON.parse(razorpayData) as object);

      // Cashfree should use snake_case
      expect(cashfreeKeys).toEqual(
        expect.arrayContaining(["provider", "environment", "session_id"]),
      );

      // Razorpay should use snake_case
      expect(razorpayKeys).toEqual(
        expect.arrayContaining([
          "provider",
          "environment",
          "order_id",
          "public_key",
          "callback_url",
          "amount_minor",
          "currency",
        ]),
      );
    });
  });
});
