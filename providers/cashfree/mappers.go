package cashfree

import (
	"math"
	"strconv"
	"time"

	cf "github.com/cashfree/cashfree_pg"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// AmountMinorToCashfree converts an amount from paisa (int64) to rupees (float64).
// Paisa is the minor currency unit in India; 100 paisa = 1 rupee.
// Example: 50000 paisa → 500.00 rupees
func AmountMinorToCashfree(minorAmount int64) float64 {
	return float64(minorAmount) / 100.0
}

// AmountCashfreeToMinor converts an amount from rupees (float64) to paisa (int64).
// Uses math.Round for proper rounding to the nearest paisa.
// Example: 500.50 rupees → 50050 paisa
func AmountCashfreeToMinor(rupeesAmount float64) int64 {
	return int64(math.Round(rupeesAmount * 100.0))
}

// MapOrderEntityToCanonical maps a Cashfree OrderEntity to the canonical domain.Order type.
// Handles type conversions and status mapping.
func MapOrderEntityToCanonical(cfOrder *cf.OrderEntity) *domain.Order {
	if cfOrder == nil {
		return nil
	}

	order := &domain.Order{
		ID:        StringPtrToStr(cfOrder.OrderId),
		Amount:    AmountCashfreeToMinor(FloatPtrToFloat64(cfOrder.OrderAmount)),
		Currency:  domain.Currency(StringPtrToStr(cfOrder.OrderCurrency)),
		Receipt:   StringPtrToStr(cfOrder.OrderNote.Get()),
		Notes:     StringPtrToStr(cfOrder.OrderNote.Get()),
		Status:    mapOrderStatus(cfOrder.OrderStatus),
		CreatedAt: time.Now(), // Cashfree SDK may not provide creation timestamp in order response
	}

	return order
}

// MapPaymentEntityToCanonical maps a Cashfree PaymentEntity to the canonical domain.Payment type.
func MapPaymentEntityToCanonical(cfPayment *cf.PaymentEntity) *domain.Payment {
	if cfPayment == nil {
		return nil
	}

	paymentID := ""
	if cfPayment.CfPaymentId != nil {
		paymentID = strconv.FormatInt(*cfPayment.CfPaymentId, 10)
	}

	payment := &domain.Payment{
		ID:        paymentID,
		OrderID:   StringPtrToStr(cfPayment.OrderId),
		Amount:    AmountCashfreeToMinor(FloatPtrToFloat64(cfPayment.PaymentAmount)),
		Status:    mapPaymentStatus(cfPayment.PaymentStatus),
		Method:    "", // PaymentMethod is a complex type in Cashfree SDK, extract as needed
		CreatedAt: TimeFromTimestamp(cfPayment.PaymentTime),
	}

	return payment
}

// mapOrderStatus converts Cashfree order status strings to canonical domain.OrderStatus.
func mapOrderStatus(statusPtr *string) domain.OrderStatus {
	status := ""
	if statusPtr != nil {
		status = *statusPtr
	}
	switch status {
	case "PAID":
		return domain.OrderPaid
	case "EXPIRED":
		return domain.OrderExpired
	case "CANCELLED":
		return domain.OrderCancelled
	case "ACTIVE", "PENDING":
		return domain.OrderCreated
	default:
		return domain.OrderCreated
	}
}

// mapPaymentStatus converts Cashfree payment status strings to canonical domain.PaymentStatus.
func mapPaymentStatus(statusPtr *string) domain.PaymentStatus {
	status := ""
	if statusPtr != nil {
		status = *statusPtr
	}
	switch status {
	case "AUTHORIZED":
		return domain.PaymentAuthorized
	case "CAPTURED":
		return domain.PaymentCaptured
	case "FAILED":
		return domain.PaymentFailed
	case "REFUNDED":
		return domain.PaymentRefunded
	default:
		return domain.PaymentFailed
	}
}

// StringPtrToStr safely dereferences a string pointer or returns empty string.
func StringPtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// FloatPtrToFloat64 safely dereferences a float32 pointer or returns 0.0.
func FloatPtrToFloat64(f *float32) float64 {
	if f == nil {
		return 0.0
	}
	return float64(*f)
}

// TimeFromTimestamp converts a Cashfree timestamp string to a Time value.
// Cashfree typically uses RFC3339 or ISO8601 formatted timestamps.
// For now, returns zero time if parsing fails (caller should handle).
func TimeFromTimestamp(ts *string) time.Time {
	if ts == nil || *ts == "" {
		return time.Time{}
	}
	// Attempt to parse RFC3339 format (Cashfree standard)
	t, err := time.Parse(time.RFC3339, *ts)
	if err != nil {
		// Fallback: return zero time on parse failure
		return time.Time{}
	}
	return t
}

// MapRefundEntityToCanonical maps a Cashfree RefundEntity to the canonical domain.Refund type.
func MapRefundEntityToCanonical(cfRefund *cf.RefundEntity) *domain.Refund {
	if cfRefund == nil {
		return nil
	}

	refundID := ""
	if cfRefund.RefundId != nil {
		refundID = *cfRefund.RefundId
	}

	refund := &domain.Refund{
		ID:        refundID,
		OrderID:   StringPtrToStr(cfRefund.OrderId),
		Amount:    AmountCashfreeToMinor(FloatPtrToFloat64(cfRefund.RefundAmount)),
		Status:    mapRefundStatus(cfRefund.RefundStatus),
		CreatedAt: TimeFromTimestamp(cfRefund.CreatedAt),
	}

	return refund
}

// mapRefundStatus converts Cashfree refund status strings to canonical domain.RefundStatus.
func mapRefundStatus(statusPtr *string) domain.RefundStatus {
	status := ""
	if statusPtr != nil {
		status = *statusPtr
	}
	switch status {
	case "PENDING":
		return domain.RefundCreated
	case "PROCESSED":
		return domain.RefundProcessed
	case "CANCELLED":
		return domain.RefundFailed
	case "FAILED":
		return domain.RefundFailed
	default:
		return domain.RefundCreated
	}
}

// MapInstrumentEntityToCanonical maps a Cashfree InstrumentEntity to the canonical domain.Instrument type.
func MapInstrumentEntityToCanonical(cfInstrument *cf.InstrumentEntity) *domain.Instrument {
	if cfInstrument == nil {
		return nil
	}

	instrumentID := ""
	if cfInstrument.InstrumentId != nil {
		instrumentID = *cfInstrument.InstrumentId
	}

	instrumentType := ""
	if cfInstrument.InstrumentType != nil {
		instrumentType = *cfInstrument.InstrumentType
	}

	instrument := &domain.Instrument{
		ID:         instrumentID,
		CustomerID: StringPtrToStr(cfInstrument.CustomerId),
		Type:       instrumentType,
		CreatedAt:  TimeFromTimestamp(cfInstrument.CreatedAt),
	}

	return instrument
}

// MapLinkEntityToCanonical maps a Cashfree LinkEntity (payment link) to the canonical domain.PaymentLinkResponse type.
func MapLinkEntityToCanonical(cfLink *cf.LinkEntity) *domain.PaymentLinkResponse {
	if cfLink == nil {
		return nil
	}

	// Parse ExpiresAt timestamp to *time.Time
	var expiresAt *time.Time
	expiresAtTime := TimeFromTimestamp(cfLink.LinkExpiryTime)
	if !expiresAtTime.IsZero() {
		expiresAt = &expiresAtTime
	}

	// Parse CreatedAt timestamp to time.Time
	createdAt := TimeFromTimestamp(cfLink.LinkCreatedAt)

	link := &domain.PaymentLinkResponse{
		ID:          StringPtrToStr(cfLink.LinkId),
		URL:         StringPtrToStr(cfLink.LinkUrl),
		Amount:      AmountCashfreeToMinor(FloatPtrToFloat64(cfLink.LinkAmount)),
		Currency:    StringPtrToStr(cfLink.LinkCurrency),
		Description: StringPtrToStr(cfLink.LinkPurpose),
		Status:      StringPtrToStr(cfLink.LinkStatus),
		NotifyEmail: "", // Cashfree LinkNotifyEntity only has boolean flags, not actual email values
		NotifyPhone: "", // Cashfree LinkNotifyEntity only has boolean flags, not actual phone values
		ExpiresAt:   expiresAt,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt, // Cashfree doesn't provide UpdatedAt, use CreatedAt as fallback
	}

	return link
}
