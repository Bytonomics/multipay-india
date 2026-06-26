import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import type React from "react";
import { PaymentPicker } from "../src/react/PaymentPicker";
import type {
  PaymentPickerProps,
  PickerProviders,
  PickerControls,
} from "../src/react/types";
import { Provider, PickerVariant, PickerTheme } from "../src/core/types";

describe("PaymentPicker Component", () => {
  // Mock window.matchMedia for theme detection
  const mockMatchMedia = vi.fn();
  let originalMatchMedia: typeof window.matchMedia;

  beforeEach(() => {
    // Store original matchMedia
    originalMatchMedia = window.matchMedia;

    // Mock matchMedia implementation
    mockMatchMedia.mockImplementation((query: string) => ({
      matches: query.includes("dark"),
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    // @ts-ignore - replacing window method for testing
    window.matchMedia = mockMatchMedia;
  });

  afterEach(() => {
    // Restore original matchMedia
    // @ts-ignore
    window.matchMedia = originalMatchMedia;
    vi.clearAllMocks();
  });

  const defaultProviders: PickerProviders = {
    cashfree: { label: "Cashfree", visible: true, enabled: true },
    razorpay: { label: "Razorpay", visible: true, enabled: true },
    payu: { label: "PayU", visible: false, enabled: true },
  };

  const defaultProps: PaymentPickerProps = {
    payment: {
      amountMinor: 5000,
      currency: "INR",
      providers: defaultProviders,
      default: Provider.CASHFREE,
    },
    appearance: {
      variant: PickerVariant.INTERACTIVE_MATRIX,
      theme: PickerTheme.LIGHT,
    },
    onSelect: vi.fn(),
  };

  describe("Assertion 1: Default picker renders exactly TWO aggregators", () => {
    it("should render Cashfree and Razorpay when both are enabled", () => {
      render(<PaymentPicker {...defaultProps} />);

      // Verify both providers are rendered
      expect(screen.getByText("Cashfree")).toBeInTheDocument();
      expect(screen.getByText("Razorpay")).toBeInTheDocument();

      // Verify default selection
      const cashfreeButton = screen
        .getByText("Cashfree")
        .closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute("aria-pressed", "true");
    });

    it("should NOT render PayU when not in payment.providers", () => {
      render(<PaymentPicker {...defaultProps} />);

      // Verify PayU is not rendered
      expect(screen.queryByText("PayU")).not.toBeInTheDocument();
    });

    it("should pre-select Cashfree when default is Provider.CASHFREE", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.CASHFREE,
        },
      };
      render(<PaymentPicker {...props} />);

      const cashfreeButton = screen
        .getByText("Cashfree")
        .closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute("aria-pressed", "true");
    });

    it("should pre-select Razorpay when default is Provider.RAZORPAY", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.RAZORPAY,
        },
      };
      render(<PaymentPicker {...props} />);

      const razorpayButton = screen
        .getByText("Razorpay")
        .closest('[role="button"]');
      expect(razorpayButton).toHaveAttribute("aria-pressed", "true");
    });

    it("should ignore invalid default value and select first enabled provider", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: "invalid" as unknown as Provider,
          providers: {
            cashfree: { label: "Cashfree", visible: true, enabled: true },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
      };
      render(<PaymentPicker {...props} />);

      // Should default to first enabled provider (cashfree)
      const cashfreeButton = screen
        .getByText("Cashfree")
        .closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute("aria-pressed", "true");

      // Verify console was warned about invalid default
      expect(consoleSpy).toHaveBeenCalled();

      consoleSpy.mockRestore();
    });

    it("should ignore disabled provider in default and select first enabled provider", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.CASHFREE,
          providers: {
            cashfree: { label: "Cashfree", visible: true, enabled: false },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
      };
      render(<PaymentPicker {...props} />);

      // Should default to first enabled provider (razorpay)
      const razorpayButton = screen
        .getByText("Razorpay")
        .closest('[role="button"]');
      expect(razorpayButton).toHaveAttribute("aria-pressed", "true");

      // Disabled card should still render but be disabled
      const cashfreeButton = screen
        .getByText("Cashfree")
        .closest('[role="button"]');
      expect(cashfreeButton).toBeDisabled();

      consoleSpy.mockRestore();
    });

    it("should ignore absent provider in default and select first enabled provider", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.PAYU,
          providers: {
            cashfree: { label: "Cashfree", visible: true, enabled: true },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
      };
      render(<PaymentPicker {...props} />);

      // Should default to first enabled provider (cashfree)
      const cashfreeButton = screen
        .getByText("Cashfree")
        .closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute("aria-pressed", "true");

      consoleSpy.mockRestore();
    });
  });

  describe("Assertion 2: Theme rendering (light, dark, auto)", () => {
    const variants = [
      PickerVariant.DYNAMIC_STACK,
      PickerVariant.INTERACTIVE_MATRIX,
      PickerVariant.SECURE_VAULT,
      PickerVariant.NEUMORPHIC_FLOW,
    ];

    variants.forEach((variant) => {
      describe(`${variant} variant`, () => {
        it("should render with light theme", () => {
          const props = {
            ...defaultProps,
            appearance: {
              ...defaultProps.appearance,
              variant,
              theme: PickerTheme.LIGHT,
            },
          };

          const { container } = render(<PaymentPicker {...props} />);

          // Verify data-theme attribute is set to light
          const picker = container.querySelector("[data-theme]");
          expect(picker).toHaveAttribute("data-theme", "light");
        });

        it("should render with dark theme", () => {
          const props = {
            ...defaultProps,
            appearance: {
              ...defaultProps.appearance,
              variant,
              theme: PickerTheme.DARK,
            },
          };

          const { container } = render(<PaymentPicker {...props} />);

          // Verify data-theme attribute is set to dark
          const picker = container.querySelector("[data-theme]");
          expect(picker).toHaveAttribute("data-theme", "dark");
        });

        it("should render with auto theme reflecting system preference (dark)", () => {
          // Mock system prefers dark
          mockMatchMedia.mockImplementation((query: string) => ({
            matches: query.includes("dark"),
            media: query,
            onchange: null,
            addListener: vi.fn(),
            removeListener: vi.fn(),
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
            dispatchEvent: vi.fn(),
          }));

          const props = {
            ...defaultProps,
            appearance: {
              ...defaultProps.appearance,
              variant,
              theme: PickerTheme.AUTO,
            },
          };

          const { container } = render(<PaymentPicker {...props} />);

          // Verify data-theme reflects system preference
          const picker = container.querySelector("[data-theme]");
          expect(picker).toHaveAttribute("data-theme", "dark");
        });

        it("should render with auto theme reflecting system preference (light)", () => {
          // Mock system prefers light
          mockMatchMedia.mockImplementation((query: string) => ({
            matches: !query.includes("dark"),
            media: query,
            onchange: null,
            addListener: vi.fn(),
            removeListener: vi.fn(),
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
            dispatchEvent: vi.fn(),
          }));

          const props = {
            ...defaultProps,
            appearance: {
              ...defaultProps.appearance,
              variant,
              theme: PickerTheme.AUTO,
            },
          };

          const { container } = render(<PaymentPicker {...props} />);

          // Verify data-theme reflects system preference
          const picker = container.querySelector("[data-theme]");
          expect(picker).toHaveAttribute("data-theme", "light");
        });

        it("should update theme when system preference changes (auto mode)", async () => {
          let mockMediaQuery: MediaQueryList | null = null;
          const callbacks: Array<
            (this: MediaQueryList, ev: MediaQueryListEvent) => void
          > = [];

          // Initial state: prefers light
          mockMatchMedia.mockImplementation((query: string): MediaQueryList => {
            const mq: MediaQueryList = {
              matches: false,
              media: query,
              onchange: null,
              addListener: vi.fn(),
              removeListener: vi.fn(),
              addEventListener(
                _event: string,
                listener: EventListenerOrEventListenerObject,
              ): void {
                if (typeof listener === "function") {
                  callbacks.push(
                    listener as (
                      this: MediaQueryList,
                      ev: MediaQueryListEvent,
                    ) => void,
                  );
                }
              },
              removeEventListener: vi.fn(),
              dispatchEvent: vi.fn(),
            };
            mockMediaQuery = mq;
            return mq;
          });

          const props = {
            ...defaultProps,
            appearance: {
              ...defaultProps.appearance,
              variant,
              theme: PickerTheme.AUTO,
            },
          };

          const { container } = render(<PaymentPicker {...props} />);

          // Initial theme should be light
          let picker = container.querySelector("[data-theme]");
          expect(picker).toHaveAttribute("data-theme", "light");

          // Simulate system preference change to dark
          const lastCallback = callbacks[callbacks.length - 1];
          if (lastCallback && mockMediaQuery) {
            const mockEvent = {
              matches: true,
              media: "screen",
            } as MediaQueryListEvent;
            lastCallback.call(mockMediaQuery, mockEvent);
          }

          // Wait for state update
          await waitFor(() => {
            picker = container.querySelector("[data-theme]");
            expect(picker).toHaveAttribute("data-theme", "dark");
          });
        });
      });
    });
  });

  describe("Assertion 3: ONE card/slot per aggregator in every variant", () => {
    it(`${PickerVariant.DYNAMIC_STACK}: should render exactly one card per provider with expandable alternatives`, async () => {
      const props = {
        ...defaultProps,
        appearance: {
          ...defaultProps.appearance,
          variant: PickerVariant.DYNAMIC_STACK,
        },
      };

      render(<PaymentPicker {...props} />);

      // DynamicStack shows primary card + accordion with alternatives
      // Initially: 1 primary card visible
      expect(screen.getByText("Cashfree")).toBeInTheDocument();

      // Expand accordion to reveal alternative cards
      const accordionTrigger = screen.getByText(/Show Alternative Options/i);
      fireEvent.click(accordionTrigger);

      // After expanding: all alternative cards visible
      await waitFor(() => {
        expect(screen.getByText("Razorpay")).toBeInTheDocument();
      });

      // Verify no method-specific text
      expect(screen.queryByText("UPI")).not.toBeInTheDocument();
      expect(screen.queryByText("Cards")).not.toBeInTheDocument();
    });

    it(`${PickerVariant.INTERACTIVE_MATRIX}: should render exactly one card per provider (no method-split elements)`, () => {
      const props = {
        ...defaultProps,
        appearance: {
          ...defaultProps.appearance,
          variant: PickerVariant.INTERACTIVE_MATRIX,
        },
      };

      render(<PaymentPicker {...props} />);

      // InteractiveMatrix shows all visible providers as cards
      expect(screen.getByText("Cashfree")).toBeInTheDocument();
      expect(screen.getByText("Razorpay")).toBeInTheDocument();

      // Verify no method-specific text
      expect(screen.queryByText("UPI")).not.toBeInTheDocument();
      expect(screen.queryByText("Cards")).not.toBeInTheDocument();
    });

    it(`${PickerVariant.SECURE_VAULT}: should render exactly one card per provider (no method-split elements)`, () => {
      const props = {
        ...defaultProps,
        appearance: {
          ...defaultProps.appearance,
          variant: PickerVariant.SECURE_VAULT,
        },
      };

      render(<PaymentPicker {...props} />);

      // SecureVault shows all visible providers as slots
      expect(screen.getByText("Cashfree")).toBeInTheDocument();
      expect(screen.getByText("Razorpay")).toBeInTheDocument();

      // Verify no method-specific text
      expect(screen.queryByText("UPI")).not.toBeInTheDocument();
      expect(screen.queryByText("Cards")).not.toBeInTheDocument();
    });

    it(`${PickerVariant.NEUMORPHIC_FLOW}: should render exactly one segment per provider (no method-split elements)`, () => {
      const props = {
        ...defaultProps,
        appearance: {
          ...defaultProps.appearance,
          variant: PickerVariant.NEUMORPHIC_FLOW,
        },
      };

      render(<PaymentPicker {...props} />);

      // NeumorphicFlow shows providers as toggle segments
      expect(screen.getByText("Cashfree")).toBeInTheDocument();
      expect(screen.getByText("Razorpay")).toBeInTheDocument();

      // Verify no method-specific text
      expect(screen.queryByText("UPI")).not.toBeInTheDocument();
      expect(screen.queryByText("Cards")).not.toBeInTheDocument();
    });
  });

  describe("Assertion 4: default Provider behavior", () => {
    it("should pre-select the named ENABLED aggregator", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.RAZORPAY,
        },
      };

      render(<PaymentPicker {...props} />);

      const razorpayCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Razorpay"));

      expect(razorpayCard).toBeDefined();
      expect(razorpayCard).toHaveAttribute("aria-pressed", "true");
    });

    it("should select first enabled provider when default is invalid value", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: "invalid_provider" as unknown as Provider,
        },
      };

      render(<PaymentPicker {...props} />);

      // Should fall back to first enabled provider
      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));
      expect(cashfreeCard).toHaveAttribute("aria-pressed", "true");

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should select first enabled provider when default is disabled provider", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.CASHFREE,
          providers: {
            cashfree: { label: "Cashfree", visible: true, enabled: false },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
      };

      render(<PaymentPicker {...props} />);

      // Should select first enabled provider (razorpay)
      const razorpayCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Razorpay"));
      expect(razorpayCard).toHaveAttribute("aria-pressed", "true");

      // Disabled card should render but be disabled
      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));
      expect(cashfreeCard).toBeDisabled();

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should select first enabled provider when default is absent from visible providers", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.PAYU,
          providers: {
            cashfree: { label: "Cashfree", visible: true, enabled: true },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
      };

      render(<PaymentPicker {...props} />);

      // Should fall back to first enabled provider (cashfree)
      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));
      expect(cashfreeCard).toHaveAttribute("aria-pressed", "true");

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should emit dev warning for invalid/disabled/absent default", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: "invalid" as unknown as Provider,
        },
      };

      render(<PaymentPicker {...props} />);

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });
  });

  describe("Assertion 5: Click behavior and onSelect callback", () => {
    it("should call onSelect with canonical Provider.CASHFREE when clicking enabled card", async () => {
      const mockOnSelect = vi.fn();

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.RAZORPAY,
        },
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));

      if (cashfreeCard) {
        fireEvent.click(cashfreeCard);

        await waitFor(() => {
          expect(mockOnSelect).toHaveBeenCalledWith(Provider.CASHFREE);
        });
      }
    });

    it("should call onSelect with canonical Provider.RAZORPAY when clicking enabled card", async () => {
      const mockOnSelect = vi.fn();

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.CASHFREE,
        },
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      const razorpayCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Razorpay"));

      if (razorpayCard) {
        fireEvent.click(razorpayCard);

        await waitFor(() => {
          expect(mockOnSelect).toHaveBeenCalledWith(Provider.RAZORPAY);
        });
      }
    });

    it("should NOT call onSelect when clicking disabled card", async () => {
      const mockOnSelect = vi.fn();

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          providers: {
            cashfree: {
              label: "Cashfree",
              visible: true,
              enabled: false,
              disabledMessage: "Maintenance",
            },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));

      if (cashfreeCard) {
        fireEvent.click(cashfreeCard);

        // onSelect should not be called
        expect(mockOnSelect).not.toHaveBeenCalled();

        // Disabled message should be visible
        expect(screen.getByText("Maintenance")).toBeInTheDocument();
      }
    });

    it("should show disabledMessage for disabled card", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          providers: {
            cashfree: {
              label: "Cashfree",
              visible: true,
              enabled: false,
              disabledMessage: "Temporarily unavailable",
            },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: true },
          },
        },
      };

      render(<PaymentPicker {...props} />);

      expect(screen.getByText("Temporarily unavailable")).toBeInTheDocument();
    });

    it("should never emit Provider.PAYU from onSelect callback (PayU is code-only placeholder)", async () => {
      const mockOnSelect = vi.fn();

      // PayU is a code-only placeholder - even if included in fixtures, it should NOT be visible
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          providers: {
            cashfree: { label: "Cashfree", visible: true, enabled: true },
            razorpay: { label: "Razorpay", visible: true, enabled: true },
            payu: { label: "PayU", visible: false, enabled: false }, // PayU never shown, never enabled
          },
        },
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      // Verify PayU is NOT rendered in the DOM
      expect(screen.queryByText("PayU")).not.toBeInTheDocument();

      // Click all available (enabled) provider cards
      const enabledCards = screen
        .getAllByRole("button")
        .filter(
          (button) =>
            button.textContent?.includes("Cashfree") ||
            button.textContent?.includes("Razorpay"),
        );

      enabledCards.forEach((card) => {
        fireEvent.click(card);
      });

      await waitFor(() => {
        // Verify onSelect was only called with canonical providers
        mockOnSelect.mock.calls.forEach((call) => {
          expect([Provider.CASHFREE, Provider.RAZORPAY]).toContain(call[0]);
        });

        // PayU should never be emitted (and never be in the fixture as visible)
        expect(mockOnSelect).not.toHaveBeenCalledWith(Provider.PAYU);
      });
    });
  });

  describe("Assertion 6: Formatted total and tax disclaimer", () => {
    it('should display formatted total (e.g., "₹50.00" from amountMinor:5000)', () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: 5000, // ₹50.00
          currency: "INR",
        },
      };

      render(<PaymentPicker {...props} />);

      expect(screen.getByText("₹50.00")).toBeInTheDocument();
    });

    it('should display formatted total for USD (e.g., "$12.34" from amountMinor:1234)', () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: 1234, // $12.34
          currency: "USD",
        },
      };

      render(<PaymentPicker {...props} />);

      expect(screen.getByText("$12.34")).toBeInTheDocument();
    });

    it("should display default tax disclaimer", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: 5000,
          currency: "INR",
        },
        appearance: {
          ...defaultProps.appearance,
          taxNote: undefined, // Use default
        },
      };

      render(<PaymentPicker {...props} />);

      expect(
        screen.getByText(
          "Final taxes are added at checkout by the payment provider.",
        ),
      ).toBeInTheDocument();
    });

    it("should display custom taxNote when appearance.taxNote is provided", () => {
      const customNote = "Including GST and service charges";
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: 5000,
          currency: "INR",
        },
        appearance: {
          ...defaultProps.appearance,
          taxNote: customNote,
        },
      };

      render(<PaymentPicker {...props} />);

      expect(screen.getByText(customNote)).toBeInTheDocument();
      expect(
        screen.queryByText(
          "Final taxes are added at checkout by the payment provider.",
        ),
      ).not.toBeInTheDocument();
    });

    it("should render total and tax note near CTA section", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: 5000,
          currency: "INR",
        },
      };

      const { container } = render(<PaymentPicker {...props} />);

      // Verify total and tax note are in the document
      expect(screen.getByText("₹50.00")).toBeInTheDocument();
      expect(
        screen.getByText(
          "Final taxes are added at checkout by the payment provider.",
        ),
      ).toBeInTheDocument();

      // Verify they appear in a structured section (not just anywhere)
      const totalSection = container.querySelector(
        '.totalSection, .total-section, [class*="total"]',
      );
      expect(totalSection).toBeInTheDocument();
    });
  });

  describe("Assertion 7: PickerControls via ref", () => {
    it("should expose the imperative control surface via ref", async () => {
      const mockRef: React.RefObject<PickerControls> = { current: null };

      const props = {
        ...defaultProps,
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      // Provider enable/disable is declarative (payment.providers[id].enabled);
      // the ref exposes only genuinely imperative actions.
      if (mockRef.current) {
        expect(typeof mockRef.current.selectProvider).toBe("function");
        expect(typeof mockRef.current.getSelectedProvider).toBe("function");
        expect(typeof mockRef.current.isSelected).toBe("function");
        expect(typeof mockRef.current.focus).toBe("function");
        expect(typeof mockRef.current.blur).toBe("function");
      }
    });

    it("should provide selectProvider control method via ref", async () => {
      const mockOnSelect = vi.fn();
      const mockRef: React.RefObject<PickerControls> = { current: null };

      const props = {
        ...defaultProps,
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        // Programmatically select razorpay
        mockRef.current.selectProvider(Provider.RAZORPAY);

        await waitFor(() => {
          expect(mockOnSelect).toHaveBeenCalledWith(Provider.RAZORPAY);
        });
      }
    });

    it("should provide getSelectedProvider control method via ref", async () => {
      const mockRef: React.RefObject<PickerControls> = { current: null };

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.RAZORPAY,
        },
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        const selected = mockRef.current.getSelectedProvider();
        expect(selected).toBe(Provider.RAZORPAY);
      }
    });

    it("should provide isSelected control method via ref", async () => {
      const mockRef: React.RefObject<PickerControls> = { current: null };

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          default: Provider.CASHFREE,
        },
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        expect(mockRef.current.isSelected(Provider.CASHFREE)).toBe(true);
        expect(mockRef.current.isSelected(Provider.RAZORPAY)).toBe(false);
      }
    });

    it("should provide focus and blur control methods via ref", async () => {
      const mockRef: React.RefObject<PickerControls> = { current: null };

      const props = {
        ...defaultProps,
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        // Focus first provider card
        mockRef.current.focus();

        // Verify focus was called (implementation-specific)
        // Blur is also available
        expect(typeof mockRef.current.focus).toBe("function");
        expect(typeof mockRef.current.blur).toBe("function");
      }
    });
  });

  describe("Error handling and edge cases", () => {
    it("should throw error when payment.amountMinor is missing", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: undefined as unknown as number, // Missing - should throw
        },
      };

      expect(() => render(<PaymentPicker {...props} />)).toThrow();
    });

    it("should throw error when payment.currency is missing", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          currency: "", // Invalid
        },
      };

      expect(() => render(<PaymentPicker {...props} />)).toThrow();
    });

    it("should handle async onSelect with loading state", async () => {
      const mockOnSelect = vi.fn(async () => {
        // Simulate async operation
        await new Promise((resolve) => setTimeout(resolve, 100));
      });

      const props = {
        ...defaultProps,
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));

      if (cashfreeCard) {
        fireEvent.click(cashfreeCard);

        // Should show loading state
        await waitFor(() => {
          expect(
            screen.getByText(/redirecting|processing/i),
          ).toBeInTheDocument();
        });

        // Should complete after async operation
        await waitFor(
          () => {
            expect(mockOnSelect).toHaveBeenCalled();
          },
          { timeout: 200 },
        );
      }
    });

    it("should handle onSelect error and show error state", async () => {
      const mockOnSelect = vi.fn(async () => {
        throw new Error("Payment failed");
      });

      const props = {
        ...defaultProps,
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));

      if (cashfreeCard) {
        fireEvent.click(cashfreeCard);

        // Should show error state
        await waitFor(() => {
          expect(screen.getByText(/payment failed|error/i)).toBeInTheDocument();
        });
      }
    });
  });
});
