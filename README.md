# MultiPay Adapter

A unified Go library for integrating payment providers (Cashfree, Razorpay) with a single, consistent API. Each client is bound to one provider. Process orders, payments, refunds, and webhooks without provider-specific code.

## Overview

MultiPay Adapter provides a abstraction layer over payment processing providers, allowing you to:

- **Use a single API** for all payment operations regardless of provider
- **Switch providers** by creating a new client — no application code changes
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
go get github.com/Bytonomics/multipay-adapter
```

**Go Version Requirement**: Go 1.26 or later

## Amounts Are Always in Minor Currency Units

**All amounts in this library use `AmountMinor` (`int64`) — the smallest unit of the currency.**

This is the single most important API contract. Getting it wrong means charging 100x too little or too much.

| Currency | Minor Unit | 1 Major Unit | To charge ₹500 / $500 / ¥500 |
|----------|-----------|-------------|-------------------------------|
| INR | paisa | 100 paisa = ₹1 | `AmountMinor: 50000` |
| USD | cent | 100 cents = $1 | `AmountMinor: 50000` |
| EUR | cent | 100 cents = €1 | `AmountMinor: 50000` |
| JPY | yen | 1 yen = ¥1 (no subdivision) | `AmountMinor: 500` |
| BHD | fils | 1000 fils = 1 BHD | `AmountMinor: 500000` |
| KWD | fils | 1000 fils = 1 KWD | `AmountMinor: 500000` |

```go
// CORRECT — charge ₹500.00
order, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 50000,  // 50000 paisa = ₹500.00
    Currency:    "INR",
    // ...
})

// WRONG — this charges ₹5.00, not ₹500!
order, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 500,    // 500 paisa = ₹5.00
    Currency:    "INR",
    // ...
})

// CORRECT — charge ¥500 (JPY has no minor unit)
order, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 500,    // 500 yen = ¥500 (exponent 0)
    Currency:    "JPY",
    // ...
})
```

The library handles provider-specific conversion internally:
- **Cashfree** receives amounts in major units (rupees/dollars) — the library converts using ISO 4217 exponents via `bojanz/currency`
- **Razorpay** receives amounts in minor units (paisa/cents) — no conversion needed

You never need to worry about provider differences. Just pass `AmountMinor` consistently.

## Quick Start

### 1. Create an Adapter

Each client is bound to a single payment provider. Choose one:

```go
// Option A: Cashfree adapter
cashfreeAdapter, err := providers.NewCashfreeAdapter(&providers.CashfreeConfig{
    ClientID:     "your-cashfree-client-id",
    ClientSecret: "your-cashfree-secret",
    Environment:  domain.EnvironmentProduction, // or domain.EnvironmentSandbox
})
if err != nil {
    log.Fatal(err)
}

// Option B: Razorpay adapter
razorpayAdapter, err := providers.NewRazorpayAdapter(&providers.RazorpayConfig{
    Key:         "your-razorpay-key",
    Secret:      "your-razorpay-secret",
    Environment: domain.EnvironmentProduction, // or domain.EnvironmentSandbox
})
if err != nil {
    log.Fatal(err)
}
```

### 2. Create Client

```go
mpClient, err := client.NewClient(&client.ClientConfig{
    Provider: cashfreeAdapter,
    Logger:   yourLogger,
})
if err != nil {
    log.Fatal(err)
}
```

### 3. Make an Order

```go
ctx := context.Background()

order, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 10000,  // 100 paisa = ₹1.00
    Currency:    "INR",
    CustomerInfo: &domain.CustomerInfo{
        ID:    "cust_123",
        Email: "user@example.com",
        Phone: "+919876543210",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order created: %s (Status: %s)\n", order.ID, order.Status)
```

### 4. Fetch and List Payments

```go
// Fetch a single order
order, err := client.Orders().GetOrder(ctx, &domain.GetOrderRequest{
    OrderID: "order_123",
})
if err != nil {
    log.Fatal(err)
}

// List all payments for an order
payments, err := client.Orders().ListOrderPayments(ctx, &domain.ListOrderPaymentsRequest{
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
cashfreeAdapter := providers.NewCashfreeAdapter(&providers.CashfreeConfig{
    ClientID:     os.Getenv("CASHFREE_CLIENT_ID"),
    ClientSecret: os.Getenv("CASHFREE_CLIENT_SECRET"),
    Environment:  domain.EnvironmentProduction,  // or domain.EnvironmentSandbox
    AccountID:    "prod",  // Optional: identifier for this account
})
```

**Environment Variables** (recommended):

```bash
export CASHFREE_CLIENT_ID="your-client-id"
export CASHFREE_CLIENT_SECRET="your-secret"
export CASHFREE_ENVIRONMENT="PROD"
```

### Razorpay Configuration

```go
razorpayAdapter := providers.NewRazorpayAdapter(&providers.RazorpayConfig{
    Key:           os.Getenv("RAZORPAY_KEY"),
    Secret:        os.Getenv("RAZORPAY_SECRET"),
    WebhookSecret: os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
    Environment:   domain.EnvironmentProduction,  // or domain.EnvironmentSandbox
    AccountID:     "prod",  // Optional: identifier for this account
})
```

**Environment Variables** (recommended):

```bash
export RAZORPAY_KEY="your-razorpay-key"
export RAZORPAY_SECRET="your-razorpay-secret"
```

**Important: API Key Format Validation**

Razorpay uses API keys to determine the environment — there is no separate environment flag like Cashfree:
- **Sandbox keys** start with `rzp_test_` (e.g., `rzp_test_1DP5u41A123456`)
- **Live keys** start with `rzp_live_` (e.g., `rzp_live_1DP5u41B654321`)

The adapter validates the API key format against the configured environment at initialization time:
- If `Environment: domain.EnvironmentSandbox`, the key MUST start with `rzp_test_`
- If `Environment: domain.EnvironmentProduction`, the key MUST start with `rzp_live_`

Misconfiguration returns an error from `providers.NewRazorpayAdapter(...)`. Error messages intentionally do not include the full key.

```go
// WRONG: Live key with sandbox environment
adapter, err := providers.NewRazorpayAdapter(&providers.RazorpayConfig{
    Key:         "rzp_live_1DP5u41B654321", // Live key
    Secret:      os.Getenv("RAZORPAY_SECRET"),
    Environment: domain.EnvironmentSandbox,   // Sandbox env
})
if err != nil {
    // err: razorpay API key must start with "rzp_test_" for environment "SANDBOX"
}

// CORRECT: Test key with sandbox environment
adapter, err = providers.NewRazorpayAdapter(&providers.RazorpayConfig{
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
order, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 500000,  // 500000 paisa = ₹5000.00
    Currency:   "INR",
    CustomerInfo: &domain.CustomerInfo{
        ID:    "cust_123",
        Email: "user@example.com",
    },
})

// Fetch an order
order, err := client.Orders().GetOrder(ctx, &domain.GetOrderRequest{
    OrderID: "order_abc123",
})

// List all payments for an order
payments, err := client.Orders().ListOrderPayments(ctx, &domain.ListOrderPaymentsRequest{
    OrderID: "order_abc123",
})
```

### Payments

Process and retrieve payment details.

```go
// Fetch a payment
payment, err := client.Payments().GetPayment(ctx, &domain.GetPaymentRequest{
    PaymentID: "pay_abc123",
})

// List payments for an order
payments, err := client.Payments().ListPayments(ctx, &domain.ListPaymentsRequest{
    OrderID: "order_abc123",
    Limit:   10,
})

// Capture a payment (for authorized payments on Razorpay)
payment, err := client.Payments().CapturePayment(ctx, &domain.CapturePaymentRequest{
    PaymentID: "pay_abc123",
    Amount:    50000,  // Capture specific amount in minor units
})
```

### Refunds

Process and manage refunds.

```go
// Create a refund
refund, err := client.Refunds().CreateRefund(ctx, &domain.CreateRefundRequest{
    PaymentID:   "pay_abc123",
    AmountMinor: 250000,  // Partial refund in minor units
})

// Fetch a refund
refund, err := client.Refunds().GetRefund(ctx, &domain.GetRefundRequest{
    RefundID: "refund_abc123",
})

// List refunds for an order
refunds, err := client.Refunds().ListRefunds(ctx, &domain.ListRefundsRequest{
    OrderID: "order_abc123",
})
```

### Instruments

Manage payment instruments (cards, wallets, etc.).

```go
// Fetch an instrument
instrument, err := client.Instruments().GetInstrument(ctx, &domain.GetInstrumentRequest{
    InstrumentID: "inst_abc123",
})

// List instruments for a customer
instruments, err := client.Instruments().ListInstruments(ctx, &domain.ListInstrumentsRequest{
    CustomerID: "cust_123",
})

// Delete an instrument
err := client.Instruments().DeleteInstrument(ctx, &domain.DeleteInstrumentRequest{
    InstrumentID: "inst_abc123",
})
```

### Payment Links

Create shareable payment links.

```go
// Create a payment link
link, err := client.PaymentLinks().CreatePaymentLink(ctx, &domain.CreatePaymentLinkRequest{
    AmountMinor: 1000000,  // 1000000 paisa = ₹10000.00
    Currency:   "INR",
    CustomerInfo: &domain.CustomerInfo{
        Email: "user@example.com",
    },
    ExpiryTime: int64(24 * 60 * 60),  // 24 hours in seconds
})

// Fetch a payment link
link, err := client.PaymentLinks().GetPaymentLink(ctx, &domain.GetPaymentLinkRequest{
    LinkID: "link_abc123",
})

// Cancel a payment link
link, err := client.PaymentLinks().CancelPaymentLink(ctx, &domain.CancelPaymentLinkRequest{
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
// Check if provider supports refund creation
caps := client.Capabilities()
if !caps.Supports(domain.ProviderCashfree, capabilities.CapRefundCreate) {
    return fmt.Errorf("Cashfree does not support refunds")
}

// Attempt operation
refund, err := client.Refunds().CreateRefund(ctx, req)
if ce := &domain.CapabilityError{}; errors.As(err, &ce) {
    // Handle unsupported capability gracefully
    log.Printf("Provider %s doesn't support %s: %s\n", ce.Provider, ce.Capability, ce.Message)
    return fallbackRefundFlow(ctx, order)
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
| 35 | Subscription Create | Create a subscription | ✗ | ✓ |
| 36 | Subscription Fetch | Retrieve subscription details | ✗ | ✓ |
| 37 | Subscription List | List all subscriptions | ✗ | ✓ |
| 38 | Plan Create | Create a billing plan | ✗ | ✓ |
| 39 | Plan Fetch | Retrieve plan details | ✗ | ✓ |
| 40 | Plan List | List all plans | ✗ | ✓ |
| 41 | Offer Create | Create offers or promotions | ✓ | ✗ |
| 42 | Offer Fetch | Retrieve offer details | ✓ | ✗ |
| 43 | Eligibility Fetch | Check eligibility for offers | ✓ | ✗ |
| 44 | Settlement Order Fetch | Fetch settlement for a specific order | ✓ | ✓ |
| 45 | Settlement List | List all settlements | ✓ | ✓ |
| 46 | Settlement Recon Fetch | Fetch settlement reconciliation data | ✓ | ✓ |
| 47 | Recon Fetch | Fetch general reconciliation data | ✓ | ✓ |
| 48 | UPI Create | Create a UPI payment request | ✗ | ✓ |
| 49 | VPA Validate | Validate a UPI VPA address | ✗ | ✓ |

**Legend:**
- ✓ = Supported
- ✗ = Not supported

Use [`client.Capabilities().Supports(provider, capability)`](./golang/multipay-adapter/orchestration/capabilities.go) to check capability availability before attempting operations.

## Webhook Setup

### How Webhook URLs Work

You choose the webhook URL and register it in your provider's dashboard:

1. **You define the endpoint path** in your server (e.g., `/webhooks/cashfree/acct_main`)
2. **You register this URL** in Cashfree/Razorpay's dashboard as the webhook endpoint
3. **The provider sends POST requests** to your URL when payment events occur
4. **Our library processes them** via the `WebhookHandler` (an `http.Handler`)

The URL pattern is `{your-base-url}/webhooks/{provider}/{accountID}`:
- `provider` = `cashfree` or `razorpay` (matches the provider that will send to this endpoint)
- `accountID` = your identifier for this provider account (e.g., `prod`, `sandbox`, `merchant_123`)

Example: Register `https://api.yourapp.com/webhooks/cashfree/prod` in Cashfree's dashboard.

### Mounting Webhook Handler

#### With `net/http`

```go
http.HandleFunc("/webhooks/payments", func(w http.ResponseWriter, r *http.Request) {
    err := client.Webhooks().Handle(r.Context(), r, w)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
})
http.ListenAndServe(":8080", nil)
```

#### With `chi` Router

```go
r := chi.NewRouter()
r.Post("/webhooks/payments", func(w http.ResponseWriter, r *http.Request) {
    err := client.Webhooks().Handle(r.Context(), r, w)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
})
```

#### With `echo` Framework

```go
e := echo.New()
e.POST("/webhooks/payments", func(c echo.Context) error {
    r := c.Request()
    w := c.Response().Writer
    return client.Webhooks().Handle(r.Context(), r, w)
})
```

#### With `gin` Framework

```go
r := gin.Default()
r.POST("/webhooks/payments", func(c *gin.Context) {
    err := client.Webhooks().Handle(c.Request.Context(), c.Request, c.Writer)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    }
})
```

### Custom Event Handlers

Register custom handlers for specific webhook events:

```go
// Create a custom webhook event handler
customHandler := func(ctx context.Context, event *domain.WebhookEvent) error {
    switch event.EventType {
    case domain.EventPaymentCaptured:
        if event.Payment != nil {
            log.Printf("Payment successful: %s\n", event.Payment.ProviderPaymentID)
        }
        // Update order status, fulfill order, send email, etc.

    case domain.EventPaymentFailed:
        if event.Payment != nil {
            log.Printf("Payment failed: %s\n", event.Payment.ProviderPaymentID)
        }
        // Notify customer, retry or cancel order

    case domain.EventRefundCreated:
        if event.Refund != nil {
            log.Printf("Refund created: %s\n", event.Refund.ProviderRefundID)
        }
        // Update refund status, notify customer

    default:
        log.Printf("Unhandled webhook event: %s\n", event.EventType)
    }
    return nil
}

// Wire the handler into the client's webhook service
// This is typically done at client initialization
```

### Deduplication and Idempotency

Webhooks are automatically deduplicated using a combination of provider, transaction ID, and event type. The webhook handler:

1. **Stores raw payload** immediately for audit
2. **Detects duplicates** and safely ACKs without re-processing
3. **Verifies signature** using provider's public key
4. **Parses event** into typed struct
5. **Calls handler** exactly once per unique event
6. **Marks processed** in deduplication store

This ensures idempotency: if a provider resends the same webhook, your handler is called exactly once.

## Accessing Provider-Specific Details

While the library provides a consistent canonical API across providers, each provider returns unique fields and metadata that may be important for your application. These provider-specific details are captured in strongly-typed structs and accessible through the `ProviderDetails` field on response objects.

### ProviderDetails Fields

All response types (Order, Payment, Refund, Instrument, PaymentLink) include a `ProviderDetails` field containing provider-specific data:

```go
order, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    AmountMinor: 10000,
    Currency:    "INR",
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
import "github.com/Bytonomics/multipay-adapter/domain"

order, err := client.Orders().GetOrder(ctx, &domain.GetOrderRequest{
    OrderID: "nonexistent",
})
if errors.Is(err, domain.ErrOrderNotFound) {
    log.Println("Order does not exist")
    // Handle missing order
}

refund, err := client.Refunds().CreateRefund(ctx, req)
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
_, err := client.Refunds().CreateRefund(ctx, req)
if ce := &domain.CapabilityError{}; errors.As(err, &ce) {
    log.Printf("Provider %s does not support %s: %s\n",
        ce.Provider, ce.Capability, ce.Message)
    // Fall back to alternative flow
}
```

### ProviderAPIError

Wraps provider-specific errors with context:

```go
_, err := client.Payments().GetPayment(ctx, &domain.GetPaymentRequest{
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
_, err := client.Orders().CreateOrder(ctx, &domain.CreateOrderRequest{
    // Missing required fields (AmountMinor, Currency)
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
    Provider: prodCashfreeAdapter,
    Logger:   yourLogger,
})
if err != nil {
    log.Fatalf("failed to create production client: %v", err) // invalid config or nil adapter
}

// Staging: Razorpay
stagingClient, err := client.NewClient(&client.ClientConfig{
    Provider: stagingRazorpayAdapter,
    Logger:   yourLogger,
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

All MultiPayClient instances are fully thread-safe and can be safely shared across goroutines:

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
prodAdapter, err := providers.NewCashfreeAdapter(&providers.CashfreeConfig{
    ClientID:     "prod-client-id",
    ClientSecret: "prod-secret",
    Environment:  domain.EnvironmentProduction,
    AccountID:    "prod",
})
if err != nil {
    log.Fatalf("failed to create prod adapter: %v", err) // invalid credentials or environment
}

sandboxAdapter, err := providers.NewCashfreeAdapter(&providers.CashfreeConfig{
    ClientID:     "sandbox-client-id",
    ClientSecret: "sandbox-secret",
    Environment:  domain.EnvironmentSandbox,
    AccountID:    "sandbox",
})
if err != nil {
    log.Fatalf("failed to create sandbox adapter: %v", err)
}

prodClient, err := client.NewClient(&client.ClientConfig{
    Provider: prodAdapter,
    Logger:   yourLogger,
})
if err != nil {
    log.Fatalf("failed to create prod client: %v", err)
}

sandboxClient, err := client.NewClient(&client.ClientConfig{
    Provider: sandboxAdapter,
    Logger:   yourLogger,
})
if err != nil {
    log.Fatalf("failed to create sandbox client: %v", err)
}
```

### Multi-Account Webhook Routing

Each provider+account combination gets its own webhook endpoint:

- `POST /webhooks/cashfree/acct_prod` → routed to your production Cashfree adapter
- `POST /webhooks/cashfree/acct_sandbox` → routed to your sandbox Cashfree adapter
- `POST /webhooks/razorpay/acct_main` → routed to your Razorpay adapter

Register each URL separately in the corresponding provider dashboard. The `EndpointRegistry` tracks which provider+account combinations are active and rejects webhooks for unregistered endpoints.

## Hook Pipeline

Hooks allow you to extend the library with custom observability and business logic.

### Built-In Hooks

The library includes built-in hooks for audit logging and metrics:

```go
// Audit hook (enabled by default)
// Logs all operations: method, provider, duration, status

// Metrics hook (enabled by default)
// Records: latency histograms, status counters, error rates
```

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
    Provider: cashfreeAdapter,
    Hooks:    []ports.Hook{customHook},
    Logger:   yourLogger,
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
    RequestData  interface{}     // Original request struct
    ResponseData interface{}     // Response struct (nil in before/error-hook)
    Error        error           // Error (nil if no error)
    StartTime    time.Time       // When operation started
}
```
## Architecture

For detailed architecture information, see [DESIGN.md](./DESIGN.md).

### Key Concepts

- **Canonical Types**: All requests/responses use library-defined types; providers return provider-specific types that are immediately mapped to canonical types
- **Hook Pipeline**: Every operation follows: validate → check capability → execute before-hooks → call adapter → execute after/error-hooks
- **Deduplication**: Webhook payloads are stored and deduplicated to ensure idempotency
- **Capability Matrix**: Check if a provider supports an operation before attempting it
- **Thread-Safe**: All components are safe to use concurrently; no global state

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

See [LICENSE](./LICENSE) file