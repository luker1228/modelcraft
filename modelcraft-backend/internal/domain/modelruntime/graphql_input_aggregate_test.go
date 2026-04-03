package modelruntime

import (
	"testing"
)

// TestNewAggregateInput tests newAggregateInput function
func TestNewAggregateInput(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "valid input with count aggregate",
			args: map[string]any{
				"where": map[string]any{
					"age": map[string]any{
						"gte": 18,
					},
				},
				"_count": map[string]any{
					"_all": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input with multiple aggregates",
			args: map[string]any{
				"_count": map[string]any{
					"id": true,
				},
				"_avg": map[string]any{
					"age": true,
				},
				"_sum": map[string]any{
					"salary": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid input with all aggregate types",
			args: map[string]any{
				"_count": map[string]any{
					"_all": true,
				},
				"_avg": map[string]any{
					"age": true,
				},
				"_sum": map[string]any{
					"salary": true,
				},
				"_min": map[string]any{
					"age": true,
				},
				"_max": map[string]any{
					"age": true,
				},
			},
			wantErr: false,
		},
		{
			name:    "missing aggregate operations",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "invalid where type",
			args: map[string]any{
				"where": "invalid",
				"_count": map[string]any{
					"_all": true,
				},
			},
			wantErr: true,
		},
		{
			name: "valid input without where",
			args: map[string]any{
				"_count": map[string]any{
					"_all": true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newAggregateInput("users", params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAggregateInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != "users" {
					t.Errorf("newAggregateInput() TableName = %v, want users", got.TableName)
				}
			}
		})
	}
}

// TestParseAggregateField tests parseAggregateField function
func TestParseAggregateField(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		key        string
		wantFields map[string]bool
	}{
		{
			name: "valid count aggregate",
			args: map[string]any{
				"_count": map[string]any{
					"_all": true,
					"id":   true,
					"name": false,
				},
			},
			key: "_count",
			wantFields: map[string]bool{
				"_all": true,
				"id":   true,
			},
		},
		{
			name: "missing key",
			args: map[string]any{
				"_avg": map[string]any{
					"age": true,
				},
			},
			key:        "_count",
			wantFields: map[string]bool{},
		},
		{
			name: "invalid aggregate type",
			args: map[string]any{
				"_count": "invalid",
			},
			key:        "_count",
			wantFields: map[string]bool{},
		},
		{
			name: "empty aggregate map",
			args: map[string]any{
				"_count": map[string]any{},
			},
			key:        "_count",
			wantFields: map[string]bool{},
		},
		{
			name: "mixed boolean values",
			args: map[string]any{
				"_avg": map[string]any{
					"age":    true,
					"salary": false,
					"score":  true,
				},
			},
			key: "_avg",
			wantFields: map[string]bool{
				"age":   true,
				"score": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := make(map[string]bool)
			parseAggregateField(tt.args, tt.key, target)

			if len(target) != len(tt.wantFields) {
				t.Errorf("parseAggregateField() got %d fields, want %d", len(target), len(tt.wantFields))
			}

			for field, want := range tt.wantFields {
				if got, exists := target[field]; !exists || got != want {
					t.Errorf("parseAggregateField() field %s = %v, want %v", field, got, want)
				}
			}
		})
	}
}

// TestHasAnyAggregate tests hasAnyAggregate function
func TestHasAnyAggregate(t *testing.T) {
	tests := []struct {
		name  string
		input *AggregateInput
		want  bool
	}{
		{
			name: "has count aggregate",
			input: &AggregateInput{
				Count: map[string]bool{"_all": true},
				Avg:   map[string]bool{},
				Sum:   map[string]bool{},
				Min:   map[string]bool{},
				Max:   map[string]bool{},
			},
			want: true,
		},
		{
			name: "has avg aggregate",
			input: &AggregateInput{
				Count: map[string]bool{},
				Avg:   map[string]bool{"age": true},
				Sum:   map[string]bool{},
				Min:   map[string]bool{},
				Max:   map[string]bool{},
			},
			want: true,
		},
		{
			name: "has sum aggregate",
			input: &AggregateInput{
				Count: map[string]bool{},
				Avg:   map[string]bool{},
				Sum:   map[string]bool{"salary": true},
				Min:   map[string]bool{},
				Max:   map[string]bool{},
			},
			want: true,
		},
		{
			name: "has min aggregate",
			input: &AggregateInput{
				Count: map[string]bool{},
				Avg:   map[string]bool{},
				Sum:   map[string]bool{},
				Min:   map[string]bool{"age": true},
				Max:   map[string]bool{},
			},
			want: true,
		},
		{
			name: "has max aggregate",
			input: &AggregateInput{
				Count: map[string]bool{},
				Avg:   map[string]bool{},
				Sum:   map[string]bool{},
				Min:   map[string]bool{},
				Max:   map[string]bool{"age": true},
			},
			want: true,
		},
		{
			name: "no aggregates",
			input: &AggregateInput{
				Count: map[string]bool{},
				Avg:   map[string]bool{},
				Sum:   map[string]bool{},
				Min:   map[string]bool{},
				Max:   map[string]bool{},
			},
			want: false,
		},
		{
			name: "multiple aggregates",
			input: &AggregateInput{
				Count: map[string]bool{"_all": true},
				Avg:   map[string]bool{"age": true},
				Sum:   map[string]bool{"salary": true},
				Min:   map[string]bool{},
				Max:   map[string]bool{},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAnyAggregate(tt.input)
			if got != tt.want {
				t.Errorf("hasAnyAggregate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseSelectArg tests parseSelectArg function
func TestParseSelectArg(t *testing.T) {
	tests := []struct {
		name      string
		selectArg any
		want      map[string]bool
		wantErr   bool
	}{
		{
			name: "valid select with true values",
			selectArg: map[string]any{
				"id":   true,
				"name": true,
			},
			want: map[string]bool{
				"id":   true,
				"name": true,
			},
			wantErr: false,
		},
		{
			name: "select with mixed boolean values",
			selectArg: map[string]any{
				"id":   true,
				"name": false,
				"age":  true,
			},
			want: map[string]bool{
				"id":  true,
				"age": true,
			},
			wantErr: false,
		},
		{
			name:      "empty select map",
			selectArg: map[string]any{},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "select with all false values",
			selectArg: map[string]any{
				"id":   false,
				"name": false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:      "invalid select type",
			selectArg: "invalid",
			want:      map[string]bool{},
			wantErr:   false,
		},
		{
			name:      "nil select arg",
			selectArg: nil,
			want:      map[string]bool{},
			wantErr:   false,
		},
		{
			name: "select with non-boolean values",
			selectArg: map[string]any{
				"id":   "yes",
				"name": 1,
				"age":  true,
			},
			want: map[string]bool{
				"age": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSelectArg(tt.selectArg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSelectArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseSelectArg() got %d fields, want %d", len(got), len(tt.want))
				}
				for field, wantVal := range tt.want {
					if gotVal, exists := got[field]; !exists || gotVal != wantVal {
						t.Errorf("parseSelectArg() field %s = %v, want %v", field, gotVal, wantVal)
					}
				}
			}
		})
	}
}

// TestNewCountInput tests newCountInput function
func TestNewCountInput(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "valid count with where",
			args: map[string]any{
				"where": map[string]any{
					"age": map[string]any{
						"gte": 18,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid count with select",
			args: map[string]any{
				"select": map[string]any{
					"_all": true,
					"id":   true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid count with where and select",
			args: map[string]any{
				"where": map[string]any{
					"status": "active",
				},
				"select": map[string]any{
					"id": true,
				},
			},
			wantErr: false,
		},
		{
			name:    "valid count without parameters",
			args:    map[string]any{},
			wantErr: false,
		},
		{
			name: "invalid where type",
			args: map[string]any{
				"where": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid select - empty map",
			args: map[string]any{
				"select": map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "invalid select - all false",
			args: map[string]any{
				"select": map[string]any{
					"id":   false,
					"name": false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newCountInput("users", params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCountInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != "users" {
					t.Errorf("newCountInput() TableName = %v, want users", got.TableName)
				}
			}
		})
	}
}

// TestAggregateInputWithComplexWhere tests aggregate input with complex where conditions
func TestAggregateInputWithComplexWhere(t *testing.T) {
	args := map[string]any{
		"where": map[string]any{
			"AND": []map[string]any{
				{"age": map[string]any{"gte": 18}},
				{"status": "active"},
			},
		},
		"_count": map[string]any{
			"_all": true,
		},
	}

	params := createMockResolveParamsNoSelection(args)
	got, err := newAggregateInput("users", params)
	if err != nil {
		t.Fatalf("newAggregateInput() unexpected error = %v", err)
	}

	if got.TableName != "users" {
		t.Errorf("newAggregateInput() TableName = %v, want users", got.TableName)
	}

	if len(got.Count) == 0 {
		t.Error("newAggregateInput() Count is empty, want at least one field")
	}

	if !got.Count["_all"] {
		t.Error("newAggregateInput() Count[_all] = false, want true")
	}
}

// BenchmarkNewAggregateInput benchmarks aggregate input creation
func BenchmarkNewAggregateInput(b *testing.B) {
	args := map[string]any{
		"where": map[string]any{
			"age": map[string]any{
				"gte": 18,
			},
		},
		"_count": map[string]any{
			"_all": true,
			"id":   true,
		},
		"_avg": map[string]any{
			"age": true,
		},
	}

	params := createMockResolveParamsNoSelection(args)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = newAggregateInput("users", params)
	}
}

// BenchmarkParseSelectArg benchmarks select argument parsing
func BenchmarkParseSelectArg(b *testing.B) {
	selectArg := map[string]any{
		"id":     true,
		"name":   true,
		"age":    true,
		"status": true,
		"email":  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseSelectArg(selectArg)
	}
}
