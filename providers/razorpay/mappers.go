package razorpay

import (
	"strconv"
	"time"
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

// getFloat64 safely extracts a float64 value from a map.
// Returns 0 if the key doesn't exist or the value cannot be converted to float64.
func getFloat64(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	val, ok := m[key]
	if !ok {
		return 0
	}

	// Handle float64 directly
	if f, ok := val.(float64); ok {
		return f
	}

	// Handle int64
	if i, ok := val.(int64); ok {
		return float64(i)
	}

	// Handle string representation
	if s, ok := val.(string); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}

	return 0
}

// getBool safely extracts a bool value from a map.
// Returns false if the key doesn't exist or the value is not a bool.
func getBool(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	val, ok := m[key]
	if !ok {
		return false
	}
	b, ok := val.(bool)
	if !ok {
		return false
	}
	return b
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
// Returns zero time if the key doesn't exist or parsing fails.
func getTime(m map[string]interface{}, key string) time.Time {
	if m == nil {
		return time.Time{}
	}
	val, ok := m[key]
	if !ok {
		return time.Time{}
	}

	// Try as Unix timestamp (float64 or int64)
	if f, ok := val.(float64); ok {
		return time.Unix(int64(f), 0)
	}
	if i, ok := val.(int64); ok {
		return time.Unix(i, 0)
	}

	// Try as string (RFC3339 or other format)
	if s, ok := val.(string); ok {
		// Try RFC3339 format
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
		// Try Unix timestamp as string
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return time.Unix(i, 0)
		}
	}

	return time.Time{}
}
