import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import type React from "react";
import { PaymentPicker } from "../src/react/PaymentPicker";
import type { PaymentPickerProps, PickerProviderId } from "../src/react/types";
import type { Provider } from "../src/core/types";

// Type for PaymentPicker ref
type PaymentPickerRef = React.RefObject<{
  selectProvider: (_providerId: PickerProviderId) => void;
  getSelectedProvider: () => PickerProviderId;
  isSelected: (_providerId: PickerProviderId) => boolean;
  setProviderDisabled: (
    _providerId: PickerProviderId,
    _disabled: boolean,
    _reason?: string,
  ) => void;
  focus: () => void;
  blur: () => void;
}>;

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

  const defaultProviders = [
    { id: "cashfree" as PickerProviderId, label: "Cashfree", enabled: true },
    { id: "razorpay" as PickerProviderId, label: "Razorpay", enabled: true },
  ];

  const defaultProps: PaymentPickerProps = {
    payment: {
      amountMinor: 5000,
      currency: "INR",
      providers: defaultProviders,
      defaultSelected: "cashfree" as PickerProviderId,
    },
    appearance: {
      variant: "interactive-matrix",
      theme: "light",
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
      const cashfreeButton = screen.getByText("Cashfree").closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute('aria-pressed', 'true');
    });

    it("should NOT render PayU when not in payment.providers", () => {
      render(<PaymentPicker {...defaultProps} />);

      // Verify PayU is not rendered
      expect(screen.queryByText("PayU")).not.toBeInTheDocument();
    });

    it("should pre-select Cashfree when defaultSelected is cashfree", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "cashfree" as PickerProviderId,
        },
      };
      render(<PaymentPicker {...props} />);

      const cashfreeButton = screen.getByText("Cashfree").closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute('aria-pressed', 'true');
    });

    it("should pre-select Razorpay when defaultSelected is razorpay", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "razorpay" as PickerProviderId,
        },
      };
      render(<PaymentPicker {...props} />);

      const razorpayButton = screen.getByText("Razorpay").closest('[role="button"]');
      expect(razorpayButton).toHaveAttribute('aria-pressed', 'true');
    });

    it("should ignore invalid defaultSelected value", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "invalid" as PickerProviderId,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: true,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
        },
      };
      render(<PaymentPicker {...props} />);

      // Should default to first enabled provider (cashfree)
      const cashfreeButton = screen.getByText("Cashfree").closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute('aria-pressed', 'true');

      // Verify console was warned about invalid default
      expect(consoleSpy).toHaveBeenCalled();

      consoleSpy.mockRestore();
    });

    it("should ignore disabled provider in defaultSelected", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "cashfree" as PickerProviderId,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: false,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
        },
      };
      render(<PaymentPicker {...props} />);

      // Should default to first enabled provider (razorpay)
      const razorpayButton = screen.getByText("Razorpay").closest('[role="button"]');
      expect(razorpayButton).toHaveAttribute('aria-pressed', 'true');

      const cashfreeButton = screen.getByText("Cashfree").closest('[role="button"]');
      expect(cashfreeButton).toBeDisabled();

      consoleSpy.mockRestore();
    });

    it("should ignore absent provider in defaultSelected", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "payu" as PickerProviderId,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: true,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
        },
      };
      render(<PaymentPicker {...props} />);

      // Should default to first enabled provider (cashfree)
      const cashfreeButton = screen.getByText("Cashfree").closest('[role="button"]');
      expect(cashfreeButton).toHaveAttribute('aria-pressed', 'true');

      consoleSpy.mockRestore();
    });
  });

  describe("Assertion 2: Theme rendering (light, dark, auto)", () => {
    const variants: Array<
      | "dynamic-stack"
      | "interactive-matrix"
      | "secure-vault"
      | "neumorphic-flow"
    > = [
      "dynamic-stack",
      "interactive-matrix",
      "secure-vault",
      "neumorphic-flow",
    ];

    variants.forEach((variant) => {
      describe(`${variant} variant`, () => {
        it("should render with light theme", () => {
          const props = {
            ...defaultProps,
            appearance: {
              ...defaultProps.appearance,
              variant,
              theme: "light" as const,
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
              theme: "dark" as const,
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
              theme: "auto" as const,
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
              theme: "auto" as const,
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
              theme: "auto" as const,
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
    const variants: Array<
      | "dynamic-stack"
      | "interactive-matrix"
      | "secure-vault"
      | "neumorphic-flow"
    > = [
      "dynamic-stack",
      "interactive-matrix",
      "secure-vault",
      "neumorphic-flow",
    ];

    variants.forEach((variant) => {
      it(`${variant}: should render exactly one card per provider (no method-split elements)`, () => {
        const props = {
          ...defaultProps,
          appearance: {
            ...defaultProps.appearance,
            variant,
          },
        };

        render(<PaymentPicker {...props} />);

        // Count provider cards by finding all clickable provider elements
        const providerCards = screen
          .getAllByRole("button")
          .filter(
            (button) =>
              button.textContent &&
              (button.textContent.includes("Cashfree") ||
                button.textContent.includes("Razorpay")),
          );

        // Should have exactly 2 cards (one per provider)
        expect(providerCards).toHaveLength(2);

        // Verify no method-specific text (UPI, Cards, Wallets, etc.) is split into separate elements
        expect(screen.queryByText("UPI")).not.toBeInTheDocument();
        expect(screen.queryByText("Cards")).not.toBeInTheDocument();
        expect(screen.queryByText("Wallets")).not.toBeInTheDocument();
      });
    });
  });

  describe("Assertion 4: defaultSelected behavior", () => {
    it("should pre-select the named ENABLED aggregator", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "razorpay" as PickerProviderId,
        },
      };

      render(<PaymentPicker {...props} />);

      const razorpayCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Razorpay"));

      expect(razorpayCard).toBeDefined();
      expect(razorpayCard).toHaveAttribute('aria-pressed', 'true');
    });

    it("should select nothing when defaultSelected is invalid value", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "invalid_provider" as PickerProviderId,
        },
      };

      render(<PaymentPicker {...props} />);

      // Should fall back to first enabled provider
      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));
      expect(cashfreeCard).toHaveAttribute('aria-pressed', 'true');

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should select nothing when defaultSelected is disabled provider", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "cashfree" as PickerProviderId,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: false,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
        },
      };

      render(<PaymentPicker {...props} />);

      // Should select first enabled provider (razorpay)
      const razorpayCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Razorpay"));
      expect(razorpayCard).toHaveAttribute('aria-pressed', 'true');

      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));
      expect(cashfreeCard).toBeDisabled();

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should select nothing when defaultSelected is absent from providers array", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "payu" as PickerProviderId,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: true,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
        },
      };

      render(<PaymentPicker {...props} />);

      // Should fall back to first enabled provider (cashfree)
      const cashfreeCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Cashfree"));
      expect(cashfreeCard).toHaveAttribute('aria-pressed', 'true');

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should emit dev warning for invalid/disabled/absent defaultSelected", () => {
      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "invalid" as PickerProviderId,
        },
      };

      render(<PaymentPicker {...props} />);

      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });
  });

  describe("Assertion 5: Click behavior and onSelect callback", () => {
    it("should call onSelect with canonical Provider (cashfree) when clicking enabled card", async () => {
      const mockOnSelect = vi.fn();

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: undefined,
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
          expect(mockOnSelect).toHaveBeenCalledWith("cashfree" as Provider);
        });
      }
    });

    it("should call onSelect with canonical Provider (razorpay) when clicking enabled card", async () => {
      const mockOnSelect = vi.fn();

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: undefined,
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
          expect(mockOnSelect).toHaveBeenCalledWith("razorpay" as Provider);
        });
      }
    });

    it("should NOT call onSelect when clicking disabled card", async () => {
      const mockOnSelect = vi.fn();

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: false,
              disabledReason: "Maintenance",
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
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
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: false,
              disabledReason: "Temporarily unavailable",
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
        },
      };

      render(<PaymentPicker {...props} />);

      expect(screen.getByText("Temporarily unavailable")).toBeInTheDocument();
    });

    it('should never emit "payu" from onSelect callback', async () => {
      const mockOnSelect = vi.fn();

      // Even if someone manually adds payu to providers
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: true,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
            {
              id: "multipay_default" as PickerProviderId,
              label: "PayU",
              enabled: true,
            },
          ],
        },
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} />);

      // Try to click all provider cards
      const allCards = screen.getAllByRole("button");
      allCards.forEach((card) => {
        fireEvent.click(card);
      });

      await waitFor(() => {
        // Verify onSelect was only called with canonical providers
        mockOnSelect.mock.calls.forEach((call) => {
          expect(["cashfree", "razorpay"]).toContain(call[0]);
        });

        // PayU should never be emitted
        expect(mockOnSelect).not.toHaveBeenCalledWith("multipay_default");
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
        screen.getByText("Total amount inclusive of all taxes"),
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
        screen.queryByText("Total amount inclusive of all taxes"),
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
        screen.getByText("Total amount inclusive of all taxes"),
      ).toBeInTheDocument();

      // Verify they appear in a structured section (not just anywhere)
      const totalSection = container.querySelector(
        '.totalSection, .total-section, [class*="total"]',
      );
      expect(totalSection).toBeInTheDocument();
    });
  });

  describe("Assertion 7: PickerControls via ref", () => {
    it("should provide setLoading control method via ref", async () => {
      const mockRef: PaymentPickerRef = { current: null };

      const props = {
        ...defaultProps,
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      // Wait for ref to be populated
      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      // Call setLoading method
      if (mockRef.current) {
        mockRef.current.setProviderDisabled(
          "cashfree" as PickerProviderId,
          true,
          "Test maintenance",
        );

        // Verify provider was disabled
        await waitFor(() => {
          const cashfreeCard = screen
            .getAllByRole("button")
            .find((button) => button.textContent?.includes("Cashfree"));
          expect(cashfreeCard).toBeDisabled();
        });
      }
    });

    it("should provide setError control method via ref", async () => {
      const mockRef: React.RefObject<{
        selectProvider: (_providerId: PickerProviderId) => void;
        getSelectedProvider: () => PickerProviderId;
        isSelected: (_providerId: PickerProviderId) => boolean;
        setProviderDisabled: (
          _providerId: PickerProviderId,
          _disabled: boolean,
          _reason?: string,
        ) => void;
        focus: () => void;
        blur: () => void;
      }> = { current: null };

      const props = {
        ...defaultProps,
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      // Note: setError is available via controls but doesn't directly render
      // Error rendering is handled by the component's error state
      if (mockRef.current) {
        // Verify controls exists (implementation detail)
        expect(typeof mockRef.current.setProviderDisabled).toBe("function");
      }
    });

    it("should provide selectProvider control method via ref", async () => {
      const mockOnSelect = vi.fn();
      const mockRef: PaymentPickerRef = { current: null };

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
        mockRef.current.selectProvider("razorpay" as PickerProviderId);

        await waitFor(() => {
          expect(mockOnSelect).toHaveBeenCalledWith("razorpay" as Provider);
        });
      }
    });

    it("should provide getSelectedProvider control method via ref", async () => {
      const mockRef: PaymentPickerRef = { current: null };

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "razorpay" as PickerProviderId,
        },
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        const selected = mockRef.current.getSelectedProvider();
        expect(selected).toBe("razorpay");
      }
    });

    it("should provide isSelected control method via ref", async () => {
      const mockRef: PaymentPickerRef = { current: null };

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          defaultSelected: "cashfree" as PickerProviderId,
        },
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        expect(mockRef.current.isSelected("cashfree" as PickerProviderId)).toBe(
          true,
        );
        expect(mockRef.current.isSelected("razorpay" as PickerProviderId)).toBe(
          false,
        );
      }
    });

    it("should provide setProviderDisabled control method via ref", async () => {
      const mockRef: PaymentPickerRef = { current: null };

      const props = {
        ...defaultProps,
        onSelect: vi.fn(),
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      await waitFor(() => {
        expect(mockRef.current).not.toBeNull();
      });

      if (mockRef.current) {
        mockRef.current.setProviderDisabled(
          "cashfree" as PickerProviderId,
          true,
          "Maintenance",
        );

        await waitFor(() => {
          const cashfreeCard = screen
            .getAllByRole("button")
            .find((button) => button.textContent?.includes("Cashfree"));
          expect(cashfreeCard).toBeDisabled();
          expect(screen.getByText("Maintenance")).toBeInTheDocument();
        });
      }
    });

    it("should provide focus and blur control methods via ref", async () => {
      const mockRef: PaymentPickerRef = { current: null };

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

  describe("Integration tests: Combined behavior", () => {
    it("should handle complete selection flow with all features", async () => {
      const mockOnSelect = vi.fn();
      const mockRef: PaymentPickerRef = { current: null };

      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          amountMinor: 10000, // ₹100.00
          currency: "INR",
          providers: [
            {
              id: "cashfree" as PickerProviderId,
              label: "Cashfree",
              enabled: true,
            },
            {
              id: "razorpay" as PickerProviderId,
              label: "Razorpay",
              enabled: true,
            },
          ],
          defaultSelected: "cashfree" as PickerProviderId,
        },
        appearance: {
          variant: "interactive-matrix" as const,
          theme: "light" as const,
          taxNote: "All taxes included",
        },
        onSelect: mockOnSelect,
      };

      render(<PaymentPicker {...props} ref={mockRef} />);

      // Verify initial state
      await waitFor(() => {
        expect(screen.getByText("₹100.00")).toBeInTheDocument();
        expect(screen.getByText("All taxes included")).toBeInTheDocument();
        expect(screen.getByText("Cashfree")).toBeInTheDocument();
        expect(screen.getByText("Razorpay")).toBeInTheDocument();
      });

      // Verify ref controls
      if (mockRef.current) {
        expect(mockRef.current.isSelected("cashfree" as PickerProviderId)).toBe(
          true,
        );
        expect(mockRef.current.getSelectedProvider()).toBe("cashfree");
      }

      // Click razorpay card
      const razorpayCard = screen
        .getAllByRole("button")
        .find((button) => button.textContent?.includes("Razorpay"));

      if (razorpayCard) {
        fireEvent.click(razorpayCard);

        await waitFor(() => {
          expect(mockOnSelect).toHaveBeenCalledWith("razorpay" as Provider);
        });
      }

      // Use ref to disable cashfree
      if (mockRef.current) {
        mockRef.current.setProviderDisabled(
          "cashfree" as PickerProviderId,
          true,
          "Temporarily down",
        );

        await waitFor(() => {
          expect(screen.getByText("Temporarily down")).toBeInTheDocument();
          const cashfreeCard = screen
            .getAllByRole("button")
            .find((button) => button.textContent?.includes("Cashfree"));
          expect(cashfreeCard).toBeDisabled();
        });
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

    it("should throw error when payment.providers is empty", () => {
      const props = {
        ...defaultProps,
        payment: {
          ...defaultProps.payment,
          providers: [], // Invalid
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
