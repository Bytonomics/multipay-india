package razorpay

import (
	"encoding/json"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

const invoiceFixtureJSON = `{"id":"inv_1","entity":"invoice","payment_id":"pay_1","amount":100,"currency":"INR","status":"paid","order_id":"order_1","paid_at":1481541600,"created_at":1481541534}`

func TestMapInvoiceToSubscriptionPayment(t *testing.T) {
	m := map[string]any{}
	if err := json.Unmarshal([]byte(invoiceFixtureJSON), &m); err != nil {
		t.Fatalf("failed to unmarshal invoice fixture: %v", err)
	}

	typed, err := decodeResponse[razorpayInvoiceResponse](m)
	if err != nil {
		t.Fatalf("failed to decode invoice response: %v", err)
	}

	pmt := mapInvoiceToSubscriptionPayment(typed, "sub_1", m)

	if pmt.PaymentID != "pay_1" {
		t.Fatalf("expected PaymentID='pay_1', got '%s'", pmt.PaymentID)
	}

	if pmt.SubscriptionID != "sub_1" {
		t.Fatalf("expected SubscriptionID='sub_1', got '%s'", pmt.SubscriptionID)
	}

	if int64(pmt.AmountMinor) != 100 {
		t.Fatalf("expected AmountMinor=100 (Razorpay native minor, no conversion), got %d", int64(pmt.AmountMinor))
	}

	if pmt.Status != domain.SubPaymentStatusSuccess {
		t.Fatalf("expected Status=SubPaymentStatusSuccess (from 'paid'), got %v", pmt.Status)
	}
}
