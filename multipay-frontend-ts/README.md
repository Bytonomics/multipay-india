# multipay-frontend-ts

The frontend TypeScript/React library for the **multipay-india** payment-aggregator monorepo — providing headless checkout redirects and a React `<PaymentPicker>` component for selecting payment aggregators (Cashfree, Razorpay).

## Package

**npm:** `@bytonomics/multipay-frontend-ts`

**TypeScript:** Strictly typed, zero `any` — external/untyped values are converted at a single typed boundary.

**React:** Optional peer dependency (v18+). Use the headless `/core` entry from vanilla JS, Vue, Angular, or Svelte without pulling in React.

---

## Two Entry Points

### 1. `/core` — Headless Checkout (Zero React)

```typescript
import { MultiPay } from '@bytonomics/multipay-frontend-ts/core'

const mpay = new MultiPay()
await mpay.checkout(payload)
```

**Pure TypeScript** — no React, no JSX. Validates the payload and performs a full-page redirect to the vendor's hosted payment page.

### 2. `/react` — PaymentPicker Component

```typescript
import { PaymentPicker } from '@bytonomics/multipay-frontend-ts/react'
import '@bytonomics/multipay-frontend-ts/styles.css'

<PaymentPicker
  payment={{
    amountMinor: 50000,
    currency: 'INR',
    providers: [
      { id: 'cashfree', label: 'Cashfree', enabled: true },
      { id: 'razorpay', label: 'Razorpay', enabled: true },
    ],
    defaultSelected: 'cashfree',
  }}
  appearance={{
    variant: 'interactive-matrix',
    theme: 'auto',
  }}
  onSelect={(provider) => {
    // provider is 'cashfree' | 'razorpay'
  }}
/>
```

**React 18+** is an **optional peer dependency** — only consumers of `/react` need React installed.

---

## Headless Usage

### `MultiPay.checkout(payload)`

Validates the payload and redirects to the vendor's hosted payment page.

```typescript
import { MultiPay } from '@bytonomics/multipay-frontend-ts/core'

const mpay = new MultiPay()

// Cashfree checkout
await mpay.checkout({
  provider: 'cashfree',
  order_id: 'order_123',
  payment_session_id: 'session_abc',
  environment: 'sandbox',
  amount: 50000,
  currency: 'INR',
  customer_id: 'cust_xyz',
  customer_phone: '+919876543210',
  customer_email: 'user@example.com',
  metadata: { source: 'web' },
})

// Razorpay checkout
await mpay.checkout({
  provider: 'razorpay',
  order_id: 'order_123',
  key_id: 'key_abc',
  public_key: 'rzp_live_xyz...',
  amount_minor: 50000,
  currency: 'INR',
  environment: 'production',
  customer_id: 'cust_xyz',
  customer_phone: '+919876543210',
  customer_email: 'user@example.com',
  callback_url: 'https://example.com/return',
  metadata: { source: 'web' },
})
```

### Provider Payloads

**Cashfree:**

```typescript
{
  provider: 'cashfree'
  order_id: string
  payment_session_id: string
  environment: 'sandbox' | 'production'
  amount: number
  currency: string
  customer_id?: string
  customer_phone?: string
  customer_email?: string
  metadata?: Record<string, string>
}
```

**Razorpay:**

```typescript
{
  provider: 'razorpay'
  order_id: string
  key_id: string
  public_key: string
  amount_minor: number
  currency: string
  environment: 'sandbox' | 'production'
  customer_id?: string
  customer_phone?: string
  customer_email?: string
  callback_url?: string
  metadata?: Record<string, string>
}
```

### Redirect Behavior

- **Cashfree:** Loads `cashfree.js` from CDN lazily (deduplicated), calls `Cashfree({ mode }).checkout({ paymentSessionId })`
- **Razorpay:** Creates a hidden form and POSTs to `https://api.razorpay.com/v1/checkout/embedded` (full-page redirect, not iframe)

Both are **full-page redirects** — the vendor handles the payment UI, method selection (UPI, cards, wallets), and 3D Secure on their own hosted page.

---

## Picker Usage

### Props

```typescript
interface PaymentPickerProps {
  payment: {
    amountMinor: number          // Amount in minor units (paisa, cents)
    currency: string              // ISO 4217 code (e.g., 'INR', 'USD')
    providers: ProviderOption[]  // Available aggregators
    defaultSelected?: PickerProviderId
  }

  appearance?: {
    variant?: PickerVariant       // 'interactive-matrix' | 'dynamic-stack' | 'secure-vault' | 'neumorphic-flow'
    theme?: PickerTheme          // 'light' | 'dark' | 'auto'
    branding?: PickerBranding    // logo/title/subtitle/footerText
    className?: string           // Additional CSS class
    taxNote?: string             // Override default tax disclaimer
  }

  onSelect: (provider: Provider) => void | Promise<void>
}

interface ProviderOption {
  id: PickerProviderId           // 'cashfree' | 'razorpay' | 'multipay_default'
  label: string                  // Display name
  description?: string           // Optional subtitle
  icon?: ReactNode               // Custom logo component
  recommended?: boolean          // Show "Recommended" badge
  disabled?: boolean             // Grey out and prevent selection
  disabledReason?: string        // Tooltip when disabled
}
```

### Basic Example

```typescript
import { PaymentPicker } from '@bytonomics/multipay-frontend-ts/react'
import '@bytonomics/multipay-frontend-ts/styles.css'

<PaymentPicker
  payment={{
    amountMinor: 50000,
    currency: 'INR',
    providers: [
      { id: 'cashfree', label: 'Cashfree Payments', enabled: true, recommended: true },
      { id: 'razorpay', label: 'Razorpay', enabled: true },
    ],
    defaultSelected: 'cashfree',
  }}
  appearance={{
    variant: 'interactive-matrix',
    theme: 'auto',
  }}
  onSelect={async (provider) => {
    // Call your backend to get checkout payload
    const payload = await fetch('/api/checkout', {
      method: 'POST',
      body: JSON.stringify({ provider }),
    }).then(r => r.json())

    // Redirect using headless MultiPay
    const mpay = new MultiPay()
    await mpay.checkout(payload)
  }}
/>
```

### With Branding

```typescript
<PaymentPicker
  payment={{ /* ... */ }}
  appearance={{
    variant: 'secure-vault',
    theme: 'light',
    branding: {
      logo: <MyLogo />,
      title: 'Checkout',
      subtitle: 'Select your payment provider',
      footerText: 'Secured by 256-bit SSL',
    },
    taxNote: 'Including GST @ 18%',
  }}
  onSelect={handleProviderSelect}
/>
```

### Default Provider

`defaultSelected` sets the initially selected provider. If the ID is invalid or the provider is disabled, the picker falls back to the first enabled provider:

```typescript
<PaymentPicker
  payment={{
    // ... amount, currency, providers
    defaultSelected: 'cashfree',  // Pre-select Cashfree
  }}
  onSelect={handleProviderSelect}
/>
```

---

## Four Visual Variants

All variants are **web-first**, **responsive** (1-up on mobile, 2-3-up on desktop), and **theme-complete** (ship both light AND dark palettes).

### `interactive-matrix` (Default)

Grid layout with hover states, click-to-select, and clear selection indicators. Best for general checkout flows.

### `dynamic-stack`

Stacked card layout with expand/collapse animations. Ideal for mobile-first flows where vertical space is limited.

### `secure-vault`

Bounded container with elevated trust indicators (shields, badges). Designed for security-conscious applications (enterprise, fintech).

### `neumorphic-flow`

Soft UI with tactile depth and smooth transitions. Best for modern, design-forward applications.

> **Note:** All variants show **one card per provider** — the vendor handles method selection (UPI, cards, wallets) on their hosted page. We only choose the aggregator.

---

## Theming

### Theme Modes

```typescript
appearance={{
  theme: 'light'   // Force light palette
  // theme: 'dark'  // Force dark palette
  // theme: 'auto'  // Follow OS prefers-color-scheme (default)
}}
```

### CSS Custom Properties

All variants use `--mpay-*` custom properties. Override per `[data-theme]` for custom palettes:

```css
/* Light theme overrides */
[data-theme='light'] {
  --mpay-primary-600: #ff6b35; /* Custom brand color */
  --mpay-neutral-100: #f8f9fa;
}

/* Dark theme overrides */
[data-theme='dark'] {
  --mpay-primary-500: #ff6b35;
  --mpay-neutral-800: #1a1a1a;
}
```

### Font Family

Inherits by default (`--mpay-font-family: inherit`). Override globally or per-component:

```css
:root {
  --mpay-font-family: 'Inter', system-ui, sans-serif;
}

/* Or per instance */
<PaymentPicker
  appearance={{
    fontFamily: "'Poppins', sans-serif",
  }}
  ...
/>
```

---

## Branding Customization

### Header/Footer

```typescript
appearance={{
  branding: {
    logo: <MyCompanyLogo />,
    title: 'Complete Your Purchase',
    subtitle: 'Choose a payment provider',
    footerText: 'Powered by MultiPay India',
  },
}}
```

### Custom Provider Icons

```typescript
const providers = [
  {
    id: 'cashfree',
    label: 'Cashfree',
    enabled: true,
    icon: <MyCashfreeLogo />,
  },
  {
    id: 'razorpay',
    label: 'Razorpay',
    enabled: true,
    icon: <MyRazorpayLogo />,
  },
]
```

### CSS Classes

Pass `className` to add custom classes (CSS Modules prevent collisions):

```typescript
<PaymentPicker
  appearance={{
    className: 'my-custom-picker',
  }}
  ...
/>
```

---

## Total Amount & Tax Disclaimer

### Amount Formatting

The picker auto-formats `amountMinor` (minor units) to a localized currency string using `Intl.NumberFormat`:

```typescript
amountMinor: 50000, currency: 'INR'  → "₹500.00"
amountMinor: 1234, currency: 'USD'  → "$12.34"
```

### Tax Note

Default disclaimer: `"Total amount inclusive of all taxes"`

Override via `appearance.taxNote`:

```typescript
<PaymentPicker
  appearance={{
    taxNote: 'Including GST @ 18%',
  }}
  ...
/>
```

---

## Provider Types & PayU Placeholder

### Canonical Providers

`Provider` = payable providers with Go backend support:

```typescript
type Provider = 'cashfree' | 'razorpay'
```

### Picker-Only Identifiers

`PickerProviderId` = picker superset (includes future placeholders):

```typescript
type PickerProviderId = 'cashfree' | 'razorpay' | 'multipay_default'
```

### PayU Placeholder

PayU (`id: 'multipay_default'`) is a **code-only placeholder** — not shown by default, always `disabled: true`, never emitted by `onSelect`. This allows future PayU support without breaking the API contract.

---

## Imperative Control Hook

`usePaymentPicker` provides imperative control for programmatic selection:

```typescript
import { usePaymentPicker } from '@bytonomics/multipay-frontend-ts/react'

function CheckoutFlow() {
  const { controls } = usePaymentPicker()

  // Programmatically select a provider
  const selectCashfree = () => {
    controls.selectProvider('cashfree')
  }

  // Check if a provider is selected
  const isCashfreeSelected = controls.isSelected('cashfree')

  // Disabling a provider is DECLARATIVE — set it on the payment.providers prop:
  //   payment.providers.razorpay = {
  //     ...entry, enabled: false, disabledMessage: 'Currently unavailable',
  //   }
  // Every variant honors `enabled`/`disabledMessage`; there is no imperative disable method.

  return (
    <PaymentPicker
      ref={pickerRef}
      payment={{ /* ... */ }}
      onSelect={handleProviderSelect}
    />
  )
}
```

---

## Error Handling

### `onSelect` Errors

If `onSelect` throws, the picker shows an inline error banner with a retry button:

```typescript
<PaymentPicker
  onSelect={async (provider) => {
    try {
      const payload = await fetchCheckoutPayload(provider)
      await mpay.checkout(payload)
    } catch (error) {
      // Error shown in picker, user can retry
      throw error
    }
  }}
/>
```

### Validation Errors

`MultiPay.checkout()` throws `MultiPayError` for invalid payloads:

```typescript
import { MultiPay, MultiPayError } from '@bytonomics/multipay-frontend-ts/core'

try {
  await mpay.checkout(invalidPayload)
} catch (error) {
  if (error instanceof MultiPayError) {
    console.error('Checkout failed:', error.message)
  }
}
```

---

## Styling & CSS

### Import Styles

```typescript
import '@bytonomics/multipay-frontend-ts/styles.css'
```

This imports:
- `variables.css` — All `--mpay-*` custom properties (light + dark themes)
- Variant-specific CSS Modules (no global collisions)

### CSS Modules

All component styles use CSS Modules (`.module.css` files) — no Tailwind, no global CSS. This prevents style collisions with consumer apps.

---

## TypeScript Types

### Core Types

```typescript
import type {
  Provider,
  Environment,
  CheckoutPayload,
  CashfreeCheckoutPayload,
  RazorpayCheckoutPayload,
} from '@bytonomics/multipay-frontend-ts/core'
```

### Picker Types

```typescript
import type {
  PickerProviderId,
  PickerVariant,
  PickerTheme,
  ProviderOption,
  PickerAppearance,
  PickerBranding,
  PaymentPickerProps,
  PickerControls,
} from '@bytonomics/multipay-frontend-ts/react'
```

---

## Integration Notes

### Backend Contract

The `CheckoutPayload` TypeScript type must match the Go `domain.CheckoutPayload` exactly:

| Field | TypeScript | Go |
|---|---|---|
| `provider` | `'cashfree' \| 'razorpay'` | `domain.Provider` (lowercase enum) |
| `environment` | `'sandbox' \| 'production'` | `domain.Environment` (lowercase enum) |
| `session_id` | `string` (cashfree) | `SessionID string` |
| `order_id` | `string` | `OrderID string` |
| `public_key` | `string` (razorpay) | `PublicKey string` |
| `callback_url` | `string` (razorpay) | `CallbackURL string` |
| `amount_minor` | `number` (razorpay) | `AmountMinor int64` |

### Cashfree SDK Loading

The Cashfree JS SDK is loaded lazily from CDN (`https://sdk.cashfree.com/js/v3/cashfree.js`). Script loading is deduplicated — multiple calls to `checkout()` reload the script only once.

### Razorpay Form POST

Razorpay checkout uses a native form POST to `https://api.razorpay.com/v1/checkout/embedded` — full-page redirect, not iframe.

---

## Build & Test

```bash
make build          # Build ESM + CJS + type declarations via rollup
make typecheck      # Strict TypeScript check (tsc --noEmit)
make lint           # ESLint (strict, no any, exhaustive deps)
make test           # Unit tests (vitest + testing-library/react)
make check          # Full pre-commit sequence: typecheck -> lint -> test
make clean          # Remove build artifacts
```

---

## License

MIT
