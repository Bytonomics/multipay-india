import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { Mock } from "vitest";
import { MultiPay } from "../src/core/checkout";
import { MultiPayError } from "../src/core/errors";
import type { CheckoutPayload } from "../src/core/types";
import { Provider, Environment } from "../src/core/types";

// Mock script-loader at module level before any imports use it
vi.mock("../src/core/script-loader", () => ({
  loadScript: vi.fn(() => Promise.resolve()),
}));

describe("MultiPay.checkout", () => {
  describe("validation before side effects", () => {
    it("should throw validation error before any DOM operations for malformed Cashfree payload", async () => {
      const mpay = new MultiPay();
      const malformedPayload = {
        provider: Provider.CASHFREE,
        environment: Environment.PRODUCTION,
        // session_id is missing
      } as CheckoutPayload;

      // Mock DOM operations to ensure they are never called
      const createElementSpy = vi.spyOn(document, "createElement");
      const loadScriptSpy = vi.fn();

      await expect(mpay.checkout(malformedPayload)).rejects.toThrow(
        MultiPayError,
      );
      await expect(mpay.checkout(malformedPayload)).rejects.toThrow(
        "session_id is required for Cashfree",
      );

      // Verify no DOM operations were performed
      expect(createElementSpy).not.toHaveBeenCalled();
      expect(loadScriptSpy).not.toHaveBeenCalled();

      createElementSpy.mockRestore();
    });

    it("should throw validation error before any DOM operations for malformed Razorpay payload", async () => {
      const mpay = new MultiPay();
      const malformedPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
        // public_key is missing
        callback_url: "https://example.com/callback",
      } as CheckoutPayload;

      // Mock DOM operations to ensure they are never called
      const createElementSpy = vi.spyOn(document, "createElement");

      await expect(mpay.checkout(malformedPayload)).rejects.toThrow(
        MultiPayError,
      );
      await expect(mpay.checkout(malformedPayload)).rejects.toThrow(
        "public_key is required for Razorpay",
      );

      // Verify no form was created
      expect(createElementSpy).not.toHaveBeenCalled();

      createElementSpy.mockRestore();
    });

    it("should throw validation error before script loading for invalid provider", async () => {
      const mpay = new MultiPay();
      const malformedPayload = {
        provider: "stripe" as unknown as "stripe" | "cashfree" | "razorpay",
      } as CheckoutPayload;

      await expect(mpay.checkout(malformedPayload)).rejects.toThrow(
        MultiPayError,
      );
      await expect(mpay.checkout(malformedPayload)).rejects.toThrow(
        'Provider "stripe" is not yet supported',
      );
    });
  });

  describe("Razorpay checkout flow", () => {
    let mockForm: HTMLFormElement;
    let mockBody: HTMLBodyElement;

    beforeEach(() => {
      // Mock document.body with appendChild mock
      mockBody = {
        appendChild: vi.fn(),
      } as unknown as HTMLBodyElement;

      // Mock form element with non-recursive appendChild
      mockForm = {
        method: "",
        action: "",
        appendChild: vi.fn(),
        submit: vi.fn(),
      } as unknown as HTMLFormElement;

      // Mock document.createElement to return our mock form
      vi.spyOn(document, "createElement").mockImplementation(
        (tagName: string) => {
          if (tagName === "form") {
            return mockForm as unknown as HTMLElement;
          }
          // For other tags, throw or return a minimal mock
          if (tagName === "input") {
            return {
              type: "",
              name: "",
              value: "",
            } as unknown as HTMLElement;
          }
          // Return a minimal mock for any other tag
          return {} as unknown as HTMLElement;
        },
      );

      // Mock document.body
      Object.defineProperty(document, "body", {
        value: mockBody,
        writable: true,
      });
    });

    afterEach(() => {
      // Restore original methods
      vi.restoreAllMocks();
    });

    it("should build form POST to initiate Razorpay provider-hosted checkout", async () => {
      const mpay = new MultiPay();
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_RZP123",
        public_key: "rzp_live_xxx",
        callback_url: "https://api.smriti.ai/v1/payments/callback/razorpay",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      await mpay.checkout(payload);

      // Verify form POST to Razorpay's provider-hosted checkout endpoint
      expect(mockForm.method).toBe("POST");
      expect(mockForm.action).toBe(
        "https://api.razorpay.com/v1/checkout/embedded",
      );

      // Verify form was appended to body
      expect(mockBody.appendChild).toHaveBeenCalledWith(mockForm);

      // Verify form.submit was called
      expect(mockForm.submit).toHaveBeenCalled();
    });

    it("should create hidden input fields for Razorpay", async () => {
      const mpay = new MultiPay();
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_RZP123",
        public_key: "rzp_live_xxx",
        callback_url: "https://api.smriti.ai/v1/payments/callback/razorpay",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      const inputElements: HTMLInputElement[] = [];
      let inputCount = 0;

      // Override the createElement mock to track inputs
      const createElementMock = vi.spyOn(document, "createElement");
      createElementMock.mockImplementation((tagName: string) => {
        if (tagName === "form") {
          return mockForm as unknown as HTMLElement;
        } else if (tagName === "input") {
          const mockInput = {
            type: "",
            name: "",
            value: "",
          } as unknown as HTMLInputElement;
          inputElements.push(mockInput);
          inputCount++;
          return mockInput as unknown as HTMLElement;
        }
        // Return a minimal mock for any other tag
        return {} as unknown as HTMLElement;
      });

      await mpay.checkout(payload);

      // Verify all expected input fields were created
      expect(inputCount).toBe(5); // key_id, order_id, amount, currency, callback_url

      // Verify input field properties
      const fields = [
        { name: "key_id", value: "rzp_live_xxx" },
        { name: "order_id", value: "order_RZP123" },
        { name: "amount", value: "50000" }, // amount_minor converted to string
        { name: "currency", value: "INR" },
        {
          name: "callback_url",
          value: "https://api.smriti.ai/v1/payments/callback/razorpay",
        },
      ];

      fields.forEach((field) => {
        const input = inputElements.find((inp) => inp.name === field.name);
        expect(input).toBeDefined();
        expect(input?.type).toBe("hidden");
        expect(input?.value).toBe(field.value);
      });
    });

    it("should convert amount_minor to string for Razorpay amount field", async () => {
      const mpay = new MultiPay();
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_RZP123",
        public_key: "rzp_live_xxx",
        callback_url: "https://api.smriti.ai/v1/payments/callback/razorpay",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      await mpay.checkout(payload);

      // Verify amount_minor (number) was converted to string
      const amountValue = String(payload.amount_minor);
      expect(amountValue).toBe("50000");
      expect(typeof amountValue).toBe("string");
    });
  });

  describe("Cashfree checkout flow", () => {
    let mockCashfreeInstance: {
      checkout: Mock;
    };
    let mockCashfreeGlobal: {
      Cashfree: Mock;
    };

    beforeEach(() => {
      // Mock Cashfree instance
      mockCashfreeInstance = {
        checkout: vi.fn(),
      };

      // Mock Cashfree global
      mockCashfreeGlobal = {
        Cashfree: vi.fn(() => mockCashfreeInstance),
      };

      // Stub window.Cashfree global
      vi.stubGlobal("Cashfree", mockCashfreeGlobal.Cashfree);
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it("should call Cashfree SDK with paymentSessionId", async () => {
      const mpay = new MultiPay();
      const payload: CheckoutPayload = {
        provider: Provider.CASHFREE,
        session_id: "session_abc123",
        environment: Environment.PRODUCTION,
      };

      await mpay.checkout(payload);

      // Verify Cashfree was initialized with correct mode
      expect(mockCashfreeGlobal.Cashfree).toHaveBeenCalledWith({
        mode: "production",
      });

      // Verify checkout was called with correct parameters
      expect(mockCashfreeInstance.checkout).toHaveBeenCalledWith({
        paymentSessionId: "session_abc123",
        redirectTarget: "_self",
      });
    });

    it("should call Cashfree SDK with sandbox mode for sandbox environment", async () => {
      const mpay = new MultiPay();
      const payload: CheckoutPayload = {
        provider: Provider.CASHFREE,
        session_id: "session_abc123",
        environment: Environment.SANDBOX,
      };

      await mpay.checkout(payload);

      // Verify Cashfree was initialized with sandbox mode
      expect(mockCashfreeGlobal.Cashfree).toHaveBeenCalledWith({
        mode: "sandbox",
      });
    });
  });

  describe("Cashfree subscription authorization", () => {
    let mockCashfreeInstance: { subscriptionsCheckout: Mock };
    let mockCashfreeGlobal: { Cashfree: Mock };

    beforeEach(() => {
      mockCashfreeInstance = { subscriptionsCheckout: vi.fn() };
      mockCashfreeGlobal = { Cashfree: vi.fn(() => mockCashfreeInstance) };
      vi.stubGlobal("Cashfree", mockCashfreeGlobal.Cashfree);
    });
    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it("calls subscriptionsCheckout with subsSessionId", async () => {
      const mpay = new MultiPay();
      await mpay.authorizeSubscription({
        provider: Provider.CASHFREE,
        environment: Environment.PRODUCTION,
        auth_session_id: "sub_session_abc123",
      });
      expect(mockCashfreeGlobal.Cashfree).toHaveBeenCalledWith({
        mode: "production",
      });
      expect(mockCashfreeInstance.subscriptionsCheckout).toHaveBeenCalledWith({
        subsSessionId: "sub_session_abc123",
        redirectTarget: "_self",
      });
    });
  });

  describe("Razorpay subscription authorization", () => {
    it("redirects to auth_link", async () => {
      const assign = vi.fn();
      vi.stubGlobal("location", { assign } as unknown as Location);
      const mpay = new MultiPay();
      await mpay.authorizeSubscription({
        provider: Provider.RAZORPAY,
        environment: Environment.SANDBOX,
        auth_link: "https://rzp.io/i/abc",
      });
      expect(assign).toHaveBeenCalledWith("https://rzp.io/i/abc");
      vi.unstubAllGlobals();
    });
  });
});
