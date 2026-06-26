/**
 * Format currency amounts from minor units to localized strings.
 * Uses Intl.NumberFormat with ISO 4217 currency codes for proper exponent handling.
 */

/**
 * Convert minor units to a localized currency string.
 * Uses ISO 4217 exponent to determine the correct divisor.
 * @param amountMinor - Amount in minor units (paisa, cents, etc.)
 * @param currency - ISO 4217 currency code (e.g., 'INR', 'USD')
 * @param locale - BCP 47 language tag (default: 'en-IN')
 * @returns Formatted currency string (e.g., '₹500.00', '$12.34')
 */
export function formatMinor(
  amountMinor: number,
  currency: string,
  locale: string = "en-IN",
): string {
  // First, create a formatter WITHOUT min/maxFractionDigits to get the true ISO exponent
  const probeFormatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency: currency,
  });

  // Get the ISO 4217 exponent (decimal places) for this currency
  // exponent 0: 1 minor = 1 major (JPY, KRW, VND)
  // exponent 2: 100 minor = 1 major (INR, USD, EUR, GBP)
  // exponent 3: 1000 minor = 1 major (BHD, KWD, OMR)
  const exponent: number =
    probeFormatter.resolvedOptions().maximumFractionDigits ?? 2;
  const divisor: number = 10 ** exponent;
  const amountMajor: number = amountMinor / divisor;

  // Now format with the default precision for this currency
  return probeFormatter.format(amountMajor);
}

/**
 * Format currency amount with custom precision.
 * Uses ISO 4217 exponent to determine the correct divisor.
 * @param amountMinor - Amount in minor units
 * @param currency - ISO 4217 currency code
 * @param fractionDigits - Number of decimal places (default: 2)
 * @param locale - BCP 47 language tag (default: 'en-IN')
 * @returns Formatted currency string
 */
export function formatMinorWithPrecision(
  amountMinor: number,
  currency: string,
  fractionDigits: number = 2,
  locale: string = "en-IN",
): string {
  // First, create a formatter WITHOUT min/maxFractionDigits to get the true ISO exponent
  const probeFormatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency: currency,
  });

  // Get the ISO 4217 exponent (decimal places) for this currency
  const exponent: number =
    probeFormatter.resolvedOptions().maximumFractionDigits ?? 2;
  const divisor: number = 10 ** exponent;
  const amountMajor: number = amountMinor / divisor;

  // Now format with the requested precision
  const displayFormatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency: currency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });

  return displayFormatter.format(amountMajor);
}
