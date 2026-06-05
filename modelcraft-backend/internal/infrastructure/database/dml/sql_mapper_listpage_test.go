package dml

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"strings"
	"testing"
)

func TestConvertListPageInputToSQL_FirstPage_NoAfter(t *testing.T) {
	input := &modelruntime.ListPageInput{
		TableName:     "products",
		SortField:     "price",
		SortDirection: "asc",
		Limit:         10,
	}
	sql, args, err := convertListPageInputToSQL(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	// Should contain ORDER BY price ASC
	if !strings.Contains(strings.ToUpper(sql), "ORDER BY") {
		t.Errorf("expected ORDER BY in SQL: %s", sql)
	}
	// Prepared statements put the LIMIT value in args, not inline.
	// Expect LIMIT ? with value 11 (limit+1 for hasNextPage detection).
	if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
		t.Errorf("expected LIMIT clause in SQL: %s", sql)
	}
	found := false
	for _, a := range args {
		if fmt.Sprintf("%v", a) == "11" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 11 in args (limit+1), got: %v | SQL: %s", args, sql)
	}
	t.Logf("SQL: %s | args: %v", sql, args)
}

func TestConvertListPageInputToSQL_WithAfter_DualField(t *testing.T) {
	after := &modelruntime.CursorData{
		SortField: "price", SortValue: "100",
		IOField: "created_at", IOValue: "2026-06-05T10:00:00Z",
	}
	input := &modelruntime.ListPageInput{
		TableName:           "products",
		SortField:           "price",
		SortDirection:       "asc",
		InsertionOrderField: "created_at",
		After:               after,
		Limit:               10,
	}
	sql, args, err := convertListPageInputToSQL(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have a WHERE clause with OR conditions for dual-field cursor
	upperSQL := strings.ToUpper(sql)
	if !strings.Contains(upperSQL, "WHERE") {
		t.Errorf("expected WHERE in SQL: %s", sql)
	}
	if !strings.Contains(upperSQL, "OR") {
		t.Errorf("expected OR in SQL for dual-field cursor: %s", sql)
	}
	t.Logf("SQL: %s | args: %v", sql, args)
}

func TestConvertListPageInputToSQL_WithAfter_SingleField(t *testing.T) {
	after := &modelruntime.CursorData{
		SortField: "id", SortValue: "abc123",
	}
	input := &modelruntime.ListPageInput{
		TableName:     "products",
		SortField:     "id",
		SortDirection: "asc",
		After:         after,
		Limit:         5,
	}
	sql, args, err := convertListPageInputToSQL(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	upperSQL := strings.ToUpper(sql)
	if !strings.Contains(upperSQL, "WHERE") {
		t.Errorf("expected WHERE in SQL: %s", sql)
	}
	// Single field = no OR, just a simple GT condition
	if strings.Contains(upperSQL, " OR ") {
		t.Errorf("single-field cursor should NOT use OR: %s", sql)
	}
	t.Logf("SQL: %s | args: %v", sql, args)
}
