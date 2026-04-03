package ddlfactory

// MySQLDataType MySQL数据类型
type MySQLDataType string

const (
	// 数值类型
	TINYINT  MySQLDataType = "TINYINT"
	SMALLINT MySQLDataType = "SMALLINT"
	INT      MySQLDataType = "INT"
	BIGINT   MySQLDataType = "BIGINT"
	DECIMAL  MySQLDataType = "DECIMAL"
	FLOAT    MySQLDataType = "FLOAT"
	DOUBLE   MySQLDataType = "DOUBLE"

	BOOL MySQLDataType = "BOOLEAN"
	// 字符串类型
	CHAR    MySQLDataType = "CHAR"
	VARCHAR MySQLDataType = "VARCHAR"
	TEXT    MySQLDataType = "TEXT"

	// 日期时间类型
	DATE      MySQLDataType = "DATE"
	DATETIME  MySQLDataType = "DATETIME"
	TIMESTAMP MySQLDataType = "TIMESTAMP"

	// JSON类型
	JSON MySQLDataType = "JSON"
)

// FieldEntity MySQL字段定义
type FieldEntity struct {
	Name          string        `json:"name"`          // 字段名
	Type          MySQLDataType `json:"type"`          // 数据类型
	Length        *int          `json:"length"`        // 长度(VARCHAR, CHAR等)
	Precision     *int          `json:"precision"`     // 精度(DECIMAL)
	Scale         *int          `json:"scale"`         // 小数位数(DECIMAL)
	Nullable      bool          `json:"nullable"`      // 是否可空
	DefaultValue  *string       `json:"defaultValue"`  // 默认值
	AutoIncrement bool          `json:"autoIncrement"` // 自增
	Primary       bool          `json:"primary"`       // 主键
	Unique        bool          `json:"unique"`        // 唯一约束
	Comment       string        `json:"comment"`       // 注释
}

// TableEntity 表定义
type TableEntity struct {
	Name    string         `json:"name"`    // 表名
	Fields  []*FieldEntity `json:"fields"`  // 字段列表
	Engine  string         `json:"engine"`  // 存储引擎，默认InnoDB
	Charset string         `json:"charset"` // 字符集，默认utf8mb4
	Comment string         `json:"comment"` // 表注释
}
