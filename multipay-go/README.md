---
type: Overview
title: MultiPay Adapter
status: stable
tags:
- readme
stale_after: 2026-12-31
generated:
  by: Tushar Dwivedi
  at: 2026-08-25T00:00:00Z
---

# MultiPay Adapter


| Section | What it covers |
|---------|----------------|
| [Overview](#overview) | MultiPay Adapter provides a abstraction layer over payment processing providers, allowing… |
| [Installation](#installation) | Installation |
| [Quick Start](#quick-start) | Quick Start |
| [Provider Configuration](#provider-configuration) | Provider Configuration |
| [Core Services](#core-services) | Core Services |
| [Capability Checking](#capability-checking) | Before attempting an operation, check if the provider supports it: |
| [Supported Capabilities Matrix](#supported-capabilities-matrix) | The following table shows all capabilities supported by Cashfree and Razorpay. Use this… |
| [ReturnURL Is a Browser Redirect Only](#returnurl-is-a-browser-redirect-only) | ReturnURL is a browser-redirect parameter only. When your payment gateway finishes collecting payment… |
| [Webhook Setup](#webhook-setup) | Webhook Setup |
| [Accessing Provider-Specific Details](#accessing-provider-specific-details) | While the library provides a consistent canonical API across providers, each provider… |
| [Error Handling](#error-handling) | Error Handling |
| [Multi-Instance Support](#multi-instance-support) | Create separate clients for different environments or providers: |
| [Hook Pipeline](#hook-pipeline) | Hooks allow you to extend the library with custom observability and business logic. |
| [Architecture](#architecture) | For detailed architecture information, see DESIGN.md. |
| [Contributing](#contributing) | Contributions are welcome! Please ensure: |
| [License](#license) | See LICENSE file |

A unified Go library for integrating payment providers (Cashfree, Razorpay) with a single, consistent API. Each client is bound to one provider. Process orders, payments, refunds, and webhooks without provider-specific code.

## Overview

MultiPay Adapter provides a abstraction layer over payment processing providers, allowing you to:

- **Use a single API** for all payment operations regardless of provider
- **Switch providers** by creating a new client — no application code changes (a few requests need a provider-specific identifier: for example Cashfree refunds and payment fetches need `OrderID`, Razorpay refunds need `PaymentID`)
- **Run multiple clients** side-by-side for different providers or accounts
- **Handle webhooks** with built-in deduplication and signature verification
- **Validate capabilities** before attempting operations
- **Extend with hooks** for custom observability and business logic

### Supported Providers

- **Cashfree Payment Gateway (PG)** — India, Southeast Asia
- **Razorpay** — India, Global

### Documentation

- **[DESIGN.md](./DESIGN.md)** — Architecture, design decisions, and flow diagrams

## Installation

```bash
go get github.com/Bytonomics/multipay-india/multipay-go
```

Requires Go 1.26 or later.

## Quick Start

### 1. Create an Adapter

Each client is bound to a single payment provider. Choose one:

```go
import (
    "github.com/Bytonomics/multipay-india/multipay-go/client"
    "github.com/Bytonomics/multipay-india/multipay-go/domain"
    "github.com/Bytonomics/multipay-india/multipay-go/providers/cashfree"
    "github.com/Bytonomics/multipay-india/multipay-go/providers/razorpay"
)
```

```go
// Option A: Cashfree adapter
cashfreeAdapter, err := cashfree.NewAdapter(&cashfree.Config{
    ClientID:     "your-cashfree-client-id",
    ClientSecret: "your-cashfree-secret",
    Environment:  domain.EnvironmentProduction, // or domain.EnvironmentSandbox
    Logger:       yourLogger, // mandatory: NewAdapter panics when nil; pass ports.NewNoopLogger() to discard logs
})
if err != nil {
    log.Fatal(err)
}

// Option B: Razorpay adapter
razorpayAdapter, err := razorpay.NewAdapter(&razorpay.Config{
    Key:         "your-razorpay-key",
    Secret:      "your-razorpay-secret",
    Environment: domain.EnvironmentProduction, // or domain.EnvironmentSandbox
})
if err != nil {
    log.Fatal(err)
}
```

### 2. Create Client

`WebhookStore` and `Logger` are required: `NewClient` panics when either is nil. The webhook store gives the client durable raw-payload capture and deduplication for incoming webhooks. You implement the `ports.WebhookStore` interface (`StoreRawPayload`, `IsDuplicate`, `MarkProcessed`) on top of your own database or cache; the library does not ship a ready-made store.

```go
mpClient, err := client.NewClient(&client.ClientConfig{
    Provider:     cashfreeAdapter,
    WebhookStore: yourWebhookStore, // mandatory: NewClient panics when nil; your implementation of ports.WebhookStore
    Logger:       yourLogger,       // mandatory: NewClient panics when nil; pass ports.NewNoopLogger() to discard logs
})
if err != nil {
    log.Fatal(err)
}
```

### 3. Make an Order

```go
ctx := context.Background()

order, err := mpClient.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 10000,  // 10000 paisa = ₹100.00
    Currency:    "INR",
    Customer: &domain.CustomerInfo{
        CustomerID: "cust_123",
        Email:      "user@example.com",
        Phone:      "+919876543210",
    },
    ReturnURL: "https://app.example.com/payment/return",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order created: %s (Status: %s)\n", order.OrderID, order.Status)
```

### 4. Fetch and List Payments

```go
// Fetch a single order
order, err := mpClient.Orders().GetOrder(ctx, &domain.GetOrderRequest{
    OrderID: "order_123",
})
if err != nil {
    log.Fatal(err)
}

// List all payments for an order
payments, err := mpClient.Orders().ListOrderPayments(ctx, &domain.ListOrderPaymentsRequest{
    OrderID: "order_123",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Payments: %d\n", len(payments))
```

## Provider Configuration

### Cashfree Configuration

```go
cashfreeAdapter, err := cashfree.NewAdapter(&cashfree.Config{
    ClientID:     os.Getenv("CASHFREE_CLIENT_ID"),
    ClientSecret: os.Getenv("CASHFREE_CLIENT_SECRET"),
    Environment:  domain.EnvironmentProduction,  // or domain.EnvironmentSandbox
    AccountID:    "prod",  // Optional: identifier for this account
    Logger:       yourLogger, // mandatory: NewAdapter panics when nil; pass ports.NewNoopLogger() to discard logs
})
if err != nil {
    log.Fatal(err)
}
```

**Environment Variables** (recommended):

```bash
export CASHFREE_CLIENT_ID="your-client-id"
export CASHFREE_CLIENT_SECRET="your-secret"
export CASHFREE_ENVIRONMENT="PRODUCTION"
```

The library does not read environment variables; your code reads them and passes the values in.

### Razorpay Configuration

```go
razorpayAdapter, err := razorpay.NewAdapter(&razorpay.Config{
    Key:           os.Getenv("RAZORPAY_KEY"),
    Secret:        os.Getenv("RAZORPAY_SECRET"),
    WebhookSecret: os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
    Environment:   domain.EnvironmentProduction,  // or domain.EnvironmentSandbox
    AccountID:     "prod",  // Optional: identifier for this account
})
if err != nil {
    log.Fatal(err)
}
```

**Environment Variables** (recommended):

```bash
export RAZORPAY_KEY="your-razorpay-key"
export RAZORPAY_SECRET="your-razorpay-secret"
export RAZORPAY_WEBHOOK_SECRET="your-webhook-secret"
```

The library does not read environment variables; your code reads them and passes the values in.

**Important: API Key Format Validation**

Razorpay uses API keys to determine the environment — there is no separate environment flag like Cashfree:
- **Sandbox keys** start with `rzp_test_` (e.g., `rzp_test_1DP5u41A123456`)
- **Live keys** start with `rzp_live_` (e.g., `rzp_live_1DP5u41B654321`)

The adapter validates the API key format against the configured environment at initialization time:
- If `Environment: domain.EnvironmentSandbox`, the key MUST start with `rzp_test_`
- If `Environment: domain.EnvironmentProduction`, the key MUST start with `rzp_live_`

The sandbox environment also accepts keys that start with `rzp_mock_`, which is meant for tests.

Misconfiguration returns an error from `razorpay.NewAdapter(...)`. Error messages intentionally do not include the full key.

```go
// WRONG: Live key with sandbox environment
adapter, err := razorpay.NewAdapter(&razorpay.Config{
    Key:         "rzp_live_1DP5u41B654321", // Live key
    Secret:      os.Getenv("RAZORPAY_SECRET"),
    Environment: domain.EnvironmentSandbox,   // Sandbox env
})
if err != nil {
    // err: razorpay API key must start with "rzp_test_" for environment "SANDBOX"
}

// CORRECT: Test key with sandbox environment
adapter, err = razorpay.NewAdapter(&razorpay.Config{
    Key:         "rzp_test_1DP5u41A123456",
    Secret:      os.Getenv("RAZORPAY_SECRET"),
    Environment: domain.EnvironmentSandbox,
})
if err != nil {
    log.Fatal(err)
}
```

This strict validation ensures configuration errors are caught immediately at startup, rather than causing payment failures at runtime.

## Core Services

### Orders

Create and manage payment orders.

```go
// Create an order
order, err := mpClient.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 500000,  // 500000 paisa = ₹5000.00
    Currency:    "INR",
    Customer: &domain.CustomerInfo{
        CustomerID: "cust_123",
        Email:      "user@example.com",
        Phone:      "+919876543210",
    },
    ReturnURL: "https://app.example.com/payment/return",
})

// Fetch an order
order, err := mpClient.Orders().GetOrder(ctx, &domain.GetOrderRequest{
    OrderID: "order_abc123",
})

// List all payments for an order
payments, err := mpClient.Orders().ListOrderPayments(ctx, &domain.ListOrderPaymentsRequest{
    OrderID: "order_abc123",
})
```

### Payments

Process and retrieve payment details.

```go
// Fetch a payment
payment, err := mpClient.Payments().GetPayment(ctx, &domain.GetPaymentRequest{
    PaymentID: "pay_abc123",
    OrderID:   "order_abc123", // required by Cashfree, optional for Razorpay
})

// List payments for an order
payments, err := mpClient.Payments().ListPayments(ctx, &domain.ListPaymentsRequest{
    OrderID: "order_abc123",
})

// Capture a payment (for authorized payments on Razorpay)
payment, err := mpClient.Payments().CapturePayment(ctx, &domain.CapturePaymentRequest{
    PaymentID:   "pay_abc123",
    AmountMinor: 50000,  // Capture specific amount in minor units
    Currency:    "INR",
})
```

### Refunds

Process and manage refunds.

```go
// Create a refund
refund, err := mpClient.Refunds().CreateRefund(ctx, &domain.CreateRefundRequest{
    // Cashfree identifies the refund by OrderID, Razorpay by PaymentID (set at least one)
    OrderID:     "order_abc123",
    PaymentID:   "pay_abc123",
    AmountMinor: 250000,  // Partial refund in minor units
    Currency:    "INR",
})

// Fetch a refund
refund, err := mpClient.Refunds().GetRefund(ctx, &domain.GetRefundRequest{
    RefundID: "refund_abc123",
})

// List refunds: set OrderID for Cashfree, PaymentID for Razorpay
refunds, err := mpClient.Refunds().ListRefunds(ctx, &domain.ListRefundsRequest{
    OrderID: "order_abc123",
})
```

### Instruments

Manage payment instruments (cards, wallets, etc.).

```go
// Fetch an instrument
instrument, err := mpClient.Instruments().GetInstrument(ctx, &domain.GetInstrumentRequest{
    CustomerID:   "cust_123",
    InstrumentID: "inst_abc123",
})

// List instruments for a customer
instruments, err := mpClient.Instruments().ListInstruments(ctx, &domain.ListInstrumentsRequest{
    CustomerID: "cust_123",
})

// Delete an instrument
deleted, err := mpClient.Instruments().DeleteInstrument(ctx, &domain.DeleteInstrumentRequest{
    CustomerID:   "cust_123",
    InstrumentID: "inst_abc123",
})
if err == nil {
    log.Printf("Deleted instrument: %s\n", deleted.InstrumentID)
}
```

### Payment Links

Create shareable payment links.

```go
// Create a payment link
exp := time.Now().Add(24 * time.Hour) // link expires in 24 hours
link, err := mpClient.PaymentLinks().CreatePaymentLink(ctx, &domain.CreatePaymentLinkRequest{
    AmountMinor: 1000000,  // 1000000 paisa = ₹10000.00
    Currency:    "INR",
    Purpose:     "Order 123",
    Customer: &domain.CustomerInfo{
        CustomerID: "cust_123",
        Email:      "user@example.com",
        Phone:      "+919876543210",
    },
    ExpiryTime: &exp,
})

// Fetch a payment link
link, err := mpClient.PaymentLinks().GetPaymentLink(ctx, &domain.GetPaymentLinkRequest{
    LinkID: "link_abc123",
})

// Cancel a payment link
link, err := mpClient.PaymentLinks().CancelPaymentLink(ctx, &domain.CancelPaymentLinkRequest{
    LinkID: "link_abc123",
})
```

### Webhooks

Handle incoming webhooks from payment providers.

```go
// See "Webhook Setup" section below
```

## Capability Checking

Before attempting an operation, check if the provider supports it:

```go
import (
    "github.com/Bytonomics/multipay-india/multipay-go/capabilities"
    "github.com/Bytonomics/multipay-india/multipay-go/domain"
)
```

```go
// Check if provider supports payment capture
caps := mpClient.Capabilities()
if !caps.Supports(ctx, domain.ProviderCashfree, capabilities.CapPaymentCapture) {
    return fmt.Errorf("Cashfree does not support payment capture")
}

// Attempt operation
_, err := mpClient.Payments().CapturePayment(ctx, captureReq)
if ce := &domain.CapabilityError{}; errors.As(err, &ce) {
    // Handle unsupported capability gracefully
    log.Printf("Provider %s doesn't support %s: %s\n", ce.Provider, ce.Capability, ce.Message)
    return fallbackCaptureFlow(ctx, order)
}
```

## Supported Capabilities Matrix

The following table shows all capabilities supported by Cashfree and Razorpay. Use this to determine which provider meets your feature requirements before integration.

| # | Capability | Description | Cashfree | Razorpay |
|---|---|---|---|---|
| 1 | Order Create | Create a new order | ✓ | ✓ |
| 2 | Order Fetch | Retrieve order details | ✓ | ✓ |
| 3 | Order List Payments | List all payments associated with an order | ✓ | ✓ |
| 4 | Order Update | Modify order details | ✗ | ✓ |
| 5 | Order List | List all orders | ✗ | ✓ |
| 6 | Payment Fetch | Retrieve payment details | ✓ | ✓ |
| 7 | Payment List | List all payments | ✓ | ✓ |
| 8 | Payment Pay | Initiate or process a payment | ✓ | ✓ |
| 9 | Payment Capture | Capture an authorized payment | ✗ | ✓ |
| 10 | Refund Create | Create a refund for a payment | ✓ | ✓ |
| 11 | Refund Fetch | Retrieve refund details | ✓ | ✓ |
| 12 | Refund List | List all refunds | ✓ | ✓ |
| 13 | Refund Update | Modify refund details | ✗ | ✓ |
| 14 | Instrument Fetch | Retrieve payment instrument details | ✓ | ✓ |
| 15 | Instrument List | List stored payment instruments | ✓ | ✓ |
| 16 | Instrument Delete | Delete a stored payment instrument | ✓ | ✓ |
| 17 | Instrument Cryptogram | Fetch cryptogram for tokenized instrument | ✓ | ✗ |
| 18 | Payment Link Create | Create a shareable payment link | ✓ | ✓ |
| 19 | Payment Link Fetch | Retrieve payment link details | ✓ | ✓ |
| 20 | Payment Link Cancel | Cancel an active payment link | ✓ | ✓ |
| 21 | Payment Link Update | Modify payment link details | ✗ | ✓ |
| 22 | Payment Link Notify | Send payment link notification to customer | ✗ | ✓ |
| 23 | Payment Link List | List all payment links | ✗ | ✓ |
| 24 | Payment Link List Orders | List orders within a payment link | ✓ | ✗ |
| 25 | Webhook Consume | Consume and verify webhook events | ✓ | ✓ |
| 26 | Webhook Create | Create a new webhook endpoint | ✗ | ✓ |
| 27 | Webhook Fetch | Retrieve webhook configuration | ✗ | ✓ |
| 28 | Webhook Edit | Modify webhook settings | ✗ | ✓ |
| 29 | Webhook Delete | Delete a webhook endpoint | ✗ | ✓ |
| 30 | Webhook List | List all webhook endpoints | ✗ | ✓ |
| 31 | Customer Create | Create a new customer record | ✗ | ✓ |
| 32 | Customer Fetch | Retrieve customer details | ✗ | ✓ |
| 33 | Customer Edit | Modify customer information | ✗ | ✓ |
| 34 | Customer List | List all customers | ✗ | ✓ |
| 35 | Subscription List | List all subscriptions | ✗ | ✓ |
| 36 | Subscription Manual Charge | Trigger a manual charge on an on-demand subscription | ✓ | ✓ |
| 37 | Subscription Upgrade Proration | Charge the prorated amount when upgrading a subscription | ✓ | ✓ |
| 38 | Subscription Eligibility | Check subscription eligibility | ✓ | ✗ |
| 39 | Subscription Cancel At Cycle End | Cancel a subscription at the end of the billing cycle | ✗ | ✓ |
| 40 | Plan List | List all plans | ✗ | ✓ |
| 41 | Offer Create | Create offers or promotions | ✓ | ✗ |
| 42 | Offer Fetch | Retrieve offer details | ✓ | ✗ |
| 43 | Eligibility Fetch | Check eligibility for offers | ✓ | ✗ |
| 44 | Settlement Order Fetch | Fetch settlement for a specific order | ✓ | ✗ |
| 45 | Settlement List | List all settlements | ✓ | ✗ |
| 46 | Settlement Recon Fetch | Fetch settlement reconciliation data | ✓ | ✗ |
| 47 | Recon Fetch | Fetch general reconciliation data | ✓ | ✗ |
| 48 | Settlement All | List all settlements | ✗ | ✓ |
| 49 | Settlement Fetch | Retrieve a single settlement | ✗ | ✓ |
| 50 | Settlement Reports | Retrieve settlement reports | ✗ | ✓ |
| 51 | Settlement On-Demand Create | Create an on-demand settlement | ✗ | ✓ |
| 52 | Settlement On-Demand Fetch | Retrieve an on-demand settlement | ✗ | ✓ |
| 53 | Settlement On-Demand List | List on-demand settlements | ✗ | ✓ |
| 54 | UPI Create | Create a UPI payment request | ✗ | ✓ |
| 55 | VPA Validate | Validate a UPI VPA address | ✗ | ✓ |

**Legend:**
- ✓ = Supported
- ✗ = Not supported

Use [`mpClient.Capabilities().Supports(ctx, provider, capability)`](./orchestration/capabilities.go) to check capability availability before attempting operations.

## ReturnURL Is a Browser Redirect Only

`ReturnURL` is required on every order request, and it is a browser-redirect parameter only. When the payment finishes, the user's browser goes to this URL so your frontend can display a success/failure page. Cashfree redirects the browser to it. For Razorpay, the library returns it to your frontend as `order.Checkout.CallbackURL`.

**The library does NOT implement a handler for this URL.** It is not a server-side event source. **Webhooks are the only reliable, server-side mechanism for tracking payment state changes.** Do not rely on the redirect for order fulfillment, subscription activation, or any state change that must happen server-side.

**Consequence:** `ReturnURL` must be a frontend page (e.g., `/payment/return`), not a server-side webhook endpoint. Your backend should listen for webhooks to perform server-side operations like updating order status or fulfilling orders.

```go
// CORRECT: ReturnURL is a frontend page
order, err := mpClient.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 50000,
    Currency:    "INR",
    Customer: &domain.CustomerInfo{
        CustomerID: "cust_123",
        Email:      "user@example.com",
        Phone:      "+919876543210",
    },
    ReturnURL: "https://app.example.com/payment/return",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order created: %s\n", order.OrderID)

// Webhook handler processes payment on the server
customHandler := func(ctx context.Context, event *domain.WebhookEvent) error {
    if event.EventType == domain.EventPaymentCaptured {
        // Update order status in your database
        // Fulfill the order
        // Activate subscription, etc.
    }
    return nil
}
```

## Webhook Setup

### Enabling Webhooks in Your Provider Dashboard

Before webhooks will fire in your application, you must enable the corresponding events in your payment provider's dashboard. This section lists the exact event names to enable so that the library's event constants actually deliver events to your handlers.

#### Razorpay — Events to Enable in Dashboard

Enable these under **Settings → Webhooks** in your Razorpay dashboard:

| Event Name (Razorpay Dashboard) | Library Constant |
|---|---|
| `subscription.authenticated` | `EventSubAuthenticated` |
| `subscription.activated` | `EventSubActivated` |
| `subscription.charged` | `EventSubCharged` |
| `subscription.payment_failed` | `EventSubPaymentFailed` |
| `subscription.pending` | `EventSubOnHold` |
| `subscription.halted` | `EventSubHalted` |
| `subscription.paused` | `EventSubPaused` |
| `subscription.resumed` | `EventSubResumed` |
| `subscription.cancelled` | `EventSubCancelled` |
| `subscription.completed` | `EventSubCompleted` |
| `subscription.updated` | `EventSubUpdated` |
| `payment.authorized` | `EventPaymentAuthorized` |
| `payment.captured` | `EventPaymentCaptured` |
| `payment.failed` | `EventPaymentFailed` |
| `refund.created` | `EventRefundCreated` |
| `refund.processed` | `EventRefundProcessed` |
| `refund.failed` | `EventRefundFailed` |

#### Cashfree — Events to Enable in Dashboard (API Version 2023-08-01)

Enable these under **Developers → Webhooks** in your Cashfree merchant dashboard:

| Event Name (Cashfree Dashboard) | Library Constant(s) |
|---|---|
| `SUBSCRIPTION_AUTH_STATUS` | `EventSubAuthenticated` |
| `SUBSCRIPTION_PAYMENT_SUCCESS` | `EventSubCharged` |
| `SUBSCRIPTION_PAYMENT_FAILED` | `EventSubPaymentFailed` |
| `SUBSCRIPTION_STATUS_CHANGED` | `EventSubActivated`, `EventSubOnHold`, `EventSubPaused`, `EventSubCancelled`, `EventSubCompleted`, `EventSubExpired`, `EventSubBankApprovalPending`, `EventSubCardExpired` |
| `SUBSCRIPTION_REFUND_STATUS` | `EventSubRefund` |
| `SUBSCRIPTION_CARD_EXPIRY_REMINDER` | `EventSubCardExpiring` |
| `SUBSCRIPTION_PAYMENT_CANCELLED` | `EventSubPaymentCancelled` |
| `SUBSCRIPTION_PAYMENT_NOTIFICATION_INITIATED` | `EventSubPreDebitNotice` |
| `ORDER.PAID` | `EventPaymentCaptured` |
| `ORDER.EXPIRED` | `EventOrderExpired` |
| `PAYMENT.AUTHORIZED` | `EventPaymentAuthorized` |
| `PAYMENT.FAILED` | `EventPaymentFailed` |
| `REFUND.PROCESSED` | `EventRefundProcessed` |
| `REFUND.FAILED` | `EventRefundFailed` |

**Important:** If you register an event handler (e.g., `EventSubCharged`) but don't enable the corresponding event in your provider's dashboard, your handler will never be called. Always verify both:
1. The event is enabled in the provider dashboard
2. Your webhook endpoint is registered with the correct URL
3. Your handler is registered for the correct event type

### How Webhook URLs Work

You choose the webhook URL and register it in your provider's dashboard:

1. **You define the endpoint path** in your server (e.g., `/webhooks/cashfree/prod`)
2. **You register this URL** in Cashfree/Razorpay's dashboard as the webhook endpoint
3. **The provider sends POST requests** to your URL when payment events occur
4. **Our library processes them** via the `WebhookHandler` (an `http.Handler`)

The URL pattern is `{your-base-url}/webhooks/{provider}/{accountID}`:
- `provider` = `cashfree` or `razorpay` (matches the provider that will send to this endpoint)
- `accountID` = your identifier for this provider account (e.g., `prod`, `sandbox`, `merchant_123`)

Example: Register `https://api.yourapp.com/webhooks/cashfree/prod` in Cashfree's dashboard.

### Mounting Webhook Handler

`mpClient.Webhooks().Handler(basePath)` returns a standard `http.Handler`, so it mounts on any Go HTTP router. The handler parses the full request path `{basePath}/{provider}/{accountID}`, so mount it on a subtree route that keeps `basePath` in the path. Do not strip the prefix, and do not mount it on the exact `basePath` alone: the real webhook URL would return a 400 `INVALID_ENDPOINT` response. The handler accepts `POST` requests only.

**IMPORTANT:** Call `RegisterHandler` for every event type (see [Custom Event Handlers](#custom-event-handlers)) *before* you call `Handler`. `Handler` copies the handlers registered at that moment, so a handler registered afterwards is never called.

With `basePath := "/webhooks"`, the mounted routes match the URL pattern `{your-base-url}/webhooks/{provider}/{accountID}` described above.

#### With net/http

```go
basePath := "/webhooks"

mux := http.NewServeMux()
mux.Handle(basePath+"/", mpClient.Webhooks().Handler(basePath)) // trailing slash mounts the subtree
http.ListenAndServe(":8080", mux)
```

#### With `chi` Router

```go
basePath := "/webhooks"

r := chi.NewRouter()
r.Handle(basePath+"/*", mpClient.Webhooks().Handler(basePath))
```

#### With `echo` Framework

```go
basePath := "/webhooks"

e := echo.New()
e.Any(basePath+"/*", echo.WrapHandler(mpClient.Webhooks().Handler(basePath)))
```

#### With `gin` Framework

```go
basePath := "/webhooks"

r := gin.Default()
r.Any(basePath+"/*path", gin.WrapH(mpClient.Webhooks().Handler(basePath)))
```

### Custom Event Handlers

Register custom handlers for specific webhook events:

```go
// Create a custom webhook event handler
customHandler := func(ctx context.Context, event *domain.WebhookEvent) error {
    if event.ParseError != "" {
        // The event could not be fully parsed; the raw body is still in event.RawPayload
        log.Printf("Webhook parse error for %s: %s\n", event.RawVendorEventType, event.ParseError)
    }

    switch event.EventType {
    case domain.EventPaymentCaptured:
        log.Printf("Payment captured (vendor event %s, account %s)\n", event.RawVendorEventType, event.AccountID)
        // Update order status, fulfill order, send email, etc.

    case domain.EventPaymentFailed:
        log.Printf("Payment failed (vendor event %s, account %s)\n", event.RawVendorEventType, event.AccountID)
        // Notify customer, retry or cancel order

    case domain.EventRefundCreated:
        log.Printf("Refund created (vendor event %s, account %s)\n", event.RawVendorEventType, event.AccountID)
        // Update refund status, notify customer

    default:
        log.Printf("Unhandled webhook event: %s (vendor event %s)\n", event.EventType, event.RawVendorEventType)
    }
    return nil
}

// Register the handler once per event type it handles.
// Do this BEFORE calling mpClient.Webhooks().Handler(basePath).
for _, eventType := range []domain.WebhookEventType{
    domain.EventPaymentCaptured,
    domain.EventPaymentFailed,
    domain.EventRefundCreated,
} {
    mpClient.Webhooks().RegisterHandler(eventType, customHandler)
}
```

Registering a second handler for the same event type replaces the first. Events with no registered handler are acknowledged with a 2xx response and logged.

Every event has `EventType`, `AccountID`, `RawPayload` and `RawHeaders` set. `ParseError` is non-empty when the body could not be fully parsed, and the event type is then `EventUnknown` unless the adapter already identified it. `RawVendorEventType` carries the provider's own event name when the adapter could read it. The provider-specific objects on the event (`Payment`, `Refund`, `Order`, `Subscription`) are filled when the adapter can parse them, so check for nil before use.

Hooks do not run for webhooks served through `Webhooks().Handler()`.

### Deduplication and Idempotency

The webhook endpoint deduplicates by the SHA-256 hash (hex) of the raw request body. Your `ports.WebhookStore` decides what counts as a duplicate: the library calls `IsDuplicate` with that hash. For each request the endpoint:

1. Accepts `POST` only (otherwise 405)
2. Matches the path `{basePath}/{provider}/{accountID}` (otherwise 400 `INVALID_ENDPOINT`)
3. Reads the body (400 `READ_BODY_FAILED` on failure)
4. Computes the body hash
5. Stores the raw payload with `StoreRawPayload` (500 `STORAGE_ERROR` on failure)
6. Calls `IsDuplicate` (500 `DEDUP_ERROR` on failure)
7. Acknowledges a duplicate with 200 `DUPLICATE_ACK` and stops; your handler is not called
8. Checks that the provider in the URL matches the client's provider (otherwise 400 `PROVIDER_MISMATCH`)
9. Verifies the signature (400 `SIGNATURE_INVALID`, or 500 `VERIFICATION_ERROR` on other errors)
10. Parses the event on a best-effort basis; a parse failure is stored in `event.ParseError` and does not stop dispatch
11. Calls the handler registered for the event type, or the default handler (it logs and returns nil)
12. On handler success calls `MarkProcessed`; on handler error logs it and does not mark the event processed
13. Responds 200 `ACK`, after signature verification, whether or not your handler succeeded

A redelivery of an already-processed payload is acknowledged without calling your handler. If your handler returns an error, the event is not marked processed and the endpoint still returns 200, so the event can be handled again on redelivery or replay. Write handlers to be idempotent. Your `IsDuplicate` implementation should return true once an event has been marked processed.

## Accessing Provider-Specific Details

While the library provides a consistent canonical API across providers, each provider returns unique fields and metadata that may be important for your application. These provider-specific details are captured in strongly-typed structs and accessible through the `ProviderDetails` field on response objects.

### ProviderDetails Fields

All response types (Order, Payment, Refund, Instrument, PaymentLink) include a `ProviderDetails` field containing provider-specific data:

```go
order, err := mpClient.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 10000,
    Currency:    "INR",
    Customer: &domain.CustomerInfo{
        CustomerID: "cust_123",
        Email:      "user@example.com",
        Phone:      "+919876543210",
    },
    ReturnURL: "https://app.example.com/payment/return",
})
if err != nil {
    log.Fatal(err)
}

// Access Cashfree-specific order details
if order.ProviderDetails != nil && order.ProviderDetails.Cashfree != nil {
    cf := order.ProviderDetails.Cashfree
    log.Printf("Cashfree Order ID: %s\n", cf.CfOrderID)
    log.Printf("Order Entity: %s\n", cf.Entity)
    if cf.OrderMeta != nil {
        log.Printf("Return URL: %s\n", cf.OrderMeta.ReturnURL)
    }
}

// Access Razorpay-specific order details
if order.ProviderDetails != nil && order.ProviderDetails.Razorpay != nil {
    rz := order.ProviderDetails.Razorpay
    log.Printf("Razorpay Receipt: %s\n", rz.Receipt)
    log.Printf("Offer ID: %s\n", rz.OfferID)
    log.Printf("Amount Paid: %d\n", rz.AmountPaid)
}
```

### Provider-Specific Fields by Type

**Order Details:**
- **Cashfree**: Order ID, Entity type, Order note, Vendor splits, Return/Notify URLs, Payment methods
- **Razorpay**: Receipt ID, Offer ID, Amount paid, Amount due, Attempt count

**Payment Details:**
- **Cashfree**: Payment ID, Order amount/currency, Payment message, Auth ID, Error details
- **Razorpay**: Description, Email, Contact, Fees, Tax, Refund status, International flag, Card/VPA/Wallet info, Acquirer data

**Refund Details:**
- **Cashfree**: Refund ID, Payment ID, Refund charge/type/mode, Status, Refund speed, Vendor splits, Forex charges
- **Razorpay**: Receipt, Speed (requested/processed), Batch ID, Acquirer data

**Instrument Details:**
- **Cashfree**: Instrument UID, Card network, Bank name, Card type
- **Razorpay**: Token, Max payment amount, Expiry timestamp, Compliance flag

**Payment Link Details:**
- **Cashfree**: Link ID, Partial payments, Min partial amount, Auto reminders, QR code, Vendor splits
- **Razorpay**: Description, Callback URL/Method, Reminder enabled, Payment count, Min partial amount

### Why Provider Details Matter

Provider-specific details are useful for:
- **Compliance Reporting**: Access settlement and reconciliation data specific to each provider
- **Advanced Features**: Use provider-specific fields like Cashfree's vendor splits or Razorpay's offers
- **Debugging**: Access error codes and provider-specific error details
- **Analytics**: Track provider-specific metrics like payment attempts or refund speeds
- **Custom UI**: Display provider-specific information to customers or admins

## Error Handling

### Sentinel Errors

Common errors are represented as sentinel errors that you can check with `errors.Is()`:

```go
import "github.com/Bytonomics/multipay-india/multipay-go/domain"

order, err := mpClient.Orders().GetOrder(ctx, &domain.GetOrderRequest{
    OrderID: "nonexistent",
})
if errors.Is(err, domain.ErrOrderNotFound) {
    log.Println("Order does not exist")
    // Handle missing order
}

refund, err := mpClient.Refunds().CreateRefund(ctx, req)
if errors.Is(err, domain.ErrProviderError) {
    log.Println("Provider returned an error")
    // Handle provider error
}
```

**Common Sentinel Errors:**

- `ErrOrderNotFound` — Order does not exist
- `ErrPaymentNotFound` — Payment does not exist
- `ErrRefundNotFound` — Refund does not exist
- `ErrPaymentLinkNotFound` — Payment link does not exist
- `ErrInstrumentNotFound` — Instrument (card, wallet) does not exist
- `ErrInvalidRequest` — Request validation failed
- `ErrProviderError` — Provider returned an API error
- `ErrUnsupportedCapability` — Provider does not support this operation

### CapabilityError

Returned when a provider doesn't support an operation:

```go
_, err := mpClient.Refunds().CreateRefund(ctx, req)
if ce := &domain.CapabilityError{}; errors.As(err, &ce) {
    log.Printf("Provider %s does not support %s: %s\n",
        ce.Provider, ce.Capability, ce.Message)
    // Fall back to alternative flow
}
```

### ProviderAPIError

Wraps provider-specific errors with context:

```go
_, err := mpClient.Payments().GetPayment(ctx, &domain.GetPaymentRequest{
    PaymentID: "pay_123",
})
if pe := &domain.ProviderAPIError{}; errors.As(err, &pe) {
    log.Printf("Provider %s error (code=%s): %s\n",
        pe.Provider, pe.ErrorCode, pe.Message)
    // Handle provider-specific error
}
```

### Validation Errors

Request validation errors from pedantigo:

```go
_, err := mpClient.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    // Missing required fields (Customer, Currency, ReturnURL)
})
if ve := &domain.ValidationError{}; errors.As(err, &ve) {
    log.Printf("Validation failed: %v\n", ve)
    // Log all validation errors and reject request
}
```

## Multi-Instance Support

Create separate clients for different environments or providers:

```go
// Production: Cashfree
prodClient, err := client.NewClient(&client.ClientConfig{
    Provider:     prodCashfreeAdapter,
    WebhookStore: yourWebhookStore, // mandatory: your implementation of ports.WebhookStore
    Logger:       yourLogger,
})
if err != nil {
    log.Fatalf("failed to create production client: %v", err) // invalid config (a nil adapter, WebhookStore or Logger panics instead)
}

// Staging: Razorpay
stagingClient, err := client.NewClient(&client.ClientConfig{
    Provider:     stagingRazorpayAdapter,
    WebhookStore: yourWebhookStore, // mandatory: your implementation of ports.WebhookStore
    Logger:       yourLogger,
})
if err != nil {
    log.Fatalf("failed to create staging client: %v", err)
}

// Route based on environment
var mpClient *client.MultiPayClient
if os.Getenv("ENV") == "prod" {
    mpClient = prodClient
} else {
    mpClient = stagingClient
}
```

### Thread Safety

Register all webhook handlers before you serve traffic; after that the client is safe for concurrent use and can be shared across goroutines:

```go
// Safe: Share client across goroutines
go func() {
    order, err := mpClient.Orders().GetOrder(ctx, &domain.GetOrderRequest{
        OrderID: "order_1",
    })
    if err != nil {
        log.Printf("failed to fetch order: %v", err) // ErrOrderNotFound, ErrProviderError, etc.
    }
    _ = order
}()

go func() {
    order, err := mpClient.Orders().GetOrder(ctx, &domain.GetOrderRequest{
        OrderID: "order_2",
    })
    if err != nil {
        log.Printf("failed to fetch order: %v", err)
    }
    _ = order
}()
```

### Multiple Cashfree Accounts

If you have multiple Cashfree accounts, create separate adapters and clients:

```go
prodAdapter, err := cashfree.NewAdapter(&cashfree.Config{
    ClientID:     "prod-client-id",
    ClientSecret: "prod-secret",
    Environment:  domain.EnvironmentProduction,
    AccountID:    "prod",
    Logger:       yourLogger,
})
if err != nil {
    log.Fatalf("failed to create prod adapter: %v", err) // missing credentials or invalid environment
}

sandboxAdapter, err := cashfree.NewAdapter(&cashfree.Config{
    ClientID:     "sandbox-client-id",
    ClientSecret: "sandbox-secret",
    Environment:  domain.EnvironmentSandbox,
    AccountID:    "sandbox",
    Logger:       yourLogger,
})
if err != nil {
    log.Fatalf("failed to create sandbox adapter: %v", err)
}

prodClient, err := client.NewClient(&client.ClientConfig{
    Provider:     prodAdapter,
    WebhookStore: yourWebhookStore, // mandatory: your implementation of ports.WebhookStore
    Logger:       yourLogger,
})
if err != nil {
    log.Fatalf("failed to create prod client: %v", err)
}

sandboxClient, err := client.NewClient(&client.ClientConfig{
    Provider:     sandboxAdapter,
    WebhookStore: yourWebhookStore, // mandatory: your implementation of ports.WebhookStore
    Logger:       yourLogger,
})
if err != nil {
    log.Fatalf("failed to create sandbox client: %v", err)
}
```

### Multi-Account Webhook Routing

Each client has its own webhook handler, bound to that client's single provider adapter. The account ID in the URL does not select an adapter, and a webhook for a different provider than the client's is rejected with 400 `PROVIDER_MISMATCH`.

To receive webhooks for several accounts of one provider, create one client per account and mount each client's `Webhooks().Handler(basePath)` under a different `basePath`. The handler expects `{basePath}/{provider}/{accountID}`:

```go
mux := http.NewServeMux()
mux.Handle("/webhooks/prod/", prodClient.Webhooks().Handler("/webhooks/prod"))
mux.Handle("/webhooks/sandbox/", sandboxClient.Webhooks().Handler("/webhooks/sandbox"))
```

The resulting webhook URLs are:

- `POST /webhooks/prod/cashfree/prod` → handled by `prodClient`
- `POST /webhooks/sandbox/cashfree/sandbox` → handled by `sandboxClient`

Register each URL separately in the corresponding provider dashboard.

## Hook Pipeline

Hooks allow you to extend the library with custom observability and business logic.

### Built-In Hooks

The library ships two hooks for audit logging and metrics. No hook is enabled automatically; add the ones you want through `ClientConfig.Hooks`:

```go
// Audit hook: logs when each operation starts, completes or fails
auditHook := hooks.NewAuditHook(yourLogger)

// Metrics hook: forwards operation start, success and error events
// to your own hooks.MetricsCollector implementation
metricsHook := hooks.NewMetricsHook(yourMetricsCollector)

mpClient, err := client.NewClient(&client.ClientConfig{
    Provider:     cashfreeAdapter,
    Hooks:        []ports.Hook{auditHook, metricsHook},
    WebhookStore: yourWebhookStore,
    Logger:       yourLogger,
})
if err != nil {
    log.Fatal(err)
}
```

The `hooks` and `ports` packages are *github.com/Bytonomics/multipay-india/multipay-go/hooks* and *github.com/Bytonomics/multipay-india/multipay-go/ports*.

### Hook Use Cases

Hooks are middleware for payment operations. Here are practical examples of what you can build:

**Before Hooks** (run before every payment operation):
- **Idempotency**: Check if this request was already processed to prevent double-charging
- **Rate Limiting**: Throttle requests per provider to stay within API limits
- **Tracing**: Inject OpenTelemetry spans into the context for distributed tracing
- **Request Logging**: Log operation start with provider, amount, and customer ID

**After Hooks** (run after successful operations):
- **Cache Invalidation**: Clear cached order/payment state after mutations
- **Notifications**: Send Slack alerts for payments above a threshold
- **Analytics**: Record business metrics (revenue by provider, avg order value)
- **Audit Trail**: Write immutable audit log entries for compliance

**OnError Hooks** (run when operations fail):
- **Alerting**: Page on-call if payment failure rate exceeds 5%
- **Retry Queuing**: Enqueue transient failures for automatic retry
- **Error Categorization**: Tag errors as provider-side vs client-side for dashboards
- **Fallback Logging**: Ensure failures are always recorded even if the main logger is down

### Custom Hooks

Custom hooks implement the `ports.Hook` interface and can be supplied via `client.ClientConfig.Hooks`.

```go
type LoggingHook struct{}

func (h *LoggingHook) Before(ctx context.Context, hookCtx *ports.HookContext) (context.Context, error) {
    log.Printf("Starting: %s on %s\n", hookCtx.RequestType, hookCtx.Provider)
    return ctx, nil
}

func (h *LoggingHook) After(ctx context.Context, hookCtx *ports.HookContext) error {
    duration := time.Since(hookCtx.StartTime)
    log.Printf("Completed: %s on %s (duration=%dms)\n",
        hookCtx.RequestType, hookCtx.Provider, duration.Milliseconds())
    return nil
}

func (h *LoggingHook) OnError(ctx context.Context, hookCtx *ports.HookContext) error {
    log.Printf("Failed: %s on %s (error=%v)\n",
        hookCtx.RequestType, hookCtx.Provider, hookCtx.Error)
    return nil
}

customHook := &LoggingHook{}

cashfreeClient, err := client.NewClient(&client.ClientConfig{
    Provider:     cashfreeAdapter,
    Hooks:        []ports.Hook{customHook},
    WebhookStore: yourWebhookStore, // mandatory: your implementation of ports.WebhookStore
    Logger:       yourLogger,
})
if err != nil {
    // handle failure
}
```

### Hook Context

Hooks receive a HookContext with details about the operation:

```go
type HookContext struct {
    Provider     domain.Provider // domain.ProviderCashfree, domain.ProviderRazorpay
    RequestType  string          // "CreateOrder", "GetPayment", "CreateRefund", etc.
    RequestData  any     // Original request struct
    ResponseData any     // Response struct (nil in before/error-hook)
    Error        error           // Error (nil if no error)
    StartTime    time.Time       // When operation started
    Metadata     map[string]any  // For hooks to pass arbitrary data down the pipeline
}
```

## Architecture

For detailed architecture information, see [DESIGN.md](./DESIGN.md).

### Key Concepts

- **Canonical Types**: All requests/responses use library-defined types; providers return provider-specific types that are immediately mapped to canonical types
- **Hook Pipeline**: Every operation follows: validate → check capability → execute before-hooks → call adapter → execute after/error-hooks
- **Deduplication**: Webhook payloads are stored and deduplicated to ensure idempotency
- **Capability Matrix**: Check if a provider supports an operation before attempting it
- **Thread-Safe**: Register all webhook handlers before you serve traffic; after that the client is safe for concurrent use. No global state

### Flow Diagrams

See [DESIGN.md](./DESIGN.md) for Mermaid diagrams showing:

- Architecture overview
- Hook pipeline flow
- Webhook routing flow
- Capability validation decision tree
- Dependency injection construction flow

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass
2. New code includes tests
3. Code follows Go conventions (gofmt, golint)
4. Commit messages are clear and descriptive

## License

See [LICENSE](../LICENSE) file