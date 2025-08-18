package validator

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	v, err := NewValidator("../../schemas")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	if v == nil {
		t.Fatal("Validator should not be nil")
	}
}

func TestValidateEnvelope(t *testing.T) {
	v, err := NewValidator("../../schemas")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test with simple valid envelope
	valid, errors := v.ValidateEnvelope("../../test-simple.json")
	if !valid {
		t.Errorf("Expected valid envelope, got errors: %v", errors)
	}
}
