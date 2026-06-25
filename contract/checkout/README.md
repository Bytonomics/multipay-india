# Checkout Slice

Frontend-bound, provider-agnostic checkout payload that enables JavaScript clients to drive vendor-hosted checkout redirects. Discriminated by `Provider`; only fields for the configured provider are populated.

## Core Types

### `CheckoutPayload`

```json
{
  "provider": "cashfree" | "razorpay",
  "environment": "SANDBOX" | "PRODUCTION",
  "session_id": "string (Cashfree only)",
  "order_id": "string (Razorpay only)",
  "public_key": "string (Razorpay only)",
  "callback_url": "string (Razorpay only)",
  "amount_minor": "integer (Razorpay only)",
  "currency": "string (Razorpay only)"
}
```

### Amount Field Contract (CRITICAL)

The canonical money key is `amount_minor` — represents minor units (paisa for INR, cents for USD).

- **Go domain**: `CheckoutPayload.AmountMinor int64` with JSON tag `amount_minor`
- **TypeScript**: `RazorpayCheckoutPayload.amount_minor: number`
- **Razorpay hosted form**: Field name is `amount`, but its VALUE comes from `amount_minor` (minor units, paisa)

**Example**: ₹500.00 → `amount_minor: 50000` (500 × 100 minor units per major)

**Never pass major units**: `amount_minor: 500` would be interpreted as ₹5.00, not ₹500.00.

### Provider/Environment Enums

#### `Provider` (lowercase)
- `"cashfree"` — Cashfree PG
- `"razorpay"` — Razorpay

#### `Environment` (UPPERCASE)
- `"SANDBOX"` — Test environment
- `"PRODUCTION"` — Live environment

### Provider-Specific Fields

#### Cashfree Payload
- `session_id` (required) — Cashfree payment session ID from order creation
- All other fields omitted

#### Razorpay Payload  
- `order_id` (required) — Razorpay order ID
- `public_key` (required) — Razorpay key ID for frontend initialization
- `callback_url` (required) — Redirect URL after payment completion
- `amount_minor` (required) — Amount in minor units (paisa)
- `currency` (required) — ISO 4217 currency code (e.g., "INR")
- `session_id` omitted

## Contract Validation

Both ports validate against canonical contract files:

- **Go backend**: `multipay-go` — validates domain types against this contract
- **TypeScript frontend**: `multipay-frontend-ts` — validates TS types against this contract

Test vectors and enum values defined in:
- `contract/checkout/vectors/*.json` — Golden test vectors for checkout redirects
- `contract/checkout/enums.json` — Canonical enum values for Provider and Environment

## Frontend Usage

1. Backend creates order, receives `CheckoutPayload`
2. Frontend reads `provider` field to determine which checkout SDK to load
3. For Razorpay: initialize with `public_key`, pass `amount` (value from `amount_minor`)
4. For Cashfree: initialize with `session_id`
5. Handle redirect/response via `callback_url` or Return URL

## Important Notes

- **Provider case sensitivity**: `provider` is lowercase (`"cashfree"`, `"razorpay"`)
- **Environment case sensitivity**: `environment` is UPPERCASE (`"SANDBOX"`, `"PRODUCTION"`)
- **Amount is always minor units**: Never divide or convert — `amount_minor` is what the checkout expects
- **Razorpay form field mismatch**: Form field is named `amount`, but value comes from `amount_minor` (minor units, not major)
