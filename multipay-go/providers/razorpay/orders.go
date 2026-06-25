package razorpay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// D17: Typed request struct for Razorpay Create Order API.
type razorpayCreateOrderRequest struct {
	Amount   int64           `json:"amount"`
	Currency string          `json:"currency"`
	Notes    domain.Metadata `json:"notes,omitempty"`
	Receipt  string          `json:"receipt,omitempty"`
}

// D17: Typed response struct for Razorpay Order API responses.
type razorpayOrderResponse struct {
	ID         string `json:"id"`
	Entity     string `json:"entity"`
	Receipt    string `json:"receipt"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"created_at"`
	OfferID    string `json:"offer_id"`
	AmountPaid int64  `json:"amount_paid"`
	AmountDue  int64  `json:"amount_due"`
	Attempts   int64  `json:"attempts"`
}

// CreateOrder creates a new order on Razorpay.
// It takes a CreateOrderRequest and returns a canonical Order domain object.
// The amount is in paisa (minor currency unit), which Razorpay uses natively.
func (a *Adapter) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}

	// Build Razorpay order creation parameters
	params, err := encodeRequest(&razorpayCreateOrderRequest{
		Amount:   int64(req.AmountMinor),
		Currency: string(req.Currency),
		Notes:    req.Metadata,
		Receipt:  req.OrderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode create order request: %w", err)
	}

	// Call Razorpay SDK to create order
	responseMap, err := a.client.Order.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayOrderResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Map typed struct to canonical domain type
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order response: %w", err)
	}
	order := mapOrderFromResponse(typed, rawJSON)
	order.Checkout = buildRazorpayCheckout(a.config, req, order)
	return order, nil
}

// buildRazorpayCheckout constructs checkout payload for frontend rendering.
func buildRazorpayCheckout(cfg *Config, req *domain.CreateOrderRequest, order *domain.Order) *domain.CheckoutPayload {
	return &domain.CheckoutPayload{
		Provider:    domain.ProviderRazorpay,
		Environment: cfg.Environment,
		OrderID:     order.ProviderOrderID,
		PublicKey:   cfg.Key,
		CallbackURL: req.ReturnURL,
		AmountMinor: order.AmountMinor,
		Currency:    order.Currency,
	}
}

// GetOrder retrieves an existing order from Razorpay.
// It takes a GetOrderRequest and returns a canonical Order domain object.
func (a *Adapter) GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch order
	responseMap, err := a.client.Order.Fetch(req.OrderID, nil, nil)
	if err != nil {
		// Check if order not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("order %s not found: %w", req.OrderID, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayOrderResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Map typed struct to canonical domain type
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order response: %w", err)
	}
	return mapOrderFromResponse(typed, rawJSON), nil
}

// D17: Typed response struct for Razorpay Payment API responses (used in list responses).
type razorpayPaymentResponse struct {
	ID               string                `json:"id"`
	Entity           string                `json:"entity"`
	OrderID          string                `json:"order_id"`
	Amount           int64                 `json:"amount"`
	Currency         string                `json:"currency"`
	Status           string                `json:"status"`
	Method           string                `json:"method"`
	Captured         bool                  `json:"captured"`
	BankAccount      string                `json:"bank_account"`
	ErrorCode        string                `json:"error_code"`
	ErrorDescription string                `json:"error_description"`
	CreatedAt        int64                 `json:"created_at"`
	Description      string                `json:"description"`
	Email            string                `json:"email"`
	Contact          string                `json:"contact"`
	Fee              int64                 `json:"fee"`
	Tax              int64                 `json:"tax"`
	AmountRefunded   int64                 `json:"amount_refunded"`
	RefundStatus     string                `json:"refund_status"`
	International    bool                  `json:"international"`
	CardID           string                `json:"card_id"`
	Bank             string                `json:"bank"`
	VPA              string                `json:"vpa"`
	Wallet           string                `json:"wallet"`
	ErrorSource      string                `json:"error_source"`
	ErrorStep        string                `json:"error_step"`
	ErrorReason      string                `json:"error_reason"`
	AcquirerData     *razorpayAcquirerData `json:"acquirer_data"`
}

type razorpayAcquirerData struct {
	BankTransactionID string `json:"bank_transaction_id"`
	AuthCode          string `json:"auth_code"`
	RRN               string `json:"rrn"`
}

// D17: Typed response struct for payment list response.
type razorpayPaymentListResponse struct {
	Items []razorpayPaymentResponse `json:"items"`
}

// ListOrderPayments retrieves all payments associated with a specific order.
// It takes a ListOrderPaymentsRequest with order ID and returns a slice of canonical Payment domain objects.
func (a *Adapter) ListOrderPayments(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch payments for the order
	paymentsData, err := a.client.Order.Payments(req.OrderID, nil, nil)
	if err != nil {
		// Check if order not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("order %s not found: %w", req.OrderID, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("failed to list order payments: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentListResponse](paymentsData)
	if err != nil {
		return nil, err
	}

	// D17: Map each typed payment response to canonical domain type
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payments response: %w", err)
	}
	payments := make([]*domain.Payment, 0, len(typed.Items))
	for i := range typed.Items {
		payment := mapPaymentFromResponse(&typed.Items[i], rawJSON)
		payments = append(payments, payment)
	}

	return payments, nil
}

// D17: Typed struct mapper for order response
func mapOrderFromResponse(r *razorpayOrderResponse, rawJSON []byte) *domain.Order {
	return &domain.Order{
		ProviderOrderID: r.ID,
		OrderID:         r.Receipt,
		AmountMinor:     domain.AmountMinor(r.Amount),
		Currency:        domain.Currency(r.Currency),
		Status:          mapOrderStatus(r.Status),
		CreatedAt:       unixPtr(r.CreatedAt),
		Raw:             domain.RawProviderResponse(rawJSON),
		ProviderDetails: &domain.OrderProviderDetail{
			Razorpay: &domain.RazorpayOrderDetail{
				Entity:     r.Entity,
				Receipt:    r.Receipt,
				OfferID:    r.OfferID,
				AmountPaid: r.AmountPaid,
				AmountDue:  r.AmountDue,
				Attempts:   r.Attempts,
			},
		},
	}
}

// D17: Typed struct mapper for payment response
func mapPaymentFromResponse(r *razorpayPaymentResponse, rawJSON []byte) *domain.Payment {
	payment := &domain.Payment{
		ProviderPaymentID: r.ID,
		OrderID:           r.OrderID,
		AmountMinor:       domain.AmountMinor(r.Amount),
		Currency:          domain.Currency(r.Currency),
		Status:            mapPaymentStatus(r.Status),
		PaymentMethod:     r.Method,
		IsCaptured:        r.Captured,
		BankReference:     r.BankAccount,
		ErrorCode:         r.ErrorCode,
		ErrorMessage:      r.ErrorDescription,
		PaymentTime:       unixPtr(r.CreatedAt),
		Raw:               domain.RawProviderResponse(rawJSON),
		ProviderDetails: &domain.PaymentProviderDetail{
			Razorpay: &domain.RazorpayPaymentDetail{
				Entity:         r.Entity,
				Description:    r.Description,
				Email:          r.Email,
				Contact:        r.Contact,
				Fee:            r.Fee,
				Tax:            r.Tax,
				AmountRefunded: r.AmountRefunded,
				RefundStatus:   r.RefundStatus,
				International:  r.International,
				CardID:         r.CardID,
				Bank:           r.Bank,
				VPA:            r.VPA,
				Wallet:         r.Wallet,
				ErrorSource:    r.ErrorSource,
				ErrorStep:      r.ErrorStep,
				ErrorReason:    r.ErrorReason,
			},
		},
	}

	// D17: Map acquirer_data if present
	if r.AcquirerData != nil {
		payment.ProviderDetails.Razorpay.AcquirerData = &domain.RazorpayAcquirerData{
			BankTransactionID: r.AcquirerData.BankTransactionID,
			AuthCode:          r.AcquirerData.AuthCode,
			RRN:               r.AcquirerData.RRN,
		}
	}

	return payment
}

// mapOrderStatus converts Razorpay order status to canonical domain OrderStatus.
// Razorpay uses: "created", "attempted", "paid", "cancelled", "expired"
// Domain uses: "created", "paid", "expired", "cancelled"
func mapOrderStatus(razorpayStatus string) domain.OrderStatus {
	switch razorpayStatus {
	case "created":
		return domain.OrderCreated
	case "attempted":
		// Razorpay has "attempted" state when payment is being processed
		// For canonical domain, we treat this as "created" (not yet paid)
		return domain.OrderCreated
	case "paid":
		return domain.OrderPaid
	case "expired":
		return domain.OrderExpired
	case "cancelled":
		return domain.OrderCancelled
	default:
		return domain.OrderCreated
	}
}

// mapPaymentStatus converts Razorpay payment status to canonical domain PaymentStatus.
// Razorpay uses: "authorized", "captured", "failed", "refunded"
// Domain uses: "authorized", "captured", "failed", "refunded"
func mapPaymentStatus(razorpayStatus string) domain.PaymentStatus {
	switch razorpayStatus {
	case "authorized":
		return domain.PaymentAuthorized
	case "captured":
		return domain.PaymentCaptured
	case "failed":
		return domain.PaymentFailed
	case "refunded":
		return domain.PaymentRefunded
	default:
		return domain.PaymentFailed
	}
}
