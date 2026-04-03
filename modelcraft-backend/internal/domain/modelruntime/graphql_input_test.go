package modelruntime

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// 辅助函数：创建模拟的 ResolveParams
func createMockResolveParams(args map[string]any, selectedFields []string) graphql.ResolveParams {
	// 创建 SelectionSet
	selections := make([]ast.Selection, 0, len(selectedFields))
	for _, fieldName := range selectedFields {
		selections = append(selections, &ast.Field{
			Name: &ast.Name{
				Value: fieldName,
			},
		})
	}

	return graphql.ResolveParams{
		Args: args,
		Info: graphql.ResolveInfo{
			FieldASTs: []*ast.Field{
				{
					SelectionSet: &ast.SelectionSet{
						Selections: selections,
					},
				},
			},
		},
	}
}

// 辅助函数：创建不带 SelectionSet 的 ResolveParams
func createMockResolveParamsNoSelection(args map[string]any) graphql.ResolveParams {
	return graphql.ResolveParams{
		Args: args,
		Info: graphql.ResolveInfo{
			FieldASTs: []*ast.Field{
				{
					SelectionSet: nil,
				},
			},
		},
	}
}

// TestGetWhere 测试 getWhere 函数
func TestGetWhere(t *testing.T) {
	tests := []struct {
		name      string
		param     map[string]any
		wantWhere map[string]any
		wantErr   bool
	}{
		{
			name: "valid where map",
			param: map[string]any{
				"where": map[string]any{
					"id": "123",
				},
			},
			wantWhere: map[string]any{
				"id": "123",
			},
			wantErr: false,
		},
		{
			name:      "no where parameter",
			param:     map[string]any{},
			wantWhere: nil,
			wantErr:   false,
		},
		{
			name: "invalid where type",
			param: map[string]any{
				"where": "invalid",
			},
			wantWhere: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getWhere(tt.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("getWhere() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.wantWhere) {
				t.Errorf("getWhere() = %v, want %v", got, tt.wantWhere)
			}
		})
	}
}

// TestNewFindUniqueInput 测试 newFindUniqueInput 函数
func TestNewFindUniqueInput(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		args      map[string]any
		wantErr   bool
	}{
		{
			name:      "valid input",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"id": "123",
				},
			},
			wantErr: false,
		},
		{
			name:      "no where parameter",
			tableName: "users",
			args:      map[string]any{},
			wantErr:   false,
		},
		{
			name:      "invalid where type",
			tableName: "users",
			args: map[string]any{
				"where": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newFindUniqueInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFindUniqueInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newFindUniqueInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
			}
		})
	}
}

// TestNewFindManyInput 测试 newFindManyInput 函数
func TestNewFindManyInput(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		args      map[string]any
		wantErr   bool
	}{
		{
			name:      "valid input with where",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"age": 25,
				},
			},
			wantErr: false,
		},
		{
			name:      "valid input without where",
			tableName: "users",
			args:      map[string]any{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newFindManyInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFindManyInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newFindManyInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
			}
		})
	}
}

// TestNewFindFirstInput 测试 newFindFirstInput 函数
func TestNewFindFirstInput(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		args      map[string]any
		wantErr   bool
	}{
		{
			name:      "valid input",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"name": "John",
				},
			},
			wantErr: false,
		},
		{
			name:      "no where parameter",
			tableName: "users",
			args:      map[string]any{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newFindFirstInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFindFirstInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newFindFirstInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
			}
		})
	}
}

// TestNewCreateOneInput 测试 newCreateOneInput 函数
func TestNewCreateOneInput(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		args           map[string]any
		selectedFields []string
		wantCreatedObj bool
		wantErr        bool
	}{
		{
			name:      "valid input with createdObj selected",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			selectedFields: []string{"createdObj"},
			wantCreatedObj: true,
			wantErr:        false,
		},
		{
			name:      "valid input without createdObj selected",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"name": "Jane",
				},
			},
			selectedFields: []string{"success"},
			wantCreatedObj: false,
			wantErr:        false,
		},
		{
			name:      "missing data parameter",
			tableName: "users",
			args:      map[string]any{},
			wantErr:   true,
		},
		{
			name:      "invalid data type",
			tableName: "users",
			args: map[string]any{
				"data": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParams(tt.args, tt.selectedFields)
			got, err := newCreateOneInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCreateOneInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newCreateOneInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
				if got.CreatedObj != tt.wantCreatedObj {
					t.Errorf("newCreateOneInput() CreatedObj = %v, want %v", got.CreatedObj, tt.wantCreatedObj)
				}
			}
		})
	}
}

// TestNewUpdateOneInput 测试 newUpdateOneInput 函数
func TestNewUpdateOneInput(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		args           map[string]any
		selectedFields []string
		wantUpdatedObj bool
		wantErr        bool
	}{
		{
			name:      "valid input with updatedObj selected",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"id": "123",
				},
				"data": map[string]any{
					"name": "Updated Name",
				},
			},
			selectedFields: []string{"updatedObj"},
			wantUpdatedObj: true,
			wantErr:        false,
		},
		{
			name:      "valid input without updatedObj selected",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"id": "123",
				},
				"data": map[string]any{
					"age": 31,
				},
			},
			selectedFields: []string{"success"},
			wantUpdatedObj: false,
			wantErr:        false,
		},
		{
			name:      "missing data parameter",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"id": "123",
				},
			},
			wantErr: true,
		},
		{
			name:      "invalid where type",
			tableName: "users",
			args: map[string]any{
				"where": "invalid",
				"data": map[string]any{
					"name": "Test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParams(tt.args, tt.selectedFields)
			got, err := newUpdateOneInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newUpdateOneInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newUpdateOneInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
				if got.UpdatedObj != tt.wantUpdatedObj {
					t.Errorf("newUpdateOneInput() UpdatedObj = %v, want %v", got.UpdatedObj, tt.wantUpdatedObj)
				}
			}
		})
	}
}

// TestNewDeleteOneInput 测试 newDeleteOneInput 函数
func TestNewDeleteOneInput(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		args           map[string]any
		selectedFields []string
		wantDeletedObj bool
		wantErr        bool
	}{
		{
			name:      "valid input with deletedObj selected",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"id": "123",
				},
			},
			selectedFields: []string{"deletedObj"},
			wantDeletedObj: true,
			wantErr:        false,
		},
		{
			name:      "valid input without deletedObj selected",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"id": "456",
				},
			},
			selectedFields: []string{"success"},
			wantDeletedObj: false,
			wantErr:        false,
		},
		{
			name:      "invalid where type",
			tableName: "users",
			args: map[string]any{
				"where": []string{"invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParams(tt.args, tt.selectedFields)
			got, err := newDeleteOneInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newDeleteOneInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newDeleteOneInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
				if got.DeletedObj != tt.wantDeletedObj {
					t.Errorf("newDeleteOneInput() DeletedObj = %v, want %v", got.DeletedObj, tt.wantDeletedObj)
				}
			}
		})
	}
}

// TestNewCreateManyInput 测试 newCreateManyInput 函数
func TestNewCreateManyInput(t *testing.T) {
	tests := []struct {
		name               string
		tableName          string
		args               map[string]any
		selectedFields     []string
		wantSkipDuplicates bool
		wantReturnIdList   bool
		wantErr            bool
	}{
		{
			name:      "valid input with idList selected",
			tableName: "users",
			args: map[string]any{
				"data": []any{
					map[string]any{"name": "User1"},
					map[string]any{"name": "User2"},
				},
				"skipDuplicates": true,
			},
			selectedFields:     []string{"idList"},
			wantSkipDuplicates: true,
			wantReturnIdList:   true,
			wantErr:            false,
		},
		{
			name:      "valid input without idList selected",
			tableName: "users",
			args: map[string]any{
				"data": []any{
					map[string]any{"name": "User3"},
				},
			},
			selectedFields:     []string{"count"},
			wantSkipDuplicates: false,
			wantReturnIdList:   false,
			wantErr:            false,
		},
		{
			name:      "empty data array",
			tableName: "users",
			args: map[string]any{
				"data": []any{},
			},
			wantErr: true,
		},
		{
			name:      "missing data parameter",
			tableName: "users",
			args:      map[string]any{},
			wantErr:   true,
		},
		{
			name:      "invalid data type",
			tableName: "users",
			args: map[string]any{
				"data": "invalid",
			},
			wantErr: true,
		},
		{
			name:      "exceeds batch size limit",
			tableName: "users",
			args: map[string]any{
				"data": func() []any {
					data := make([]any, MaxCreateManyBatchSize+1)
					for i := range data {
						data[i] = map[string]any{"name": "User"}
					}
					return data
				}(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParams(tt.args, tt.selectedFields)
			got, err := newCreateManyInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCreateManyInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newCreateManyInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
				if got.SkipDuplicates != tt.wantSkipDuplicates {
					t.Errorf(
						"newCreateManyInput() SkipDuplicates = %v, want %v",
						got.SkipDuplicates,
						tt.wantSkipDuplicates,
					)
				}
				if got.ReturnIdList != tt.wantReturnIdList {
					t.Errorf("newCreateManyInput() ReturnIdList = %v, want %v", got.ReturnIdList, tt.wantReturnIdList)
				}
			}
		})
	}
}

// TestNewUpdateManyInput 测试 newUpdateManyInput 函数
func TestNewUpdateManyInput(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		args      map[string]any
		wantTake  uint
		wantErr   bool
	}{
		{
			name:      "valid input with where",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"age": 25,
				},
				"data": map[string]any{
					"status": "active",
				},
				"take": 10,
			},
			wantTake: 10,
			wantErr:  false,
		},
		{
			name:      "valid input without where",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"status": "inactive",
				},
				"take": 5,
			},
			wantTake: 5,
			wantErr:  false,
		},
		{
			name:      "missing take parameter",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"status": "active",
				},
			},
			wantErr: true,
		},
		{
			name:      "invalid take type",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"status": "active",
				},
				"take": "invalid",
			},
			wantErr: true,
		},
		{
			name:      "take below minimum",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"status": "active",
				},
				"take": 0,
			},
			wantErr: true,
		},
		{
			name:      "take exceeds maximum",
			tableName: "users",
			args: map[string]any{
				"data": map[string]any{
					"status": "active",
				},
				"take": MaxCreateManyBatchSize + 1,
			},
			wantErr: true,
		},
		{
			name:      "missing data parameter",
			tableName: "users",
			args: map[string]any{
				"take": 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newUpdateManyInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newUpdateManyInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newUpdateManyInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
				if got.Take != tt.wantTake {
					t.Errorf("newUpdateManyInput() Take = %v, want %v", got.Take, tt.wantTake)
				}
			}
		})
	}
}

// TestNewDeleteManyInput 测试 newDeleteManyInput 函数
func TestNewDeleteManyInput(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		args      map[string]any
		wantTake  uint
		wantErr   bool
	}{
		{
			name:      "valid input with where",
			tableName: "users",
			args: map[string]any{
				"where": map[string]any{
					"status": "inactive",
				},
				"take": 20,
			},
			wantTake: 20,
			wantErr:  false,
		},
		{
			name:      "valid input without where",
			tableName: "users",
			args: map[string]any{
				"take": 15,
			},
			wantTake: 15,
			wantErr:  false,
		},
		{
			name:      "missing take parameter",
			tableName: "users",
			args:      map[string]any{},
			wantErr:   true,
		},
		{
			name:      "invalid take type",
			tableName: "users",
			args: map[string]any{
				"take": "invalid",
			},
			wantErr: true,
		},
		{
			name:      "take below minimum",
			tableName: "users",
			args: map[string]any{
				"take": 0,
			},
			wantErr: true,
		},
		{
			name:      "take exceeds maximum",
			tableName: "users",
			args: map[string]any{
				"take": MaxCreateManyBatchSize + 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParamsNoSelection(tt.args)
			got, err := newDeleteManyInput(tt.tableName, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("newDeleteManyInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.TableName != tt.tableName {
					t.Errorf("newDeleteManyInput() TableName = %v, want %v", got.TableName, tt.tableName)
				}
				if got.Take != tt.wantTake {
					t.Errorf("newDeleteManyInput() Take = %v, want %v", got.Take, tt.wantTake)
				}
			}
		})
	}
}

// TestHasSelectedField 测试 hasSelectedField 函数
func TestHasSelectedField(t *testing.T) {
	tests := []struct {
		name           string
		selectedFields []string
		checkField     string
		want           bool
	}{
		{
			name:           "field is selected",
			selectedFields: []string{"createdObj", "success"},
			checkField:     "createdObj",
			want:           true,
		},
		{
			name:           "field is not selected",
			selectedFields: []string{"success"},
			checkField:     "createdObj",
			want:           false,
		},
		{
			name:           "empty selection",
			selectedFields: []string{},
			checkField:     "createdObj",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := createMockResolveParams(map[string]any{}, tt.selectedFields)
			got := hasSelectedField(params, tt.checkField)
			if got != tt.want {
				t.Errorf("hasSelectedField() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHasSelectedFieldWithNilFieldASTs 测试 hasSelectedField 在 FieldASTs 为 nil 时的行为
func TestHasSelectedFieldWithNilFieldASTs(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]any{},
		Info: graphql.ResolveInfo{
			FieldASTs: nil,
		},
	}

	got := hasSelectedField(params, "anyField")
	if got != false {
		t.Errorf("hasSelectedField() with nil FieldASTs = %v, want false", got)
	}
}

// TestHasSelectedFieldWithNilSelectionSet 测试 hasSelectedField 在 SelectionSet 为 nil 时的行为
func TestHasSelectedFieldWithNilSelectionSet(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]any{},
		Info: graphql.ResolveInfo{
			FieldASTs: []*ast.Field{
				{
					SelectionSet: nil,
				},
			},
		},
	}

	got := hasSelectedField(params, "anyField")
	if got != false {
		t.Errorf("hasSelectedField() with nil SelectionSet = %v, want false", got)
	}
}

// TestHasSelectedFieldWithNonFieldSelection 测试 hasSelectedField 处理非 Field 类型的 Selection
func TestHasSelectedFieldWithNonFieldSelection(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]any{},
		Info: graphql.ResolveInfo{
			FieldASTs: []*ast.Field{
				{
					SelectionSet: &ast.SelectionSet{
						Selections: []ast.Selection{
							// 使用 FragmentSpread 而不是 Field
							&ast.FragmentSpread{
								Name: &ast.Name{
									Value: "someFragment",
								},
							},
						},
					},
				},
			},
		},
	}

	got := hasSelectedField(params, "anyField")
	if got != false {
		t.Errorf("hasSelectedField() with non-Field selection = %v, want false", got)
	}
}

// TestHasSelectedFieldWithNilFieldName 测试 hasSelectedField 处理 Field.Name 为 nil 的情况
func TestHasSelectedFieldWithNilFieldName(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]any{},
		Info: graphql.ResolveInfo{
			FieldASTs: []*ast.Field{
				{
					SelectionSet: &ast.SelectionSet{
						Selections: []ast.Selection{
							&ast.Field{
								Name: nil,
							},
						},
					},
				},
			},
		},
	}

	got := hasSelectedField(params, "anyField")
	if got != false {
		t.Errorf("hasSelectedField() with nil Field.Name = %v, want false", got)
	}
}

// TestCreateManyInputDataConversion 测试 CreateManyInput 的数据转换
func TestCreateManyInputDataConversion(t *testing.T) {
	args := map[string]any{
		"data": []any{
			map[string]any{
				"name": "User1",
				"age":  25,
			},
			map[string]any{
				"name": "User2",
				"age":  30,
			},
		},
	}

	params := createMockResolveParamsNoSelection(args)
	got, err := newCreateManyInput("users", params)
	if err != nil {
		t.Fatalf("newCreateManyInput() unexpected error = %v", err)
	}

	if len(got.Data) != 2 {
		t.Errorf("newCreateManyInput() Data length = %v, want 2", len(got.Data))
	}

	if got.Data[0]["name"] != "User1" {
		t.Errorf("newCreateManyInput() Data[0][name] = %v, want User1", got.Data[0]["name"])
	}
}

// TestUpdateManyInputWithNilWhere 测试 UpdateManyInput 在 where 为 nil 时的行为
func TestUpdateManyInputWithNilWhere(t *testing.T) {
	args := map[string]any{
		"data": map[string]any{
			"status": "active",
		},
		"take": 10,
	}

	params := createMockResolveParamsNoSelection(args)
	got, err := newUpdateManyInput("users", params)
	if err != nil {
		t.Fatalf("newUpdateManyInput() unexpected error = %v", err)
	}

	if got.Where == nil || len(got.Where) != 0 {
		t.Errorf("newUpdateManyInput() Where = %v, want empty map", got.Where)
	}
}

// TestDeleteManyInputWithNilWhere 测试 DeleteManyInput 在 where 为 nil 时的行为
func TestDeleteManyInputWithNilWhere(t *testing.T) {
	args := map[string]any{
		"take": 10,
	}

	params := createMockResolveParamsNoSelection(args)
	got, err := newDeleteManyInput("users", params)
	if err != nil {
		t.Fatalf("newDeleteManyInput() unexpected error = %v", err)
	}

	if got.Where == nil || len(got.Where) != 0 {
		t.Errorf("newDeleteManyInput() Where = %v, want empty map", got.Where)
	}
}

// BenchmarkNewCreateManyInput 性能测试
func BenchmarkNewCreateManyInput(b *testing.B) {
	data := make([]any, 100)
	for i := range data {
		data[i] = map[string]any{
			"name": "User",
			"age":  25,
		}
	}

	args := map[string]any{
		"data": data,
	}

	params := createMockResolveParamsNoSelection(args)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = newCreateManyInput("users", params)
	}
}

// BenchmarkHasSelectedField 性能测试
func BenchmarkHasSelectedField(b *testing.B) {
	params := createMockResolveParams(map[string]any{}, []string{"field1", "field2", "field3", "field4", "field5"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasSelectedField(params, "field3")
	}
}
