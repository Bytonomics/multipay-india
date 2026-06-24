# multipay-frontend-ts

The frontend (checkout picker + redirect) arm of the **multipay-india** payment-aggregator monorepo —
the browser-side companion to the backend ports (`multipay-go`, `multipay-ts`, `multipay-py`). It provides
a headless checkout redirect plus a React `<PaymentPicker>` for selecting a payment aggregator.

- npm package: `@bytonomics/multipay-frontend-ts` (planned)
- Strictly typed TypeScript: no `any`; external/untyped values (vendor SDK globals, JSON from the backend)
  are converted at a single typed boundary.

Status: planned — not yet implemented.
