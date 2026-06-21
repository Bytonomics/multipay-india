package ports

import "github.com/Bytonomics/multipay-adapter/domain"

// MetadataMapper translates provider-specific metadata into the adapter's canonical format.
// Different payment providers have different metadata structures and requirements.
// Implementations of this interface handle provider-specific translation logic.
type MetadataMapper interface {
	// Map translates provider-specific metadata into the adapter's canonical Metadata format.
	// Returns an error if the translation fails or if required metadata is missing.
	Map(metadata domain.Metadata) error
}
