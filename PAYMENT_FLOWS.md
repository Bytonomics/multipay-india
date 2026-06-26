# Payment Flows

Sequence diagrams for every workflow `multipay-india` supports. These are **provider-neutral** — the same
flow holds whether the client is bound to Cashfree, Razorpay, or any future provider. Examples use the Go
port (`multipay-go`); the other language ports expose the same surface.

**Actors**

- **Frontend** — your UI (optionally using `multipay-frontend-ts` for the redirect/picker).
- **Backend** — your server, which imports `multipay-go`.
- **multipay-go** — the library. The backend → library call is an **in-process function call**, not a
  network hop.
- **Provider** — the payment gateway (Cashfree / Razorpay / …).

Two invariants worth repeating:

- **One client is bound to one provider** — chosen at construction, no runtime branching.
- **The webhook is the source of truth.** A post-payment browser redirect is UX; the webhook is what you
  trust. The handler must always answer `2xx` after the signature check, or the provider will auto-disable
  the endpoint.

> Field names shown in the diagrams (e.g. `{ amount, currency, customer }`) are illustrative. The exact
> request/response structs live in `multipay-go/domain` — see [`multipay-go/README.md`](./multipay-go/README.md).

---

## The big picture — who imports what

```mermaid
flowchart LR
  subgraph app[Your application]
    direction TB
    FE[Frontend<br/>imports multipay-frontend-ts]
    BE[Backend<br/>imports multipay-go]
  end
  MP{{multipay-go<br/>1 client = 1 provider}}
  PG[(Payment provider)]

  FE -- CheckoutPayload JSON --> BE
  BE -- Orders / Payments / Refunds / Instruments / PaymentLinks / Plans / Subscriptions / Webhooks --> MP
  MP -- provider SDK calls --> PG
  FE -. redirect to hosted page .-> PG
  PG -. webhook = source of truth .-> BE
```

---

## 1. Standard checkout (orders)

Create an order, hand the typed `CheckoutPayload` to the frontend, redirect to the provider's hosted page.
The return redirect is UX; the webhook is the truth.

```mermaid
sequenceDiagram
  autonumber
  participant FE as Frontend (multipay-frontend-ts)
  participant BE as Your backend (imports multipay-go)
  participant MP as multipay-go
  participant PG as Provider

  FE->>BE: POST /checkout { amount, currency, customer, return_url }
  Note over BE,MP: in-process Go call (same binary)
  BE->>MP: client.Orders().CreateOrder(req)
  MP->>PG: create order (provider SDK, HTTPS)
  PG-->>MP: provider order id + payment session
  MP-->>BE: domain.Order{ Checkout: CheckoutPayload }
  BE-->>FE: CheckoutPayload { provider, environment, session/order id }
  FE->>PG: MultiPay.checkout(payload) — redirect to hosted page
  Note over FE,PG: customer pays on the provider's hosted page
  PG-->>BE: redirect to return_url, then webhook /webhooks/{provider}/{account}
  BE->>MP: client.Orders().GetOrder(id) to confirm status
  MP-->>BE: domain.Order{ Status: OrderPaid }
  BE-->>PG: 200 ACK to the webhook (after signature check)
  Note over BE: your handler marks the order paid / fulfils
```

---

## 2. Payments (fetch, list, capture)

Inspect a payment, list the payments on an order, or capture a previously **authorized** payment (the
auth → capture two-step, when you authorize first and capture later).

```mermaid
sequenceDiagram
  autonumber
  participant BE as Your backend
  participant MP as multipay-go
  participant PG as Provider

  Note over BE,PG: fetch a single payment
  BE->>MP: client.Payments().GetPayment(req)
  MP->>PG: fetch payment (SDK)
  PG-->>MP: payment { status, amount, ... }
  MP-->>BE: domain.Payment

  Note over BE,PG: capture an authorized payment
  BE->>MP: client.Payments().CapturePayment(req)
  MP->>PG: capture (SDK)
  PG-->>MP: payment { status: captured }
  MP-->>BE: domain.Payment

  Note over BE,PG: list payments (e.g. for an order)
  BE->>MP: client.Payments().ListPayments(req)
  MP->>PG: list payments (SDK)
  PG-->>MP: [ payment, ... ]
  MP-->>BE: []domain.Payment
```

---

## 3. Refunds

Refund a captured payment by its provider payment id. The refund settles asynchronously — its terminal state
arrives as a webhook.

```mermaid
sequenceDiagram
  autonumber
  participant BE as Your backend
  participant MP as multipay-go
  participant PG as Provider

  BE->>MP: client.Refunds().CreateRefund({ PaymentID, AmountMinor, RefundID })
  MP->>PG: create refund (SDK, HTTPS)
  PG-->>MP: refund { id, status }
  MP-->>BE: domain.Refund{ status }
  Note over PG,BE: refund settles asynchronously
  PG-->>BE: webhook refund.processed / refund.failed
  BE-->>PG: 200 ACK
  opt look up later
    BE->>MP: client.Refunds().GetRefund(req) / ListRefunds(req)
    MP->>PG: fetch (SDK)
    PG-->>MP: refund(s)
    MP-->>BE: domain.Refund / []domain.Refund
  end
```

---

## 4. Instruments (saved cards / mandate tokens)

List the instruments saved against a customer, fetch one, or delete/revoke one. The provider owns the
instrument lifecycle; the library is a typed window onto it.

```mermaid
sequenceDiagram
  autonumber
  participant BE as Your backend
  participant MP as multipay-go
  participant PG as Provider

  BE->>MP: client.Instruments().ListInstruments(req)
  MP->>PG: list saved instruments for a customer (SDK)
  PG-->>MP: [ instrument, ... ]
  MP-->>BE: []domain.Instrument

  BE->>MP: client.Instruments().GetInstrument(req)
  MP->>PG: fetch one instrument (SDK)
  PG-->>MP: instrument
  MP-->>BE: domain.Instrument

  BE->>MP: client.Instruments().DeleteInstrument(req)
  MP->>PG: delete / revoke instrument (SDK)
  PG-->>MP: deleted instrument
  MP-->>BE: domain.Instrument
```

---

## 5. Payment links

A hosted pay-now URL you share over email/SMS/WhatsApp — useful for recovery (abandoned cart, failed
renewal). No browser comes back to your app, so the webhook is the *only* thing that closes the loop.

```mermaid
sequenceDiagram
  autonumber
  participant BE as Your backend
  participant MP as multipay-go
  participant PG as Provider
  participant U as Customer

  BE->>MP: client.PaymentLinks().CreatePaymentLink(req)
  MP->>PG: create payment link (SDK, HTTPS)
  PG-->>MP: hosted link_url
  MP-->>BE: domain.PaymentLink{ LinkURL }
  BE->>U: share link_url (email / SMS / WhatsApp)
  U->>PG: open link, pay on the hosted page
  Note over BE,PG: no redirect back to your app — fire-and-forget
  PG-->>BE: webhook (payment captured)
  BE-->>PG: 200 ACK
  opt manage the link
    BE->>MP: client.PaymentLinks().GetPaymentLink(req) / CancelPaymentLink(req)
    MP->>PG: fetch / cancel (SDK)
    PG-->>MP: payment link
    MP-->>BE: domain.PaymentLink
  end
```

---

## 6. Plans

Create or fetch a recurring **plan** (both providers are plan-first — a plan is created, then reused when
creating subscriptions). A plan id from here can be passed to `CreateSubscription({ PlanID })`.

```mermaid
sequenceDiagram
  autonumber
  participant BE as Your backend
  participant MP as multipay-go
  participant PG as Provider

  BE->>MP: client.Plans().CreatePlan(req)
  MP->>PG: create plan (SDK, HTTPS)
  PG-->>MP: plan { id, ... }
  MP-->>BE: domain.Plan

  BE->>MP: client.Plans().GetPlan(req)
  MP->>PG: fetch plan (SDK)
  PG-->>MP: plan
  MP-->>BE: domain.Plan
  Note over BE,PG: reuse plan id in CreateSubscription({ PlanID }) — see Subscriptions
```

---

## 7. Subscriptions (recurring billing)

Create once (with an inline plan or an existing plan id), the customer authorizes a mandate on the hosted
page, and then the provider auto-charges every cycle with **no call from you** — each charge arrives as a
webhook. Lifecycle actions are on-demand calls.

```mermaid
sequenceDiagram
  autonumber
  participant FE as Frontend
  participant BE as Your backend
  participant MP as multipay-go
  participant PG as Provider

  FE->>BE: subscribe to a plan
  BE->>MP: client.Subscriptions().CreateSubscription({ PlanDetails or PlanID })
  MP->>PG: create plan + subscription (SDK)
  PG-->>MP: subscription + mandate authorization handle
  MP-->>BE: domain.Subscription{ status }
  BE-->>FE: redirect to authorize the mandate
  FE->>PG: authorize mandate on the hosted page
  PG-->>BE: webhook subscription.activated
  BE-->>PG: 200 ACK
  Note over PG: every billing cycle — NO call from you
  PG-->>BE: webhook subscription.charged / payment_failed
  BE-->>PG: 200 ACK
  Note over BE: update invoice / run dunning policy
  Note over BE,PG: lifecycle actions, on demand
  BE->>MP: Cancel / Pause / Resume / ChangePlan / GetSubscriptionPayments
  MP->>PG: SDK call
  PG-->>MP: updated subscription
  MP-->>BE: domain.Subscription
```

---

## 8. Webhook consumption (how every confirmation arrives)

The one flow the *provider* starts — and the source of truth behind every flow above. The library's
`WebhookHandler` runs a fixed, durable, idempotent pipeline and **always answers `2xx` after the signature
check**. A `5xx` would make the provider auto-disable your endpoint, so handler errors are logged and the
event is left in the store for replay, never bubbled up as a 5xx.

```mermaid
sequenceDiagram
  autonumber
  participant PG as Provider
  participant WH as WebhookHandler (multipay-go)
  participant ST as WebhookStore (your durable store)
  participant H as Your event handler

  PG->>WH: POST /webhooks/{provider}/{account} (raw body + signature)
  WH->>ST: StoreRawPayload (audit first)
  WH->>ST: IsDuplicate? (dedupe key)
  alt duplicate
    WH-->>PG: 200 DUPLICATE_ACK
  else new event
    WH->>WH: VerifySignature (webhook secret)
    WH->>WH: ParseEvent → domain.WebhookEvent
    WH->>H: dispatch by event type (or default handler)
    H-->>WH: result (handler errors are logged, never propagated)
    WH->>ST: MarkProcessed
    WH-->>PG: 200 ACK (always, after signature check)
  end
```

> **What you do with a confirmed webhook is yours, not the library's.** The library hands you a typed
> `domain.WebhookEvent`; turning that into "mark the order paid", "activate the subscription", "grant access",
> or "send a receipt" is entirely your application's business logic.

---

## 9. Capabilities (introspection guard)

Not a money flow — a synchronous, **in-memory** check against the immutable support matrix. Use it to gate a
provider-specific operation *before* you call it, so unsupported combinations fail fast with a clear error
instead of a surprise at the provider.

```mermaid
sequenceDiagram
  autonumber
  participant BE as Your backend
  participant MP as multipay-go

  BE->>MP: client.Capabilities().Supports(provider, capability)
  MP-->>BE: bool (support matrix lookup — no provider/network call)
  alt supported
    Note over BE: proceed to call the operation
  else not supported
    Note over BE: skip / show a clear "not available" message
  end
  BE->>MP: client.Capabilities().AllCapabilities(provider)
  MP-->>BE: []Capability
```
