package ports

import (
	"context"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// MetadataMapper translates provider-specific metadata into the adapter's canonical format.
// Different payment providers have different metadata structures and requirements.
// Implementations of this interface handle provider-specific translation logic.
type MetadataMapper interface {
	MapOrderMetadata(ctx context.Context, metadata domain.Metadata) (map[string]interface{}, error)
	MapRefundMetadata(ctx context.Context, metadata domain.Metadata) (map[string]interface{}, error)
	MapPaymentLinkMetadata(ctx context.Context, metadata domain.Metadata) (map[string]interface{}, error)
}
