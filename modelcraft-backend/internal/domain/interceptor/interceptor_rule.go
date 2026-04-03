package interceptor

import (
	"modelcraft/pkg/bizerrors"
	"time"
)

// InterceptorRule represents a JavaScript-based interceptor configuration
// that executes before GraphQL operations to modify inputs or enforce rules.
type InterceptorRule struct {
	ID          uint64    `json:"id"`
	ModelID     uint64    `json:"modelId"`
	Operation   string    `json:"operation"` // findFirst, findMany, createOne, etc.
	Script      string    `json:"script"`    // JavaScript code
	Priority    int       `json:"priority"`  // Higher priority executes first
	Enabled     bool      `json:"enabled"`
	TimeoutMs   int       `json:"timeoutMs"` // Timeout in milliseconds
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Validate validates the interceptor rule fields
func (r *InterceptorRule) Validate() error {
	if r.ModelID == 0 {
		return bizerrors.New("modelId is required")
	}
	if r.Operation == "" {
		return bizerrors.New("operation is required")
	}
	if r.Script == "" {
		return bizerrors.New("script is required")
	}
	if r.TimeoutMs <= 0 {
		r.TimeoutMs = 100 // Default timeout: 100ms
	}
	if r.TimeoutMs > 10000 {
		return bizerrors.New("timeoutMs cannot exceed 10000ms")
	}

	// Validate operation name
	validOperations := []string{
		"findFirst", "findMany", "findUnique", "aggregate", "count",
		"createOne", "createMany", "updateOne", "updateMany", "deleteOne", "deleteMany",
	}
	isValid := false
	for _, op := range validOperations {
		if r.Operation == op {
			isValid = true
			break
		}
	}
	if !isValid {
		return bizerrors.Errorf("invalid operation: %s", r.Operation)
	}

	return nil
}

// NewInterceptorRule creates a new interceptor rule with validation
func NewInterceptorRule(modelID uint64, operation, script string, priority int) (*InterceptorRule, error) {
	rule := &InterceptorRule{
		ModelID:   modelID,
		Operation: operation,
		Script:    script,
		Priority:  priority,
		Enabled:   true,
		TimeoutMs: 100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := rule.Validate(); err != nil {
		return nil, err
	}

	return rule, nil
}
