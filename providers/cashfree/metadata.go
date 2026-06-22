package cashfree

import (
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// MapOrderMetadata extracts provider-specific metadata from a Cashfree OrderEntity.
// Preserves Cashfree-specific fields (OrderId, OrderNote, OrderStatus, CfOrderId, timestamps, tags)
// that don't fit into the canonical Order type but may be needed for auditing or reconciliation.
func MapOrderMetadata(cfOrder *cf.OrderEntity) *domain.Metadata {
	if cfOrder == nil {
		return nil
	}

	metadata := domain.Metadata{}

	// Store OrderId (merchant's order identifier)
	if cfOrder.OrderId != nil {
		metadata["order_id"] = *cfOrder.OrderId
	}

	// Store CfOrderId (Cashfree's internal order identifier)
	if cfOrder.CfOrderId != nil {
		metadata["cf_order_id"] = *cfOrder.CfOrderId
	}

	// Store OrderNote (additional notes for the order)
	if cfOrder.OrderNote != nil {
		metadata["order_note"] = *cfOrder.OrderNote
	}

	// Store OrderStatus (ACTIVE, PAID, EXPIRED, CANCELLED)
	if cfOrder.OrderStatus != nil {
		metadata["order_status"] = *cfOrder.OrderStatus
	}

	// Store PaymentSessionId
	if cfOrder.PaymentSessionId != nil {
		metadata["payment_session_id"] = *cfOrder.PaymentSessionId
	}

	// Store creation timestamp (Cashfree format)
	if cfOrder.CreatedAt != nil {
		metadata["created_at"] = *cfOrder.CreatedAt
	}

	// Store order expiry time if set
	if cfOrder.OrderExpiryTime != nil {
		metadata["order_expiry_time"] = *cfOrder.OrderExpiryTime
	}

	// Store Entity type (typically "order")
	if cfOrder.Entity != nil {
		metadata["entity"] = *cfOrder.Entity
	}

	// Store custom order tags as serialized key-value pairs
	// Tags are stored with "tag_" prefix for easy identification
	if cfOrder.OrderTags != nil && len(*cfOrder.OrderTags) > 0 {
		for key, value := range *cfOrder.OrderTags {
			metadata["tag_"+key] = value
		}
	}

	return &metadata
}

// MapRefundMetadata extracts provider-specific metadata from a Cashfree RefundEntity.
// Preserves Cashfree-specific fields (RefundId, RefundNote, RefundStatus, RefundAmount,
// CfRefundId, RefundArn, timestamps, metadata) that don't fit into the canonical Refund type
// but may be needed for auditing, reconciliation, or compliance tracking.
func MapRefundMetadata(cfRefund *cf.RefundEntity) *domain.Metadata {
	if cfRefund == nil {
		return nil
	}

	metadata := domain.Metadata{}

	// Store RefundId (merchant's refund identifier)
	if cfRefund.RefundId != nil {
		metadata["refund_id"] = *cfRefund.RefundId
	}

	// Store CfRefundId (Cashfree's internal refund identifier)
	if cfRefund.CfRefundId != nil {
		metadata["cf_refund_id"] = *cfRefund.CfRefundId
	}

	// Store CfPaymentId (Cashfree's internal payment identifier for this refund)
	if cfRefund.CfPaymentId != nil {
		metadata["cf_payment_id"] = *cfRefund.CfPaymentId
	}

	// Store OrderId
	if cfRefund.OrderId != nil {
		metadata["order_id"] = *cfRefund.OrderId
	}

	// Store RefundNote (notes for the refund)
	if cfRefund.RefundNote != nil {
		metadata["refund_note"] = *cfRefund.RefundNote
	}

	// Store RefundStatus (SUCCESS, PENDING, CANCELLED, ONHOLD, FAILED)
	if cfRefund.RefundStatus != nil {
		metadata["refund_status"] = *cfRefund.RefundStatus
	}

	// Store RefundAmount in the original Cashfree format (rupees as float)
	if cfRefund.RefundAmount != nil {
		metadata["refund_amount"] = fmt.Sprintf("%.2f", *cfRefund.RefundAmount)
	}

	// Store RefundCurrency
	if cfRefund.RefundCurrency != nil {
		metadata["refund_currency"] = *cfRefund.RefundCurrency
	}

	// Store RefundArn (bank reference number for the refund)
	if cfRefund.RefundArn != nil {
		metadata["refund_arn"] = *cfRefund.RefundArn
	}

	// Store RefundCharge (processing charge for the refund in INR)
	if cfRefund.RefundCharge != nil {
		metadata["refund_charge"] = fmt.Sprintf("%.2f", *cfRefund.RefundCharge)
	}

	// Store status description
	if cfRefund.StatusDescription != nil {
		metadata["status_description"] = *cfRefund.StatusDescription
	}

	// Store RefundType (PAYMENT_AUTO_REFUND, MERCHANT_INITIATED, UNRECONCILED_AUTO_REFUND)
	if cfRefund.RefundType != nil {
		metadata["refund_type"] = *cfRefund.RefundType
	}

	// Store RefundMode (method or speed of processing)
	if cfRefund.RefundMode != nil {
		metadata["refund_mode"] = *cfRefund.RefundMode
	}

	// Store Entity type (typically "refund")
	if cfRefund.Entity != nil {
		metadata["entity"] = *cfRefund.Entity
	}

	// Store creation timestamp (Cashfree format)
	if cfRefund.CreatedAt != nil {
		metadata["created_at"] = *cfRefund.CreatedAt
	}

	// Store processing timestamp if available
	if cfRefund.ProcessedAt != nil {
		metadata["processed_at"] = *cfRefund.ProcessedAt
	}

	// Store custom metadata from Cashfree's refund metadata field
	// These are stored with "meta_" prefix to distinguish from standard fields
	if len(cfRefund.Metadata) > 0 {
		for key, value := range cfRefund.Metadata {
			metadata["meta_"+key] = fmt.Sprintf("%v", value)
		}
	}

	return &metadata
}
