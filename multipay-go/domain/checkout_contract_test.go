package domain

import (
	"encoding/json"
	"os"
	"path"
	"runtime"
	"testing"
)

// TestCheckoutPayloadRoundTrip verifies that CheckoutPayload round-trips correctly
// through JSON serialization/deserialization for all provider vectors.
func TestCheckoutPayloadRoundTrip(t *testing.T) {
	// Get the directory of this test file
	_, filename, _, _ := runtime.Caller(0)
	vectorsDir := path.Join(path.Dir(filename), "..", "..", "contract", "checkout", "vectors")

	vectorFiles := []struct {
		name     string
		filename string
	}{
		{"cashfree checkout vector", "cashfree.checkout.json"},
		{"razorpay checkout vector", "razorpay.checkout.json"},
	}

	for _, tt := range vectorFiles {
		t.Run(tt.name, func(t *testing.T) {
			vectorPath := path.Join(vectorsDir, tt.filename)
			vectorData, err := os.ReadFile(vectorPath)
			if err != nil {
				t.Fatalf("failed to read vector file %s: %v", vectorPath, err)
			}

			// First unmarshal into the CheckoutPayload
			var original CheckoutPayload
			if unmarshalErr := json.Unmarshal(vectorData, &original); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal vector data: %v\nVector content: %s", unmarshalErr, string(vectorData))
			}

			// Marshal it back to JSON
			jsonData, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("failed to marshal CheckoutPayload: %v", err)
			}

			// Unmarshal again to verify round-trip
			var roundTripped CheckoutPayload
			if err := json.Unmarshal(jsonData, &roundTripped); err != nil {
				t.Fatalf("failed to unmarshal round-tripped data: %v", err)
			}

			// Verify all fields match
			if roundTripped.Provider != original.Provider {
				t.Errorf("Provider mismatch: got %q, want %q", roundTripped.Provider, original.Provider)
			}
			if roundTripped.Environment != original.Environment {
				t.Errorf("Environment mismatch: got %q, want %q", roundTripped.Environment, original.Environment)
			}
			if roundTripped.SessionID != original.SessionID {
				t.Errorf("SessionID mismatch: got %q, want %q", roundTripped.SessionID, original.SessionID)
			}
			if roundTripped.OrderID != original.OrderID {
				t.Errorf("OrderID mismatch: got %q, want %q", roundTripped.OrderID, original.OrderID)
			}
			if roundTripped.PublicKey != original.PublicKey {
				t.Errorf("PublicKey mismatch: got %q, want %q", roundTripped.PublicKey, original.PublicKey)
			}
			if roundTripped.CallbackURL != original.CallbackURL {
				t.Errorf("CallbackURL mismatch: got %q, want %q", roundTripped.CallbackURL, original.CallbackURL)
			}
			if roundTripped.AmountMinor != original.AmountMinor {
				t.Errorf("AmountMinor mismatch: got %d, want %d", roundTripped.AmountMinor, original.AmountMinor)
			}
			if roundTripped.Currency != original.Currency {
				t.Errorf("Currency mismatch: got %q, want %q", roundTripped.Currency, original.Currency)
			}
		})
	}
}

// TestCheckoutPayloadAmountMinorFieldExists verifies that the amount_minor field
// is present in the CheckoutPayload struct and can be serialized/deserialized.
func TestCheckoutPayloadAmountMinorFieldExists(t *testing.T) {
	// Get the directory of this test file
	_, filename, _, _ := runtime.Caller(0)
	vectorsDir := path.Join(path.Dir(filename), "..", "..", "contract", "checkout", "vectors")

	// Test with razorpay vector which has amount_minor
	vectorPath := path.Join(vectorsDir, "razorpay.checkout.json")
	vectorData, err := os.ReadFile(vectorPath)
	if err != nil {
		t.Fatalf("failed to read razorpay vector: %v", err)
	}

	var payload CheckoutPayload
	if unmarshalErr := json.Unmarshal(vectorData, &payload); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal razorpay vector: %v", unmarshalErr)
	}

	// Verify that amount_minor is correctly deserialized
	// The razorpay vector has "amount_minor": 50000
	if payload.AmountMinor != 50000 {
		t.Errorf("AmountMinor field not correctly deserialized: got %d, want 50000", payload.AmountMinor)
	}

	// Verify it marshals back to JSON correctly
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	var unmarshaled CheckoutPayload
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal marshaled data: %v", err)
	}

	if unmarshaled.AmountMinor != 50000 {
		t.Errorf("AmountMinor field not preserved after round-trip: got %d, want 50000", unmarshaled.AmountMinor)
	}
}

// TestCheckoutPayloadFieldPresence verifies that provider-specific fields
// are correctly populated based on the provider type.
func TestCheckoutPayloadFieldPresence(t *testing.T) {
	// Get the directory of this test file
	_, filename, _, _ := runtime.Caller(0)
	vectorsDir := path.Join(path.Dir(filename), "..", "..", "contract", "checkout", "vectors")

	t.Run("cashfree session_id field", func(t *testing.T) {
		vectorPath := path.Join(vectorsDir, "cashfree.checkout.json")
		vectorData, err := os.ReadFile(vectorPath)
		if err != nil {
			t.Fatalf("failed to read cashfree vector: %v", err)
		}

		var payload CheckoutPayload
		if err := json.Unmarshal(vectorData, &payload); err != nil {
			t.Fatalf("failed to unmarshal cashfree vector: %v", err)
		}

		// Cashfree should have session_id populated
		if payload.SessionID != "session_abc123" {
			t.Errorf("Cashfree SessionID incorrect: got %q, want %q", payload.SessionID, "session_abc123")
		}

		// These Razorpay-specific fields should be empty for Cashfree
		if payload.OrderID != "" {
			t.Errorf("Cashfree payload should not have OrderID, got %q", payload.OrderID)
		}
		if payload.PublicKey != "" {
			t.Errorf("Cashfree payload should not have PublicKey, got %q", payload.PublicKey)
		}
	})

	t.Run("razorpay fields", func(t *testing.T) {
		vectorPath := path.Join(vectorsDir, "razorpay.checkout.json")
		vectorData, err := os.ReadFile(vectorPath)
		if err != nil {
			t.Fatalf("failed to read razorpay vector: %v", err)
		}

		var payload CheckoutPayload
		if err := json.Unmarshal(vectorData, &payload); err != nil {
			t.Fatalf("failed to unmarshal razorpay vector: %v", err)
		}

		// Razorpay should have these fields populated
		if payload.OrderID != "order_RZP123" {
			t.Errorf("Razorpay OrderID incorrect: got %q, want %q", payload.OrderID, "order_RZP123")
		}
		if payload.PublicKey != "rzp_live_xxx" {
			t.Errorf("Razorpay PublicKey incorrect: got %q, want %q", payload.PublicKey, "rzp_live_xxx")
		}
		if payload.CallbackURL != "https://api.smriti.ai/v1/payments/callback/razorpay" {
			t.Errorf("Razorpay CallbackURL incorrect: got %q, want %q", payload.CallbackURL, "https://api.smriti.ai/v1/payments/callback/razorpay")
		}
		if payload.AmountMinor != 50000 {
			t.Errorf("Razorpay AmountMinor incorrect: got %d, want 50000", payload.AmountMinor)
		}
		if payload.Currency != "INR" {
			t.Errorf("Razorpay Currency incorrect: got %q, want INR", payload.Currency)
		}

		// Cashfree-specific field should be empty for Razorpay
		if payload.SessionID != "" {
			t.Errorf("Razorpay payload should not have SessionID, got %q", payload.SessionID)
		}
	})
}
