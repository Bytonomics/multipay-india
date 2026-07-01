package fake

import (
	"context"
	"testing"
	"time"

	domain "github.com/Bytonomics/multipay-india/multipay-go/domain"
)

func TestHarness_CreateOrder_ReachesFakeAdapter(t *testing.T) {
	h := NewHarness(&Config{})
	h.Adapter.CreateOrderFunc = func(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
		return &domain.Order{OrderID: req.OrderID, ProviderOrderID: "cf_1", Status: domain.OrderCreated, SessionID: "sess_1", Checkout: &domain.CheckoutPayload{Provider: domain.ProviderCashfree, Environment: domain.EnvironmentProduction, SessionID: "sess_1"}}, nil
	}
	order, err := h.Client.Orders().CreateOrder(context.Background(), &domain.CreateOrderRequest{OrderID: "order_1", AmountMinor: 50000, Currency: "INR", Customer: &domain.CustomerInfo{CustomerID: "org_1", Phone: "+919876543210"}, ReturnURL: "https://app.example.com/return", Note: "test"})
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}
	if order.OrderID != "order_1" {
		t.Errorf("expected OrderID order_1, got %q", order.OrderID)
	}
	if order.Status != domain.OrderCreated {
		t.Errorf("expected status %q, got %q", domain.OrderCreated, order.Status)
	}
	if order.Checkout == nil {
		t.Errorf("expected non-nil Checkout")
	}
	if len(h.Adapter.CreateOrderCalls) != 1 {
		t.Errorf("expected 1 CreateOrder call, got %d", len(h.Adapter.CreateOrderCalls))
	}
}

func TestHarness_CreatePaymentLink_ReachesFakeAdapter(t *testing.T) {
	h := NewHarness(&Config{})
	h.Adapter.CreatePaymentLinkFunc = func(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
		return &domain.PaymentLink{LinkID: "pl_1", ProviderLinkID: "cf_pl_1", Status: domain.PaymentLinkStatusActive, LinkURL: "https://cashfree.com/link/pl_1"}, nil
	}
	link, err := h.Client.PaymentLinks().CreatePaymentLink(context.Background(), &domain.CreatePaymentLinkRequest{AmountMinor: 75000, Currency: "INR", Purpose: "plan purchase", Customer: &domain.CustomerInfo{CustomerID: "org_1", Name: "A", Email: "a@example.com", Phone: "+919876543210"}})
	if err != nil {
		t.Fatalf("CreatePaymentLink returned error: %v", err)
	}
	if link.LinkURL == "" {
		t.Errorf("expected non-empty LinkURL")
	}
	if link.Status != domain.PaymentLinkStatusActive {
		t.Errorf("expected status %q, got %q", domain.PaymentLinkStatusActive, link.Status)
	}
}

func TestHarness_CreateSubscription_ReachesFakeAdapter(t *testing.T) {
	h := NewHarness(&Config{})
	h.Adapter.CreateSubscriptionFunc = func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
		return &domain.Subscription{SubscriptionID: req.SubscriptionID, ProviderSubscriptionID: "cf_sub_1", Status: domain.SubscriptionStatusInitialized, AuthLink: "https://auth.link"}, nil
	}
	sub, err := h.Client.Subscriptions().CreateSubscription(context.Background(), &domain.CreateSubscriptionRequest{SubscriptionID: "sub_1", PlanID: "plan_1", CustomerEmail: "a@example.com", CustomerPhone: "+919876543210", ReturnURL: "https://app.example.com/return"})
	if err != nil {
		t.Fatalf("CreateSubscription returned error: %v", err)
	}
	if sub.AuthLink != "https://auth.link" {
		t.Errorf("expected AuthLink https://auth.link, got %q", sub.AuthLink)
	}
	if len(h.Adapter.CreateSubscriptionCalls) != 1 {
		t.Errorf("expected 1 CreateSubscription call, got %d", len(h.Adapter.CreateSubscriptionCalls))
	}
}

func TestHarness_EmitWebhookAfter_DeliversToRegisteredHandler(t *testing.T) {
	h := NewHarness(&Config{WebhookDelays: map[domain.WebhookEventType]time.Duration{domain.EventSubCharged: 10 * time.Millisecond}})
	var got *domain.WebhookEvent
	handled := false
	h.Client.Webhooks().RegisterHandler(domain.EventSubCharged, func(ctx context.Context, event *domain.WebhookEvent) error {
		got = event
		handled = true
		return nil
	})
	h.EmitWebhookAfter(&WebhookEnvelope{EventType: domain.EventSubCharged, DedupeKey: "dk_1", RawVendorStatus: "ACTIVE", Subscription: &domain.Subscription{SubscriptionID: "sub_1"}})
	h.WaitForWebhooks()
	if errs := h.WebhookErrors(); len(errs) != 0 {
		t.Fatalf("expected no webhook delivery errors, got %v", errs)
	}
	if !handled {
		t.Fatalf("expected handler to be invoked")
	}
	if got == nil {
		t.Fatalf("expected non-nil event")
	}
	if got.EventType != domain.EventSubCharged {
		t.Errorf("expected event type %q, got %q", domain.EventSubCharged, got.EventType)
	}
	if got.DedupeKey != "dk_1" {
		t.Errorf("expected dedupe key dk_1, got %q", got.DedupeKey)
	}
	if got.RawVendorStatus != "ACTIVE" {
		t.Errorf("expected raw vendor status ACTIVE, got %q", got.RawVendorStatus)
	}
	if got.Subscription == nil || got.Subscription.SubscriptionID != "sub_1" {
		t.Errorf("expected subscription sub_1, got %+v", got.Subscription)
	}
}

func TestHarness_Webhook_DedupeSuppressesSecondDelivery(t *testing.T) {
	h := NewHarness(&Config{DefaultWebhookDelay: 0})
	count := 0
	h.Client.Webhooks().RegisterHandler(domain.EventSubCharged, func(ctx context.Context, event *domain.WebhookEvent) error {
		count++
		return nil
	})
	env := &WebhookEnvelope{EventType: domain.EventSubCharged, DedupeKey: "dk_dup", Subscription: &domain.Subscription{SubscriptionID: "sub_1"}}
	h.EmitWebhookAfter(env)
	h.WaitForWebhooks()
	env2 := &WebhookEnvelope{EventType: domain.EventSubCharged, DedupeKey: "dk_dup", Subscription: &domain.Subscription{SubscriptionID: "sub_1"}}
	h.EmitWebhookAfter(env2)
	h.WaitForWebhooks()
	if errs := h.WebhookErrors(); len(errs) != 0 {
		t.Fatalf("expected no webhook delivery errors, got %v", errs)
	}
	if count != 1 {
		t.Errorf("expected handler invoked once (second is deduped), got %d", count)
	}
}
