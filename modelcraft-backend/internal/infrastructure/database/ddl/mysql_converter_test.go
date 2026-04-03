package ddl

import (
	"context"
	"fmt"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ddlfactory"
	"testing"

	entity "modelcraft/internal/domain/modeldesign"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DDLConverterTestSuite 测试套件
type DDLConverterTestSuite struct {
	suite.Suite
	converter *DDLConverter
}

// SetupTest 每个测试前的设置
func (suite *DDLConverterTestSuite) SetupTest() {
	suite.converter = newDdlConverter(context.Background())
}

// TestNewSchemaConverter 测试构造函数
func (suite *DDLConverterTestSuite) TestNewSchemaConverter() {
	converter := newDdlConverter(context.Background())
	assert.NotNil(suite.T(), converter, "转换器不应该为nil")
}

// TestGenerateColumnDefinition 测试列定义生成
func (suite *DDLConverterTestSuite) TestGenerateColumnDefinition() {
	tests := []struct {
		name     string
		column   entity.FieldDefinition
		expected func(*testing.T, *ddlfactory.FieldEntity)
	}{
		{
			name: "uuid类型字段",
			column: entity.FieldDefinition{
				Name:        "id",
				Title:       "主键ID",
				Description: "主键ID",
				Type:        entity.GetFieldTypeByFormat(entity.FormatUUID),
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "id", field.Name)
				assert.Equal(t, ddlfactory.CHAR, field.Type)
				assert.True(t, field.Primary)
				assert.Equal(t, "主键ID", field.Comment)
			},
		},
		{
			name: "整数类型字段",
			column: entity.FieldDefinition{
				Name:        "id",
				Title:       "主键ID",
				Description: "主键ID",
				Type:        entity.GetFieldTypeByFormat(entity.FormatInteger),
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "id", field.Name)
				assert.Equal(t, ddlfactory.INT, field.Type)
				assert.True(t, field.Primary)
				assert.Equal(t, "主键ID", field.Comment)
			},
		},
		{
			name: "布尔类型字段",
			column: entity.FieldDefinition{
				Name:        "is_active",
				Title:       "是否激活",
				Description: "是否激活",
				Type:        entity.GetFieldTypeByFormat(entity.FormatBoolean),
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "is_active", field.Name)
				assert.Equal(t, ddlfactory.BOOL, field.Type)
				assert.False(t, field.Primary)
				assert.Equal(t, "是否激活", field.Comment)
			},
		},
		{
			name: "数字类型字段",
			column: entity.FieldDefinition{
				Name:        "price",
				Title:       "价格",
				Description: "价格",
				Type:        entity.GetFieldTypeByFormat(entity.FormatNumber),
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "price", field.Name)
				assert.Equal(t, ddlfactory.DOUBLE, field.Type)
				assert.Equal(t, "价格", field.Comment)
			},
		},
		{
			name: "短文本字段带长度",
			column: entity.FieldDefinition{
				Name:        "username",
				Title:       "用户名",
				Description: "用户名",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
				Validation:  &entity.ValidationConfig{MaxLength: bizutils.IntPtr(50)},
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "username", field.Name)
				assert.Equal(t, ddlfactory.VARCHAR, field.Type)
				assert.Equal(t, 50, *field.Length)
				assert.Equal(t, "用户名", field.Comment)
			},
		},
		{
			name: "短文本字段无长度（使用TEXT）",
			column: entity.FieldDefinition{
				Name:        "name",
				Title:       "用户名",
				Description: "用户名",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "name", field.Name)
				assert.Equal(t, ddlfactory.TEXT, field.Type)
				assert.Nil(t, field.Length)
			},
		},
		{
			name: "长文本字段",
			column: entity.FieldDefinition{
				Name:        "description",
				Title:       "描述信息",
				Description: "描述信息",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
				Validation:  &entity.ValidationConfig{MaxLength: bizutils.IntPtr(1000)},
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "description", field.Name)
				assert.Equal(t, ddlfactory.TEXT, field.Type)
				assert.Nil(t, field.Length)
				assert.Equal(t, "描述信息", field.Comment)
			},
		},
		{
			name: "字符串字段带较短长度",
			column: entity.FieldDefinition{
				Name:        "title",
				Title:       "标题",
				Description: "标题",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
				Validation:  &entity.ValidationConfig{MaxLength: bizutils.IntPtr(200)},
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "title", field.Name)
				assert.Equal(t, ddlfactory.VARCHAR, field.Type)
				assert.Equal(t, "标题", field.Comment)
				assert.Equal(t, 200, *field.Length)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			field, err := suite.converter.generateColumnDefinition(&tt.column)

			require.NoError(suite.T(), err, "生成列定义不应该出错")
			tt.expected(suite.T(), field)
		})
	}
}

// TestPrimaryKeyFunctionality 专门测试主键功能
func (suite *DDLConverterTestSuite) TestPrimaryKeyFunctionality() {
	tests := []struct {
		name            string
		column          entity.FieldDefinition
		expected        func(*testing.T, *ddlfactory.FieldEntity)
		shouldBePrimary bool
	}{
		{
			name: "整数主键字段",
			column: entity.FieldDefinition{
				Name:        "id",
				Title:       "主键ID",
				Description: "主键ID",
				Type:        entity.GetFieldTypeByFormat(entity.FormatInteger),
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "id", field.Name)
				assert.Equal(t, ddlfactory.INT, field.Type)
				assert.True(t, field.Primary, "整数主键字段应该被标记为主键")
				assert.Equal(t, "主键ID", field.Comment)
			},
			shouldBePrimary: true,
		},
		{
			name: "UUIDV7主键字段",
			column: entity.FieldDefinition{
				Name:        "uuid",
				Title:       "UUIDV7主键",
				Description: "UUIDV7主键（天然有序）",
				Type:        entity.GetFieldTypeByFormat(entity.FormatUUID),
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "uuid", field.Name)
				assert.Equal(t, ddlfactory.CHAR, field.Type)
				assert.True(t, field.Primary, "UUIDV7主键字段应该被标记为主键")
				assert.Equal(t, "UUIDV7主键（天然有序）", field.Comment)
			},
			shouldBePrimary: true,
		},
		{
			name: "非主键字段",
			column: entity.FieldDefinition{
				Name:        "name",
				Title:       "名称",
				Description: "名称",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
				IsPrimary:   false,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "name", field.Name)
				assert.Equal(t, ddlfactory.TEXT, field.Type)
				assert.False(t, field.Primary, "非主键字段不应该被标记为主键")
				assert.Equal(t, "名称", field.Comment)
			},
			shouldBePrimary: false,
		},
		{
			name: "默认非主键字段",
			column: entity.FieldDefinition{
				Name:        "description",
				Title:       "描述",
				Description: "描述",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
				// IsPrimary 默认为 false
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "description", field.Name)
				assert.Equal(t, ddlfactory.TEXT, field.Type)
				assert.False(t, field.Primary, "默认字段不应该被标记为主键")
				assert.Equal(t, "描述", field.Comment)
			},
			shouldBePrimary: false,
		},
		{
			name: "字符串主键字段",
			column: entity.FieldDefinition{
				Name:        "code",
				Title:       "代码主键",
				Description: "代码主键",
				Type:        entity.GetFieldTypeByFormat(entity.FormatString),
				Validation:  &entity.ValidationConfig{MaxLength: bizutils.IntPtr(50)},
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "code", field.Name)
				assert.Equal(t, ddlfactory.VARCHAR, field.Type)
				assert.True(t, field.Primary, "字符串主键字段应该被标记为主键")
				assert.Equal(t, "代码主键", field.Comment)
				assert.Equal(t, 50, *field.Length)
			},
			shouldBePrimary: true,
		},
		{
			name: "数字主键字段",
			column: entity.FieldDefinition{
				Name:        "serial_no",
				Title:       "序列号主键",
				Description: "序列号主键",
				Type:        entity.GetFieldTypeByFormat(entity.FormatNumber),
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "serial_no", field.Name)
				assert.Equal(t, ddlfactory.DOUBLE, field.Type)
				assert.True(t, field.Primary, "数字主键字段应该被标记为主键")
				assert.Equal(t, "序列号主键", field.Comment)
			},
			shouldBePrimary: true,
		},
		{
			name: "布尔主键字段",
			column: entity.FieldDefinition{
				Name:        "flag",
				Title:       "标志主键",
				Description: "标志主键",
				Type:        entity.GetFieldTypeByFormat(entity.FormatBoolean),
				IsPrimary:   true,
			},
			expected: func(t *testing.T, field *ddlfactory.FieldEntity) {
				assert.Equal(t, "flag", field.Name)
				assert.Equal(t, ddlfactory.BOOL, field.Type)
				assert.True(t, field.Primary, "布尔主键字段应该被标记为主键")
				assert.Equal(t, "标志主键", field.Comment)
			},
			shouldBePrimary: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			field, err := suite.converter.generateColumnDefinition(&tt.column)

			require.NoError(suite.T(), err, "生成列定义不应该出错")

			// 验证主键标记是否正确
			if tt.shouldBePrimary {
				assert.True(suite.T(), field.Primary, "字段应该被标记为主键")
			} else {
				assert.False(suite.T(), field.Primary, "字段不应该被标记为主键")
			}

			// 执行具体的验证函数
			tt.expected(suite.T(), field)
		})
	}
}

// TestFieldTypeMapping 测试字段类型映射
func (suite *DDLConverterTestSuite) TestFieldTypeMapping() {
	mappings := map[string]ddlfactory.MySQLDataType{
		string(entity.FormatBoolean): ddlfactory.BOOL,
		string(entity.FormatNumber):  ddlfactory.DOUBLE,
		string(entity.FormatInteger): ddlfactory.INT,
	}

	for formatStr, expectedType := range mappings {
		suite.Run(formatStr, func() {
			column := entity.FieldDefinition{
				Name:        "",
				Title:       "",
				Description: "",
				Type:        entity.GetFieldTypeByFormat(entity.FormatType(formatStr)),
			}

			field, err := suite.converter.generateColumnDefinition(&column)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), expectedType, field.Type, "字段类型映射不正确")
		})
	}
}

// TestGenerateTablePrimaryKeyDDL 测试表主键DDL生成
func (suite *DDLConverterTestSuite) TestGenerateTablePrimaryKeyDDL() {
	tests := []struct {
		name     string
		table    entity.DataModel
		expected func(*testing.T, string, error)
	}{
		{
			name: "用户表",
			table: entity.DataModel{
				ModelMeta: entity.ModelMeta{
					ModelLocator: entity.ModelLocator{
						ModelName: "users",
					},
					Title:       "用户表",
					Description: "用户表",
				},
				Fields: []*entity.FieldDefinition{
					{
						Name:        "id",
						Title:       "主键ID",
						Description: "主键ID",
						Type:        entity.GetFieldTypeByFormat(entity.FormatInteger),
						IsPrimary:   true,
					},
				},
			},
			expected: func(t *testing.T, ddl string, err error) {
				require.NoError(t, err, "生成DDL不应该出错")
				assert.Contains(t, ddl, "CREATE TABLE IF NOT EXISTS `users`", "DDL应该包含正确的表名")
				assert.Contains(t, ddl, "`id` INT NOT NULL", "DDL应该包含整数主键字段")
				assert.Contains(t, ddl, "PRIMARY KEY (`id`)", "DDL应该包含主键约束")
				assert.Contains(t, ddl, "ENGINE=InnoDB", "DDL应该包含正确的引擎")
				assert.Contains(t, ddl, "CHARSET=utf8mb4", "DDL应该包含正确的字符集")
				assert.Contains(t, ddl, "COMMENT='用户表'", "DDL应该包含正确的注释")
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ddl, err := suite.converter.GenerateCreateTableDDL(&tt.table)
			fmt.Printf("ddl=%s", ddl)
			tt.expected(suite.T(), ddl, err)
		})
	}
}

// TestGenerateTableDDL 测试表DDL生成
func (suite *DDLConverterTestSuite) TestGenerateTableDDL() {
	tests := []struct {
		name     string
		table    entity.DataModel
		expected func(*testing.T, string, error)
	}{
		{
			name: "用户表",
			table: entity.DataModel{
				ModelMeta: entity.ModelMeta{
					ModelLocator: entity.ModelLocator{
						ModelName: "users",
					},
					Title:       "用户表",
					Description: "用户表",
				},
				Fields: []*entity.FieldDefinition{
					{
						Name:        "id",
						Title:       "主键ID",
						Description: "主键ID",
						Type:        entity.GetFieldTypeByFormat(entity.FormatInteger),
						IsPrimary:   true,
					},
					{
						Name:        "username",
						Title:       "用户名",
						Description: "用户名",
						Type:        entity.GetFieldTypeByFormat(entity.FormatString),
						Validation:  &entity.ValidationConfig{MaxLength: bizutils.IntPtr(50)},
					},
					{
						Name:        "email",
						Title:       "邮箱",
						Description: "邮箱",
						Type:        entity.GetFieldTypeByFormat(entity.FormatString),
						Validation:  &entity.ValidationConfig{MaxLength: bizutils.IntPtr(100)},
					},
					{
						Name:        "is_active",
						Title:       "是否激活",
						Description: "是否激活",
						Type:        entity.GetFieldTypeByFormat(entity.FormatBoolean),
					},
					{
						Name:        "created_at",
						Title:       "创建时间",
						Description: "创建时间",
						Type:        entity.GetFieldTypeByFormat(entity.FormatDateTime),
					},
				},
			},
			expected: func(t *testing.T, ddl string, err error) {
				require.NoError(t, err, "生成DDL不应该出错")
				assert.Contains(t, ddl, "CREATE TABLE IF NOT EXISTS `users`", "DDL应该包含正确的表名")
				assert.Contains(t, ddl, "`id` INT NOT NULL", "DDL应该包含整数主键字段")
				assert.Contains(t, ddl, "`username` VARCHAR(50)", "DDL应该包含用户名字段")
				assert.Contains(t, ddl, "`email` VARCHAR(100)", "DDL应该包含邮箱字段")
				assert.Contains(t, ddl, "`is_active` BOOL", "DDL应该包含布尔字段")
				assert.Contains(t, ddl, "`created_at` DATETIME", "DDL应该包含时间字段")
				assert.Contains(t, ddl, "PRIMARY KEY (`id`)", "DDL应该包含主键约束")
				assert.Contains(t, ddl, "ENGINE=InnoDB", "DDL应该包含正确的引擎")
				assert.Contains(t, ddl, "CHARSET=utf8mb4", "DDL应该包含正确的字符集")
				assert.Contains(t, ddl, "COMMENT='用户表'", "DDL应该包含正确的注释")
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ddl, err := suite.converter.GenerateCreateTableDDL(&tt.table)
			fmt.Printf("ddl=%s", ddl)
			tt.expected(suite.T(), ddl, err)
		})
	}
}

// 运行测试套件
func TestSchemaColumnHolderTestSuite(t *testing.T) {
	suite.Run(t, new(DDLConverterTestSuite))
}
