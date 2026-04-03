//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/api"
	_ "github.com/99designs/gqlgen/codegen/config"
	_ "github.com/99designs/gqlgen/internal/imports"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
