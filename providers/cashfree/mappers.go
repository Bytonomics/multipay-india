package cashfree

import (
	"encoding/json"
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
		ProviderOrderID: strconv.FormatInt(derefInt64(cfOrder.CfOrderId), 10),
		OrderID:         StringPtrToStr(cfOrder.OrderId),
		Status:          mapOrderStatus(cfOrder.OrderStatus),
		AmountMinor:     domain.AmountMinor(AmountCashfreeToMinor(FloatPtrToFloat64(cfOrder.OrderAmount))),
		Currency:        domain.Currency(StringPtrToStr(cfOrder.OrderCurrency)),
		SessionID:       StringPtrToStr(cfOrder.PaymentSessionId),
		ExpiryTime:      cfOrder.OrderExpiryTime,
		CreatedAt:       cfOrder.CreatedAt,
		Customer:        nil,
		Metadata:        domain.Metadata(cfOrder.OrderTags),
		Raw:             rawResponse(cfOrder),
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
		ProviderPaymentID: paymentID,
		OrderID:           StringPtrToStr(cfPayment.OrderId),
		Status:            mapPaymentStatus(cfPayment.PaymentStatus),
		AmountMinor:       domain.AmountMinor(AmountCashfreeToMinor(FloatPtrToFloat64(cfPayment.PaymentAmount))),
		Currency:          domain.Currency(StringPtrToStr(cfPayment.PaymentCurrency)),
		PaymentGroup:      StringPtrToStr(cfPayment.PaymentGroup),
		PaymentMethod:     "", // PaymentMethod is complex; extract as needed
		PaymentTime:       timePtr(TimeFromTimestamp(cfPayment.PaymentTime)),
		CompletionTime:    timePtr(TimeFromTimestamp(cfPayment.PaymentCompletionTime)),
		IsCaptured:        derefBool(cfPayment.IsCaptured),
		BankReference:     StringPtrToStr(cfPayment.BankReference),
		ErrorCode:         "",
		ErrorMessage:      "",
		Raw:               rawResponse(cfPayment),
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

// timePtr converts a time.Time value to a pointer to time.Time.
// Returns nil if the input time is zero.
func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
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
		ProviderRefundID: refundID,
		RefundID:         refundID,
		OrderID:          StringPtrToStr(cfRefund.OrderId),
		PaymentID:        strconv.FormatInt(derefInt64(cfRefund.CfPaymentId), 10),
		Status:           mapRefundStatus(cfRefund.RefundStatus),
		AmountMinor:      domain.AmountMinor(AmountCashfreeToMinor(FloatPtrToFloat64(cfRefund.RefundAmount))),
		Currency:         domain.Currency(StringPtrToStr(cfRefund.RefundCurrency)),
		Reason:           StringPtrToStr(cfRefund.RefundNote),
		ARN:              StringPtrToStr(cfRefund.RefundArn),
		CreatedAt:        timePtr(TimeFromTimestamp(cfRefund.CreatedAt)),
		ProcessedAt:      timePtr(TimeFromTimestamp(cfRefund.ProcessedAt)),
		Raw:              rawResponse(cfRefund),
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
		CustomerID:     StringPtrToStr(cfInstrument.CustomerId),
		InstrumentID:   instrumentID,
		InstrumentType: instrumentType,
		DisplayValue:   StringPtrToStr(cfInstrument.InstrumentDisplay),
		Status:         StringPtrToStr(cfInstrument.InstrumentStatus),
		CreatedAt:      timePtr(TimeFromTimestamp(cfInstrument.CreatedAt)),
		Raw:            rawResponse(cfInstrument),
	}

	return instrument
}

// MapLinkEntityToCanonical maps a Cashfree LinkEntity (payment link) to the canonical domain.PaymentLink type.
func MapLinkEntityToCanonical(cfLink *cf.LinkEntity) *domain.PaymentLink {
	if cfLink == nil {
		return nil
	}

	// Parse LinkStatus to PaymentLinkStatus
	var linkStatus domain.PaymentLinkStatus
	if cfLink.LinkStatus != nil {
		linkStatus = domain.PaymentLinkStatus(*cfLink.LinkStatus)
	}

	link := &domain.PaymentLink{
		ProviderLinkID: strconv.FormatInt(derefInt64(cfLink.CfLinkId), 10),
		LinkID:         StringPtrToStr(cfLink.LinkId),
		Status:         linkStatus,
		AmountMinor:    domain.AmountMinor(AmountCashfreeToMinor(FloatPtrToFloat64(cfLink.LinkAmount))),
		AmountPaid:     domain.AmountMinor(AmountCashfreeToMinor(FloatPtrToFloat64(cfLink.LinkAmountPaid))),
		Currency:       domain.Currency(StringPtrToStr(cfLink.LinkCurrency)),
		Purpose:        StringPtrToStr(cfLink.LinkPurpose),
		LinkURL:        StringPtrToStr(cfLink.LinkUrl),
		Customer:       nil, // Customer info not directly available from LinkEntity
		CreatedAt:      timePtr(TimeFromTimestamp(cfLink.LinkCreatedAt)),
		ExpiryTime:     timePtr(TimeFromTimestamp(cfLink.LinkExpiryTime)),
		Metadata:       nil,
		Raw:            rawResponse(cfLink),
	}

	return link
}

// derefInt64 safely dereferences an int64 pointer or returns 0.
func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// derefBool safely dereferences a bool pointer or returns false.
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// rawResponse marshals a Cashfree SDK entity to JSON and returns it as RawProviderResponse.
func rawResponse(v interface{}) domain.RawProviderResponse {
	data, err := json.Marshal(v)
	if err != nil {
		// If marshaling fails, return empty response
		return domain.RawProviderResponse(nil)
	}
	return domain.RawProviderResponse(data)
}
