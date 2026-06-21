package razorpay

// MapOrderMetadata extracts provider-specific metadata from a Razorpay order response.
// Preserves Razorpay-specific fields (receipt, notes, first_payment_min_amount, amount_paid,
// amount_due, payments_count, customer_id, created_at) that don't fit into the canonical
// Order type but may be needed for auditing or reconciliation.
// Returns non-nil empty map if no metadata is found; never returns nil.
func MapOrderMetadata(rzOrder map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	if rzOrder == nil {
		return metadata
	}

	// Store receipt (order receipt/reference)
	if receipt := getString(rzOrder, "receipt"); receipt != "" {
		metadata["receipt"] = receipt
	}

	// Store notes (any notes/tags attached to order)
	if notes := getString(rzOrder, "notes"); notes != "" {
		metadata["notes"] = notes
	}

	// Store first_payment_min_amount (minimum for partial payments if enabled)
	if minAmount := getInt64(rzOrder, "first_payment_min_amount"); minAmount > 0 {
		metadata["first_payment_min_amount"] = minAmount
	}

	// Store amount_paid (already paid amount)
	if amountPaid := getInt64(rzOrder, "amount_paid"); amountPaid >= 0 {
		metadata["amount_paid"] = amountPaid
	}

	// Store amount_due (remaining amount)
	if amountDue := getInt64(rzOrder, "amount_due"); amountDue >= 0 {
		metadata["amount_due"] = amountDue
	}

	// Store payments_count (number of payments on this order)
	if paymentsCount := getInt64(rzOrder, "payments_count"); paymentsCount >= 0 {
		metadata["payments_count"] = paymentsCount
	}

	// Store customer_id (if present)
	if customerID := getString(rzOrder, "customer_id"); customerID != "" {
		metadata["customer_id"] = customerID
	}

	// Store created_at (Unix timestamp)
	if createdAt := getInt64(rzOrder, "created_at"); createdAt > 0 {
		metadata["created_at"] = createdAt
	}

	return metadata
}

// MapRefundMetadata extracts provider-specific metadata from a Razorpay refund response.
// Preserves Razorpay-specific refund fields (receipt, notes, batch_id, status_details,
// speed_processed, created_at, first_min_partial_amount) that don't fit into the canonical
// Refund type but may be needed for auditing, reconciliation, or compliance tracking.
// Returns non-nil empty map if no metadata is found; never returns nil.
func MapRefundMetadata(rzRefund map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	if rzRefund == nil {
		return metadata
	}

	// Store receipt (refund receipt if available)
	if receipt := getString(rzRefund, "receipt"); receipt != "" {
		metadata["receipt"] = receipt
	}

	// Store notes (notes attached to refund)
	if notes := getString(rzRefund, "notes"); notes != "" {
		metadata["notes"] = notes
	}

	// Store batch_id (for batch refunds)
	if batchID := getString(rzRefund, "batch_id"); batchID != "" {
		metadata["batch_id"] = batchID
	}

	// Store status_details (status reason/description)
	if statusDetails := getString(rzRefund, "status_details"); statusDetails != "" {
		metadata["status_details"] = statusDetails
	}

	// Store speed_processed (processing speed)
	if speedProcessed := getString(rzRefund, "speed_processed"); speedProcessed != "" {
		metadata["speed_processed"] = speedProcessed
	}

	// Store created_at (Unix timestamp)
	if createdAt := getInt64(rzRefund, "created_at"); createdAt > 0 {
		metadata["created_at"] = createdAt
	}

	// Store first_min_partial_amount (if partial refunds enabled)
	if firstMinPartial := getInt64(rzRefund, "first_min_partial_amount"); firstMinPartial > 0 {
		metadata["first_min_partial_amount"] = firstMinPartial
	}

	return metadata
}
