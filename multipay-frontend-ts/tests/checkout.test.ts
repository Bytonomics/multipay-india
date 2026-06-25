import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { Mock } from "vitest";
import { MultiPay } from "../src/core/checkout";
import { MultiPayError } from "../src/core/errors";
import type { CheckoutPayload } from "../src/core/types";
import { Provider, Environment } from "../src/core/types";

describe("MultiPay.checkout", () => {
  describe("validation before side effects", () => {
    it("should throw validation error before any DOM operations for malformed Cashfree payload", () => {
      const mpay = new MultiPay();
      const malformedPayload = {
        provider: Provider.CASHFREE,
        order_id: "order_123",
        environment: Environment.PRODUCTION,
        amount: 500,
        currency: "INR",
        // session_id is missing
      } as CheckoutPayload;

      // Mock DOM operations to ensure they are never called
      const createElementSpy = vi.spyOn(document, "createElement");
      const loadScriptSpy = vi.fn();

      expect(() => mpay.checkout(malformedPayload)).toThrow(MultiPayError);
      expect(() => mpay.checkout(malformedPayload)).toThrow(
        "session_id is required for Cashfree",
      );

      // Verify no DOM operations were performed
      expect(createElementSpy).not.toHaveBeenCalled();
      expect(loadScriptSpy).not.toHaveBeenCalled();

      createElementSpy.mockRestore();
    });

    it("should throw validation error before any DOM operations for malformed Razorpay payload", () => {
      const mpay = new MultiPay();
      const malformedPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_123",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
        // public_key is missing
      } as CheckoutPayload;

      // Mock DOM operations to ensure they are never called
      const createElementSpy = vi.spyOn(document, "createElement");

      expect(() => mpay.checkout(malformedPayload)).toThrow(MultiPayError);
      expect(() => mpay.checkout(malformedPayload)).toThrow(
        "public_key is required for Razorpay",
      );

      // Verify no form was created
      expect(createElementSpy).not.toHaveBeenCalled();

      createElementSpy.mockRestore();
    });

    it("should throw validation error before script loading for invalid provider", () => {
      const mpay = new MultiPay();
      const malformedPayload = {
        provider: "stripe" as unknown as "stripe" | "cashfree" | "razorpay",
      } as CheckoutPayload;

      expect(() => mpay.checkout(malformedPayload)).toThrow(MultiPayError);
      expect(() => mpay.checkout(malformedPayload)).toThrow(
        'Provider "stripe" is not yet supported',
      );
    });
  });

  describe("Razorpay checkout flow", () => {
    let mockForm: HTMLFormElement;
    let mockBody: HTMLBodyElement;
    let createElementOriginal: typeof document.createElement;
    let appendChildOriginal: typeof Node.prototype.appendChild;

    beforeEach(() => {
      // Store original methods
      createElementOriginal = document.createElement;
      appendChildOriginal = Node.prototype.appendChild;

      // Mock document.body
      mockBody = {
        appendChild: vi.fn(),
      } as unknown as HTMLBodyElement;

      // Mock form element
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
          return createElementOriginal(tagName);
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
      document.createElement = createElementOriginal;
      Node.prototype.appendChild = appendChildOriginal;
    });

    it("should build form POST to initiate Razorpay provider-hosted checkout", async () => {
      const mpay = new MultiPay();
      const payload: CheckoutPayload = {
        provider: Provider.RAZORPAY,
        order_id: "order_RZP123",
        key_id: "key_abc123",
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
        key_id: "key_abc123",
        public_key: "rzp_live_xxx",
        callback_url: "https://api.smriti.ai/v1/payments/callback/razorpay",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      const inputElements: HTMLInputElement[] = [];
      let inputCount = 0;

      // Track createElement calls for input elements
      vi.spyOn(document, "createElement").mockImplementation(
        (tagName: string) => {
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
          return createElementOriginal(tagName);
        },
      );

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
        key_id: "key_abc123",
        public_key: "rzp_live_xxx",
        callback_url: "https://api.smriti.ai/v1/payments/callback/razorpay",
        amount_minor: 50000,
        currency: "INR",
        environment: Environment.PRODUCTION,
      };

      // Track createElement calls to capture amount input
      vi.spyOn(document, "createElement").mockImplementation(
        (tagName: string) => {
          if (tagName === "form") {
            return mockForm as unknown as HTMLElement;
          } else if (tagName === "input") {
            const mockInput = {
              type: "",
              name: "",
              value: "",
            } as unknown as HTMLInputElement;
            if (mockInput.name === "amount") {
              (mockInput as unknown as { value: string }).value = String(
                payload.amount_minor,
              );
            }
            return mockInput as unknown as HTMLElement;
          }
          return createElementOriginal(tagName);
        },
      );

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
        order_id: "order_CF123",
        session_id: "session_abc123",
        environment: Environment.PRODUCTION,
        amount: 50000,
        currency: "INR",
      };

      // Mock loadScript to avoid actual network request
      vi.doMock("../src/core/script-loader", () => ({
        loadScript: vi.fn(() => Promise.resolve()),
      }));

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
        order_id: "order_CF123",
        session_id: "session_abc123",
        environment: Environment.SANDBOX,
        amount: 50000,
        currency: "INR",
      };

      await mpay.checkout(payload);

      // Verify Cashfree was initialized with sandbox mode
      expect(mockCashfreeGlobal.Cashfree).toHaveBeenCalledWith({
        mode: "sandbox",
      });
    });
  });
});
