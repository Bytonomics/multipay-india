# CLAUDE.md — multipay-india (monorepo)

`multipay-india` is a multi-language **payment-gateway aggregator** for Indian providers (Cashfree,
Razorpay; PayU planned): one canonical contract implemented across several language ports that behave
identically. This umbrella file is intentionally **language-agnostic** — each port carries its own
`CLAUDE.md` / rules; read the one for the port you are working in. (The Go-library-specific architecture,
build commands, linters, and rules now live in [`multipay-go/CLAUDE.md`](./multipay-go/CLAUDE.md).)

## Ports

| Folder | What it is | Port rules |
|--------|-----------|-----------|
| [`multipay-go/`](./multipay-go) | Go library (reference port) | [`multipay-go/CLAUDE.md`](./multipay-go/CLAUDE.md) + [`multipay-go/.claude/rules/`](./multipay-go/.claude/rules) |
| [`multipay-frontend-ts/`](./multipay-frontend-ts) | Frontend checkout / picker library (TS + React) | [`multipay-frontend-ts/CLAUDE.md`](./multipay-frontend-ts/CLAUDE.md) |
| [`multipay-ts/`](./multipay-ts) | Backend TypeScript port | planned |
| [`multipay-py/`](./multipay-py) | Backend Python port | planned |
| [`contract/`](./contract) | Shared canonical contract + cross-language golden test vectors | — |

## Cross-language rules (apply to EVERY port)

These hold regardless of language. Port-specific *implementations* (helpers, validators, linters) live in
each port's own `CLAUDE.md`.

- **The canonical contract is the source of truth.** Request/response field names, enum string VALUES and
  casing, the webhook event taxonomy, and the checkout/redirect payload must be identical across ports and
  match [`contract/`](./contract). Any change to these is applied to every affected port in the same effort.
- **Money is always in minor units** (paisa/cents/fils), converted via the ISO-4217 exponent. Never hardcode
  `/100` or `*100`; never pass a major-unit value where a minor-unit one is expected.
- **Webhook endpoints always return 2xx** after signature verification (backend ports) — never 5xx, or the
  vendor auto-disables the endpoint; persist-and-replay instead.
- **Strict typing, no loose maps / `any`.** Build requests as typed structs/interfaces; the only untyped step
  is decoding a raw vendor response at a single boundary, then mapping straight to typed shapes.

For the high-level "how it works" diagrams and quick-start, see [`README.md`](./README.md).
