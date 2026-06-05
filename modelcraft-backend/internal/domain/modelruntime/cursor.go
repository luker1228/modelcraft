package modelruntime

import (
	"encoding/base64"
	"encoding/json"

	bizerrors "modelcraft/pkg/bizerrors"
)

// CursorData holds the internal cursor state (opaque to API consumers).
type CursorData struct {
	SortField string `json:"sf"`            // user-specified sort field name
	SortValue string `json:"sv"`            // its value on the last returned record
	IOField   string `json:"iof,omitempty"` // insertion-order field name (when configured)
	IOValue   string `json:"iov,omitempty"` // its value on the last returned record
}

// encodeCursor serialises a CursorData to a base64url string.
func encodeCursor(c CursorData) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

// decodeCursor deserialises a base64url string into CursorData.
func decodeCursor(s string) (CursorData, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return CursorData{}, bizerrors.Errorf("invalid cursor: %w", err)
	}
	var c CursorData
	if err := json.Unmarshal(b, &c); err != nil {
		return CursorData{}, bizerrors.Errorf("invalid cursor format: %w", err)
	}
	if c.SortField == "" {
		return CursorData{}, bizerrors.Errorf("invalid cursor: sortField is empty")
	}
	return c, nil
}
