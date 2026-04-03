package modelruntime

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// GraphQLDate is a custom GraphQL scalar type for Date values (ISO 8601 YYYY-MM-DD)
var GraphQLDate = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Date",
	Description: "Date scalar type represents a date in ISO 8601 format (YYYY-MM-DD), e.g., \"2024-01-15\"",
	Serialize: func(value interface{}) interface{} {
		switch v := value.(type) {
		case time.Time:
			return v.Format("2006-01-02")
		case *time.Time:
			if v == nil {
				return nil
			}
			return v.Format("2006-01-02")
		case string:
			// 验证字符串格式
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return nil
			}
			return v
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		switch v := value.(type) {
		case string:
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				return nil
			}
			return t
		case *string:
			if v == nil {
				return nil
			}
			t, err := time.Parse("2006-01-02", *v)
			if err != nil {
				return nil
			}
			return t
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			t, err := time.Parse("2006-01-02", valueAST.Value)
			if err != nil {
				return nil
			}
			return t
		default:
			return nil
		}
	},
})

// GraphQLTime is a custom GraphQL scalar type for Time values (HH:MM:SS)
var GraphQLTime = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Time",
	Description: "Time scalar type represents a time in 24-hour format (HH:MM:SS), e.g., \"14:30:00\"",
	Serialize: func(value interface{}) interface{} {
		switch v := value.(type) {
		case time.Time:
			return v.Format("15:04:05")
		case *time.Time:
			if v == nil {
				return nil
			}
			return v.Format("15:04:05")
		case string:
			// 验证字符串格式
			if _, err := time.Parse("15:04:05", v); err != nil {
				return nil
			}
			return v
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		switch v := value.(type) {
		case string:
			t, err := time.Parse("15:04:05", v)
			if err != nil {
				return nil
			}
			return t
		case *string:
			if v == nil {
				return nil
			}
			t, err := time.Parse("15:04:05", *v)
			if err != nil {
				return nil
			}
			return t
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			t, err := time.Parse("15:04:05", valueAST.Value)
			if err != nil {
				return nil
			}
			return t
		default:
			return nil
		}
	},
})
