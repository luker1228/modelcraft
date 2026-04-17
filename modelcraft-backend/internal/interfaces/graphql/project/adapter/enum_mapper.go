package adapter

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
)

// EnumMapper is the singleton enum mapper for project domain
var EnumMapper *enumMapper

func init() {
	EnumMapper = &enumMapper{}
}

type enumMapper struct{}

// ConvertEnumDefinitionToGraphQL converts a domain EnumDefinition to GraphQL type
func (e *enumMapper) ConvertEnumDefinitionToGraphQL(
	enum *modeldesign.EnumDefinition,
) *generated.EnumDefinition {
	if enum == nil {
		return nil
	}

	options := make([]*generated.EnumOption, 0, len(enum.Options))
	for _, opt := range enum.Options {
		description := opt.Description
		options = append(options, &generated.EnumOption{
			Code:        opt.Code,
			Label:       opt.Label,
			Order:       opt.Order,
			Description: &description,
		})
	}

	description := enum.Description
	return &generated.EnumDefinition{
		ID:            enum.ID,
		OrgName:       enum.OrgName,
		ProjectSlug:   enum.ProjectSlug,
		Name:          enum.Name,
		DisplayName:   enum.DisplayName,
		Description:   &description,
		Options:       options,
		IsMultiSelect: enum.IsMultiSelect,
		CreatedAt:     enum.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     enum.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
