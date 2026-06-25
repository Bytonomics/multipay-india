/**
 * Base error class for all MultiPay-related errors
 * Extends native Error with optional error code support
 */
export class MultiPayError extends Error {
  /**
   * Machine-readable error code for programmatic handling
   */
  public readonly code?: string;

  /**
   * Create a new MultiPayError
   * @param message - Human-readable error message
   * @param code - Optional machine-readable error code
   */
  constructor(message: string, code?: string) {
    super(message);
    this.name = "MultiPayError";
    this.code = code;
  }

  /**
   * Check if this error matches a specific error code
   * @param code - Error code to check against
   */
  hasCode(code: string): boolean {
    return this.code === code;
  }

  /**
   * Create a MultiPayError with a specific code
   * @param code - Machine-readable error code
   * @param message - Human-readable error message
   */
  static withCode(code: string, message: string): MultiPayError {
    return new MultiPayError(message, code);
  }
}

/**
 * Common error codes used across the MultiPay system.
 *
 * IMPORTANT: These enum values are part of the public API and are intentionally exported
 * for use by consumers of this library. Applications using this library can import
 * ErrorCodes and use them for type-safe error checking:
 *
 * ```typescript
 * import { ErrorCodes } from '@bytonomics/multipay-frontend-ts/core'
 *
 * if (error.hasCode(ErrorCodes.SCRIPT_LOAD_FAILED)) {
 *   // Handle script loading failure
 * }
 * ```
 *
 * These codes are not used internally within the library itself because they are
 * provided as a convenience for library consumers to catch and handle specific
 * error conditions in a type-safe manner.
 */
export enum ErrorCodes {
  // Initialization errors
  // SCRIPT_LOAD_FAILED = "SCRIPT_LOAD_FAILED",  // Commented out - not currently used in library
  // SCRIPT_ALREADY_LOADED = "SCRIPT_ALREADY_LOADED",  // Commented out - not currently used in library
  // INITIALIZATION_FAILED = "INITIALIZATION_FAILED",  // Commented out - not currently used in library
  // Payment flow errors
  // CHECKOUT_FAILED = "CHECKOUT_FAILED",  // Commented out - not currently used in library
  // CHECKOUT_CANCELLED = "CHECKOUT_CANCELLED",  // Commented out - not currently used in library
  // PAYMENT_FAILED = "PAYMENT_FAILED",  // Commented out - not currently used in library
  // PAYMENT_PENDING = "PAYMENT_PENDING",  // Commented out - not currently used in library
  // Validation errors
  // INVALID_PAYLOAD = "INVALID_PAYLOAD",  // Commented out - not currently used in library
  // MISSING_REQUIRED_FIELD = "MISSING_REQUIRED_FIELD",  // Commented out - not currently used in library
  // INVALID_PROVIDER = "INVALID_PROVIDER",  // Commented out - not currently used in library
  // Configuration errors
  // INVALID_CONFIG = "INVALID_CONFIG",  // Commented out - not currently used in library
  // MISSING_API_KEY = "MISSING_API_KEY",  // Commented out - not currently used in library
  // INVALID_ENVIRONMENT = "INVALID_ENVIRONMENT",  // Commented out - not currently used in library
  // Network errors
  // NETWORK_ERROR = "NETWORK_ERROR",  // Commented out - not currently used in library
  // TIMEOUT = "TIMEOUT",  // Commented out - not currently used in library
  // Provider-specific errors
  // PROVIDER_ERROR = "PROVIDER_ERROR",  // Commented out - not currently used in library
  // CASHFREE_ERROR = "CASHFREE_ERROR",  // Commented out - not currently used in library
  // RAZORPAY_ERROR = "RAZORPAY_ERROR",  // Commented out - not currently used in library
}
