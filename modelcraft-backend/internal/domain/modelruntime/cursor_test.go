package modelruntime

import (
	"testing"
)

func TestEncodeCursor_SingleField(t *testing.T) {
	c := CursorData{SortField: "price", SortValue: "100", IOField: "", IOValue: ""}
	encoded := encodeCursor(c)
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("decodeCursor error: %v", err)
	}
	if decoded.SortField != "price" || decoded.SortValue != "100" {
		t.Errorf("got %+v, want SortField=price SortValue=100", decoded)
	}
}

func TestEncodeCursor_DualField(t *testing.T) {
	c := CursorData{SortField: "price", SortValue: "100", IOField: "created_at", IOValue: "2026-06-05T10:00:00Z"}
	encoded := encodeCursor(c)
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("decodeCursor error: %v", err)
	}
	if decoded.IOField != "created_at" || decoded.IOValue != "2026-06-05T10:00:00Z" {
		t.Errorf("got %+v", decoded)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	_, err := decodeCursor("not-valid-base64!!")
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}
