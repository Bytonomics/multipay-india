package cashfree

import (
	"encoding/json"
	"fmt"
	"time"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

// MapOrderEntityToCanonical maps a Cashfree OrderEntity to the canonical domain.Order type.
// Handles type conversions and status mapping.
func MapOrderEntityToCanonical(cfOrder *cf.OrderEntity) (*domain.Order, error) {
	if cfOrder == nil {
		return nil, fmt.Errorf("order entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(cfOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order: %w", err)
	}

	metadata := domain.Metadata(nil)
	if cfOrder.OrderTags != nil {
		metadata = domain.Metadata(*cfOrder.OrderTags)
	}

	order := &domain.Order{
		ProviderOrderID: StringPtrToStr(cfOrder.CfOrderId),
		OrderID:         StringPtrToStr(cfOrder.OrderId),
		Status:          mapOrderStatus(cfOrder.OrderStatus),
		AmountMinor:     domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(cfOrder.OrderAmount), StringPtrToStr(cfOrder.OrderCurrency))),
		Currency:        domain.Currency(StringPtrToStr(cfOrder.OrderCurrency)),
		SessionID:       StringPtrToStr(cfOrder.PaymentSessionId),
		ExpiryTime:      timePtr(TimeFromTimestamp(cfOrder.OrderExpiryTime)),
		CreatedAt:       timePtr(TimeFromTimestamp(cfOrder.CreatedAt)),
		Customer:        nil,
		Metadata:        metadata,
		Raw:             raw,
	}

	order.ProviderDetails = &domain.OrderProviderDetail{
		Cashfree: &domain.CashfreeOrderDetail{
			CfOrderID: StringPtrToStr(cfOrder.CfOrderId),
			Entity:    StringPtrToStr(cfOrder.Entity),
			OrderNote: StringPtrToStr(cfOrder.OrderNote),
		},
	}
	// Populate OrderSplits if present
	if len(cfOrder.OrderSplits) > 0 {
		splits := make([]domain.CashfreeVendorSplit, 0, len(cfOrder.OrderSplits))
		for i := range cfOrder.OrderSplits {
			s := &cfOrder.OrderSplits[i]
			splits = append(splits, domain.CashfreeVendorSplit{
				VendorID:   StringPtrToStr(s.VendorId),
				Amount:     float64(FloatPtrToFloat64(s.Amount)),
				Percentage: float64(FloatPtrToFloat64(s.Percentage)),
			})
		}
		order.ProviderDetails.Cashfree.OrderSplits = splits
	}
	// Populate OrderMeta if present
	if cfOrder.OrderMeta != nil {
		order.ProviderDetails.Cashfree.OrderMeta = &domain.CashfreeOrderMeta{
			ReturnURL: StringPtrToStr(cfOrder.OrderMeta.ReturnUrl),
			NotifyURL: StringPtrToStr(cfOrder.OrderMeta.NotifyUrl),
		}
	}

	return order, nil
}

// MapPaymentEntityToCanonical maps a Cashfree PaymentEntity to the canonical domain.Payment type.
func MapPaymentEntityToCanonical(cfPayment *cf.PaymentEntity) (*domain.Payment, error) {
	if cfPayment == nil {
		return nil, fmt.Errorf("payment entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(cfPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment: %w", err)
	}

	payment := &domain.Payment{
		ProviderPaymentID: StringPtrToStr(cfPayment.CfPaymentId),
		OrderID:           StringPtrToStr(cfPayment.OrderId),
		Status:            mapPaymentStatus(cfPayment.PaymentStatus),
		AmountMinor:       domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(cfPayment.PaymentAmount), StringPtrToStr(cfPayment.PaymentCurrency))),
		Currency:          domain.Currency(StringPtrToStr(cfPayment.PaymentCurrency)),
		PaymentGroup:      StringPtrToStr(cfPayment.PaymentGroup),
		PaymentMethod:     "",
		PaymentTime:       timePtr(TimeFromTimestamp(cfPayment.PaymentTime)),
		CompletionTime:    timePtr(TimeFromTimestamp(cfPayment.PaymentCompletionTime)),
		IsCaptured:        derefBool(cfPayment.IsCaptured),
		BankReference:     StringPtrToStr(cfPayment.BankReference),
		ErrorCode:         "",
		ErrorMessage:      "",
		Raw:               raw,
	}

	payment.ProviderDetails = &domain.PaymentProviderDetail{
		Cashfree: &domain.CashfreePaymentDetail{
			CfPaymentID:    StringPtrToStr(cfPayment.CfPaymentId),
			Entity:         StringPtrToStr(cfPayment.Entity),
			OrderAmount:    float64(FloatPtrToFloat64(cfPayment.OrderAmount)),
			OrderCurrency:  StringPtrToStr(cfPayment.OrderCurrency),
			PaymentMessage: StringPtrToStr(cfPayment.PaymentMessage),
			AuthID:         StringPtrToStr(cfPayment.AuthId),
		},
	}
	if cfPayment.ErrorDetails != nil {
		payment.ProviderDetails.Cashfree.ErrorDetails = &domain.CashfreePaymentErrorDetails{
			ErrorCode:           StringPtrToStr(cfPayment.ErrorDetails.ErrorCode),
			ErrorDescription:    StringPtrToStr(cfPayment.ErrorDetails.ErrorDescription),
			ErrorReason:         StringPtrToStr(cfPayment.ErrorDetails.ErrorReason),
			ErrorSource:         StringPtrToStr(cfPayment.ErrorDetails.ErrorSource),
			ErrorCodeRaw:        StringPtrToStr(cfPayment.ErrorDetails.ErrorCodeRaw),
			ErrorDescriptionRaw: StringPtrToStr(cfPayment.ErrorDetails.ErrorDescriptionRaw),
			ErrorSubcodeRaw:     StringPtrToStr(cfPayment.ErrorDetails.ErrorSubcodeRaw),
		}
	}

	return payment, nil
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
func MapRefundEntityToCanonical(cfRefund *cf.RefundEntity) (*domain.Refund, error) {
	if cfRefund == nil {
		return nil, fmt.Errorf("refund entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(cfRefund)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund: %w", err)
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
		AmountMinor:      domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(cfRefund.RefundAmount), StringPtrToStr(cfRefund.RefundCurrency))),
		Currency:         domain.Currency(StringPtrToStr(cfRefund.RefundCurrency)),
		Reason:           StringPtrToStr(cfRefund.RefundNote),
		ARN:              StringPtrToStr(cfRefund.RefundArn),
		CreatedAt:        timePtr(TimeFromTimestamp(cfRefund.CreatedAt)),
		ProcessedAt:      timePtr(TimeFromTimestamp(cfRefund.ProcessedAt)),
		Raw:              raw,
	}

	refund.ProviderDetails = &domain.RefundProviderDetail{
		Cashfree: &domain.CashfreeRefundDetail{
			CfRefundID:        StringPtrToStr(cfRefund.CfRefundId),
			CfPaymentID:       StringPtrToStr(cfRefund.CfPaymentId),
			Entity:            StringPtrToStr(cfRefund.Entity),
			RefundCharge:      float64(FloatPtrToFloat64(cfRefund.RefundCharge)),
			RefundType:        StringPtrToStr(cfRefund.RefundType),
			RefundMode:        StringPtrToStr(cfRefund.RefundMode),
			StatusDescription: StringPtrToStr(cfRefund.StatusDescription),
			ChargesCurrency:   StringPtrToStr(cfRefund.ChargesCurrency),
			ForexRate:         float64(FloatPtrToFloat64(cfRefund.ForexConversionRate)),
			ForexCharge:       float64(FloatPtrToFloat64(cfRefund.ForexConversionHandlingCharge)),
			ForexTax:          float64(FloatPtrToFloat64(cfRefund.ForexConversionHandlingTax)),
			ProviderMetadata:  cfRefund.Metadata,
		},
	}
	if cfRefund.RefundSpeed != nil {
		refund.ProviderDetails.Cashfree.RefundSpeed = &domain.CashfreeRefundSpeed{
			Requested: StringPtrToStr(cfRefund.RefundSpeed.Requested),
			Accepted:  StringPtrToStr(cfRefund.RefundSpeed.Accepted),
			Processed: StringPtrToStr(cfRefund.RefundSpeed.Processed),
			Message:   StringPtrToStr(cfRefund.RefundSpeed.Message),
		}
	}
	// Populate RefundSplits if present
	if len(cfRefund.RefundSplits) > 0 {
		splits := make([]domain.CashfreeVendorSplit, 0, len(cfRefund.RefundSplits))
		for i := range cfRefund.RefundSplits {
			s := &cfRefund.RefundSplits[i]
			splits = append(splits, domain.CashfreeVendorSplit{
				VendorID:   StringPtrToStr(s.VendorId),
				Amount:     float64(FloatPtrToFloat64(s.Amount)),
				Percentage: float64(FloatPtrToFloat64(s.Percentage)),
			})
		}
		refund.ProviderDetails.Cashfree.RefundSplits = splits
	}

	return refund, nil
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
func MapInstrumentEntityToCanonical(cfInstrument *cf.InstrumentEntity) (*domain.Instrument, error) {
	if cfInstrument == nil {
		return nil, fmt.Errorf("instrument entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(cfInstrument)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal instrument: %w", err)
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
		Raw:            raw,
	}

	instrument.ProviderDetails = &domain.InstrumentProviderDetail{
		Cashfree: &domain.CashfreeInstrumentDetail{
			AfaReference:  StringPtrToStr(cfInstrument.AfaReference),
			InstrumentUID: StringPtrToStr(cfInstrument.InstrumentUid),
		},
	}
	if cfInstrument.InstrumentMeta != nil {
		instrument.ProviderDetails.Cashfree.InstrumentMeta = &domain.CashfreeInstrumentMeta{
			CardNetwork:  StringPtrToStr(cfInstrument.InstrumentMeta.CardNetwork),
			CardBankName: StringPtrToStr(cfInstrument.InstrumentMeta.CardBankName),
			CardCountry:  StringPtrToStr(cfInstrument.InstrumentMeta.CardCountry),
			CardType:     StringPtrToStr(cfInstrument.InstrumentMeta.CardType),
			CardSubType:  StringPtrToStr(cfInstrument.InstrumentMeta.CardSubType),
		}
	}

	return instrument, nil
}

// MapInstrumentEntityForAllSavedCardToCanonical maps a Cashfree InstrumentEntityForAllSavedCard
// (returned from list operations) to the canonical domain.Instrument type.
func MapInstrumentEntityForAllSavedCardToCanonical(cfInstrument *cf.InstrumentEntityForAllSavedCard) (*domain.Instrument, error) {
	if cfInstrument == nil {
		return nil, fmt.Errorf("instrument entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(cfInstrument)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal instrument: %w", err)
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
		Raw:            raw,
	}

	instrument.ProviderDetails = &domain.InstrumentProviderDetail{
		Cashfree: &domain.CashfreeInstrumentDetail{
			AfaReference:  StringPtrToStr(cfInstrument.AfaReference),
			InstrumentUID: StringPtrToStr(cfInstrument.InstrumentUid),
		},
	}
	if cfInstrument.InstrumentMeta != nil {
		instrument.ProviderDetails.Cashfree.InstrumentMeta = &domain.CashfreeInstrumentMeta{
			CardNetwork:  StringPtrToStr(cfInstrument.InstrumentMeta.CardNetwork),
			CardBankName: StringPtrToStr(cfInstrument.InstrumentMeta.CardBankName),
			CardCountry:  StringPtrToStr(cfInstrument.InstrumentMeta.CardCountry),
			CardType:     StringPtrToStr(cfInstrument.InstrumentMeta.CardType),
			CardSubType:  StringPtrToStr(cfInstrument.InstrumentMeta.CardSubType),
		}
	}

	return instrument, nil
}

// MapLinkEntityToCanonical maps a Cashfree LinkEntity (payment link) to the canonical domain.PaymentLink type.
func MapLinkEntityToCanonical(cfLink *cf.LinkEntity) (*domain.PaymentLink, error) {
	if cfLink == nil {
		return nil, fmt.Errorf("link entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal link: %w", err)
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
		AmountMinor:    domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(cfLink.LinkAmount), StringPtrToStr(cfLink.LinkCurrency))),
		AmountPaid:     domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(cfLink.LinkAmountPaid), StringPtrToStr(cfLink.LinkCurrency))),
		Currency:       domain.Currency(StringPtrToStr(cfLink.LinkCurrency)),
		Purpose:        StringPtrToStr(cfLink.LinkPurpose),
		LinkURL:        StringPtrToStr(cfLink.LinkUrl),
		Customer:       nil, // Customer info not directly available from LinkEntity
		CreatedAt:      timePtr(TimeFromTimestamp(cfLink.LinkCreatedAt)),
		ExpiryTime:     timePtr(TimeFromTimestamp(cfLink.LinkExpiryTime)),
		Metadata:       metadata,
		Raw:            raw,
	}

	link.ProviderDetails = &domain.PaymentLinkProviderDetail{
		Cashfree: &domain.CashfreePaymentLinkDetail{
			CfLinkID:         StringPtrToStr(cfLink.CfLinkId),
			PartialPayments:  derefBool(cfLink.LinkPartialPayments),
			MinPartialAmount: float64(FloatPtrToFloat64(cfLink.LinkMinimumPartialAmount)),
			AutoReminders:    derefBool(cfLink.LinkAutoReminders),
			LinkQRCode:       StringPtrToStr(cfLink.LinkQrcode),
		},
	}
	if len(cfLink.OrderSplits) > 0 {
		splits := make([]domain.CashfreeVendorSplit, 0, len(cfLink.OrderSplits))
		for i := range cfLink.OrderSplits {
			s := &cfLink.OrderSplits[i]
			splits = append(splits, domain.CashfreeVendorSplit{
				VendorID:   StringPtrToStr(s.VendorId),
				Amount:     float64(FloatPtrToFloat64(s.Amount)),
				Percentage: float64(FloatPtrToFloat64(s.Percentage)),
			})
		}
		link.ProviderDetails.Cashfree.OrderSplits = splits
	}

	return link, nil
}

// derefBool safely dereferences a bool pointer or returns false.
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// rawResponse marshals a Cashfree SDK entity to JSON and returns it as RawProviderResponse.
func rawResponse(v any) (domain.RawProviderResponse, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return domain.RawProviderResponse(data), nil
}

// MapPlanEntityToCanonical maps a Cashfree PlanEntity to the canonical domain.Plan type.
// Handles type conversions, currency lookup, and status mapping.
func MapPlanEntityToCanonical(entity *cf.PlanEntity) (*domain.Plan, error) {
	if entity == nil {
		return nil, fmt.Errorf("plan entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plan: %w", err)
	}

	// Extract currency from the entity (if available in the response)
	currency := ""
	if entity.PlanCurrency != nil {
		currency = *entity.PlanCurrency
	}

	// Map Cashfree status string to domain status (if status field exists in PlanEntity)
	status := ""
	if entity.PlanStatus != nil {
		status = *entity.PlanStatus
	}

	plan := &domain.Plan{
		PlanID:         StringPtrToStr(entity.PlanId),
		PlanName:       StringPtrToStr(entity.PlanName),
		PlanType:       mapPlanType(entity.PlanType),
		Currency:       domain.Currency(currency),
		AmountMinor:    domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(entity.PlanRecurringAmount), currency)),
		MaxAmountMinor: domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(entity.PlanMaxAmount), currency)),
		Interval:       getInt32OrDefault(entity.PlanIntervals),
		IntervalType:   mapPlanIntervalType(entity.PlanIntervalType),
		MaxCycles:      getInt32OrDefault(entity.PlanMaxCycles),
		Status:         status,
		Note:           StringPtrToStr(entity.PlanNote),
		Provider:       domain.ProviderCashfree,
		Raw:            raw,
	}

	return plan, nil
}

// MapSubscriptionEntityToCanonical maps a Cashfree SubscriptionEntity to the canonical domain.Subscription type.
// Handles status mapping and timestamp conversion.
func MapSubscriptionEntityToCanonical(entity *cf.SubscriptionEntity) (*domain.Subscription, error) {
	if entity == nil {
		return nil, fmt.Errorf("subscription entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription: %w", err)
	}

	// Extract PlanID from nested PlanDetails
	planID := ""
	if entity.PlanDetails != nil {
		planID = StringPtrToStr(entity.PlanDetails.PlanId)
	}

	// Extract customer details from nested CustomerDetails
	customerEmail := ""
	customerPhone := ""
	if entity.CustomerDetails != nil {
		customerEmail = entity.CustomerDetails.CustomerEmail
		customerPhone = entity.CustomerDetails.CustomerPhone
	}

	// Cashfree returns the mandate-authorization handle as the subscription session id
	// (used by the Cashfree JS SDK for the auth step); map it to the canonical AuthLink.
	authLink := StringPtrToStr(entity.SubscriptionSessionId)

	subscription := &domain.Subscription{
		SubscriptionID:         StringPtrToStr(entity.SubscriptionId),
		ProviderSubscriptionID: StringPtrToStr(entity.CfSubscriptionId),
		PlanID:                 planID,
		Status:                 mapSubscriptionStatus(entity.SubscriptionStatus),
		CustomerEmail:          customerEmail,
		CustomerPhone:          customerPhone,
		AuthLink:               authLink,
		ExpiresAt:              timePtr(TimeFromTimestamp(entity.SubscriptionExpiryTime)),
		FirstChargeTime:        timePtr(TimeFromTimestamp(entity.SubscriptionFirstChargeTime)),
		NextChargeDate:         timePtr(TimeFromTimestamp(entity.NextScheduleDate)),
		Provider:               domain.ProviderCashfree,
		Raw:                    raw,
	}

	return subscription, nil
}

// MapSubscriptionPaymentEntityToCanonical maps a Cashfree SubscriptionPaymentEntity to the canonical domain.SubscriptionPayment type.
// currency is the ISO-4217 code of the parent subscription's plan, used to convert the major-unit payment_amount to minor units; it must be a non-empty code resolved by the caller.
func MapSubscriptionPaymentEntityToCanonical(entity *cf.SubscriptionPaymentEntity, currency string) (*domain.SubscriptionPayment, error) {
	if entity == nil {
		return nil, fmt.Errorf("subscription payment entity is required: %w", domain.ErrInvalidRequest)
	}

	raw, err := rawResponse(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription payment: %w", err)
	}

	payment := &domain.SubscriptionPayment{
		PaymentID:      StringPtrToStr(entity.PaymentId),
		SubscriptionID: StringPtrToStr(entity.SubscriptionId),
		AmountMinor:    domain.AmountMinor(currencyutils.AmountMajorToMinor(FloatPtrToFloat64(entity.PaymentAmount), currency)),
		Status:         mapSubscriptionPaymentStatus(entity.PaymentStatus),
		PaymentType:    mapSubscriptionPaymentType(entity.PaymentType),
		ScheduledDate:  timePtr(TimeFromTimestamp(entity.PaymentScheduleDate)),
		InitiatedDate:  timePtr(TimeFromTimestamp(entity.PaymentInitiatedDate)),
		RetryAttempts:  int(getInt32OrDefault(entity.RetryAttempts)),
		Provider:       domain.ProviderCashfree,
		Raw:            raw,
	}

	return payment, nil
}

// mapPlanType converts Cashfree plan type strings to canonical domain.PlanType.
func mapPlanType(typePtr *string) domain.PlanType {
	planType := ""
	if typePtr != nil {
		planType = *typePtr
	}
	switch planType {
	case "PERIODIC":
		return domain.PlanTypePeriodic
	case "ON_DEMAND":
		return domain.PlanTypeOnDemand
	default:
		return domain.PlanTypePeriodic
	}
}

// mapPlanIntervalType converts Cashfree plan interval type strings to canonical domain.PlanIntervalType.
func mapPlanIntervalType(typePtr *string) domain.PlanIntervalType {
	intervalType := ""
	if typePtr != nil {
		intervalType = *typePtr
	}
	switch intervalType {
	case "DAY":
		return domain.PlanIntervalDay
	case "WEEK":
		return domain.PlanIntervalWeek
	case "MONTH":
		return domain.PlanIntervalMonth
	case "YEAR":
		return domain.PlanIntervalYear
	default:
		return domain.PlanIntervalMonth
	}
}

// mapSubscriptionStatus converts Cashfree subscription status strings to canonical domain.SubscriptionStatus.
func mapSubscriptionStatus(statusPtr *string) domain.SubscriptionStatus {
	status := ""
	if statusPtr != nil {
		status = *statusPtr
	}
	switch status {
	case "INITIALIZED":
		return domain.SubscriptionStatusInitialized
	case "BANK_APPROVAL_PENDING":
		return domain.SubscriptionStatusBankApprovalPending
	case "AUTHENTICATED":
		return domain.SubscriptionStatusAuthenticated
	case "ACTIVE":
		return domain.SubscriptionStatusActive
	case "PENDING":
		return domain.SubscriptionStatusPending
	case "ON_HOLD":
		return domain.SubscriptionStatusOnHold
	case "HALTED":
		return domain.SubscriptionStatusHalted
	case "PAUSED":
		return domain.SubscriptionStatusPaused
	case "CUSTOMER_PAUSED":
		return domain.SubscriptionStatusCustomerPaused
	case "CANCELLED":
		return domain.SubscriptionStatusCancelled
	case "CUSTOMER_CANCELLED":
		return domain.SubscriptionStatusCustomerCancelled
	case "COMPLETED":
		return domain.SubscriptionStatusCompleted
	case "EXPIRED":
		return domain.SubscriptionStatusExpired
	default:
		return domain.SubscriptionStatusInitialized
	}
}

// mapSubscriptionPaymentStatus converts Cashfree subscription payment status strings to canonical domain.SubscriptionPaymentStatus.
func mapSubscriptionPaymentStatus(statusPtr *string) domain.SubscriptionPaymentStatus {
	status := ""
	if statusPtr != nil {
		status = *statusPtr
	}
	switch status {
	case "SCHEDULED":
		return domain.SubPaymentStatusScheduled
	case "PENDING":
		return domain.SubPaymentStatusPending
	case "SUCCESS":
		return domain.SubPaymentStatusSuccess
	case "FAILED":
		return domain.SubPaymentStatusFailed
	case "CANCELLED":
		return domain.SubPaymentStatusCancelled
	default:
		return domain.SubPaymentStatusScheduled
	}
}

// mapSubscriptionPaymentType converts Cashfree subscription payment type strings to canonical domain.SubscriptionPaymentType.
func mapSubscriptionPaymentType(typePtr *string) domain.SubscriptionPaymentType {
	paymentType := ""
	if typePtr != nil {
		paymentType = *typePtr
	}
	switch paymentType {
	case "AUTH":
		return domain.SubPaymentTypeAuth
	case "CHARGE":
		return domain.SubPaymentTypeCharge
	default:
		return domain.SubPaymentTypeCharge
	}
}

// getInt32OrDefault safely converts a *int32 pointer to int32 or returns 0.
func getInt32OrDefault(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
