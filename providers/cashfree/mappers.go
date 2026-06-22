package cashfree

import (
	"encoding/json"
	"math"
	"time"

	"github.com/bojanz/currency"
	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// AmountMinorToMajor converts minor units (paisa/cents/fils) to major units (rupees/dollars/dinars)
// using the ISO 4217 minor unit exponent for the given currency.
// Examples:
//
//	AmountMinorToMajor(50000, "INR") → 500.0  (exponent 2, factor 100)
//	AmountMinorToMajor(500, "JPY")   → 500.0  (exponent 0, factor 1)
//	AmountMinorToMajor(500000, "BHD") → 500.0  (exponent 3, factor 1000)
func AmountMinorToMajor(minorAmount int64, currencyCode string) float64 {
	digits, ok := currency.GetDigits(currencyCode)
	if !ok {
		// Unknown currency — fall back to exponent 2 (most common)
		digits = 2
	}
	if digits == 0 {
		return float64(minorAmount)
	}
	factor := math.Pow(10, float64(digits))
	return float64(minorAmount) / factor
}

// AmountMajorToMinor converts major units (rupees/dollars/dinars) to minor units (paisa/cents/fils)
// using the ISO 4217 minor unit exponent for the given currency.
// Examples:
//
//	AmountMajorToMinor(500.0, "INR")  → 50000  (exponent 2, factor 100)
//	AmountMajorToMinor(500.0, "JPY")  → 500    (exponent 0, factor 1)
//	AmountMajorToMinor(500.0, "BHD")  → 500000 (exponent 3, factor 1000)
func AmountMajorToMinor(majorAmount float64, currencyCode string) int64 {
	digits, ok := currency.GetDigits(currencyCode)
	if !ok {
		// Unknown currency — fall back to exponent 2 (most common)
		digits = 2
	}
	if digits == 0 {
		return int64(math.Round(majorAmount))
	}
	factor := math.Pow(10, float64(digits))
	return int64(math.Round(majorAmount * factor))
}

// MapOrderEntityToCanonical maps a Cashfree OrderEntity to the canonical domain.Order type.
// Handles type conversions and status mapping.
func MapOrderEntityToCanonical(cfOrder *cf.OrderEntity) *domain.Order {
	if cfOrder == nil {
		return nil
	}

	metadata := domain.Metadata(nil)
	if cfOrder.OrderTags != nil {
		metadata = domain.Metadata(*cfOrder.OrderTags)
	}

	order := &domain.Order{
		ProviderOrderID: StringPtrToStr(cfOrder.CfOrderId),
		OrderID:         StringPtrToStr(cfOrder.OrderId),
		Status:          mapOrderStatus(cfOrder.OrderStatus),
		AmountMinor:     domain.AmountMinor(AmountMajorToMinor(FloatPtrToFloat64(cfOrder.OrderAmount), StringPtrToStr(cfOrder.OrderCurrency))),
		Currency:        domain.Currency(StringPtrToStr(cfOrder.OrderCurrency)),
		SessionID:       StringPtrToStr(cfOrder.PaymentSessionId),
		ExpiryTime:      timePtr(TimeFromTimestamp(cfOrder.OrderExpiryTime)),
		CreatedAt:       timePtr(TimeFromTimestamp(cfOrder.CreatedAt)),
		Customer:        nil,
		Metadata:        metadata,
		Raw:             rawResponse(cfOrder),
	}

	return order
}

// MapPaymentEntityToCanonical maps a Cashfree PaymentEntity to the canonical domain.Payment type.
func MapPaymentEntityToCanonical(cfPayment *cf.PaymentEntity) *domain.Payment {
	if cfPayment == nil {
		return nil
	}

	payment := &domain.Payment{
		ProviderPaymentID: StringPtrToStr(cfPayment.CfPaymentId),
		OrderID:           StringPtrToStr(cfPayment.OrderId),
		Status:            mapPaymentStatus(cfPayment.PaymentStatus),
		AmountMinor:       domain.AmountMinor(AmountMajorToMinor(FloatPtrToFloat64(cfPayment.PaymentAmount), StringPtrToStr(cfPayment.PaymentCurrency))),
		Currency:          domain.Currency(StringPtrToStr(cfPayment.PaymentCurrency)),
		PaymentGroup:      StringPtrToStr(cfPayment.PaymentGroup),
		PaymentMethod:     "",
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
		ProviderRefundID: StringPtrToStr(cfRefund.CfRefundId),
		RefundID:         refundID,
		OrderID:          StringPtrToStr(cfRefund.OrderId),
		PaymentID:        StringPtrToStr(cfRefund.CfPaymentId),
		Status:           mapRefundStatus(cfRefund.RefundStatus),
		AmountMinor:      domain.AmountMinor(AmountMajorToMinor(FloatPtrToFloat64(cfRefund.RefundAmount), StringPtrToStr(cfRefund.RefundCurrency))),
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

// MapInstrumentEntityForAllSavedCardToCanonical maps a Cashfree InstrumentEntityForAllSavedCard
// (returned from list operations) to the canonical domain.Instrument type.
func MapInstrumentEntityForAllSavedCardToCanonical(cfInstrument *cf.InstrumentEntityForAllSavedCard) *domain.Instrument {
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

	metadata := domain.Metadata(nil)
	if cfLink.LinkNotes != nil {
		metadata = domain.Metadata(*cfLink.LinkNotes)
	}

	link := &domain.PaymentLink{
		ProviderLinkID: StringPtrToStr(cfLink.CfLinkId),
		LinkID:         StringPtrToStr(cfLink.LinkId),
		Status:         linkStatus,
		AmountMinor:    domain.AmountMinor(AmountMajorToMinor(FloatPtrToFloat64(cfLink.LinkAmount), StringPtrToStr(cfLink.LinkCurrency))),
		AmountPaid:     domain.AmountMinor(AmountMajorToMinor(FloatPtrToFloat64(cfLink.LinkAmountPaid), StringPtrToStr(cfLink.LinkCurrency))),
		Currency:       domain.Currency(StringPtrToStr(cfLink.LinkCurrency)),
		Purpose:        StringPtrToStr(cfLink.LinkPurpose),
		LinkURL:        StringPtrToStr(cfLink.LinkUrl),
		Customer:       nil, // Customer info not directly available from LinkEntity
		CreatedAt:      timePtr(TimeFromTimestamp(cfLink.LinkCreatedAt)),
		ExpiryTime:     timePtr(TimeFromTimestamp(cfLink.LinkExpiryTime)),
		Metadata:       metadata,
		Raw:            rawResponse(cfLink),
	}

	return link
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
