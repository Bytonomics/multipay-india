package razorpay

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// getString safely extracts a string value from a map.
// Returns empty string if the key doesn't exist or the value is not a string.
func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	val, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// getInt64 safely extracts an int64 value from a map.
// Handles conversion from float64 (JSON number type).
// Returns 0 if the key doesn't exist or the value cannot be converted to int64.
func getInt64(m map[string]interface{}, key string) int64 {
	if m == nil {
		return 0
	}
	val, ok := m[key]
	if !ok {
		return 0
	}

	// Handle int64 directly
	if i, ok := val.(int64); ok {
		return i
	}

	// Handle float64 (JSON numbers come as float64)
	if f, ok := val.(float64); ok {
		return int64(f)
	}

	// Handle string representation
	if s, ok := val.(string); ok {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
	}

	return 0
}

// getMap safely extracts a nested map[string]interface{} from a map.
// Returns an empty map if the key doesn't exist or the value is not a map.
func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return make(map[string]interface{})
	}
	val, ok := m[key]
	if !ok {
		return make(map[string]interface{})
	}
	innerMap, ok := val.(map[string]interface{})
	if !ok {
		return make(map[string]interface{})
	}
	return innerMap
}

// getTime safely extracts and parses a time value from a map.
// Supports both Unix timestamp (float64/int64) and RFC3339 string formats.
// Returns nil if the key doesn't exist or parsing fails.
func getTime(m map[string]interface{}, key string) *time.Time {
	if m == nil {
		return nil
	}
	val, ok := m[key]
	if !ok {
		return nil
	}

	// Try as Unix timestamp (float64 or int64)
	if f, ok := val.(float64); ok {
		t := time.Unix(int64(f), 0)
		return &t
	}
	if i, ok := val.(int64); ok {
		t := time.Unix(i, 0)
		return &t
	}

	// Try as string (RFC3339 or other format)
	if s, ok := val.(string); ok {
		// Try RFC3339 format
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return &t
		}
		// Try Unix timestamp as string
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			t := time.Unix(i, 0)
			return &t
		}
	}

	return nil
}

// rawMapResponse converts a map to a RawProviderResponse (JSON bytes).
func rawMapResponse(m map[string]interface{}) domain.RawProviderResponse {
	data, err := json.Marshal(m)
	if err != nil {
		// If marshaling fails, return empty response
		return domain.RawProviderResponse{}
	}
	return domain.RawProviderResponse(data)
}

// getBool safely extracts a bool value from a map.
// Returns false if the key doesn't exist or the value is not a bool.
func getBool(m map[string]interface{}, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

// isNotFoundError checks if an error indicates a "not found" condition.
// This is specific to Razorpay's error response pattern.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found")
}
