package logfacade

import (
	"context"
	"fmt"
	"testing"
)

func TestErrFieldWithLogger(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		OutputPath: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	testErr := fmt.Errorf("test error")

	logger.Errorf(context.Background(), testErr, "Operation failed")
}
