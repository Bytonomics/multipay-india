# MultiPay Adapter

A unified Go library for integrating multiple payment providers (Cashfree, Razorpay) with a single, consistent API. Process orders, payments, refunds, and webhooks across any supported provider without provider-specific code.

## Overview

MultiPay Adapter provides a abstraction layer over payment processing providers, allowing you to:

- **Use a single API** for all payment operations across providers
- **Switch providers** without changing application code
- **Support multiple providers** simultaneously
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

## Quick Start

### 1. Create Adapters

```go
// Create Cashfree adapter
cashfreeAdapter := adapters.NewCashfreeAdapter(&adapters.CashfreeConfig{
    ClientID:    "your-cashfree-client-id",
    ClientSecret: "your-cashfree-secret",
    Environment: "PROD", // or "SANDBOX"
})

// Create Razorpay adapter
razorpayAdapter := adapters.NewRazorpayAdapter(&adapters.RazorpayConfig{
    Key:    "your-razorpay-key",
    Secret: "your-razorpay-secret",
})
```

### 2. Create Client

```go
client, err := multipay.NewClient(&multipay.ClientConfig{
    Adapters: map[string]multipay.Adapter{
        "cashfree": cashfreeAdapter,
        "razorpay": razorpayAdapter,
    },
    DefaultProvider: "cashfree",
})
if err != nil {
    log.Fatal(err)
}
```

### 3. Make an Order

```go
ctx := context.Background()

order, err := client.Orders().Create(ctx, &multipay.Order{
    ID:       "order_123",
    Amount:   10000, // Amount in smallest unit (e.g., paise)
    Currency: "INR",
    Customer: &multipay.Customer{
        ID:    "cust_123",
        Email: "user@example.com",
        Phone: "+919876543210",
    },
    Notes: map[string]string{
        "order_ref": "OD-2024-001",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order created: %s (Status: %s)\n", order.ID, order.Status)
```

### 4. Fetch and List Orders

```go
// Fetch a single order
order, err := client.Orders().Fetch(ctx, "order_123")
if err != nil {
    log.Fatal(err)
}

// List all payments for an order
payments, err := client.Orders().ListPayments(ctx, "order_123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Payments: %d\n", len(payments))
```

## Provider Configuration

### Cashfree Configuration

```go
cashfreeAdapter := adapters.NewCashfreeAdapter(&adapters.CashfreeConfig{
    ClientID:    os.Getenv("CASHFREE_CLIENT_ID"),
    ClientSecret: os.Getenv("CASHFREE_CLIENT_SECRET"),
    Environment: "PROD",  // "PROD" or "SANDBOX"
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
razorpayAdapter := adapters.NewRazorpayAdapter(&adapters.RazorpayConfig{
    Key:    os.Getenv("RAZORPAY_KEY"),
    Secret: os.Getenv("RAZORPAY_SECRET"),
})
```

**Environment Variables** (recommended):

```bash
export RAZORPAY_KEY="your-razorpay-key"
export RAZORPAY_SECRET="your-razorpay-secret"
```

## Core Services

### Orders

Create and manage payment orders.

```go
// Create an order
order, err := client.Orders().Create(ctx, &multipay.Order{
    ID:       "order_abc123",
    Amount:   5000,    // 50.00 INR
    Currency: "INR",
    Customer: &multipay.Customer{
        ID:    "cust_123",
        Email: "user@example.com",
    },
})

// Fetch an order
order, err := client.Orders().Fetch(ctx, "order_abc123")

// List all payments for an order
payments, err := client.Orders().ListPayments(ctx, "order_abc123")
```

### Payments

Process and retrieve payment details.

```go
// Fetch a payment
payment, err := client.Payments().Fetch(ctx, "pay_abc123")

// List payments
payments, err := client.Payments().List(ctx, &multipay.PaymentFilter{
    Status:    "captured",
    Limit:     10,
    Offset:    0,
})

// Capture a payment (for authorized payments)
payment, err := client.Payments().Capture(ctx, "pay_abc123", 5000)
```

### Refunds

Process and manage refunds.

```go
// Create a refund
refund, err := client.Refunds().Create(ctx, &multipay.RefundRequest{
    PaymentID: "pay_abc123",
    Amount:    2500,  // Partial refund
    Notes: map[string]string{
        "reason": "Customer requested",
    },
})

// Fetch a refund
refund, err := client.Refunds().Fetch(ctx, "refund_abc123")

// List refunds
refunds, err := client.Refunds().List(ctx, &multipay.RefundFilter{
    Status: "processed",
})
```

### Instruments

Manage payment instruments (cards, wallets, etc.).

```go
// Fetch an instrument
instrument, err := client.Instruments().Fetch(ctx, "inst_abc123")

// List instruments
instruments, err := client.Instruments().List(ctx, &multipay.InstrumentFilter{
    Type: "card",
})

// Delete an instrument
err := client.Instruments().Delete(ctx, "inst_abc123")
```

### Payment Links

Create shareable payment links.

```go
// Create a payment link
link, err := client.PaymentLinks().Create(ctx, &multipay.PaymentLinkRequest{
    Amount:   10000,
    Currency: "INR",
    Customer: &multipay.Customer{
        Email: "user@example.com",
    },
    ExpiresAt: time.Now().Add(24 * time.Hour),
})

// Fetch a payment link
link, err := client.PaymentLinks().Fetch(ctx, "link_abc123")

// Cancel a payment link
link, err := client.PaymentLinks().Cancel(ctx, "link_abc123")
```

### Webhooks

Handle incoming webhooks from payment providers.

```go
// See "Webhook Setup" section below
```

## Capability Checking

Before attempting an operation, check if the provider supports it:

```go
// Check if provider supports refunds
caps := client.Capabilities()
if !caps.Supports("cashfree", "refund") {
    return fmt.Errorf("Cashfree does not support refunds")
}

// Attempt operation
refund, err := client.Refunds().Create(ctx, req)
if ce := &multipay.CapabilityError{}; errors.As(err, &ce) {
    // Handle unsupported capability gracefully
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
// Handle payment completion
client.Webhooks().OnPaymentCreated(func(ctx context.Context, payment *multipay.Payment) error {
    log.Printf("Payment created: %s (Amount: %d %s)\n", payment.ID, payment.Amount, payment.Currency)
    // Fulfill order, update database, send email, etc.
    return nil
})

// Handle refund completion
client.Webhooks().OnRefundCreated(func(ctx context.Context, refund *multipay.Refund) error {
    log.Printf("Refund processed: %s\n", refund.ID)
    // Update order status, notify customer, etc.
    return nil
})

// Handle payment failure
client.Webhooks().OnPaymentFailed(func(ctx context.Context, payment *multipay.Payment) error {
    log.Printf("Payment failed: %s (%s)\n", payment.ID, payment.FailureReason)
    return nil
})
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

## Error Handling

### Sentinel Errors

Common errors are represented as sentinel errors that you can check with `errors.Is()`:

```go
import "github.com/Bytonomics/multipay-adapter/multipay"

order, err := client.Orders().Fetch(ctx, "nonexistent")
if errors.Is(err, multipay.ErrOrderNotFound) {
    log.Println("Order does not exist")
    // Handle missing order
}

refund, err := client.Refunds().Create(ctx, req)
if errors.Is(err, multipay.ErrRefundAlreadyProcessed) {
    log.Println("Refund was already processed")
    // Skip duplicate refund
}
```

**Common Sentinel Errors:**

- `ErrOrderNotFound` — Order does not exist
- `ErrPaymentNotFound` — Payment does not exist
- `ErrRefundNotFound` — Refund does not exist
- `ErrPaymentLinkNotFound` — Payment link does not exist
- `ErrInstrumentNotFound` — Instrument (card, wallet) does not exist
- `ErrRefundAlreadyProcessed` — Refund cannot be re-created
- `ErrAmountExceedsRemaining` — Refund amount exceeds remaining balance
- `ErrInvalidRequest` — Request validation failed

### CapabilityError

Returned when a provider doesn't support an operation:

```go
_, err := client.CreateRefund(ctx, req)
if ce := &multipay.CapabilityError{}; errors.As(err, &ce) {
    log.Printf("Provider %s does not support %s: %s\n",
        ce.Provider, ce.Capability, ce.Reason)
    // Fall back to alternative flow
}
```

### ProviderAPIError

Wraps provider-specific errors with context:

```go
_, err := client.Payments().Fetch(ctx, "pay_123")
if pe := &multipay.ProviderAPIError{}; errors.As(err, &pe) {
    log.Printf("Provider %s error (code=%s): %s\n",
        pe.Provider, pe.Code, pe.Message)
    // Handle provider-specific error
}
```

### Validation Errors

Request validation errors from pedantigo:

```go
_, err := client.Orders().Create(ctx, &multipay.Order{
    // Missing required fields
})
if ve := &pedantigo.ValidationError{}; errors.As(err, &ve) {
    for field, err := range ve.Errors {
        log.Printf("Field %s: %s\n", field, err)
    }
}
```

## Multi-Instance Support

Create separate clients for different regions or providers:

```go
// Production: Cashfree
prodClient, _ := multipay.NewClient(&multipay.ClientConfig{
    Adapters: map[string]multipay.Adapter{
        "cashfree": prodCashfreeAdapter,
    },
    DefaultProvider: "cashfree",
})

// Staging: Razorpay
stagingClient, _ := multipay.NewClient(&multipay.ClientConfig{
    Adapters: map[string]multipay.Adapter{
        "razorpay": stagingRazorpayAdapter,
    },
    DefaultProvider: "razorpay",
})

// Route based on environment
var client *multipay.Client
if os.Getenv("ENV") == "prod" {
    client = prodClient
} else {
    client = stagingClient
}
```

### Thread Safety

All MultiPayClient instances are fully thread-safe and can be safely shared across goroutines:

```go
// Safe: Share client across goroutines
go func() {
    order, _ := client.Orders().Fetch(ctx, "order_1")
    // ...
}()

go func() {
    order, _ := client.Orders().Fetch(ctx, "order_2")
    // ...
}()
```

### Multiple Cashfree Accounts

If you have multiple Cashfree accounts, create separate adapters and clients:

```go
cashfreeAccount1 := adapters.NewCashfreeAdapter(&adapters.CashfreeConfig{
    ClientID:     "account1-client-id",
    ClientSecret: "account1-secret",
})

cashfreeAccount2 := adapters.NewCashfreeAdapter(&adapters.CashfreeConfig{
    ClientID:     "account2-client-id",
    ClientSecret: "account2-secret",
})

client1, _ := multipay.NewClient(&multipay.ClientConfig{
    Adapters: map[string]multipay.Adapter{
        "cashfree": cashfreeAccount1,
    },
})

client2, _ := multipay.NewClient(&multipay.ClientConfig{
    Adapters: map[string]multipay.Adapter{
        "cashfree": cashfreeAccount2,
    },
})
```

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

### Custom Hooks

Add custom hooks for observability:

```go
// Register before-hook
client.Hooks().RegisterBefore(func(ctx context.Context, hookCtx *multipay.HookContext) error {
    log.Printf("Starting: %s on %s\n", hookCtx.Method, hookCtx.Provider)
    hookCtx.StartTime = time.Now()
    return nil
})

// Register after-hook
client.Hooks().RegisterAfter(func(ctx context.Context, hookCtx *multipay.HookContext) error {
    duration := time.Since(hookCtx.StartTime)
    log.Printf("Completed: %s on %s (duration=%dms, status=%s)\n",
        hookCtx.Method, hookCtx.Provider, duration.Milliseconds(), hookCtx.Status)
    return nil
})

// Register error-hook
client.Hooks().RegisterOnError(func(ctx context.Context, hookCtx *multipay.HookContext, err error) error {
    log.Printf("Failed: %s on %s (error=%v)\n",
        hookCtx.Method, hookCtx.Provider, err)
    // Send alert, record metrics, etc.
    return nil
})
```

### Hook Context

Hooks receive a HookContext with details about the operation:

```go
type HookContext struct {
    Method      string                 // "Orders.Create", "Payments.Fetch", etc.
    Provider    string                 // "cashfree" or "razorpay"
    Request     interface{}            // Original request
    Response    interface{}            // Response (nil in before-hook)
    Status      string                 // "started", "success", "error"
    Error       error                  // Error (nil if no error)
    StartTime   time.Time              // When operation started
    UserContext map[string]interface{} // Custom context from caller
}
```

### Example: OpenTelemetry Integration

```go
import "go.opentelemetry.io/otel"

client.Hooks().RegisterBefore(func(ctx context.Context, hookCtx *multipay.HookContext) error {
    tracer := otel.Tracer("multipay")
    ctx, hookCtx.Span = tracer.Start(ctx, hookCtx.Method)
    return nil
})

client.Hooks().RegisterAfter(func(ctx context.Context, hookCtx *multipay.HookContext) error {
    if hookCtx.Span != nil {
        hookCtx.Span.End()
    }
    return nil
})
```

## Architecture

For detailed architecture information, see [DESIGN.md](./DESIGN.md).

### Key Concepts

- **Canonical Types**: All requests/responses use library-defined types; providers return provider-specific types that are immediately mapped to canonical types
- **Hook Pipeline**: Every operation follows: validate → check capability → resolve provider → execute before-hooks → call adapter → execute after/error-hooks
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