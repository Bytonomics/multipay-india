# contract

Shared canonical contract for every multipay-* port: request/response types, enum string values,
webhook event taxonomy, and cross-language golden test vectors. Every backend port (multipay-go,
multipay-ts, multipay-py) and the frontend (multipay-frontend-ts) must stay in parity with this.

## Slices

### checkout/ — CheckoutPayload + Provider/Environment enums + redirect golden vectors
Frontend-bound checkout contract for provider-agnostic hosted checkout redirects. Documents the critical `amount_minor` field contract (minor units, paisa) where Razorpay's hosted form field is named `amount` but its value comes from `amount_minor`. See [checkout/README.md](checkout/README.md).
