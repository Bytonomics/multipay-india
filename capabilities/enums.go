package capabilities

// Capability represents a feature that may or may not be supported by a payment provider.
type Capability string

// Core Shared Capabilities (supported by both Cashfree and Razorpay)
const (
	CapOrderCreate       Capability = "order_create"
	CapOrderFetch        Capability = "order_fetch"
	CapOrderListPayments Capability = "order_list_payments"
	CapPaymentFetch      Capability = "payment_fetch"
	CapPaymentList       Capability = "payment_list"
	CapPaymentPay        Capability = "payment_pay"
	CapRefundCreate      Capability = "refund_create"
	CapRefundFetch       Capability = "refund_fetch"
	CapRefundList        Capability = "refund_list"
	CapInstrumentFetch   Capability = "instrument_fetch"
	CapInstrumentList    Capability = "instrument_list"
	CapInstrumentDelete  Capability = "instrument_delete"
	CapPaymentLinkCreate Capability = "payment_link_create"
	CapPaymentLinkFetch  Capability = "payment_link_fetch"
	CapPaymentLinkCancel Capability = "payment_link_cancel"
	CapWebhookConsume    Capability = "webhook_consume"
)

// Cashfree-specific Capabilities
const (
	CapInstrumentCryptogram  Capability = "instrument_cryptogram"
	CapOfferCreate           Capability = "offer_create"
	CapOfferFetch            Capability = "offer_fetch"
	CapEligibilityFetch      Capability = "eligibility_fetch"
	CapPaymentLinkListOrders Capability = "payment_link_list_orders"
	CapSettlementOrderFetch  Capability = "settlement_order_fetch"
	CapSettlementList        Capability = "settlement_list"
	CapSettlementReconFetch  Capability = "settlement_recon_fetch"
	CapReconFetch            Capability = "recon_fetch"
)

// Razorpay-specific Capabilities
const (
	CapOrderUpdate        Capability = "order_update"
	CapOrderList          Capability = "order_list"
	CapPaymentCapture     Capability = "payment_capture"
	CapRefundUpdate       Capability = "refund_update"
	CapCustomerCreate     Capability = "customer_create"
	CapCustomerFetch      Capability = "customer_fetch"
	CapCustomerEdit       Capability = "customer_edit"
	CapCustomerList       Capability = "customer_list"
	CapWebhookCreate      Capability = "webhook_create"
	CapWebhookFetch       Capability = "webhook_fetch"
	CapWebhookEdit        Capability = "webhook_edit"
	CapWebhookDelete      Capability = "webhook_delete"
	CapWebhookList        Capability = "webhook_list"
	CapSubscriptionCreate Capability = "subscription_create"
	CapSubscriptionFetch  Capability = "subscription_fetch"
	CapSubscriptionList   Capability = "subscription_list"
	CapPlanCreate         Capability = "plan_create"
	CapPlanFetch          Capability = "plan_fetch"
	CapPlanList           Capability = "plan_list"
	CapPaymentLinkUpdate  Capability = "payment_link_update"
	CapPaymentLinkNotify  Capability = "payment_link_notify"
	CapPaymentLinkList    Capability = "payment_link_list"
	CapUPICreate          Capability = "upi_create"
	CapVPAValidate        Capability = "vpa_validate"

	// Razorpay-specific settlement capabilities
	CapSettlementAll            Capability = "settlement_all"
	CapSettlementFetch          Capability = "settlement_fetch"
	CapSettlementReports        Capability = "settlement_reports"
	CapSettlementOnDemandCreate Capability = "settlement_ondemand_create"
	CapSettlementOnDemandFetch  Capability = "settlement_ondemand_fetch"
	CapSettlementOnDemandList   Capability = "settlement_ondemand_list"
)
