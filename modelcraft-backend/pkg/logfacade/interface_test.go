package logfacade

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

// TestStack tests the Stack field creation function
func TestStack(t *testing.T) {
	t.Run("WithStackTraceError", func(t *testing.T) {
		// Create an error with stack trace
		originalErr := pkgerrors.New("original error")
		wrappedErr := pkgerrors.Wrap(originalErr, "wrapped error")

		// Create Stack field
		field := Stack(wrappedErr)

		// Verify field key
		if field.Key != StackFieldKey {
			t.Errorf("Expected field.Key to be '%s', got '%s'", StackFieldKey, field.Key)
		}

		// Verify field value is a string
		stackStr, ok := field.Value.(string)
		if !ok {
			t.Errorf("Expected field.Value to be string, got %T", field.Value)
		}

		// Verify stack trace contains expected information
		if !strings.Contains(stackStr, "interface_test.go") {
			t.Errorf("Expected stack trace to contain 'interface_test.go', got:\n%s", stackStr)
		}

		// Verify stack trace contains error message
		if !strings.Contains(stackStr, "wrapped error") {
			t.Errorf("Expected stack trace to contain error message, got:\n%s", stackStr)
		}
	})

	t.Run("WithoutStackTraceError", func(t *testing.T) {
		// Create an error without stack trace
		err := fmt.Errorf("plain error")

		// Create Stack field
		field := Stack(err)

		// Verify field key
		if field.Key != StackFieldKey {
			t.Errorf("Expected field.Key to be '%s', got '%s'", StackFieldKey, field.Key)
		}

		// Verify field value is the error itself
		if field.Value != err {
			t.Errorf("Expected field.Value to be the error itself")
		}
	})

	t.Run("NilError", func(t *testing.T) {
		// Test with nil error
		field := Stack(nil)

		// Verify field key
		if field.Key != StackFieldKey {
			t.Errorf("Expected field.Key to be '%s', got '%s'", StackFieldKey, field.Key)
		}

		// Verify field value is nil
		if field.Value != nil {
			t.Errorf("Expected field.Value to be nil, got %v", field.Value)
		}
	})
}

// TestStackWithLogger tests Stack field with actual logger
func TestStackWithLogger(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		OutputPath: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create an error with stack trace
	originalErr := pkgerrors.New("database connection failed")
	wrappedErr := pkgerrors.Wrap(originalErr, "failed to initialize database")

	// Log with Stack field
	logger.With(Stack(wrappedErr)).Errorf(context.Background(), nil, "Operation failed")

	// This test mainly verifies that the logger can handle Stack field without panic
	// Visual inspection of output is needed to verify the stack trace is printed
}
