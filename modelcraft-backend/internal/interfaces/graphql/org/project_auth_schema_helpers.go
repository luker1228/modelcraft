package orggraphql

import (
	domainRLS "modelcraft/internal/domain/rls"
	"modelcraft/internal/interfaces/graphql/org/generated"
)

func toDomainAuthVarType(t generated.AuthVariableType) domainRLS.AuthVarType {
	switch t {
	case generated.AuthVariableTypeUUID:
		return domainRLS.AuthVarTypeUUID
	case generated.AuthVariableTypeInteger:
		return domainRLS.AuthVarTypeInteger
	case generated.AuthVariableTypeString:
		fallthrough
	default:
		return domainRLS.AuthVarTypeString
	}
}

func toGraphQLAuthVarType(t domainRLS.AuthVarType) generated.AuthVariableType {
	switch t {
	case domainRLS.AuthVarTypeUUID:
		return generated.AuthVariableTypeUUID
	case domainRLS.AuthVarTypeInteger:
		return generated.AuthVariableTypeInteger
	case domainRLS.AuthVarTypeString:
		fallthrough
	default:
		return generated.AuthVariableTypeString
	}
}

func toGraphQLProjectAuthSchema(authSchema *domainRLS.AuthSchema) *generated.ProjectAuthSchema {
	if authSchema == nil {
		return &generated.ProjectAuthSchema{Variables: []*generated.AuthVariable{}}
	}

	variables := make([]*generated.AuthVariable, 0, len(authSchema.Variables))
	for _, v := range authSchema.Variables {
		variables = append(variables, &generated.AuthVariable{
			Name:   v.Name,
			Source: v.Source,
			Type:   toGraphQLAuthVarType(v.Type),
		})
	}

	return &generated.ProjectAuthSchema{Variables: variables}
}
