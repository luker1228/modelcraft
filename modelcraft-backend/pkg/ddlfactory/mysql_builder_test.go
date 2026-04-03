package ddlfactory

import (
	"context"
	"modelcraft/pkg/bizutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MySQLDDLBuilderTestSuite 测试套件
type MySQLDDLBuilderTestSuite struct {
	suite.Suite
	builder *MySQLDDLBuilder
}

// SetupTest 每个测试前的设置
func (suite *MySQLDDLBuilderTestSuite) SetupTest() {
	suite.builder = NewMySQLDDLBuilder(context.Background())
}

// TestCreateTableSuccess 测试成功创建表的场景
func (suite *MySQLDDLBuilderTestSuite) TestCreateTableSuccess() {
	tests := []struct {
		name     string
		table    *TableEntity
		expected string
	}{
		{
			name: "简单用户表",
			table: &TableEntity{
				Name: "users",
				Fields: []*FieldEntity{
					{
						Name:          "id",
						Type:          BIGINT,
						Nullable:      false,
						AutoIncrement: true,
						Primary:       true,
						Comment:       "用户ID",
					},
					{
						Name:     "username",
						Type:     VARCHAR,
						Length:   bizutils.IntPtr(50),
						Nullable: false,
						Unique:   true,
						Comment:  "用户名",
					},
					{
						Name:     "email",
						Type:     VARCHAR,
						Length:   bizutils.IntPtr(100),
						Nullable: true,
						Comment:  "邮箱",
					},
				},
				Comment: "用户表",
			},
			expected: "CREATE TABLE IF NOT EXISTS `users` (\n" +
				"  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户ID',\n" +
				"  `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',\n" +
				"  `email` VARCHAR(100) NULL COMMENT '邮箱',\n" +
				"  PRIMARY KEY (`id`)\n" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表'",
		},
		{
			name: "带DECIMAL字段的表",
			table: &TableEntity{
				Name: "products",
				Fields: []*FieldEntity{
					{
						Name:     "id",
						Type:     INT,
						Nullable: false,
						Primary:  true,
					},
					{
						Name:      "price",
						Type:      DECIMAL,
						Precision: bizutils.IntPtr(10),
						Scale:     bizutils.IntPtr(2),
						Nullable:  false,
					},
				},
			},
			expected: "CREATE TABLE IF NOT EXISTS `products` (\n" +
				"  `id` INT NOT NULL,\n" +
				"  `price` DECIMAL(10,2) NOT NULL,\n" +
				"  PRIMARY KEY (`id`)\n" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		},
		{
			name: "带默认值的表",
			table: &TableEntity{
				Name: "logs",
				Fields: []*FieldEntity{
					{
						Name:     "id",
						Type:     BIGINT,
						Nullable: false,
						Primary:  true,
					},
					{
						Name:         "created_at",
						Type:         TIMESTAMP,
						Nullable:     false,
						DefaultValue: stringPtr("CURRENT_TIMESTAMP"),
					},
					{
						Name:         "status",
						Type:         VARCHAR,
						Length:       bizutils.IntPtr(20),
						Nullable:     true,
						DefaultValue: stringPtr("active"),
					},
				},
			},
			expected: "CREATE TABLE IF NOT EXISTS `logs` (\n" +
				"  `id` BIGINT NOT NULL,\n" +
				"  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,\n" +
				"  `status` VARCHAR(20) NULL DEFAULT 'active',\n" +
				"  PRIMARY KEY (`id`)\n" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.builder.BuildCreateTable(tt.table)

			require.NoError(suite.T(), err, "不应该出现错误")
			assert.Equal(suite.T(), tt.expected, result, "DDL语句应该匹配")
		})
	}
}

// TestCreateTableErrors 测试错误场景
func (suite *MySQLDDLBuilderTestSuite) TestCreateTableErrors() {
	tests := []struct {
		name          string
		table         *TableEntity
		expectedError string
	}{
		{
			name: "空表名应该报错",
			table: &TableEntity{
				Name: "",
				Fields: []*FieldEntity{
					{Name: "id", Type: INT},
				},
			},
			expectedError: "表名不能为空",
		},
		{
			name: "空字段列表应该报错",
			table: &TableEntity{
				Name:   "test",
				Fields: []*FieldEntity{},
			},
			expectedError: "字段列表不能为空",
		},
		{
			name: "VARCHAR未指定长度应该报错",
			table: &TableEntity{
				Name: "test",
				Fields: []*FieldEntity{
					{
						Name: "name",
						Type: VARCHAR,
						// Length未指定
					},
				},
			},
			expectedError: "VARCHAR类型必须指定长度",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.builder.BuildCreateTable(tt.table)

			assert.Error(suite.T(), err, "应该出现错误")
			assert.Empty(suite.T(), result, "出错时结果应该为空")
			assert.Contains(suite.T(), err.Error(), tt.expectedError, "错误信息应该包含期望的内容")
		})
	}
}

// TestDataTypes 测试各种数据类型
func (suite *MySQLDDLBuilderTestSuite) TestDataTypes() {
	table := &TableEntity{
		Name: "test_types",
		Fields: []*FieldEntity{
			{Name: "id", Type: INT, Primary: true, Nullable: false},
			{Name: "tiny_col", Type: TINYINT, Nullable: true},
			{Name: "small_col", Type: SMALLINT, Nullable: true},
			{Name: "big_col", Type: BIGINT, Nullable: true},
			{
				Name:      "decimal_col",
				Type:      DECIMAL,
				Precision: bizutils.IntPtr(10),
				Scale:     bizutils.IntPtr(2),
				Nullable:  true,
			},
			{Name: "float_col", Type: FLOAT, Nullable: true},
			{Name: "double_col", Type: DOUBLE, Nullable: true},
			{Name: "char_col", Type: CHAR, Length: bizutils.IntPtr(10), Nullable: true},
			{Name: "varchar_col", Type: VARCHAR, Length: bizutils.IntPtr(255), Nullable: true},
			{Name: "text_col", Type: TEXT, Nullable: true},
			{Name: "date_col", Type: DATE, Nullable: true},
			{Name: "datetime_col", Type: DATETIME, Nullable: true},
			{Name: "timestamp_col", Type: TIMESTAMP, Nullable: true},
			{Name: "json_col", Type: JSON, Nullable: true},
		},
	}

	result, err := suite.builder.BuildCreateTable(table)
	require.NoError(suite.T(), err)

	// 验证DDL包含所有字段
	expectedFields := []string{
		"`id` INT NOT NULL",
		"`tiny_col` TINYINT NULL",
		"`decimal_col` DECIMAL(10,2) NULL",
		"`varchar_col` VARCHAR(255) NULL",
		"`json_col` JSON NULL",
		"PRIMARY KEY (`id`)",
	}

	for _, expected := range expectedFields {
		assert.Contains(suite.T(), result, expected, "DDL应该包含字段: %s", expected)
	}
}

// TestTableOptions 测试表选项
func (suite *MySQLDDLBuilderTestSuite) TestTableOptions() {
	table := &TableEntity{
		Name:    "test_options",
		Engine:  "MyISAM",
		Charset: "latin1",
		Comment: "测试表选项",
		Fields: []*FieldEntity{
			{Name: "id", Type: INT, Primary: true, Nullable: false},
		},
	}

	result, err := suite.builder.BuildCreateTable(table)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), result, "ENGINE=MyISAM")
	assert.Contains(suite.T(), result, "DEFAULT CHARSET=latin1")
	assert.Contains(suite.T(), result, "COMMENT='测试表选项'")
}

// TestDropTable 测试DropTable
func (suite *MySQLDDLBuilderTestSuite) TestDropTable() {
	tests := []struct {
		name          string
		tableName     string
		expectedError string
	}{
		{
			name:          "测试正常",
			tableName:     "abc",
			expectedError: "",
		},
		{
			name:          "测试空表",
			tableName:     "",
			expectedError: "表名不能为空",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_, err := suite.builder.BuildDropTable(tt.tableName)
			if err != nil {
				assert.Error(suite.T(), err, "应该出现错误")
				assert.Contains(suite.T(), err.Error(), tt.expectedError, "错误信息应该包含期望的内容")
			}
		})
	}
}

// TestDefaultValues 测试默认值处理
func (suite *MySQLDDLBuilderTestSuite) TestDefaultValues() {
	table := &TableEntity{
		Name: "test_defaults",
		Fields: []*FieldEntity{
			{Name: "id", Type: INT, Primary: true, Nullable: false},
			{
				Name:         "int_default",
				Type:         INT,
				DefaultValue: stringPtr("100"),
				Nullable:     true,
			},
			{
				Name:         "str_default",
				Type:         VARCHAR,
				Length:       bizutils.IntPtr(50),
				DefaultValue: stringPtr("hello"),
				Nullable:     true,
			},
			{
				Name:         "timestamp_default",
				Type:         TIMESTAMP,
				DefaultValue: stringPtr("CURRENT_TIMESTAMP"),
				Nullable:     true,
			},
			{
				Name:         "null_default",
				Type:         VARCHAR,
				Length:       bizutils.IntPtr(50),
				DefaultValue: stringPtr("NULL"),
				Nullable:     true,
			},
		},
	}

	result, err := suite.builder.BuildCreateTable(table)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), result, "DEFAULT 100", "数值默认值不应该有引号")
	assert.Contains(suite.T(), result, "DEFAULT 'hello'", "字符串默认值应该有引号")
	assert.Contains(suite.T(), result, "DEFAULT CURRENT_TIMESTAMP", "特殊函数不应该有引号")
	assert.Contains(suite.T(), result, "DEFAULT NULL", "NULL值不应该有引号")
}

// 运行测试套件
func TestMySQLDDLBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(MySQLDDLBuilderTestSuite))
}

func stringPtr(s string) *string {
	return &s
}
