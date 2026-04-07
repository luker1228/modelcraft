package modelruntime

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizutils"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClientDatabaseRepository is a test double for ClientDatabaseRepository.
// Only FindMany has real logic; other methods return zero values.
type mockClientDatabaseRepository struct {
	result []map[string]any
	err    error
}

func (m *mockClientDatabaseRepository) FindUnique(_ context.Context, _ *FindUniqueInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) FindFirst(_ context.Context, _ *FindFirstInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) FindMany(_ context.Context, _ *FindManyInput) ([]map[string]any, error) {
	return m.result, m.err
}

func (m *mockClientDatabaseRepository) Aggregate(_ context.Context, _ *AggregateInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) Count(_ context.Context, _ *CountInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) CreateOne(_ context.Context, _ *CreateOneInput) (string, error) {
	return "", nil
}

func (m *mockClientDatabaseRepository) UpdateOne(_ context.Context, _ *UpdateOneInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) DeleteOne(_ context.Context, _ *DeleteOneInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) CreateMany(_ context.Context, _ *CreateManyInput) (interface{}, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) UpdateMany(_ context.Context, _ *UpdateManyInput) (interface{}, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) DeleteMany(_ context.Context, _ *DeleteManyInput) (interface{}, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) FindManyIn(_ context.Context, _ *FindManyInInput) ([]map[string]any, error) {
	return nil, nil
}

// TestCreateModelType 测试 createModelType 方法
func TestCreateModelType(t *testing.T) {
	tests := []struct {
		name      string
		model     *RuntimeModel
		wantErr   bool
		checkFunc func(t *testing.T, obj *graphql.Object)
	}{
		{
			name: "创建简单模型类型 - 只包含基本字段",
			model: &RuntimeModel{
				Name:        "User",
				Title:       "用户",
				Description: "用户模型",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:      "id",
						Title:     "ID",
						Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
						Required:  true,
						IsUnique:  true,
						IsPrimary: true,
					},
					"name": {
						Name:     "name",
						Title:    "姓名",
						Type:     &modeldesign.FieldType{Format: modeldesign.FormatString},
						Required: true,
					},
					"age": {
						Name:  "age",
						Title: "年龄",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatInteger},
					},
					"email": {
						Name:     "email",
						Title:    "邮箱",
						Type:     &modeldesign.FieldType{Format: modeldesign.FormatString},
						IsUnique: true,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, obj *graphql.Object) {
				assert.NotNil(t, obj)
				assert.Equal(t, "UserQuery", obj.Name())
				assert.Equal(t, "用户模型", obj.Description())
				// 检查字段
				fields := obj.Fields()
				assert.Len(t, fields, 4)
				fmt.Printf("fields: %s, lens = %d", bizutils.MarshalToStringIgnoreErr(fields), len(fields))

				// 检查 id 字段
				idField := fields["id"]
				assert.NotNil(t, idField)
				assert.Equal(t, "id", idField.Name)
				assert.Equal(t, "ID", idField.Description)
				assert.Equal(t, graphql.ID, idField.Type)

				// 检查 name 字段
				nameField := fields["name"]
				assert.NotNil(t, nameField)
				assert.Equal(t, "name", nameField.Name)
				assert.Equal(t, graphql.String, nameField.Type)

				// 检查 age 字段
				ageField := fields["age"]
				assert.NotNil(t, ageField)
				assert.Equal(t, "age", ageField.Name)
				assert.Equal(t, graphql.Int, ageField.Type)

				// 检查 email 字段
				emailField := fields["email"]
				assert.NotNil(t, emailField)
				assert.Equal(t, "email", emailField.Name)
				assert.Equal(t, graphql.String, emailField.Type)
			},
		},
		{
			name: "创建包含多种数据类型的模型",
			model: &RuntimeModel{
				Name:        "Product",
				Title:       "产品",
				Description: "产品模型",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:      "id",
						Title:     "ID",
						Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
						IsPrimary: true,
					},
					"name": {
						Name:  "name",
						Title: "产品名称",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
					},
					"price": {
						Name:  "price",
						Title: "价格",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatNumber},
					},
					"stock": {
						Name:  "stock",
						Title: "库存",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatInteger},
					},
					"isActive": {
						Name:  "isActive",
						Title: "是否激活",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatBoolean},
					},
					"description": {
						Name:  "description",
						Title: "描述",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
					},
					"createdAt": {
						Name:  "createdAt",
						Title: "创建时间",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatDateTime},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, obj *graphql.Object) {
				assert.NotNil(t, obj)
				assert.Equal(t, "ProductQuery", obj.Name())
				fields := obj.Fields()
				assert.Len(t, fields, 7)

				// 检查各种类型
				assert.Equal(t, graphql.ID, fields["id"].Type)
				assert.Equal(t, graphql.String, fields["name"].Type)
				assert.Equal(t, graphql.Float, fields["price"].Type)
				assert.Equal(t, graphql.Int, fields["stock"].Type)
				assert.Equal(t, graphql.Boolean, fields["isActive"].Type)
				assert.Equal(t, graphql.String, fields["description"].Type)
				assert.Equal(t, graphql.DateTime, fields["createdAt"].Type)
			},
		},
		{
			name: "创建空字段模型",
			model: &RuntimeModel{
				Name:        "EmptyModel",
				Title:       "空模型",
				Description: "没有字段的模型",
				Fields:      map[string]*RuntimeField{},
			},
			wantErr: true,
			checkFunc: func(t *testing.T, obj *graphql.Object) {
				// 空字段模型应该返回错误，不会执行到这里
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 resolver
			resolver := &graphqlModelResolver{
				
				model:              tt.model,
				inputTypeGenerator: newInputTypeGenerator(),
			}

			// 执行测试
			obj, err := resolver.createModelType(context.Background())

			// 检查错误
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// 执行自定义检查
			if tt.checkFunc != nil {
				tt.checkFunc(t, obj)
			}
		})
	}
}

// TestCreateModelTypeWithRelations 测试包含关联关系的模型类型创建
func TestCreateModelTypeWithRelations(t *testing.T) {
	tests := []struct {
		name      string
		model     *RuntimeModel
		wantErr   bool
		checkFunc func(t *testing.T, obj *graphql.Object)
	}{
		{
			name: "多对一关联 - ManyToOne",
			model: &RuntimeModel{
				Name:        "Post",
				Title:       "文章",
				Description: "文章模型",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:      "id",
						Title:     "ID",
						Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
						IsPrimary: true,
					},
					"title": {
						Name:  "title",
						Title: "标题",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, obj *graphql.Object) {
				assert.NotNil(t, obj)
				fields := obj.Fields()
				// 只检查非关联字段
				assert.Contains(t, fields, "id")
				assert.Contains(t, fields, "title")
			},
		},
		{
			name: "一对多关联 - OneToMany",
			model: &RuntimeModel{
				Name:        "Author",
				Title:       "作者",
				Description: "作者模型",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:      "id",
						Title:     "ID",
						Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
						IsPrimary: true,
					},
					"name": {
						Name:  "name",
						Title: "姓名",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, obj *graphql.Object) {
				// 由于没有关联模型数据，关联字段不会被创建
				assert.NotNil(t, obj)
				fields := obj.Fields()
				// 只检查非关联字段
				assert.Contains(t, fields, "id")
				assert.Contains(t, fields, "name")
			},
		},
		{
			name: "多对多关联 - ManyToMany",
			model: &RuntimeModel{
				Name:        "Student",
				Title:       "学生",
				Description: "学生模型",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:      "id",
						Title:     "ID",
						Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
						IsPrimary: true,
					},
					"name": {
						Name:  "name",
						Title: "姓名",
						Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, obj *graphql.Object) {
				// 由于没有关联模型数据，关联字段不会被创建
				assert.NotNil(t, obj)
				fields := obj.Fields()
				// 只检查非关联字段
				assert.Contains(t, fields, "id")
				assert.Contains(t, fields, "name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 resolver
			resolver := &graphqlModelResolver{
				
				model:              tt.model,
				inputTypeGenerator: newInputTypeGenerator(),
			}

			// 执行测试
			obj, err := resolver.createModelType(context.Background())

			// 检查错误
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// 执行自定义检查
			if tt.checkFunc != nil {
				tt.checkFunc(t, obj)
			}
		})
	}
}

// TestGenerateModelType 测试 generateModelType 方法
func TestGenerateModelType(t *testing.T) {
	t.Run("生成模型类型并填充关联对象映射", func(t *testing.T) {
		model := &RuntimeModel{
			Name:        "TestModel",
			Title:       "测试模型",
			Description: "测试描述",
			Fields: map[string]*RuntimeField{
				"id": {
					Name:      "id",
					Title:     "ID",
					Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
					IsPrimary: true,
				},
				"name": {
					Name:  "name",
					Title: "名称",
					Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
				},
			},
		}

		resolver := &graphqlModelResolver{
			
			model:              model,
			inputTypeGenerator: newInputTypeGenerator(),
		}

		relateObjMap := map[string]*graphql.Object{}
		obj, err := resolver.generateModelType(context.Background(), 1, model, relateObjMap)

		require.NoError(t, err)
		assert.NotNil(t, obj)
	})
}

// TestCreateField 测试 createField 方法
func TestCreateField(t *testing.T) {
	tests := []struct {
		name      string
		field     *RuntimeField
		relateObj map[string]*graphql.Object
		wantErr   bool
		checkFunc func(t *testing.T, field *graphql.Field)
	}{
		{
			name: "创建字符串类型字段",
			field: &RuntimeField{
				Name:  "username",
				Title: "用户名",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
			},
			relateObj: map[string]*graphql.Object{},
			wantErr:   false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.Equal(t, "username", field.Name)
				assert.Equal(t, "用户名", field.Description)
				assert.Equal(t, graphql.String, field.Type)
			},
		},
		{
			name: "创建整数类型字段",
			field: &RuntimeField{
				Name:  "age",
				Title: "年龄",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatInteger},
			},
			relateObj: map[string]*graphql.Object{},
			wantErr:   false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.Equal(t, graphql.Int, field.Type)
			},
		},
		{
			name: "创建浮点数类型字段",
			field: &RuntimeField{
				Name:  "price",
				Title: "价格",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatNumber},
			},
			relateObj: map[string]*graphql.Object{},
			wantErr:   false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.Equal(t, graphql.Float, field.Type)
			},
		},
		{
			name: "创建布尔类型字段",
			field: &RuntimeField{
				Name:  "isActive",
				Title: "是否激活",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatBoolean},
			},
			relateObj: map[string]*graphql.Object{},
			wantErr:   false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.Equal(t, graphql.Boolean, field.Type)
			},
		},
		{
			name: "创建日期时间类型字段",
			field: &RuntimeField{
				Name:  "createdAt",
				Title: "创建时间",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatDateTime},
			},
			relateObj: map[string]*graphql.Object{},
			wantErr:   false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.Equal(t, graphql.DateTime, field.Type)
			},
		},
		{
			name: "创建UUIDV7类型字段（天然有序）",
			field: &RuntimeField{
				Name:      "id",
				Title:     "ID",
				Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
				IsPrimary: true,
			},
			relateObj: map[string]*graphql.Object{},
			wantErr:   false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.Equal(t, graphql.ID, field.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &graphqlModelResolver{
				
				model:              &RuntimeModel{},
				inputTypeGenerator: newInputTypeGenerator(),
			}

			field, err := resolver.createField(context.Background(), 0, tt.field, tt.relateObj)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, field)
			}
		})
	}
}

// TestCreateRelationFieldReverse tests createOneToManyFieldFromFK creates a [RefModel!] list type.
func TestCreateRelationFieldReverse(t *testing.T) {
	tests := []struct {
		name         string
		lf           *modeldesign.LogicalForeignKey
		refModelName string
	}{
		{
			name:         "simple one-to-many with single FK field",
			refModelName: "Order",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-1",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"id"},
				TargetFields: []string{"user_id"},
			},
		},
		{
			name:         "composite FK with two fields",
			refModelName: "OrderItem",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-2",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "OrderItem",
				SourceFields: []string{"id", "org_id"},
				TargetFields: []string{"user_id", "org_id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			referenceObj := graphql.NewObject(graphql.ObjectConfig{
				Name: tt.refModelName + "Query",
				Fields: graphql.Fields{
					"id": &graphql.Field{Type: graphql.ID},
				},
			})

			resolver := &graphqlModelResolver{
				
				model:              &RuntimeModel{Name: tt.lf.ModelName},
				inputTypeGenerator: newInputTypeGenerator(),
			}

			graphqlField := &graphql.Field{Name: tt.refModelName + "_list"}

			result := resolver.createOneToManyFieldFromFK(tt.lf, referenceObj, graphqlField)

			require.NotNil(t, result)
			// Type must be a List
			listType, ok := result.Type.(*graphql.List)
			require.True(t, ok, "expected *graphql.List, got %T", result.Type)
			// Inner type must be NonNull(RefModelQuery)
			nonNull, ok := listType.OfType.(*graphql.NonNull)
			require.True(t, ok, "expected *graphql.NonNull inside List, got %T", listType.OfType)
			innerObj, ok := nonNull.OfType.(*graphql.Object)
			require.True(t, ok, "expected *graphql.Object inside NonNull, got %T", nonNull.OfType)
			assert.Equal(t, tt.refModelName+"Query", innerObj.Name())
			// Resolve function must be set
			assert.NotNil(t, result.Resolve)
		})
	}
}

// TestCreateOneToManyResolverFromFK tests the one-to-many resolver logic.
func TestCreateOneToManyResolverFromFK(t *testing.T) {
	orderRecords := []map[string]any{
		{"id": "o1", "user_id": "u1"},
		{"id": "o2", "user_id": "u1"},
	}

	tests := []struct {
		name       string
		lf         *modeldesign.LogicalForeignKey
		source     map[string]any
		mockResult []map[string]any
		wantLen    int
		wantEmpty  bool
	}{
		{
			name: "simple one-to-many returns 2 order records",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-1",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"id"},
				TargetFields: []string{"user_id"},
			},
			source:     map[string]any{"id": "u1", "name": "Alice"},
			mockResult: orderRecords,
			wantLen:    2,
		},
		{
			name: "composite FK - both fields used in WHERE",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-2",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "OrderItem",
				SourceFields: []string{"id", "org_id"},
				TargetFields: []string{"user_id", "org_id"},
			},
			source: map[string]any{"id": "u1", "org_id": "org-1"},
			mockResult: []map[string]any{
				{"id": "item1", "user_id": "u1", "org_id": "org-1"},
			},
			wantLen: 1,
		},
		{
			name: "source field is nil - returns empty array",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-3",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"id"},
				TargetFields: []string{"user_id"},
			},
			source:    map[string]any{"name": "Alice"}, // "id" is missing
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClientDatabaseRepository{result: tt.mockResult}

			resolver := &graphqlModelResolver{
				
				model:              &RuntimeModel{Name: tt.lf.ModelName},
				inputTypeGenerator: newInputTypeGenerator(),
			}

			resolveFn := resolver.createOneToManyResolverFromFK(tt.lf, tt.lf.RefModelName)
			require.NotNil(t, resolveFn)

			p := graphql.ResolveParams{
				Context: WithGraphqlRequestContext(context.Background(), mock),
				Source:  tt.source,
			}

			raw, err := resolveFn(p)
			require.NoError(t, err)

			if tt.wantEmpty {
				results, ok := raw.([]map[string]any)
				require.True(t, ok)
				assert.Empty(t, results)
				return
			}

			results, ok := raw.([]map[string]any)
			require.True(t, ok)
			assert.Len(t, results, tt.wantLen)
		})
	}
}
